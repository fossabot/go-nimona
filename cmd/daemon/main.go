package main

import (
	"context"
	"log"
	"os"
	"strings"

	"github.com/nimona/go-nimona/dht"
	"github.com/nimona/go-nimona/mesh"
	"github.com/nimona/go-nimona/net"
	"github.com/nimona/go-nimona/net/protocol"

	ishell "gopkg.in/abiosoft/ishell.v2"
)

func main() {
	peerID := os.Getenv("PEER_ID")
	if peerID == "" {
		log.Fatal("Missing PEER_ID")
	}

	bs := []string{}
	port := 0

	if peerID == "bootstrap" {
		port = 26801
	} else {
		bs = append(bs, "tcp:localhost:26801/router/messaging")
	}

	ctx := context.Background()
	tcp := net.NewTransportTCPWithUPNP("0.0.0.0", port)

	nn := net.New(ctx)
	rt := protocol.NewRouter()

	ps, _ := mesh.NewPubSub()
	rg, _ := mesh.NewRegisty(peerID, ps)
	ms, _ := mesh.NewMesh(nn, ps, rg)
	mg, _ := mesh.NewMessenger(ms)
	ds, _ := dht.NewDHT(ps, peerID, bs...)

	nn.AddProtocols(mg)

	rt.AddRoute(mg)

	nn.AddTransport(tcp, rt)

	if peerID == "bootstrap" {
		// ds.Put(ctx, "a", "a", map[string]string{})
	}

	shell := ishell.New()
	shell.Println("Nimona DHT")

	// handle get
	shell.AddCmd(&ishell.Cmd{
		Name: "get",
		Func: func(c *ishell.Context) {
			c.ShowPrompt(false)
			defer c.ShowPrompt(true)

			if len(c.Args) == 0 {
				c.Println("Missing key")
				return
			}

			key := c.Args[0]
			ctx := context.Background()
			rs, err := ds.Get(ctx, key)
			if err != nil {
				c.Printf("Could not get %s\n", key)
				c.Printf("Error: %s\n", err)
			}
			c.ProgressBar().Indeterminate(true)
			c.ProgressBar().Start()
			for rv := range rs {
				c.Println("  - " + rv.GetValue())
			}
			c.ProgressBar().Stop()
		},
		Help: "get a value from the dht",
	})

	// handle put
	shell.AddCmd(&ishell.Cmd{
		Name: "put",
		Func: func(c *ishell.Context) {
			c.ShowPrompt(false)
			defer c.ShowPrompt(true)

			if len(c.Args) < 2 {
				c.Println("Missing key and value")
				return
			}

			key := c.Args[0]
			val := strings.Join(c.Args[1:], " ")
			ctx := context.Background()
			if err := ds.Put(ctx, key, val, map[string]string{}); err != nil {
				c.Printf("Could not get %s\n", key)
				c.Printf("Error: %s\n", err)
			}
		},
		Help: "put a value on the dht",
	})

	// handle put
	shell.AddCmd(&ishell.Cmd{
		Name: "list",
		Func: func(c *ishell.Context) {
			c.ShowPrompt(false)
			defer c.ShowPrompt(true)

			ps, _ := ds.GetLocalPairs()
			for key, vals := range ps {
				c.Println("* " + key)
				for _, val := range vals {
					c.Printf("  - %s (%#v)\n", val.GetValue(), val.GetLabels())
				}
			}
		},
		Help: "list all values stored in our local dht",
	})

	// when started with "exit" as first argument, assume non-interactive execution
	if len(os.Args) > 1 && os.Args[1] == "exit" {
		shell.Process(os.Args[2:]...)
	} else {
		// start shell
		shell.Run()
		// teardown
		shell.Close()
	}
}