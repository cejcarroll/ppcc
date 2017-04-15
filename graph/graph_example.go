package main

import (
    "github.com/hm16083/ppcc/lib"
    "fmt"
)

func main() {
    nodeList := [6]lib.AgencyPair {
        lib.AgencyPair{"1234567890", 0},
        lib.AgencyPair{"1234567891", 1},
        lib.AgencyPair{"1234567892", 2},
        lib.AgencyPair{"1234567893", 3},
        lib.AgencyPair{"1234567894", 4},
        lib.AgencyPair{"1234567895", 5},
    }

    fmt.Println()
    fmt.Println("Creating graph:")
    myGraph := lib.NewGraph(nodeList[:])
    //fmt.Println(myGraph)

    numErrs := 0
    //fmt.Println()
    fmt.Println("Making sure all nodes are contained in graph...")
    for _, val := range nodeList {
        if !myGraph.ContainsNode(val) {
            fmt.Printf("ERROR: Graph doesn't contain", val)
            numErrs++
        }
    }

    if numErrs == 0 {
        fmt.Println("PASS")
    } else {
        fmt.Println("FAIL: ", numErrs, " errors")
    }

    //fmt.Println()
    fmt.Println("Adding Edges:")
    myGraph.AddEdge(nodeList[0], nodeList[1], 1)
    myGraph.AddEdge(nodeList[1], nodeList[3], 1)
    myGraph.AddEdge(nodeList[1], nodeList[2], 1)
    myGraph.AddEdge(nodeList[5], nodeList[4], 1)

    //fmt.Println(myGraph)

    //fmt.Println()
    fmt.Println("Making sure edges are contained in graph:")

    if !myGraph.ContainsEdge(nodeList[0], nodeList[1]) ||
       !myGraph.ContainsEdge(nodeList[1], nodeList[3]) ||
       !myGraph.ContainsEdge(nodeList[1], nodeList[2]) ||
       !myGraph.ContainsEdge(nodeList[5], nodeList[4]) ||
       !myGraph.ContainsEdge(nodeList[1], nodeList[0]) ||
       !myGraph.ContainsEdge(nodeList[3], nodeList[1]) ||
       !myGraph.ContainsEdge(nodeList[2], nodeList[1]) ||
       !myGraph.ContainsEdge(nodeList[4], nodeList[5]) {
            fmt.Println("ERROR: Graph doesn't contain edge")
            numErrs++
   }

   fmt.Println("Neighbors of node 1: ", myGraph.Neighbors(nodeList[1]))

    if numErrs == 0 {
        fmt.Println("PASS")
    } else {
        fmt.Println("FAIL: ", numErrs, " errors")
    }

    fmt.Println()
}

