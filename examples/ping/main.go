package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"

	fabric "github.com/nimona/go-nimona-fabric"
)

func main() {
	peerA, err := newPeer("PeerA")
	if err != nil {
		log.Fatal("Could not create peer A", err)
	}

	peerB, err := newPeer("PeerB")
	if err != nil {
		log.Fatal("Could not create peer B", err)
	}

	log.Println("Peer A address:", peerA.GetAddresses())

	for _, addr := range peerA.GetAddresses() {
		endpoint := addr + "/tls/router/ping"
		log.Println("-------- Dialing", endpoint)
		if _, _, err := peerB.DialContext(context.Background(), endpoint); err != nil {
			log.Println("Dial error", err)
		}
	}
}

func newPeer(peerID string) (*fabric.Fabric, error) {
	ctx := context.Background()
	crt, err := GenX509KeyPair()
	if err != nil {
		fmt.Println("Cert creation error", err)
		return nil, err
	}

	yamux := &fabric.YamuxProtocol{}
	router := &fabric.RouterProtocol{}
	identity := &fabric.IdentityProtocol{Local: peerID}
	tls := &fabric.SecProtocol{
		Config: tls.Config{
			Certificates:       []tls.Certificate{crt},
			InsecureSkipVerify: true,
		},
	}
	ping := &Ping{}

	f := fabric.New(tls, router)
	f.AddTransport(fabric.NewTransportTCP("0.0.0.0", 0))
	f.AddTransport(fabric.NewTransportWebsocket("0.0.0.0", 0))

	f.AddProtocol(yamux)
	f.AddProtocol(identity)
	f.AddProtocol(ping)

	if err := f.Listen(ctx); err != nil {
		log.Fatal("Could not listen for peer A", err)
	}

	return f, nil
}
