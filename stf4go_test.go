package stf4go

import (
	"testing"

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
	Size:       multiaddr.LengthPrefixedVarSize,
}

func init() {
	if err := multiaddr.AddProtocol(protoKCP); err != nil {
		panic(err)
	}
}
func TestMultiAddr(t *testing.T) {
	addr, err := multiaddr.NewMultiaddr("/ip4/127.0.0.1/udp/1812/kcp/xxxxxxxxxxx")

	require.NoError(t, err)

	require.NotNil(t, addr)

	addrs := multiaddr.Split(addr)

	for _, addr := range addrs {
		print(addr.String(), "|")
		for _, protocol := range addr.Protocols() {
			println(protocol.Name)
		}
	}
}
