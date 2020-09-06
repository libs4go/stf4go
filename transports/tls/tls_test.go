package tls

import (
	"bytes"
	"context"
	"testing"

	"github.com/libs4go/bcf4go/key"
	"github.com/libs4go/scf4go"
	"github.com/libs4go/scf4go/reader/memory"
	"github.com/libs4go/slf4go"
	_ "github.com/libs4go/slf4go/backend/console" //
	"github.com/libs4go/stf4go"
	_ "github.com/libs4go/stf4go/transports/kcp" //
	_ "github.com/libs4go/stf4go/transports/tcp" //
	"github.com/multiformats/go-multiaddr"
	"github.com/stretchr/testify/require"
)

var loggerjson = `
{
	"default":{
		"backend":"console",
		"level":"debug"
	},
	"backend":{
		"console":{
			"formatter":{
				"output": "@t @l @s @m"
			}
		}
	}
}
`

func init() {
	config := scf4go.New()

	err := config.Load(memory.New(memory.Data(loggerjson, "json")))

	if err != nil {
		panic(err)
	}

	err = slf4go.Config(config)

	if err != nil {
		panic(err)
	}
}

func newKeyStore(t *testing.T) stf4go.Option {
	k, err := key.RandomKey("eth")

	require.NoError(t, err)

	var buff bytes.Buffer

	err = key.Encode("web3.standard", k.PriKey(), key.Property{
		"password": "test",
	}, &buff)

	require.NoError(t, err)

	return KeyWeb3(buff.Bytes())
}

func TestListenConnect(t *testing.T) {

	laddr, err := multiaddr.NewMultiaddr("/ip4/127.0.0.1/udp/1813/kcp/tls")

	require.NoError(t, err)

	listener, err := stf4go.Listen(laddr, KeyProvider("eth"), KeyPassword("test"), newKeyStore(t))

	require.NoError(t, err)

	println(listener.Addr().String())

	go func() {

		_, err := stf4go.Dial(context.Background(), laddr, KeyProvider("eth"), KeyPassword("test"), newKeyStore(t))

		require.NoError(t, err)

		// _, err = conn.Write([]byte("hello world"))

		// require.NoError(t, err)
	}()

	conn, err := listener.Accept()

	require.NoError(t, err)

	// var buff [32]byte

	// _, err = conn.Read(buff[:])

	// require.NoError(t, err)

	<-conn.(Conn).RemoteKey()
}
