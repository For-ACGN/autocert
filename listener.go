package autocert

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"os"
	"strings"
	"time"

	"github.com/For-ACGN/autocert/acme"
	"github.com/For-ACGN/autocert/certmgr"
)

// Config contains configuration about ACME Manager.
type Config struct {
	Domains []string
	IPAddrs []string
	Cache   certmgr.Cache
	Client  *acme.Client
}

type acListener struct {
	hosts []string

	listener  net.Listener
	manager   *certmgr.Manager
	tlsConfig *tls.Config

	portmap *portmap
	http01  *http01

	ctx    context.Context
	cancel context.CancelFunc
}

// Listen is used to listen a TLS listener with ACME.
func Listen(network, address string, config *Config) (net.Listener, error) {
	return ListenContext(context.Background(), network, address, config)
}

// ListenContext is used to listen a TLS listener with context.
func ListenContext(ctx context.Context, network, address string, config *Config) (net.Listener, error) {
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
	// prepare certificate cache directory
	cache := config.Cache
	if cache == nil {
		err = os.MkdirAll("certs", 0700)
		if err == nil {
			cache = certmgr.DirCache("certs")
		}
	}
	listener, err := listenContext(ctx, network, address)
	if err != nil {
		return nil, err
	}
	tl := &acListener{
		hosts:    allowList,
		listener: listener,
	}
	manager := &certmgr.Manager{
		Prompt:       certmgr.AcceptTOS,
		HostPolicy:   certmgr.HostWhitelist(allowList...),
		Cache:        cache,
		Client:       config.Client,
		BeforeVerify: tl.startChallenge,
		AfterVerify:  tl.stopChallenge,
	}
	tl.manager = manager
	tl.tlsConfig = manager.TLSConfig()
	tl.ctx, tl.cancel = context.WithCancel(context.Background())
	// dont need port map or HTTP01
	if strings.Contains(address, ":443") {
		go tl.trigger()
		return tl, nil
	}
	err = tryBindPort(ctx, "443")
	if err == nil {
		tl.portmap = newPortmap(network, port)
		go tl.trigger()
		return tl, nil
	}
	err = tryBindPort(ctx, "80")
	if err == nil {
		tl.http01 = newHTTP01(manager.HTTPHandler(nil))
		go tl.trigger()
		return tl, nil
	}
	return nil, errors.New("failed to bind port 443 and 80 for ACME")
}

func tryBindPort(ctx context.Context, port string) error {
	rd := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < 3; i++ {
		listener, err := listenContext(ctx, "tcp", ":"+port)
		if err != nil {
			wait := time.Duration(1+rd.Intn(5)) * time.Second
			select {
			case <-time.After(wait):
			case <-ctx.Done():
				return ctx.Err()
			}
			continue
		}
		return listener.Close()
	}
	return fmt.Errorf("failed to bind port: %s", port)
}

func tryBindListener(ctx context.Context, port string) (net.Listener, error) {
	rd := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < 3; i++ {
		listener, err := listenContext(ctx, "tcp", ":"+port)
		if err != nil {
			wait := time.Duration(1+rd.Intn(5)) * time.Second
			select {
			case <-time.After(wait):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
			continue
		}
		return listener, nil
	}
	return nil, fmt.Errorf("failed to bind listener on port: %s", port)
}

func listenContext(ctx context.Context, network, address string) (net.Listener, error) {
	var lc net.ListenConfig
	return lc.Listen(ctx, network, address)
}

func (l *acListener) startChallenge(ctx context.Context) error {
	if l.portmap != nil {
		return l.portmap.Start(ctx)
	}
	if l.http01 != nil {
		return l.http01.Start(ctx)
	}
	return nil
}

func (l *acListener) stopChallenge(ctx context.Context) error {
	if l.portmap != nil {
		return l.portmap.Stop()
	}
	if l.http01 != nil {
		return l.http01.Stop(ctx)
	}
	return nil
}

func (l *acListener) trigger() {
	l.preprovision()
	rd := rand.New(rand.NewSource(time.Now().UnixNano()))
	for {
		wait := time.Duration(rd.Intn(300)) * time.Second
		select {
		case <-time.After(wait):
			l.preprovision()
		case <-l.ctx.Done():
			return
		}
	}
}

func (l *acListener) preprovision() {
	for _, host := range l.hosts {
		hello := &tls.ClientHelloInfo{
			ServerName: host,
			CipherSuites: []uint16{
				tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			},
		}
		_, _ = l.manager.GetCertificate(hello)
	}
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
	l.cancel()
	return l.listener.Close()
}
