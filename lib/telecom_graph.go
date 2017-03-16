package lib

type AgencyPair struct {
    Node        string
    Telecom     int
}

type TelecomGraph struct {
    NumNodes    int
    Nodes       map[AgencyPair]bool
    Graph       map[AgencyPair][]AgencyPair
}

func NewGraph(nodeList []AgencyPair) *TelecomGraph {
    contains := make(map[AgencyPair]bool)
    for _, item := range nodeList {
        contains[item] = true
    }

    return &TelecomGraph {
        NumNodes:   len(nodeList),
        Nodes:      contains,
        Graph:      make(map[AgencyPair][]AgencyPair),
    }
}

func (g *TelecomGraph) ContainsNode(node AgencyPair) bool {
    return g.Nodes[node]
}

func (g *TelecomGraph) ContainsEdge(node1 AgencyPair, node2 AgencyPair) bool {
    for _, neighbor := range g.Graph[node1] {
        if neighbor == node2 {
            return true
        }
    }
    return false
}

func (g *TelecomGraph) AddEdge(node1 AgencyPair, node2 AgencyPair) {
    if !g.ContainsNode(node1) || !g.ContainsNode(node2) {
        return
    }

    g.Graph[node1] = append(g.Graph[node1], node2)
    g.Graph[node2] = append(g.Graph[node2], node1)
}

func (g *TelecomGraph) AddNode(node AgencyPair) {
    if g.ContainsNode(node) {
        return
    }

    g.NumNodes++
    g.Nodes[node] = true
}
