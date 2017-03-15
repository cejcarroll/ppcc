package protocol

import (
	"fmt"
	"github.com/hm16083/ppcc/lib"
	"github.com/hm16083/ppcc/protocol"
	"testing"
	"time"

	"gopkg.in/dedis/onet.v1"
	"gopkg.in/dedis/onet.v1/log"
	"gopkg.in/dedis/onet.v1/network"
)

func TestMain(m *testing.M) {
	log.MainTest(m)
}

// Tests a 2-node system
func TestNode(t *testing.T) {
	local := onet.NewLocalTest()
	nbrNodes := 2
	_, _, tree := local.GenTree(nbrNodes, true)
	defer local.CloseAll()

	p, err := local.StartProtocol("PPCC", tree)
	if err != nil {
		t.Fatal("Couldn't start protocol:", err)
	}
	protocolInstance := p.(*protocol.PPCC)
	timeout := network.WaitRetry * time.Duration(network.MaxRetryConnect*nbrNodes*2) * time.Millisecond
	select {
	case children := <-protocolInstance.ChildCount:
		log.Lvl2("Instance 1 is done")
        /*
		if children != nbrNodes {
			t.Fatal("Didn't get a child-cound of", nbrNodes)
		}
        */
	case <-time.After(timeout):
		t.Fatal("Didn't finish in time")
	}
}

}
