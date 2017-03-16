package main

import (
	//"fmt"
    //"errors"
    //"strconv"
	"github.com/BurntSushi/toml"
	"github.com/hm16083/ppcc/protocol"
	"github.com/hm16083/ppcc/lib"
	"gopkg.in/dedis/onet.v1"
	"gopkg.in/dedis/onet.v1/log"
	//"gopkg.in/dedis/onet.v1/network"
	"gopkg.in/dedis/onet.v1/simul/monitor"
	//"gopkg.in/dedis/onet.v1/simul"
)

/*
This is a simple ExampleChannels-protocol with two steps:
- announcement - which sends a message to all children
- reply - used for counting the number of children
*/

func init() {
	onet.SimulationRegister("PPCC", NewChannelSimulation)
}

// ChannelSimulation implements onet.Simulation.
type ChannelSimulation struct {
	onet.SimulationBFTree
}

// NewChannelSimulation is used internally to register the simulation.
func NewChannelSimulation(config string) (onet.Simulation, error) {
	es := &ChannelSimulation{}
	_, err := toml.Decode(config, es)
	if err != nil {
		return nil, err
	}
	return es, nil
}

// Setup implements onet.Simulation.
func (e *ChannelSimulation) Setup(dir string, hosts []string) (
	*onet.SimulationConfig, error) {
	sc := &onet.SimulationConfig{}
	e.CreateRoster(sc, hosts, 2000)
	err := e.CreateTree(sc)
	if err != nil {
		return nil, err
	}
	return sc, nil
}

var initSize int = 5

// Run implements onet.Simulation.
func (e *ChannelSimulation) Run(config *onet.SimulationConfig) error {
	size := config.Tree.Size()
	log.Lvl2("Size is:", size, "rounds:", e.Rounds)
	for round := 0; round < e.Rounds; round++ {
		log.Lvl1("Starting round", round)
		round := monitor.NewTimeMeasure("round")
		p, err := config.Overlay.CreateProtocol("PPCC", config.Tree, onet.NilServiceID)
		if err != nil {
			return err
		}

        warrant := lib.NewTriple("1234567890", 0, 0)
        warrant2 := lib.NewTriple("2031234567", 0, 0)
        queue := lib.NewQueue(initSize)
        queue.Push(warrant)
        queue.Push(warrant2)

        rh := p.(*protocol.PPCC)
        rh.Queue = queue

		go p.Start()
		children := <-p.(*protocol.PPCC).ChildCount
        //finished := <-p.(*protocol.PPCC).ProtocolDone
		round.Record()
        //log.Lvl1("Finished: ", finished)
        log.Lvl1("ChildCount returned:", children)
        /*
		if children != size {
			return errors.New("Didn't get " + strconv.Itoa(size) +
				" children")
		}
        */
	}

    log.Lvl1("Exiting Run()")

	return nil
}
