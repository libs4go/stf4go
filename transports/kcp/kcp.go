package kcp

import (
	"context"
	"net"

	"github.com/libs4go/errors"
	"github.com/libs4go/scf4go"
	"github.com/libs4go/slf4go"
	"github.com/libs4go/stf4go"
	"github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr-net"
	kcpgo "github.com/xtaci/kcp-go"
)

const protocolKCPID = 482

var protoKCP = multiaddr.Protocol{
	Name:  "kcp",
	Code:  protocolKCPID,
	VCode: multiaddr.CodeToVarint(protocolKCPID),
}

var kcpMultiAddr multiaddr.Multiaddr

func init() {

	if err := multiaddr.AddProtocol(protoKCP); err != nil {
		panic(err)
	}

	var err error
	kcpMultiAddr, err = multiaddr.NewMultiaddr("/kcp")
	if err != nil {
		panic(err)
	}
}

type kcpTransport struct {
	slf4go.Logger
}

func newKCPTransport() *kcpTransport {
	return &kcpTransport{
		Logger: slf4go.Get("stf4go-transport-kcp"),
	}
}

func (transport *kcpTransport) Name() string {
	return "stf4go-transport-kcp"
}

func (transport *kcpTransport) Protocols() []multiaddr.Protocol {
	return []multiaddr.Protocol{
		protoKCP,
	}
}

func (transport *kcpTransport) Listen(laddr multiaddr.Multiaddr, config scf4go.Config) (stf4go.Listener, error) {

	network, host, err := manet.DialArgs(laddr)

	if err != nil {
		return nil, errors.Wrap(err, "parser laddr %s error", laddr.String())
	}

	addr, err := net.ResolveUDPAddr(network, host)

	if err != nil {
		return nil, errors.Wrap(err, "resolve udp addr %s %s error", network, host)
	}

	transport.I("listen on {@laddr}", addr.String())

	listener, err := kcpgo.Listen(addr.String())

	if err != nil {
		return nil, errors.Wrap(err, "listen %s error", addr.String())
	}

	maddr, err := manet.FromNetAddr(listener.Addr())

	if err != nil {
		return nil, errors.Wrap(err, "convert laddr %s to multiaddr error", listener.Addr().String())
	}

	maddr = maddr.Encapsulate(kcpMultiAddr)

	return &kcpListener{
		Logger:   transport.Logger,
		listener: listener,
		addr:     maddr,
	}, nil
}

func (transport *kcpTransport) Dial(ctx context.Context, raddr multiaddr.Multiaddr, config scf4go.Config) (stf4go.Conn, error) {

	network, host, err := manet.DialArgs(raddr)

	if err != nil {
		return nil, errors.Wrap(err, "parser laddr %s error", raddr.String())
	}

	addr, err := net.ResolveUDPAddr(network, host)

	if err != nil {
		return nil, errors.Wrap(err, "resolve udp addr %s %s error", network, host)
	}

	transport.I("dial to {@laddr}", addr.String())

	conn, err := kcpgo.Dial(addr.String())

	if err != nil {
		return nil, errors.Wrap(err, "kcp dial to %s error", addr.String())
	}

	transport.I("dial to {@laddr} -- success", addr.String())

	return newKCPConn(conn)
}

type kcpListener struct {
	slf4go.Logger
	listener net.Listener
	addr     multiaddr.Multiaddr
}

func (listener *kcpListener) Close() error {
	return listener.listener.Close()
}

func (listener *kcpListener) Accept() (stf4go.Conn, error) {
	listener.I("listener {@laddr} start accept", listener.listener.Addr().String())

	conn, err := listener.listener.Accept()

	listener.I("listener {@laddr} recv conn", listener.listener.Addr().String())

	if err != nil {
		return nil, errors.Wrap(err, "call accept on listener %s error", listener.addr.String())
	}

	return newKCPConn(conn)
}

func (listener *kcpListener) Addr() multiaddr.Multiaddr {
	return listener.addr
}

type kcpConn struct {
	net.Conn
	laddr multiaddr.Multiaddr
	raddr multiaddr.Multiaddr
}

func newKCPConn(conn net.Conn) (*kcpConn, error) {

	laddr, err := manet.FromNetAddr(conn.LocalAddr())

	if err != nil {
		return nil, errors.Wrap(err, "convert laddr %s to multiaddr error", conn.LocalAddr().String())
	}

	laddr = laddr.Encapsulate(kcpMultiAddr)

	raddr, err := manet.FromNetAddr(conn.RemoteAddr())

	if err != nil {
		return nil, errors.Wrap(err, "convert laddr %s to multiaddr error", conn.LocalAddr().String())
	}

	raddr = raddr.Encapsulate(kcpMultiAddr)

	return &kcpConn{
		Conn:  conn,
		laddr: laddr,
		raddr: raddr,
	}, nil
}

func (conn *kcpConn) LocalAddr() multiaddr.Multiaddr {
	return conn.laddr
}

func (conn *kcpConn) RemoteAddr() multiaddr.Multiaddr {
	return conn.raddr
}

func (conn *kcpConn) Underlying() *stf4go.Conn {
	return nil
}

func init() {
	stf4go.RegisterTransport(newKCPTransport())
}
