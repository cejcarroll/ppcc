package protocol

import (
	"errors"
	"fmt"
	"github.com/hm16083/ppcc/lib"
    //"gopkg.in/dedis/crypto.v0/abstract"
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
}

// NewPPCC initialises the structure for use in one round
func NewPPCC(n *onet.TreeNodeInstance) (onet.ProtocolInstance, error) {
	c := &PPCC{
		TreeNodeInstance: n,
		ChildCount:       make(chan int),
	}

    totalNodes := len(c.List())
    numTelecoms := totalNodes - 1
    telecoms := make([]*onet.TreeNode, numTelecoms)

    j := 0
    for _, tn := range n.List() {
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

    c.NodeDone = false
    c.Telecoms = telecoms
    c.NumTelecoms = numTelecoms
    c.OutstandingPackets = 0

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

    // Start protocol by handling the first message (the warrant)
    warrant := p.Queue.Pop()
    telecomIdx := warrant.Telecom
    p.CurrentDepth = warrant.Depth
    if telecomIdx >= p.NumTelecoms {
        return fmt.Errorf("invalid telecom number")
    }

    out := &AuthorityQuery {
        Query:      warrant.Node,
        Telecom:    warrant.Telecom,
        Depth:      warrant.Depth,
    }

    err := p.SendTo(p.Telecoms[telecomIdx], out)
    p.OutstandingPackets++
    if err != nil {
        log.Error("failed to send initial warant", err)
    }

    return nil
}

func (p *PPCC) handleReply(in *Reply) error {
    log.Lvl1("In HandleReply")

    if !p.IsRoot() {
        return fmt.Errorf("non-root received reply")
    }

    p.OutstandingPackets--
    for _, pair := range in.Response {
        triple := lib.NewTriple(pair.Node, pair.Telecom, p.CurrentDepth - 1)
        if p.CurrentDepth > 0 {
            p.Queue.Push(triple)
        }
    }

    if p.OutstandingPackets == 0 && p.Queue.IsEmpty() {
        p.NodeDone = true
        return nil
    }

    if !p.Queue.IsEmpty() {
        warrant := p.Queue.Pop()
        telecomIdx := warrant.Telecom
        if telecomIdx >= p.NumTelecoms {
            return fmt.Errorf("invalid telecom number")
        }

        out := &AuthorityQuery {
            Query:      warrant.Node,
            Telecom:    warrant.Telecom,
            Depth:      warrant.Depth,
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
    log.Lvl1("Node ", p.TelecomIdx, "in HandleReply")

    if p.IsRoot() {
        log.Lvl1("ERROR: Root received AuthorityQuery")
        return nil
    }

    var err error

    // Hardcoded response to simulate telecom response without implementing maps
    if in.Query == "1234567890" && in.Telecom == 0 {
        newArr := make([]lib.AgencyPair, 1)
        newArr[0] = lib.AgencyPair{"1234567891", 1}
        err = p.SendTo(p.Agency, &Reply{newArr})
    } else {
        err = p.SendTo(p.Agency, &Reply{make([]lib.AgencyPair, 0)})
    }

    if err != nil {
        log.Lvl1("ERROR sending to parent")
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
