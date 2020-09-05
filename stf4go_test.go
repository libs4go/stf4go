package stf4go

import (
	"context"
	"testing"

	"github.com/libs4go/scf4go"
	"github.com/multiformats/go-multiaddr"
	"github.com/stretchr/testify/require"
)

func kcpS2B(val string) ([]byte, error) {
	return []byte(val), nil
}

func kcpB2S(s []byte) (string, error) {
	return string(s), nil
}

func kcpValue([]byte) error {
	return nil
}

var TranscoderKCP = multiaddr.NewTranscoderFromFunctions(kcpS2B, kcpB2S, kcpValue)

const protocolKCPID = 482

var protoKCP = multiaddr.Protocol{
	Name:       "kcp",
	Code:       protocolKCPID,
	VCode:      multiaddr.CodeToVarint(protocolKCPID),
	Transcoder: TranscoderKCP,
	Size:       0,
}

const protocolP2PID = 483

var protoP2P = multiaddr.Protocol{
	Name:       "p2p2",
	Code:       protocolP2PID,
	VCode:      multiaddr.CodeToVarint(protocolP2PID),
	Transcoder: TranscoderKCP,
	Size:       multiaddr.LengthPrefixedVarSize,
}

type testKCPTransport struct {
}

func (transport *testKCPTransport) Name() string {
	return "kcp"
}

func (transport *testKCPTransport) Protocols() []multiaddr.Protocol {
	return []multiaddr.Protocol{protoKCP}
}

func (transport *testKCPTransport) Listen(laddr multiaddr.Multiaddr, config scf4go.Config) (Listener, error) {
	return nil, nil
}

func (transport *testKCPTransport) Dial(ctx context.Context, raddr multiaddr.Multiaddr, config scf4go.Config) (Conn, error) {
	return nil, nil
}

type testP2PTransport struct {
}

func (transport *testP2PTransport) Name() string {
	return "p2p2"
}

func (transport *testP2PTransport) Protocols() []multiaddr.Protocol {
	return []multiaddr.Protocol{protoP2P}
}

func (transport *testP2PTransport) Client(conn Conn, raddr multiaddr.Multiaddr, config scf4go.Config) (Conn, error) {
	return nil, nil
}

func (transport *testP2PTransport) Server(conn Conn, laddr multiaddr.Multiaddr, config scf4go.Config) (Conn, error) {
	return nil, nil
}

func init() {
	if err := multiaddr.AddProtocol(protoKCP); err != nil {
		panic(err)
	}

	if err := multiaddr.AddProtocol(protoP2P); err != nil {
		panic(err)
	}

	RegisterTransport(&testKCPTransport{})
	RegisterTransport(&testP2PTransport{})
}

func TestLookupTransports(t *testing.T) {
	addr, err := multiaddr.NewMultiaddr("/ip4/127.0.0.1/udp/1812/kcp/p2p2/xxxxxxxxxxx")

	require.NoError(t, err)

	require.NotNil(t, addr)

	addrs, native, tunnels, err := lookupTransports(addr)

	require.NoError(t, err)

	require.Equal(t, len(addrs), 2)

	for _, addr := range addrs {
		println(addr.String())
	}

	require.Equal(t, native.Name(), "kcp")

	require.Equal(t, len(tunnels), 1)

	require.Equal(t, tunnels[0].Name(), "p2p2")

}

func TestLookupTransportException(t *testing.T) {
	addr, err := multiaddr.NewMultiaddr("/ip4/127.0.0.1/udp/1812/")

	require.NoError(t, err)

	require.NotNil(t, addr)

	_, _, _, err = lookupTransports(addr)

	require.Error(t, err, "detect native transport test failed")

	_, err = multiaddr.NewMultiaddr("/ip4/127.0.0.1/udp/1812/fs")

	require.Error(t, err, "")
}
