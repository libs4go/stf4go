package stf4go

import (
	"context"
	"net"

	"github.com/libs4go/errors"
	"github.com/libs4go/scf4go"
	"github.com/libs4go/slf4go"
	"github.com/multiformats/go-multiaddr"
)

// ScopeOfAPIError .
const errVendor = "stf4go"

// errors
var (
	ErrTransport = errors.New("transport load error", errors.WithVendor(errVendor), errors.WithCode(-1))
)

var log = slf4go.Get("stf4go")

// Conn stf4go connection object
type Conn interface {
	net.Conn
	Underlying() *Conn
}

// Listener .
type Listener interface {
	Close() error
	Accept() (Conn, error)
	Addr() multiaddr.Multiaddr
}

// Transport stf4go tunnel transport base protocol
type Transport interface {
	// Name transport name
	Name() string
	// Transport support protocols
	Protocols() []multiaddr.Protocol
}

// NativeTransport .
type NativeTransport interface {
	Transport
	Listen(laddr multiaddr.Multiaddr, config scf4go.Config) (Listener, error)
	Dial(ctx context.Context, raddr multiaddr.Multiaddr, config scf4go.Config) (Conn, error)
}

// TunnelTransport .
type TunnelTransport interface {
	Transport
	Client(conn Conn, raddr multiaddr.Multiaddr, config scf4go.Config) (Conn, error)
	Server(conn Conn, laddr multiaddr.Multiaddr, config scf4go.Config) (Conn, error)
}

func lookupTransports(addr multiaddr.Multiaddr) ([]multiaddr.Multiaddr, NativeTransport, []TunnelTransport, error) {
	addrs := multiaddr.Split(addr)

	protocol := addrs[0].Protocols()[0]

	transport, ok := globalRegister.get(protocol.Name)

	if !ok {
		return nil, nil, nil, errors.Wrap(ErrTransport, "unsupport transport %s", protocol.Name)
	}

	nativeTransport, ok := transport.(NativeTransport)

	if !ok {
		return nil, nil, nil, errors.Wrap(ErrTransport, "first transport %s is not native transport", addrs[0].String())
	}

	var transports []TunnelTransport

	for _, addr := range addrs[1:] {

		protocol := addr.Protocols()[0]

		transport, ok := globalRegister.get(protocol.Name)

		if !ok {
			return nil, nil, nil, errors.Wrap(ErrTransport, "unsupport transport %s", protocol.Name)
		}

		tunnelTransport, ok := transport.(TunnelTransport)

		if !ok {
			return nil, nil, nil, errors.Wrap(ErrTransport, "first transport %s is not native transport", addrs[0].String())
		}

		transports = append(transports, tunnelTransport)
	}

	return addrs, nativeTransport, transports, nil
}
