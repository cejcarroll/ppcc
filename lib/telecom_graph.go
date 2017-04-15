package lib

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
    Visited     map[AgencyPair]bool
    Graph       map[AgencyPair][]Edge
}

func NewGraph(nodeList []AgencyPair) *TelecomGraph {
    contains := make(map[AgencyPair]bool)
    for _, item := range nodeList {
        contains[item] = true
    }

    return &TelecomGraph {
        NumNodes:   len(nodeList),
        Nodes:      contains,
        Visited:    make(map[AgencyPair]bool),
        Graph:      make(map[AgencyPair][]Edge),
    }
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
