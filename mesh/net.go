package mesh

import (
	context "context"
	"errors"
	"fmt"
	"net"
)

var (
	ErrAllAddressesFailed = errors.New("all addresses failed to dial")
)

type Net struct {
	registry Registry
	accepted chan net.Conn
	reusable map[string]*reusableConn
	handlers map[string]Handler
	close    chan bool
}

func New(registry Registry) *Net {
	return &Net{
		registry: registry,
		close:    make(chan bool),
		accepted: make(chan net.Conn),
		reusable: map[string]*reusableConn{},
		handlers: map[string]Handler{
			"id":    &ID{},
			"yamux": &Yamux{},
		},
	}
}

func (n *Net) Dial(ctx context.Context, peerID string, commands ...string) (net.Conn, error) {
	peerInfo, err := n.registry.GetPeerInfo(peerID)
	if err != nil {
		return nil, err
	}
	var conn net.Conn
	if reusableConn, ok := n.reusable[peerID]; ok {
		newConn, err := reusableConn.NewConn()
		if err != nil {
			// TODO remove reusable conn
		} else {
			fmt.Println("dial reusing conn")
			conn = newConn
		}
	}
	if conn == nil {
		for _, addr := range peerInfo.Addresses {
			fmt.Println("dial dialing new conn")
			newConn, err := net.Dial("tcp", addr)
			if err != nil {
				// TODO blacklist address for a bit
				// TODO hold error maybe?
				// return nil, err
				continue
			}
			localAddress := peerAddress{
				network: "tcp",
				peerID:  n.registry.GetLocalPeerInfo().ID,
			}
			remoteAddress := peerAddress{
				network: "tcp",
				peerID:  peerID,
			}
			conn = NewAddressableConn(newConn, localAddress, remoteAddress)
			break
		}
	}
	if conn == nil {
		return nil, ErrAllAddressesFailed
	}

	finalConn, err := n.Select(conn, commands...)
	if err != nil {
		return nil, err
	}

	return finalConn, nil
}

func (n *Net) Select(conn net.Conn, commands ...string) (net.Conn, error) {
	for _, command := range commands {
		if err := WriteToken(conn, []byte(command)); err != nil {
			return nil, err
		}
		token, err := ReadToken(conn)
		if err != nil {
			return nil, err
		}
		fmt.Printf("Dialer got token %s for command %s\n", string(token), command)
		if string(token) != "ok" {
			return nil, errors.New("unexpected token response")
		}
		handler, ok := n.handlers[command]
		if !ok {
			return nil, errors.New("no such handler")
		}
		newConn, err := handler.Initiate(conn)
		if err != nil {
			return nil, err
		}
		// TODO maybe use a switch
		if reusableConn, ok := newConn.(*reusableConn); ok {
			go reusableConn.Accepted(n.accepted)
			// TODO lock
			fmt.Println("client storing reusable")
			n.reusable[conn.RemoteAddr().String()] = reusableConn
		}
		conn = newConn
	}
	return conn, nil
}

// TODO do we need to return a listener?
func (n *Net) Listen(addr string) (Listener, string, error) {
	tcpListener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, "", err
	}

	addresses := GetAddresses(tcpListener)

	n.registry.PutLocalPeerInfo(&PeerInfo{
		ID:        n.registry.GetLocalPeerInfo().ID,
		Addresses: addresses,
	})

	go func() {
		for {
			conn := <-n.accepted
			go n.HandleSelection(conn)
		}
	}()

	closed := false

	go func() {
		closed = true
		<-n.close
		fmt.Println("Closing")
		tcpListener.Close()
	}()

	go func() {
		for {
			conn, err := tcpListener.Accept()
			if err != nil {
				if closed {
					return
				}
				fmt.Println("Error accepting: ", err.Error())
				// TODO check conn is still alive and return
				return
			}
			localAddress := peerAddress{
				network: "tcp",
				peerID:  n.registry.GetLocalPeerInfo().ID,
			}
			remoteAddress := peerAddress{
				network: "tcp",
				peerID:  conn.RemoteAddr().String(),
			}
			conn = NewAddressableConn(conn, localAddress, remoteAddress)
			n.accepted <- conn
		}
	}()

	return nil, tcpListener.Addr().String(), nil
}

func (n *Net) Close() error {
	n.close <- true
	return nil
}

func (n *Net) HandleSelection(conn net.Conn) (net.Conn, error) {
	for {
		token, err := ReadToken(conn)
		if err != nil {
			return nil, err
		}
		fmt.Println("selection handler got token", string(token))
		handler, ok := n.handlers[string(token)]
		if !ok {
			if err := WriteToken(conn, []byte("error")); err != nil {
				return nil, err
			}
			return nil, errors.New("no such handler")
		}
		if err := WriteToken(conn, []byte("ok")); err != nil {
			return nil, err
		}
		newConn, err := handler.Handle(conn)
		if err != nil {
			return nil, err
		}
		// TODO maybe use a switch
		if reusableConn, ok := newConn.(*reusableConn); ok {
			go reusableConn.Accepted(n.accepted)
			// TODO lock
			fmt.Println("server storing reusable")
			n.reusable[conn.RemoteAddr().String()] = reusableConn
		}
		// TODO missing peerID to reuse conn
		// TODO maybe use a switch
		// if reusableConn, ok := newConn.(ReusableConn); ok {
		// 	// TODO lock
		// 	n.reusable[peerID] = reusableConn
		// }
		conn = newConn
	}
}