package autocert

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/For-ACGN/autocert/certmgr"
)

func TestTLS01(t *testing.T) {
	manager := &certmgr.Manager{
		Prompt:     certmgr.AcceptTOS,
		HostPolicy: certmgr.HostWhitelist("example.com"),
	}
	tls01 := newTLS01(manager.GetCertificate)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := tls01.Start(ctx)
	require.NoError(t, err)

	err = tls01.Stop()
	require.NoError(t, err)
}
