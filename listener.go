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

type acListener struct {
	network string
	port    string

	listener  net.Listener
	manager   *autocert.Manager
	tlsConfig *tls.Config

	portmap *portmap
	http01  *http01
}

// Listen is used to listen a TLS listener with ACME.
func Listen(network, address string, config *Config) (net.Listener, error) {
	_, port, err := net.SplitHostPort(address)
	if err != nil {
		return nil, err
	}
	var allowList []string
	allowList = append(allowList, config.Domains...)
	allowList = append(allowList, config.IPAddrs...)
	if len(allowList) < 1 {
		return nil, errors.New("must provide at least one domain or ip address")
	}
	listener, err := net.Listen(network, address)
	if err != nil {
		return nil, err
	}
	manager := &autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(allowList...),
	}
	tl := &acListener{
		network:   network,
		port:      port,
		listener:  listener,
		manager:   manager,
		tlsConfig: manager.TLSConfig(),
	}
	// dont need port map or HTTP01
	if strings.Contains(address, ":443") {
		return tl, nil
	}
	err = tryBindPort("443")
	if err == nil {
		tl.portmap = newPortmap(network, port)
		return tl, nil
	}
	err = tryBindPort("80")
	if err == nil {
		tl.http01 = newHTTP01(manager.HTTPHandler(nil))
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

func (l *acListener) Accept() (net.Conn, error) {
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

func (l *acListener) Addr() net.Addr {
	return l.listener.Addr()
}

func (l *acListener) Close() error {
	return l.listener.Close()
}
