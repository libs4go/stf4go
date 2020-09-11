package tcp

import (
	"context"
	"net"

	"github.com/libs4go/errors"
	"github.com/libs4go/slf4go"
	"github.com/libs4go/stf4go"
	"github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr-net"
)

type tcpTransport struct {
	slf4go.Logger
}

func newTCPTransport() *tcpTransport {
	return &tcpTransport{
		Logger: slf4go.Get("stf4go-transport-tcp"),
	}
}

func (transport *tcpTransport) String() string {
	return "stf4go-transport-tcp"
}

func (transport *tcpTransport) Protocols() []multiaddr.Protocol {
	return []multiaddr.Protocol{
		multiaddr.ProtocolWithName("tcp"),
	}
}

func (transport *tcpTransport) Listen(laddr multiaddr.Multiaddr, options *stf4go.Options) (stf4go.Listener, error) {

	transport.I("listen on {@laddr}", laddr.String())

	network, host, err := manet.DialArgs(laddr)

	if err != nil {
		return nil, errors.Wrap(err, "parser laddr %s error", laddr.String())
	}

	listener, err := net.Listen(network, host)

	if err != nil {
		return nil, errors.Wrap(err, "call net.Listen(%s,%s) error", network, host)
	}

	return &tcpListener{
		listener: listener,
		addr:     laddr,
	}, nil
}

func (transport *tcpTransport) Dial(ctx context.Context, raddr multiaddr.Multiaddr, options *stf4go.Options) (stf4go.Conn, error) {

	network, host, err := manet.DialArgs(raddr)

	if err != nil {
		return nil, errors.Wrap(err, "parser laddr %s error", raddr.String())
	}

	var dialer net.Dialer

	conn, err := dialer.DialContext(ctx, network, host)

	if err != nil {
		return nil, errors.Wrap(err, "call net.Dial(%s,%s) error", network, host)
	}

	return newTCPConn(conn)
}

type tcpListener struct {
	listener net.Listener
	addr     multiaddr.Multiaddr
}

func (listener *tcpListener) Close() error {
	return listener.listener.Close()
}

func (listener *tcpListener) Accept() (stf4go.Conn, error) {
	conn, err := listener.listener.Accept()

	if err != nil {
		return nil, errors.Wrap(err, "call accept on listener %s error", listener.addr.String())
	}

	return newTCPConn(conn)
}

func (listener *tcpListener) Addr() multiaddr.Multiaddr {
	return listener.addr
}

type tcpConn struct {
	net.Conn
	laddr multiaddr.Multiaddr
	raddr multiaddr.Multiaddr
}

func newTCPConn(conn net.Conn) (*tcpConn, error) {

	laddr, err := manet.FromNetAddr(conn.LocalAddr())

	if err != nil {
		return nil, errors.Wrap(err, "convert laddr %s to multiaddr error", conn.LocalAddr().String())
	}

	raddr, err := manet.FromNetAddr(conn.RemoteAddr())

	if err != nil {
		return nil, errors.Wrap(err, "convert laddr %s to multiaddr error", conn.LocalAddr().String())
	}

	return &tcpConn{
		Conn:  conn,
		laddr: laddr,
		raddr: raddr,
	}, nil
}

func (conn *tcpConn) LocalAddr() multiaddr.Multiaddr {
	return conn.laddr
}

func (conn *tcpConn) RemoteAddr() multiaddr.Multiaddr {
	return conn.raddr
}

func (conn *tcpConn) Underlying() stf4go.Conn {
	return nil
}

func init() {
	stf4go.RegisterTransport(newTCPTransport())
}
