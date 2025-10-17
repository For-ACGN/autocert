package autocert

import (
	"context"
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

func (h *http01) Start(ctx context.Context) error {
	listener, err := tryBindListener(ctx, "80")
	if err != nil {
		return err
	}
	go func() {
		_ = h.server.Serve(listener)
	}()
	return nil
}

func (h *http01) Stop(ctx context.Context) error {
	return h.server.Shutdown(ctx)
}
