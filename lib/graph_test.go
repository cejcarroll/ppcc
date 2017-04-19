package lib

import (
    "testing"
)

func TestTelecomGraph(t *testing.T) {
    nodeList := [6]AgencyPair {
        AgencyPair{"1234567890", 0},
        AgencyPair{"1234567891", 1},
        AgencyPair{"1234567892", 2},
        AgencyPair{"1234567893", 3},
        AgencyPair{"1234567894", 4},
        AgencyPair{"1234567895", 5},
    }

    myGraph := NewGraph(nodeList[:])

    for _, val := range nodeList {
        if !myGraph.ContainsNode(val) {
            panic("ERROR: Graph doesn't contain added node")
            //numErrs++
        }
    }

    myGraph.AddEdge(nodeList[0], nodeList[1], 1)
    myGraph.AddEdge(nodeList[1], nodeList[3], 1)
    myGraph.AddEdge(nodeList[1], nodeList[2], 1)
    myGraph.AddEdge(nodeList[5], nodeList[4], 1)

    if !myGraph.ContainsEdge(nodeList[0], nodeList[1]) ||
       !myGraph.ContainsEdge(nodeList[1], nodeList[3]) ||
       !myGraph.ContainsEdge(nodeList[1], nodeList[2]) ||
       !myGraph.ContainsEdge(nodeList[5], nodeList[4]) ||
       !myGraph.ContainsEdge(nodeList[1], nodeList[0]) ||
       !myGraph.ContainsEdge(nodeList[3], nodeList[1]) ||
       !myGraph.ContainsEdge(nodeList[2], nodeList[1]) ||
       !myGraph.ContainsEdge(nodeList[4], nodeList[5]) {
            panic("ERROR: Graph doesn't contain edge")
            //numErrs++
   }

   println("PASS: Graph test")
}

func TestTGFReader(t *testing.T) {
    graph, _ := ReadGraph("tgf_example.tgf")

    nodeList := [6]AgencyPair {
        AgencyPair{"1234567890", 0},
        AgencyPair{"1234567891", 1},
        AgencyPair{"1234567892", 2},
        AgencyPair{"1234567893", 3},
        AgencyPair{"1234567894", 4},
        AgencyPair{"1234567895", 5},
    }

    for _, val := range nodeList {
        if !graph.ContainsNode(val) {
            panic("ERROR:  graph doesn't contain node")
        }
    }

    println("PASS: TGF node test")

    if !graph.ContainsEdge(nodeList[0], nodeList[1]) ||
       !graph.ContainsEdge(nodeList[1], nodeList[3]) ||
       !graph.ContainsEdge(nodeList[1], nodeList[2]) ||
       !graph.ContainsEdge(nodeList[5], nodeList[4]) ||
       !graph.ContainsEdge(nodeList[1], nodeList[0]) ||
       !graph.ContainsEdge(nodeList[3], nodeList[1]) ||
       !graph.ContainsEdge(nodeList[2], nodeList[1]) ||
       !graph.ContainsEdge(nodeList[4], nodeList[5]) {
            panic("ERROR: Graph doesn't contain edge")
   }

    println("PASS: TGF edge test")
}
