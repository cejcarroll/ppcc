package lib

import (
    "gopkg.in/dedis/crypto.v0/abstract"
)

type Ciphertext struct {
    K   abstract.Point
    C   abstract.Point
}

type AgencyTriple struct {
    EncPhone    Ciphertext
    Telecom     int
    Depth       int
}

// https://gist.github.com/moraes/2141121
type AgencyQueue struct {
    nodes   []*AgencyTriple
    size    int
    head    int
    tail    int
    count   int
}

// Push adds a triple to the queue.
func (q *AgencyQueue) Push(n *AgencyTriple) {
	if q.head == q.tail && q.count > 0 {
		nodes := make([]*AgencyTriple, len(q.nodes)+q.size)
		copy(nodes, q.nodes[q.head:])
		copy(nodes[len(q.nodes)-q.head:], q.nodes[:q.head])
		q.head = 0
		q.tail = len(q.nodes)
		q.nodes = nodes
	}
	q.nodes[q.tail] = n
	q.tail = (q.tail + 1) % len(q.nodes)
	q.count++
}

// Pop removes and returns a triple from the queue
func (q *AgencyQueue) Pop() *AgencyTriple {
	if q.count == 0 {
		return nil
	}
	node := q.nodes[q.head]
	q.head = (q.head + 1) % len(q.nodes)
	q.count--
	return node
}

func (q *AgencyQueue) IsEmpty() bool {
	return q.count == 0
}

// NewQueue returns a new queue with the given initial size.
func NewQueue(size int) *AgencyQueue {
	return &AgencyQueue{
		nodes: make([]*AgencyTriple, size),
		size:  size,
	}
}

func NewTriple(node Ciphertext, telecom int, depth int) *AgencyTriple {
    return &AgencyTriple {
        EncPhone: node,
        Telecom: telecom,
        Depth: depth,
    }
}
