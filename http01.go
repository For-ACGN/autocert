package autocert

import (
	"net"
	"net/http"
)

type http01 struct {
	server *http.Server
}

func newHTTP01(handler http.Handler) *http01 {
	return &http01{
		server: &http.Server{Handler: handler},
	}
}

func (h *http01) Start() error {
	listener, err := net.Listen("tcp", ":80")
	if err != nil {
		return err
	}
	go func() {
		_ = h.server.Serve(listener)
	}()
	return nil
}

func (h *http01) Stop() error {
	return h.server.Close()
}
