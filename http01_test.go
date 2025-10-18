package autocert

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHTTP01(t *testing.T) {
	mux := http.NewServeMux()
	http01 := newHTTP01(mux)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := http01.Start(ctx)
	require.NoError(t, err)

	err = http01.Stop(ctx)
	require.NoError(t, err)
}
