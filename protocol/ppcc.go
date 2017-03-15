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
	network.RegisterMessage(Announce{})
	network.RegisterMessage(Reply{})
	network.RegisterMessage(Init{})
	onet.GlobalProtocolRegister("PPCC", NewPPCC)
}

// PPCC just holds a message that is passed to all children.
// It also defines a channel that will receive the number of children. Only the
// root-node will write to the channel.
type PPCC struct {
	*onet.TreeNodeInstance

	ChildCount      chan int

    ChannelInit     chan StructInit
	ChannelAnnounce chan StructAnnounce
    ChannelReply    chan StructReply

    NodeDone        bool
    ProtocolDone    chan bool
    Replies         int
	Queue           *lib.AgencyQueue

    NumTelecoms     int
    Telecoms        []*onet.TreeNode
}

// NewPPCC initialises the structure for use in one round
func NewPPCC(n *onet.TreeNodeInstance) (onet.ProtocolInstance, error) {
	c := &PPCC{
		TreeNodeInstance: n,
		ChildCount:       make(chan int),
	}
    c.NodeDone = false

	err := c.RegisterChannel(&c.ChannelAnnounce)
	if err != nil {
		return nil, errors.New("couldn't register announcement-channel: " + err.Error())
	}
	err = c.RegisterChannel(&c.ChannelReply)
	if err != nil {
		return nil, errors.New("couldn't register reply-channel: " + err.Error())
	}
	err = c.RegisterChannel(&c.ChannelInit)
	if err != nil {
		return nil, errors.New("couldn't register init-channel: " + err.Error())
	}
	return c, nil
}

// Start sends the Announce message to all children
func (p *PPCC) Start() error {
	log.Lvl1("Starting PPCC!!")
	//p.ChannelAnnounce <- StructAnnounce{nil, Announce{"Example is here"}}
    warrant := p.Queue.Pop()
    log.Lvl1("Warrant for phone: ", warrant.Node, "Telecom:", warrant.Telecom, "Depth:", warrant.Depth)

    out := &Init{}
	return p.handleInit(out)
}

func (p *PPCC) Dispatch() error {
    for {
        var err error
        select {
            case packet := <-p.ChannelAnnounce:
                err = p.handleAnnounce(&packet.Announce)
            case packet := <-p.ChannelReply:
                err = p.handleReply(&packet.Reply)
            case packet := <-p.ChannelInit:
                err = p.handleInit(&packet.Init)
        }

        if err != nil {
            log.Error("%v", err)
        }

        if p.NodeDone {
            if p.IsRoot() {
                log.Lvl1("Root is DONE")
                p.ChildCount <- 1
            } else {
                log.Lvl1("Child is DONE")
            }
            return nil
        }
    }
}

func (p *PPCC) handleInit (in *Init) error {

    if !p.IsRoot() {
        return fmt.Errorf("non-root node received Init message")
    }

    log.Lvl1("Root in handleAnnounce")
    for _, c := range p.Children() {
        err := p.SendTo(c, &Announce{"Hello from the NSA!"})
        if err != nil {
            log.Error(p.Info(), "failed to send to", c.Name(), err)
        }
    }

    return nil
}

func (p *PPCC) handleAnnounce (in *Announce) error {

    if p.IsRoot() {
        return fmt.Errorf("root received Announce message")
    }

    log.Lvl1("Child in handleAnnounce")
    err := p.SendTo(p.Parent(), &Reply{1})
    if err != nil {
        log.Lvl1("ERROR sending to parent")
        //log.Error(p.Info(), "failed to reply to", p.Parent().Name(), err)
        return err
    }
    p.NodeDone = true

    return nil
}

func (p *PPCC) handleReply(in *Reply) error {
    log.Lvl1("In handleReply")

    if !p.IsRoot() {
        return fmt.Errorf("non-root received reply")
    }

    p.Replies += 1
    if p.Replies == len(p.Children()) {
        log.Lvl1("Root received replies from all children")
        p.NodeDone = true
    }

    return nil
}
