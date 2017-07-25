package functionaltests

import (
	"context"
	"testing"

	"github.com/coreos/fleet/log"
	dht "github.com/nimona/go-nimona-kad-dht"
)

func setupTest() (*dht.DHTNode, *dht.DHTNode, *dht.DHTNode) {
	net1 := &dht.UDPNet{}
	net2 := &dht.UDPNet{}
	net3 := &dht.UDPNet{}
	// Start bootstrap node
	peer1 := &dht.Peer{dht.ID("a1"), []string{"127.0.0.1:8889"}}
	peer2 := &dht.Peer{dht.ID("a2"), []string{"127.0.0.1:8890"}}
	peer3 := &dht.Peer{dht.ID("a3"), []string{"127.0.0.1:8891"}}
	peer5 := &dht.Peer{dht.ID("a5"), []string{"127.0.0.1:8893"}}

	rt1 := dht.NewSimpleRoutingTable()
	rt2 := dht.NewSimpleRoutingTable()
	rt3 := dht.NewSimpleRoutingTable()

	rt1.Add(*peer2)
	rt2.Add(*peer1)
	rt2.Add(*peer3)
	rt3.Add(*peer5)
	rt3.Add(*peer2)

	node1 := dht.NewDHTNode([]*dht.Peer{peer2}, peer1, rt1, net1, "127.0.0.1:8889")
	node2 := dht.NewDHTNode([]*dht.Peer{peer1}, peer2, rt2, net2, "127.0.0.1:8890")
	node3 := dht.NewDHTNode([]*dht.Peer{peer2}, peer3, rt3, net3, "127.0.0.1:8891")

	return node1, node2, node3
}

// TODO: Create a node factory

func TestSuccessFindNode(t *testing.T) {
	n1, _, _ := setupTest()
	ctx := context.Background()

	peer, err := n1.Find(ctx, "a5")
	if err != nil {
		log.Error(err)
		t.Fail()
	}
	if peer.ID == "" {
		t.Fail()
	}

	log.Info(peer)
}

// func TestSuccessFindNodeLocal(t *testing.T) {
// 	n1, _, _ := setupTest()
// 	ctx := context.Background()

// 	peer, err := n1.Find(ctx, "a2")
// 	if err != nil {
// 		log.Error(err)
// 		t.Fail()
// 	}
// 	if peer.ID == "" {
// 		t.Fail()
// 	}

// 	log.Info(peer)
// }

// func TestGetNearNodes(t *testing.T) {
// 	n1, _, _ := setupTest()
// 	ctx := context.Background()

// 	peer, err := n1.Find(ctx, "a100")
// 	if err != nil {
// 		log.Error(err)
// 		t.Fail()
// 	}
// 	if peer.ID == "" {
// 		t.Fail()
// 	}
// }
