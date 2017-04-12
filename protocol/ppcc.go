package protocol

import (
	"errors"
	"fmt"
    "strconv"
	"github.com/hm16083/ppcc/lib"
    "gopkg.in/dedis/crypto.v0/abstract"
	"gopkg.in/dedis/onet.v1"
	"gopkg.in/dedis/onet.v1/network"
	"gopkg.in/dedis/onet.v1/log"
)

func init() {
	network.RegisterMessage(Reply{})
	network.RegisterMessage(Init{})
	network.RegisterMessage(Done{})
	network.RegisterMessage(AuthorityQuery{})
	onet.GlobalProtocolRegister("PPCC", NewPPCC)
}

// Global variable representing the subgraphs for each telecom
var globalGraphs []lib.TelecomGraph

func SetGraphs (subgraphs []lib.TelecomGraph) {
    globalGraphs = subgraphs
}

var numAuthorities int = 1

type Warrant struct {
    Phone   string
    Telecom int
    Depth   int
}

// PPCC defines the channels and variables associated with the contact-chaining protocol
type PPCC struct {
	*onet.TreeNodeInstance

    ChannelInit             chan StructInit
    ChannelDone             chan StructDone
    ChannelReply            chan StructReply
    ChannelAuthorityQuery   chan StructAuthorityQuery

    NodeDone                bool
    ProtocolDone            chan bool

    OutstandingPackets      int
	Queue                   *lib.AgencyQueue
    CurrentDepth            int

    NumTelecoms             int
    Telecoms                []*onet.TreeNode
    Agency                  *onet.TreeNode
    InitWarrant             Warrant

    OutputList              map[string]bool
	TelecomIdx				int
    LocalSubgraph           *lib.TelecomGraph

    ppcc                    *lib.PPCC
    publics                 []abstract.Point
    private                 abstract.Scalar
    verifyKey               abstract.Point
}

// NewPPCC initialises the structure for use in one round
func NewPPCC(n *onet.TreeNodeInstance) (onet.ProtocolInstance, error) {

    if (numAuthorities > 1) {
		return nil, errors.New("Protocol does not yet support multiple authorities")
    }

	c := &PPCC{
		TreeNodeInstance:   n,
        ProtocolDone:       make(chan bool),
	}

    // Assign node number, public/private keys, and telecom subgraph
    totalNodes := len(n.List())
    numTelecoms := totalNodes - numAuthorities
    telecoms := make([]*onet.TreeNode, numTelecoms)
    publics := make([]abstract.Point, totalNodes)
    j := 0
    for i, tn := range n.List() {
        publics[i] = tn.ServerIdentity.Public
        if (tn.IsRoot()) {
            c.Agency = tn
            continue
        }

        if tn.ServerIdentity.Public.Equal(n.Public()) {
			c.TelecomIdx = j
            local := globalGraphs[j]
            c.LocalSubgraph = &local
		}

        telecoms[j] = tn
        j++

        if j > totalNodes {
            return nil, errors.New("too many telecoms")
        }
    }

    c.ppcc = lib.NewPPCC(n.Suite(), n.Private(), publics)
    c.NodeDone = false
    c.Telecoms = telecoms
    c.NumTelecoms = numTelecoms
    c.OutstandingPackets = 0

    // Register channels
    err := c.RegisterChannel(&c.ChannelReply)
	if err != nil {
		return nil, errors.New("couldn't register reply-channel: " + err.Error())
	}
	err = c.RegisterChannel(&c.ChannelInit)
	if err != nil {
		return nil, errors.New("couldn't register init-channel: " + err.Error())
	}
	err = c.RegisterChannel(&c.ChannelAuthorityQuery)
	if err != nil {
		return nil, errors.New("couldn't register authquery-channel: " + err.Error())
	}
	err = c.RegisterChannel(&c.ChannelDone)
	if err != nil {
		return nil, errors.New("couldn't register done-channel: " + err.Error())
	}
	return c, nil
}

// Start begins the protocol by giving the Agency an init message
func (p *PPCC) Start() error {
    out := &Init{}
	return p.handleInit(out)
}

func (p *PPCC) Dispatch() error {
    for {
        // Invoke the handler function associated with the received packet
        var err error
        select {
            case packet := <-p.ChannelReply:
                err = p.handleReply(&packet.Reply)
            case packet := <-p.ChannelInit:
                err = p.handleInit(&packet.Init)
            case packet := <-p.ChannelAuthorityQuery:
                err = p.handleAuthorityQuery(&packet.AuthorityQuery)
            case packet := <-p.ChannelDone:
                err = p.handleDone(&packet.Done)
        }

        if err != nil {
            log.Error("%v", err)
        }

        if p.NodeDone && p.IsRoot() {
            log.Lvl3("Root is DONE")
            p.ProtocolDone <- true

            for _, tn := range p.Telecoms {
                p.SendTo(tn, &Done{})
            }
            return nil
        }

        if p.NodeDone {
            return nil
        }
    }
}

var initSize int = 5

// Begins the protocol by dequeueing the first message (the warrant)
func (p *PPCC) handleInit (in *Init) error {

    if !p.IsRoot() {
        return fmt.Errorf("non-root node received Init message")
    }

    // If we have more nodes than graphs, "truncate" the graph
    numGraphs := len(globalGraphs)
    if numGraphs < p.NumTelecoms {
        p.NumTelecoms = numGraphs
        p.Telecoms = p.Telecoms[0:numGraphs]
        log.Lvl3("Truncated Telecoms to length: ", len(p.Telecoms))
    }

    // Initialize output list for agency
    p.OutputList = make(map[string]bool)
    p.Queue = lib.NewQueue(initSize)

    // Start protocol by handling the first message (the warrant)
    warrant := p.InitWarrant
    telecomIdx := warrant.Telecom
    p.CurrentDepth = warrant.Depth
    log.Lvl1("Started protocol with depth ", warrant.Depth)
    if telecomIdx >= p.NumTelecoms {
        return fmt.Errorf("invalid telecom number")
    }

    // Encrypt components of the message under the telecoms public key
    K, C, _ := p.ppcc.EncryptTelecomMessage(warrant.Phone, numAuthorities + telecomIdx)

    // Build authority packet to send to telecom
    out := &AuthorityQuery {
        EncQuery:   lib.Ciphertext{K, C},
        Telecom:    warrant.Telecom,
        Depth:      warrant.Depth,
    }

    // Sign the fields of the message and attach the signature to the packet
    str := fmt.Sprintf("%+v%+v%+v", out.EncQuery, out.Telecom, out.Depth)
    out.Signature = p.ppcc.SignMessage(str)
    out.VerifyKey = p.ppcc.VerifyKey

    // Send to telecom
    err := p.SendTo(p.Telecoms[telecomIdx], out)
    p.OutstandingPackets++
    if err != nil {
        log.Error("failed to send initial warant", err)
    }

    return nil
}

func (p *PPCC) handleReply(in *Reply) error {
    log.Lvl3("In HandleReply")

    if !p.IsRoot() {
        return fmt.Errorf("non-root received reply")
    }

    decryptedNode, _ := p.ppcc.DecryptTelecomMessage(in.EncQuery.K, in.EncQuery.C)
    log.Lvl3("Decrypted node: ", decryptedNode)
    p.OutputList[decryptedNode] = true

    p.OutstandingPackets--
    for i, s := range(in.Telecoms) {
        if s == "" {
            continue
        }

        // Decrypt message and telecom information
        telecom, _ := strconv.Atoi(in.Telecoms[i])
        message := lib.Ciphertext{in.EncPhones[2 * i], in.EncPhones[2 * i + 1]}

        // Push to queue
        triple := lib.NewTriple(message, telecom, p.CurrentDepth - 1)
        p.Queue.Push(triple)
    }

    // See if the protocol has terminated
    if p.OutstandingPackets == 0 && p.Queue.IsEmpty() {
        p.NodeDone = true
        return nil
    }

    // Dequeue and send next triple 
    if !p.Queue.IsEmpty() {
        warrant := p.Queue.Pop()
        telecomIdx := warrant.Telecom
        if telecomIdx >= p.NumTelecoms {
            return fmt.Errorf("invalid telecom number")
        }
        p.CurrentDepth = warrant.Depth

        // Construct packet without signature
        out := &AuthorityQuery {
            EncQuery:   lib.Ciphertext{warrant.EncPhone.K, warrant.EncPhone.C},
            Telecom:    warrant.Telecom,
            Depth:      warrant.Depth,
        }

        // Sign the fields of the message and attach the signature to the packet
        str := fmt.Sprintf("%+v%+v%+v", out.EncQuery, out.Telecom, out.Depth)
        out.Signature = p.ppcc.SignMessage(str)
        out.VerifyKey = p.ppcc.VerifyKey

        err := p.SendTo(p.Telecoms[telecomIdx], out)
        p.OutstandingPackets++
        if err != nil {
            log.Error("failed to send warant", err)
        }
    }

    return nil
}

func (p *PPCC) handleAuthorityQuery (in *AuthorityQuery) error {
    var err error

    if p.IsRoot() {
        log.Lvl1("ERROR: Root received AuthorityQuery")
        return nil
    }

    if p.TelecomIdx != in.Telecom {
        log.Lvl1("ERROR: Node ", p.TelecomIdx, " received msg intended for ", in.Telecom)
        return nil
    }

    // Verify the authorities' signature
    p.verifyKey = in.VerifyKey
    str := fmt.Sprintf("%+v%+v%+v", in.EncQuery, in.Telecom, in.Depth)
    verify := p.ppcc.VerifyMessage(str, p.verifyKey, in.Signature)
    if verify != nil {
        log.Lvl1("ERROR: Could not verify signature: ", verify)
    }

    // Decrypt the message and reencrypt it under the agency's public key
    nodeQuery, _ := p.ppcc.DecryptTelecomMessage(in.EncQuery.K, in.EncQuery.C)
    log.Lvl3("Node ", p.TelecomIdx, " handling query for ", nodeQuery)
    K, C, _ := p.ppcc.EncryptTelecomMessage(nodeQuery, 0)
    encQuery := lib.Ciphertext{K, C}

    // Prepare to iterate over neighbors
    query := lib.AgencyPair{nodeQuery, p.TelecomIdx}
    graph := p.LocalSubgraph
    encPhones := make([]abstract.Point, 0)
    telecoms  := make([]string, 0)

    // Iterate over neighbors of the node, and create encrypted sets to send back to agency
    if in.Depth > 0 && graph.ContainsNode(query) {
        neighbors := graph.Neighbors(query)
        for _, pair := range neighbors {
            if !graph.HasVisited(pair) {
                K, C, _ = p.ppcc.EncryptTelecomMessage(pair.Node, numAuthorities + pair.Telecom)
                encPhones = append(encPhones, []abstract.Point{K, C}...)
                telecoms  = append(telecoms, strconv.Itoa(pair.Telecom))
                graph.MarkVisited(pair)
            }
        }
        log.Lvl3("Found ", len(telecoms), " unvisited neighbors: ")
    }

    // Send original query (encrypted with agency pubkey) and neighbors (under telecom pubkeys)
    err = p.SendTo(p.Agency, &Reply{encQuery, encPhones, telecoms})
    if err != nil {
        log.Lvl1("ERROR while sending to agency:", err)
        return err
    }
    return nil
}

func (p *PPCC) handleDone (in *Done) error {
    if p.IsRoot() {
        return fmt.Errorf("root received done message")
    }

    p.NodeDone = true
    return nil
}
