package autocert

import (
	"context"
	"crypto/tls"
	"net"

	"github.com/For-ACGN/autocert/acme"
)

type getCertificate = func(*tls.ClientHelloInfo) (*tls.Certificate, error)

type tls01 struct {
	getCert  getCertificate
	listener net.Listener
}

func newTLS01(getCert getCertificate) *tls01 {
	return &tls01{getCert: getCert}
}

func (p *tls01) Start(ctx context.Context) error {
	listener, err := tryBindListener(ctx, "443")
	if err != nil {
		return err
	}
	cfg := tls.Config{
		GetCertificate: p.getCert,
		NextProtos:     []string{acme.ALPNProto},
	}
	p.listener = tls.NewListener(listener, &cfg)
	go func() {
		for {
			conn, err := p.listener.Accept()
			if err != nil {
				return
			}
			go p.handshake(conn)
		}
	}()
	return nil
}

func (p *tls01) handshake(conn net.Conn) {
	c := conn.(*tls.Conn)
	_ = c.Handshake()
	_ = c.Close()
}

func (p *tls01) Stop() error {
	return p.listener.Close()
}
