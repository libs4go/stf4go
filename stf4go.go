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

	count := len(addrs)

	var tunnels []TunnelTransport

	for i := 1; i < count; i++ {
		current := addrs[count-i]

		transport, ok := globalRegister.get(current.Protocols()[0].Name)

		if !ok {
			return nil, nil, nil, errors.Wrap(ErrTransport, "protocol %s not found", current.Protocols()[0].Name)
		}

		nativeTransport, ok := transport.(NativeTransport)

		if ok {
			addrs = append([]multiaddr.Multiaddr{
				multiaddr.Join(addrs[0 : count-i+1]...),
			}, addrs[count-i+1:]...)

			return addrs, nativeTransport, tunnels, nil
		}

		tunnelTransport, ok := transport.(TunnelTransport)

		if !ok {
			return nil, nil, nil, errors.Wrap(ErrTransport, "protocol %s must be tunnel transport", current.Protocols()[0].Name)
		}

		tunnels = append(tunnels, tunnelTransport)
	}

	return nil, nil, nil, errors.Wrap(ErrTransport, "expect native transport")
}
