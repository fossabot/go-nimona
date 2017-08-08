package dht

import (
	"fmt"
	"testing"
)

func TestFindPeersNear(t *testing.T) {
	peer1 := &Peer{ID("a1"), []string{"127.0.0.1:8889"}}
	peer2 := &Peer{ID("a2"), []string{"127.0.0.1:8890"}}
	peer3 := &Peer{ID("a3"), []string{"127.0.0.1:8891"}}
	peer4 := &Peer{ID("a4"), []string{"127.0.0.1:8821"}}
	peer5 := &Peer{ID("a5"), []string{"127.0.0.1:8841"}}
	peer6 := &Peer{ID("a6"), []string{"127.0.0.1:8861"}}

	rt := NewSimpleRoutingTable()
	rt.Save(*peer2)
	rt.Save(*peer3)
	rt.Save(*peer4)
	rt.Save(*peer5)
	rt.Save(*peer6)

	net := &UDPNet{}

	node := NewDHTNode([]*Peer{peer2}, peer1, rt, net, "127.0.0.1:8889")
	peers, err := node.findPeersNear(peer6.ID, 3)
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}

	if len(peers) == 0 {
		t.Fail()
	}

	if peers[0].ID != peer6.ID {
		t.Fail()
	}
}
