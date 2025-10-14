package autocert

import (
	"crypto/tls"
	"errors"
	"net"
	"strings"
	"time"

	"github.com/For-ACGN/autocert/internel/acme"
	"github.com/For-ACGN/autocert/internel/autocert"
)

// Config contains configuration about ACME Manager.
type Config struct {
	Domains []string
	IPAddrs []string
	Client  *acme.Client
}

type tlsListener struct {
	listener  net.Listener
	manager   *autocert.Manager
	tlsConfig *tls.Config
}

// Listen is used to listen a TLS listener with ACME.
func Listen(network, address string) (net.Listener, error) {
	listener, err := net.Listen(network, address)
	if err != nil {
		return nil, err
	}
	manager := &autocert.Manager{
		Prompt: autocert.AcceptTOS,
	}
	tl := &tlsListener{
		listener:  listener,
		manager:   manager,
		tlsConfig: manager.TLSConfig(),
	}
	// dont need port forward
	if strings.Contains(address, ":443") {
		return tl, nil
	}

	err = tryBindPort("443")
	if err == nil {
		return tl, nil
	}
	err = tryBindPort("80")
	if err == nil {
		return tl, nil
	}
	return nil, errors.New("failed to bind port 443 and 80 for ACME")
}

func tryBindPort(port string) error {
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}
	return listener.Close()
}

func (l *tlsListener) Accept() (net.Conn, error) {
	conn, err := l.listener.Accept()
	if err != nil {
		return nil, err
	}
	tcpConn := conn.(*net.TCPConn)

	// Because Listener is a convenience function, help out with
	// this too.  This is not possible for the caller to set once
	// we return a *tcp.Conn wrapping an inaccessible net.Conn.
	// If callers don't want this, they can do things the manual
	// way and tweak as needed. But this is what net/http does
	// itself, so copy that. If net/http changes, we can change
	// here too.
	_ = tcpConn.SetKeepAlive(true)
	_ = tcpConn.SetKeepAlivePeriod(3 * time.Minute)

	return tls.Server(tcpConn, l.tlsConfig), nil
}

func (l *tlsListener) Addr() net.Addr {
	return l.listener.Addr()
}

func (l *tlsListener) Close() error {
	return l.listener.Close()
}
