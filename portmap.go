package autocert

import (
	"context"
	"crypto/tls"
	"net"

	"github.com/For-ACGN/autocert/acme"
)

type getCertificate = func(*tls.ClientHelloInfo) (*tls.Certificate, error)

type portmap struct {
	getCert  getCertificate
	listener net.Listener
}

func newPortmap(getCert getCertificate) *portmap {
	return &portmap{getCert: getCert}
}

func (p *portmap) Start(ctx context.Context) error {
	listener, err := tryBindListener(ctx, "443")
	if err != nil {
		return err
	}
	cfg := tls.Config{
		GetCertificate: p.getCert,
		NextProtos:     []string{acme.ALPNProto},
	}
	listener = tls.NewListener(listener, &cfg)
	p.listener = listener
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go p.sendCertificate(conn)
		}
	}()
	return nil
}

func (p *portmap) sendCertificate(conn net.Conn) {
	c := conn.(*tls.Conn)
	_ = c.Handshake()
}

func (p *portmap) Stop() error {
	return p.listener.Close()
}
