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
	"gopkg.in/dedis/onet.v1/simul"
)

func init() {
	onet.SimulationRegister("PPCC", NewSimulation)
}

// ChannelSimulation implements onet.Simulation.
type Simulation struct {
	onet.SimulationBFTree
}

// NewSimulation is used internally to register the simulation.
func NewSimulation(config string) (onet.Simulation, error) {
	jvs := &Simulation{}
	_, err := toml.Decode(config, jvs)
	if err != nil {
		return nil, err
	}
	return jvs, nil
}

// Setup implements onet.Simulation.
func (jvs *Simulation) Setup(dir string, hosts []string) (*onet.SimulationConfig, error) {
	sim := &onet.SimulationConfig{}
	jvs.CreateRoster(sim, hosts, 2000)
	err := jvs.CreateTree(sim)
	if err != nil {
		return nil, err
	}
	return sim, nil
}

var initSize int = 5

// Run implements onet.Simulation.
func (e *Simulation) Run(config *onet.SimulationConfig) error {
	size := config.Tree.Size()
	log.Lvl2("Size is:", size, "rounds:", e.Rounds)
	for round := 0; round < e.Rounds; round++ {

        // Initialize Queue
        warrant := protocol.Warrant{"1234567890", 0, 3}

        // Read in graph files to use in simulation
        graph0, err0 := lib.ReadGraph("../graph0.tgf")
        if err0 != nil { return err0 }
        graph1, err1 := lib.ReadGraph("../graph1.tgf")
        if err1 != nil { return err1 }
        graph2, err2 := lib.ReadGraph("../graph2.tgf")
        if err2 != nil { return err2 }

        graphArr := [3]lib.TelecomGraph{*graph0, *graph1, *graph2}
        protocol.SetGraphs(graphArr[:])

		log.Lvl1("Starting round", round)
		round := monitor.NewTimeMeasure("round")
		p, err := config.Overlay.CreateProtocol("PPCC", config.Tree, onet.NilServiceID)
		if err != nil {
			return err
		}

        rh := p.(*protocol.PPCC)
        rh.InitWarrant = warrant

		go p.Start()
        done := <-rh.ProtocolDone
		round.Record()

        if done {
            log.Lvl1("Terminated successfully, output list: ", rh.OutputList)
        } else {
            log.Lvl1("ERROR: ProtocolDone returned false")
        }
    }

    log.Lvl3("Exiting Run()")
	return nil
}

func main() {
	simul.Start()
}
