package autocert

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/For-ACGN/autocert/certmgr"
)

func TestPortmap(t *testing.T) {
	manager := &certmgr.Manager{
		Prompt:     certmgr.AcceptTOS,
		HostPolicy: certmgr.HostWhitelist("example.com"),
	}
	portmap := newPortmap(manager.GetCertificate)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := portmap.Start(ctx)
	require.NoError(t, err)

	err = portmap.Stop()
	require.NoError(t, err)
}
