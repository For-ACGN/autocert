package autocert

import (
	"net"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestListen(t *testing.T) {
	defer func() {
		err := os.RemoveAll("certs")
		require.NoError(t, err)
	}()

	config := &Config{
		Domains: []string{"example.com"},
	}

	t.Run("acme-tls", func(t *testing.T) {
		listener, err := Listen("tcp", "127.0.0.1:4000", config)
		require.NoError(t, err)

		err = listener.Close()
		require.NoError(t, err)
	})

	t.Run("http01", func(t *testing.T) {
		used, err := net.Listen("tcp", ":443")
		require.NoError(t, err)

		listener, err := Listen("tcp", "127.0.0.1:4000", config)
		require.NoError(t, err)

		err = listener.Close()
		require.NoError(t, err)

		err = used.Close()
		require.NoError(t, err)
	})
}
