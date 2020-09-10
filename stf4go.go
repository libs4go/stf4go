package stf4go

import (
	"context"
	"time"

	"github.com/libs4go/errors"
	"github.com/libs4go/scf4go"
	_ "github.com/libs4go/scf4go/codec" //
	"github.com/libs4go/slf4go"
	"github.com/multiformats/go-multiaddr"
)

// ScopeOfAPIError .
const errVendor = "stf4go"

// errors
var (
	ErrTransport = errors.New("transport load error", errors.WithVendor(errVendor), errors.WithCode(-1))
	ErrMultiAddr = errors.New("multiaddr error", errors.WithVendor(errVendor), errors.WithCode(-2))
	ErrPassword  = errors.New("password error", errors.WithVendor(errVendor), errors.WithCode(-3))
	ErrSign      = errors.New("signature invalid", errors.WithVendor(errVendor), errors.WithCode(-4))
)

var log = slf4go.Get("stf4go")

// Conn stf4go connection object
type Conn interface {
	// Read reads data from the connection.
	// Read can be made to time out and return an Error with Timeout() == true
	// after a fixed time limit; see SetDeadline and SetReadDeadline.
	Read(b []byte) (n int, err error)

	// Write writes data to the connection.
	// Write can be made to time out and return an Error with Timeout() == true
	// after a fixed time limit; see SetDeadline and SetWriteDeadline.
	Write(b []byte) (n int, err error)

	// Close closes the connection.
	// Any blocked Read or Write operations will be unblocked and return errors.
	Close() error

	// LocalAddr returns the local network address.
	LocalAddr() multiaddr.Multiaddr

	// RemoteAddr returns the remote network address.
	RemoteAddr() multiaddr.Multiaddr

	// SetDeadline sets the read and write deadlines associated
	// with the connection. It is equivalent to calling both
	// SetReadDeadline and SetWriteDeadline.
	//
	// A deadline is an absolute time after which I/O operations
	// fail with a timeout (see type Error) instead of
	// blocking. The deadline applies to all future and pending
	// I/O, not just the immediately following call to Read or
	// Write. After a deadline has been exceeded, the connection
	// can be refreshed by setting a deadline in the future.
	//
	// An idle timeout can be implemented by repeatedly extending
	// the deadline after successful Read or Write calls.
	//
	// A zero value for t means I/O operations will not time out.
	//
	// Note that if a TCP connection has keep-alive turned on,
	// which is the default unless overridden by Dialer.KeepAlive
	// or ListenConfig.KeepAlive, then a keep-alive failure may
	// also return a timeout error. On Unix systems a keep-alive
	// failure on I/O can be detected using
	// errors.Is(err, syscall.ETIMEDOUT).
	SetDeadline(t time.Time) error

	// SetReadDeadline sets the deadline for future Read calls
	// and any currently-blocked Read call.
	// A zero value for t means Read will not time out.
	SetReadDeadline(t time.Time) error

	// SetWriteDeadline sets the deadline for future Write calls
	// and any currently-blocked Write call.
	// Even if write times out, it may return n > 0, indicating that
	// some of the data was successfully written.
	// A zero value for t means Write will not time out.
	SetWriteDeadline(t time.Time) error
	// Underlying get underlying wrap conn
	Underlying() Conn
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
	String() string
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

			for i, j := 0, len(tunnels)-1; i < j; i, j = i+1, j-1 {
				tunnels[i], tunnels[j] = tunnels[j], tunnels[i]
			}

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
