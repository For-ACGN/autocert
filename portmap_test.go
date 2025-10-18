package autocert

import (
	"context"
	"io"
	"net"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPortmap(t *testing.T) {
	t.Run("ipv4", func(t *testing.T) {
		listener, err := net.Listen("tcp", "127.0.0.1:4000")
		require.NoError(t, err)
		go func() {
			conn, err := listener.Accept()
			require.NoError(t, err)

			_, err = conn.Write([]byte("hello"))
			require.NoError(t, err)

			err = conn.Close()
			require.NoError(t, err)

			err = listener.Close()
			require.NoError(t, err)
		}()

		portmap := newPortmap("tcp", "4000")

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		err = portmap.Start(ctx)
		require.NoError(t, err)

		conn, err := net.Dial("tcp", "127.0.0.1:443")
		require.NoError(t, err)

		buf := make([]byte, 5)
		_, err = io.ReadFull(conn, buf)
		require.NoError(t, err)
		require.Equal(t, "hello", string(buf))

		err = conn.Close()
		require.NoError(t, err)

		err = portmap.Stop()
		require.NoError(t, err)
	})

	t.Run("ipv6", func(t *testing.T) {
		listener, err := net.Listen("tcp6", "[::1]:4000")
		require.NoError(t, err)
		go func() {
			conn, err := listener.Accept()
			require.NoError(t, err)

			_, err = conn.Write([]byte("hello"))
			require.NoError(t, err)

			err = conn.Close()
			require.NoError(t, err)

			err = listener.Close()
			require.NoError(t, err)
		}()

		portmap := newPortmap("tcp6", "4000")

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		err = portmap.Start(ctx)
		require.NoError(t, err)

		conn, err := net.Dial("tcp6", "[::1]:443")
		require.NoError(t, err)

		buf := make([]byte, 5)
		_, err = io.ReadFull(conn, buf)
		require.NoError(t, err)
		require.Equal(t, "hello", string(buf))

		err = conn.Close()
		require.NoError(t, err)

		err = portmap.Stop()
		require.NoError(t, err)
	})
}
