package autocert

import (
	"context"
	"io"
	"net"
)

type portmap struct {
	network  string
	port     string
	listener net.Listener
}

func newPortmap(network, port string) *portmap {
	return &portmap{
		network: network,
		port:    port,
	}
}

func (p *portmap) Start(ctx context.Context) error {
	listener, err := tryBindListener(ctx, "443")
	if err != nil {
		return err
	}
	p.listener = listener
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go p.forward(conn)
		}
	}()
	return nil
}

func (p *portmap) forward(conn net.Conn) {
	var host string
	switch p.network {
	case "tcp", "tcp4":
		host = "127.0.0.1"
	case "tcp6":
		host = "::1"
	}
	address := net.JoinHostPort(host, p.port)
	remote, err := net.Dial(p.network, address)
	if err != nil {
		return
	}
	go func() {
		defer func() { _ = remote.Close() }()
		_, _ = io.Copy(conn, remote)
	}()
	go func() {
		defer func() { _ = conn.Close() }()
		_, _ = io.Copy(remote, conn)
	}()
}

func (p *portmap) Stop() error {
	return p.listener.Close()
}
