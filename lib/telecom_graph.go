package lib

import (
    "os"
    "bufio"
    "strings"
    "strconv"
    "fmt"
)

type AgencyPair struct {
    Node        string
    Telecom     int
}

type Edge struct {
    Pair    AgencyPair
    Weight  int
}

type TelecomGraph struct {
    NumNodes    int
    Nodes       map[AgencyPair]bool
    telecoms    map[string]int
    Visited     map[AgencyPair]bool
    Graph       map[AgencyPair][]Edge
}

func NewGraph(nodeList []AgencyPair) *TelecomGraph {
    contains := make(map[AgencyPair]bool)
    tcoms := make(map[string]int)
    for _, item := range nodeList {
        contains[item] = true
        tcoms[item.Node] = item.Telecom
    }

    return &TelecomGraph {
        NumNodes:   len(nodeList),
        Nodes:      contains,
        telecoms:   tcoms,
        Visited:    make(map[AgencyPair]bool),
        Graph:      make(map[AgencyPair][]Edge),
    }
}

func (g *TelecomGraph) Telecom (phone string) int {
    return g.telecoms[phone]
}

func (g *TelecomGraph) ContainsNode(node AgencyPair) bool {
    return g.Nodes[node]
}

func (g *TelecomGraph) HasVisited(node AgencyPair) bool {
    return g.Visited[node]
}

func (g *TelecomGraph) MarkVisited (node AgencyPair) {
    g.Visited[node] = true;
}

func (g *TelecomGraph) ContainsEdge(node1 AgencyPair, node2 AgencyPair) bool {
    for _, neighbor := range g.Graph[node1] {
        if neighbor.Pair == node2 {
            return true
        }
    }
    return false
}

func (g *TelecomGraph) Neighbors(node AgencyPair) []Edge {
    return g.Graph[node]
}

func (g *TelecomGraph) AddEdge(node1 AgencyPair, node2 AgencyPair, weight int) {
    if !g.ContainsNode(node1) || !g.ContainsNode(node2) {
        return
    }

    g.Graph[node1] = append(g.Graph[node1], Edge{node2, weight})
    g.Graph[node2] = append(g.Graph[node2], Edge{node1, weight})
}

func (g *TelecomGraph) AddNode(node AgencyPair) {
    if g.ContainsNode(node) {
        return
    }

    g.NumNodes++
    g.Nodes[node] = true
}

func ReadGraph (path string) (*TelecomGraph, error) {
	file, err := os.Open(path)
	if err != nil {
		fmt.Printf("%v\n", err)
		return nil, err
	}
	defer file.Close()

    var list []AgencyPair
    scanner := bufio.NewScanner(file)

    // First half of the TGF format:  Node definitions
    for scanner.Scan() {

        // Check for seperator
        if scanner.Text() == "#" {
            break
        }

        // Otherwise split into its components and form a pair
        data := strings.Fields(scanner.Text())
        phone := data[0]
        telecom, _ := strconv.Atoi(data[1])
        list = append(list, AgencyPair{phone, telecom})
    }

    graph := NewGraph(list)

    // Second half of TGF format: Edge definitions
    for scanner.Scan() {

        data := strings.Fields(scanner.Text())
        node1 := AgencyPair{data[0], graph.Telecom(data[0])}
        node2 := AgencyPair{data[1], graph.Telecom(data[1])}
        weight, _ := strconv.Atoi(data[2])

        graph.AddEdge(node1, node2, weight)
    }

    return graph, scanner.Err()
}
