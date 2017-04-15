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

