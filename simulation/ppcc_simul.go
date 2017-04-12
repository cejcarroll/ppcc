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
        //warrant := lib.NewTriple("1234567890", 0, 3)
        //queue := lib.NewQueue(initSize)
        //queue.Push(warrant)

        // Graph0
        nodeList0 := [6]lib.AgencyPair {
            lib.AgencyPair{"1234567890", 0},
            lib.AgencyPair{"1234567891", 0},
            lib.AgencyPair{"1234567892", 1},
            lib.AgencyPair{"1234567893", 0},
            lib.AgencyPair{"1234567894", 1},
            lib.AgencyPair{"1234567895", 0},
        }

        graph0 := lib.NewGraph(nodeList0[:])
        graph0.AddEdge(nodeList0[0], nodeList0[1])
        graph0.AddEdge(nodeList0[0], nodeList0[2])
        graph0.AddEdge(nodeList0[0], nodeList0[3])
        graph0.AddEdge(nodeList0[0], nodeList0[4])
        graph0.AddEdge(nodeList0[0], nodeList0[5])

        // Graph1
        nodeList1 := [5]lib.AgencyPair {
            lib.AgencyPair{"1234567896", 1},
            lib.AgencyPair{"1234567892", 1},
            lib.AgencyPair{"1234567897", 2},
            lib.AgencyPair{"1234567894", 1},
            lib.AgencyPair{"1234567898", 2},
        }

        graph1 := lib.NewGraph(nodeList1[:])
        graph1.AddEdge(nodeList1[0], nodeList1[1])
        graph1.AddEdge(nodeList1[0], nodeList1[2])
        graph1.AddEdge(nodeList1[0], nodeList1[3])
        graph1.AddEdge(nodeList1[0], nodeList1[4])

        // Graph2
        nodeList2 := [3]lib.AgencyPair {
            lib.AgencyPair{"1234567899", 2},
            lib.AgencyPair{"1234567897", 2},
            lib.AgencyPair{"1234567898", 2},
        }

        graph2 := lib.NewGraph(nodeList2[:])
        graph2.AddEdge(nodeList2[0], nodeList2[1])
        graph2.AddEdge(nodeList2[0], nodeList2[2])

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
