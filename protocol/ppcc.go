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

// PPCC defines the channels and variables associated with the contact-chaining protocol
type PPCC struct {
	*onet.TreeNodeInstance

	ChildCount              chan int

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

    OutputList              map[string]bool
	TelecomIdx				int
    LocalSubgraph           *lib.TelecomGraph

    ppcc                    *lib.PPCC
    publics                 []abstract.Point
    private                 abstract.Scalar
}

// NewPPCC initialises the structure for use in one round
func NewPPCC(n *onet.TreeNodeInstance) (onet.ProtocolInstance, error) {

    if (numAuthorities > 1) {
		return nil, errors.New("Protocol does not yet support multiple authorities")
    }

	c := &PPCC{
		TreeNodeInstance:   n,
		ChildCount:         make(chan int),
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
            log.Lvl1("Root is DONE")
            p.ChildCount <- 1

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

// Begins the protocol by dequeueing the first maessage (the warrant)
func (p *PPCC) handleInit (in *Init) error {

    if !p.IsRoot() {
        return fmt.Errorf("non-root node received Init message")
    }

    // If we have more nodes than graphs, "truncate" the graph
    numGraphs := len(globalGraphs)
    if numGraphs < p.NumTelecoms {
        p.NumTelecoms = numGraphs
        p.Telecoms = p.Telecoms[0:numGraphs]
        log.Lvl1("Truncated Telecoms to length: ", len(p.Telecoms))
    }

    // Initialize output list for agency
    p.OutputList = make(map[string]bool)

    // Start protocol by handling the first message (the warrant)
    warrant := p.Queue.Pop()
    telecomIdx := warrant.Telecom
    p.CurrentDepth = warrant.Depth
    log.Lvl1("Started protocol with depth ", warrant.Depth)
    if telecomIdx >= p.NumTelecoms {
        return fmt.Errorf("invalid telecom number")
    }

    // Encrypt components of the message under the telecoms public key
    K1, C1, _ := p.ppcc.EncryptTelecomMessage(warrant.Node, numAuthorities + telecomIdx)
    K2, C2, _ := p.ppcc.EncryptTelecomMessage(strconv.Itoa(warrant.Telecom), numAuthorities + telecomIdx)
    K3, C3, _ := p.ppcc.EncryptTelecomMessage(strconv.Itoa(warrant.Depth), numAuthorities + telecomIdx)

    // Build authority packet to send to telecom
    out := &AuthorityQuery {
        Query:      warrant.Node,
        Telecom:    warrant.Telecom,
        Depth:      warrant.Depth,
        EncQuery:   []abstract.Point{K1, C1},
        EncTelecom: []abstract.Point{K2, C2},
        EncDepth:   []abstract.Point{K3, C3},
    }

    // Send to telecom
    err := p.SendTo(p.Telecoms[telecomIdx], out)
    p.OutstandingPackets++
    if err != nil {
        log.Error("failed to send initial warant", err)
    }

    return nil
}

func (p *PPCC) handleReply(in *Reply) error {
    log.Lvl1("In HandleReply, len encPhones, encTelecoms:", len(in.EncPhones), ",", len(in.EncTelecoms))

    if !p.IsRoot() {
        return fmt.Errorf("non-root received reply")
    }

    p.OutstandingPackets--
    for i, _ := range(in.EncPhones) {
        if (i % 2 == 1) {
            continue
        }

        // Decrypt message and telecom information
        message, _ := p.ppcc.DecryptTelecomMessage(in.EncPhones[i],   in.EncPhones[i + 1])
        telecom, _ := p.ppcc.DecryptTelecomMessage(in.EncTelecoms[i], in.EncTelecoms[i + 1])
        num, _ := strconv.Atoi(telecom)

        // Push to queue and add to output list
        triple := lib.NewTriple(message, num, p.CurrentDepth - 1)
        p.OutputList[message] = true
        if p.CurrentDepth > 0 {
            p.Queue.Push(triple)
        }
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

        K1, C1, _ := p.ppcc.EncryptTelecomMessage(warrant.Node, numAuthorities + telecomIdx)
        K2, C2, _ := p.ppcc.EncryptTelecomMessage(strconv.Itoa(warrant.Telecom), numAuthorities + telecomIdx)
        K3, C3, _ := p.ppcc.EncryptTelecomMessage(strconv.Itoa(warrant.Depth), numAuthorities + telecomIdx)
        out := &AuthorityQuery {
            Query:      warrant.Node,
            Telecom:    warrant.Telecom,
            Depth:      warrant.Depth,
            EncQuery:   []abstract.Point{K1, C1},
            EncTelecom: []abstract.Point{K2, C2},
            EncDepth:   []abstract.Point{K3, C3},
        }

        err := p.SendTo(p.Telecoms[telecomIdx], out)
        p.OutstandingPackets++
        if err != nil {
            log.Error("failed to send warant", err)
        }
    }

    return nil
}

func (p *PPCC) handleAuthorityQuery (in *AuthorityQuery) error {
    log.Lvl1("Node ", p.TelecomIdx, "in HandleReply for ", in.Query)
    log.Lvl1("len query, tcom, depth: ", len(in.EncQuery), ",", len(in.EncTelecom), ",", len(in.EncDepth))

    if p.IsRoot() {
        log.Lvl1("ERROR: Root received AuthorityQuery")
        return nil
    }

    var err error

    query := lib.AgencyPair{in.Query, in.Telecom}
    graph := p.LocalSubgraph

    encPhones   := make([]abstract.Point, 0)
    encTelecoms := make([]abstract.Point, 0)

    if graph.ContainsNode(query) {
        neighbors := graph.Neighbors(query)
        for _, pair := range neighbors {
            if !graph.HasVisited(pair) {
                K, C, _ := p.ppcc.EncryptTelecomMessage(pair.Node, 0)
                encPhones =   append(encPhones, []abstract.Point{K, C}...)
                K, C, _  = p.ppcc.EncryptTelecomMessage(strconv.Itoa(pair.Telecom), 0)
                encTelecoms = append(encTelecoms, []abstract.Point{K, C}...)
                graph.MarkVisited(pair)
            }
        }
        log.Lvl1("Found ", len(encPhones) / 2, " unvisited neighbors: ")
    }

    err = p.SendTo(p.Agency, &Reply{encPhones, encTelecoms})

    if err != nil {
        log.Lvl1("ERROR sending to agency")
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
