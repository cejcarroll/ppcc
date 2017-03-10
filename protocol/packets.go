package protocol

import (
	"gopkg.in/dedis/crypto.v0/abstract"
	"gopkg.in/dedis/onet.v1"
)

type Init struct {
}

type ElgEncryptedMessage struct {
	Content   []map[int]abstract.Point
	Users     map[int]int
	NumPhones int
	Sets      int
	ID        int
	NumAuthorities int
}

type FullyPhEncryptedMessage struct {
	Content []abstract.Point
	Users   map[int]int
	Mode    int
	Sets    int
	ID      int
}

type PartiallyPhDecryptedMessage struct {
	Content []abstract.Point
	Users   map[int]int
	Mode    int
	Sets    int
	ID      int
	WantTodec int
}

type PlainMessage struct {
	Content []string
	Users   map[int]int
	Mode    int
	ID      int
}

type DoneMessage struct {
	Status int
	Src  int
	Sets int
}


type chanElgEncryptedMessage struct {
	*onet.TreeNode
	ElgEncryptedMessage
}

type chanFullyPhEncryptedMessage struct {
	*onet.TreeNode
	FullyPhEncryptedMessage
}
type chanPartiallyPhDecryptedMessage struct {
	*onet.TreeNode
	PartiallyPhDecryptedMessage
}

type chanPlainMessage struct {
	*onet.TreeNode
	PlainMessage
}
type chanDoneMessage struct {
	*onet.TreeNode
	DoneMessage
}

type chanInitiateRequest struct {
	*onet.TreeNode
	Init
}

// Announce is used to pass a message to all children
type Announce struct {
	Message string
}

// StructAnnounce contains Announce and the data necessary to identify the
// message in the onet framework.
type StructAnnounce struct {
	*onet.TreeNode
	Announce
}

// Reply returns the count of all children.
type Reply struct {
	ChildrenCount int
}

// StructReply contains Reply and the data necessary to identify the
// message in the onet framework.
type StructReply struct {
	*onet.TreeNode
	Reply
}
