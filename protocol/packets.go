package protocol

import (
	"gopkg.in/dedis/crypto.v0/abstract"
	"gopkg.in/dedis/onet.v1"
    //"github.com/hm16083/ppcc/lib"
)

type Init struct {
}

type StructInit struct {
	*onet.TreeNode
	Init
}

type Done struct {
}

type StructDone struct {
    *onet.TreeNode
    Done
}

type Reply struct {
    EncPhones   []abstract.Point
    EncTelecoms []abstract.Point
}

type StructReply struct {
	*onet.TreeNode
	Reply
}

type AuthorityQuery struct {
    Query   string
    Telecom int
    Depth   int
}

type StructAuthorityQuery struct {
    *onet.TreeNode
    AuthorityQuery
}
