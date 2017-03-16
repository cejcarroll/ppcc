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

    NumTelecoms             int
    Telecoms                []*onet.TreeNode
    Agency                  *onet.TreeNode
}

// NewPPCC initialises the structure for use in one round
func NewPPCC(n *onet.TreeNodeInstance) (onet.ProtocolInstance, error) {
	c := &PPCC{
		TreeNodeInstance: n,
		ChildCount:       make(chan int),
	}
    totalNodes := len(c.List())
    telecoms := make([]*onet.TreeNode, totalNodes - 1)
    j := 0

    for _, tn := range n.List() {
        if (tn.IsRoot()) {
            c.Agency = tn
            continue
        }

        telecoms[j] = tn
        j++
        if j > totalNodes {
            return nil, errors.New("too many telecoms")
        }
    }

    c.NodeDone = false
    c.Telecoms = telecoms
    c.NumTelecoms = totalNodes - 1
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
	log.Lvl1("Starting PPCC!!")
    //warrant := p.Queue.Pop()
    //log.Lvl1("Warrant for phone: ", warrant.Node, "Telecom:", warrant.Telecom, "Depth:", warrant.Depth)
    //p.Queue.Push(warrant)

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

    log.Lvl1("Root in handleInit")

    if !p.IsRoot() {
        return fmt.Errorf("non-root node received Init message")
    }

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
        log.Error("failed to send initial warant", err)
    }

    return nil
}

func (p *PPCC) handleReply(in *Reply) error {
    log.Lvl1("In handleReply")

    if !p.IsRoot() {
        return fmt.Errorf("non-root received reply")
    }

    p.OutstandingPackets--
    if p.OutstandingPackets == 0 && p.Queue.IsEmpty() {
        p.NodeDone = true
    }

    if !p.Queue.IsEmpty() {
        log.Lvl1("In queueNotEmpty")
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
    log.Lvl1("Node", p.Name(), " with Info ", p.Info(), " Received QUERY")

    if p.IsRoot() {
        log.Lvl1("ROOT")
        return nil
    }

    err := p.SendTo(p.Agency, &Reply{make([]string, 0)})
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
