package openairt_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	openairt "github.com/WqyJh/go-openai-realtime"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"
)

func TestGorillaWebSocket(t *testing.T) {
	s := NewTestServer(t, time.Millisecond)
	defer s.Server.Close()

	dialer := openairt.NewGorillaWebSocketDialer(openairt.GorillaWebSocketOptions{})

	conn, err := dialer.Dial(context.Background(), s.URL, nil)
	require.NoError(t, err)
	require.NotNil(t, conn)
	defer func() {
		err = conn.Close()
		require.NoError(t, err)
	}()

	err = conn.WriteMessage(context.Background(), openairt.MessageBinary+1, []byte("hello"))
	require.ErrorIs(t, err, openairt.ErrUnsupportedMessageType)

	err = conn.WriteMessage(context.Background(), openairt.MessageBinary, []byte("hello"))
	require.NoError(t, err)

	msgType, data, err := conn.ReadMessage(context.Background())
	require.NoError(t, err)
	require.Equal(t, openairt.MessageBinary, msgType)
	require.Equal(t, []byte("hello"), data)

	err = conn.WriteMessage(context.Background(), openairt.MessageText, []byte("world"))
	require.NoError(t, err)

	msgType, data, err = conn.ReadMessage(context.Background())
	require.NoError(t, err)
	require.Equal(t, openairt.MessageText, msgType)
	require.Equal(t, []byte("world"), data)

	ctx, cancel := context.WithTimeout(context.Background(), time.Microsecond*5)
	defer cancel()

	_, _, err = conn.ReadMessage(ctx)
	var permanent *openairt.PermanentError
	require.ErrorAs(t, err, &permanent)
	require.ErrorContains(t, permanent.Err, "i/o timeout")
}

func TestGorillaWebSocketReadLimitError(t *testing.T) {
	s := NewTestServer(t, time.Millisecond)
	defer s.Server.Close()

	dialer := openairt.NewGorillaWebSocketDialer(openairt.GorillaWebSocketOptions{
		ReadLimit: 5,
	})

	conn, err := dialer.Dial(context.Background(), s.URL, nil)
	require.NoError(t, err)
	require.NotNil(t, conn)

	err = conn.WriteMessage(context.Background(), openairt.MessageBinary, []byte("hello world"))
	require.NoError(t, err)

	_, _, err = conn.ReadMessage(context.Background())
	require.Error(t, err)
	var permanent *openairt.PermanentError
	require.ErrorAs(t, err, &permanent)
	t.Logf("error: %v", permanent.Err)
	require.ErrorIs(t, permanent.Err, websocket.ErrReadLimit)

	err = conn.Close()
	require.NoError(t, err)
}

func TestGorillaWebSocketReadLimitOK(t *testing.T) {
	s := NewTestServer(t, time.Millisecond)
	defer s.Server.Close()

	dialer := openairt.NewGorillaWebSocketDialer(openairt.GorillaWebSocketOptions{
		ReadLimit: 11,
	})

	conn, err := dialer.Dial(context.Background(), s.URL, nil)
	require.NoError(t, err)
	require.NotNil(t, conn)
	defer func() {
		err = conn.Close()
		require.NoError(t, err)
	}()

	err = conn.WriteMessage(context.Background(), openairt.MessageBinary, []byte("hello world"))
	require.NoError(t, err)

	msgType, data, err := conn.ReadMessage(context.Background())
	require.NoError(t, err)
	require.Equal(t, openairt.MessageBinary, msgType)
	require.Equal(t, []byte("hello world"), data)
}

func TestGorillaWebSocketDialOptions(t *testing.T) {
	s := NewTestServer(t, time.Millisecond)
	defer s.Server.Close()

	dialer := openairt.NewGorillaWebSocketDialer(openairt.GorillaWebSocketOptions{
		Dialer: &websocket.Dialer{},
	})

	conn, err := dialer.Dial(context.Background(), s.URL, http.Header{
		"X-Test":  {"test"},
		"X-Test2": {"test2", "test3"},
	})
	require.NoError(t, err)
	require.NotNil(t, conn)

	header := conn.Response().Header
	require.Equal(t, "test", header.Get("X-Test"))
	require.Equal(t, []string{"test2", "test3"}, header["X-Test2"])
}
