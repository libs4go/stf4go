package tcp

import (
	"context"
	"net"
	"sync"
	"testing"

	"github.com/libs4go/scf4go"
	"github.com/libs4go/scf4go/reader/memory"
	"github.com/libs4go/slf4go"
	_ "github.com/libs4go/slf4go/backend/console" //
	"github.com/libs4go/stf4go"
	"github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr/net"
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

func TestAddr(t *testing.T) {
	addr, err := net.ResolveTCPAddr("tcp", "[::1]:1812")

	require.NoError(t, err)

	maddr, err := manet.FromNetAddr(addr)

	require.NoError(t, err)

	println(maddr.String())
}

func TestListenConnect(t *testing.T) {

	laddr, err := multiaddr.NewMultiaddr("/ip4/127.0.0.1/tcp/1812")

	require.NoError(t, err)

	listener, err := stf4go.Listen(laddr)

	require.NoError(t, err)

	var wg sync.WaitGroup

	wg.Add(1)

	go func() {
		_, err := listener.Accept()

		require.NoError(t, err)

		wg.Done()
	}()

	_, err = stf4go.Dial(context.Background(), laddr)

	require.NoError(t, err)

	wg.Wait()
}
