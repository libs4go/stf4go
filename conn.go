package stf4go

import (
	"context"
	"net"
	"time"

	"github.com/libs4go/errors"
	"github.com/multiformats/go-multiaddr"
)

type wrapConn struct {
	conn  Conn
	laddr net.Addr
	raddr net.Addr
}

// WrapConn .
func WrapConn(conn Conn) (net.Conn, error) {

	laddr, err := ToNetAddr(conn.LocalAddr())

	if err != nil {
		return nil, err
	}

	raddr, err := ToNetAddr(conn.RemoteAddr())

	if err != nil {
		return nil, err
	}

	return &wrapConn{
		conn:  conn,
		laddr: laddr,
		raddr: raddr,
	}, nil
}

func (conn *wrapConn) Read(b []byte) (n int, err error) {
	return conn.conn.Read(b)
}
func (conn *wrapConn) Write(b []byte) (n int, err error) {
	return conn.conn.Write(b)
}
func (conn *wrapConn) Close() error {
	return conn.conn.Close()
}
func (conn *wrapConn) LocalAddr() net.Addr {
	return conn.laddr
}
func (conn *wrapConn) RemoteAddr() net.Addr {
	return conn.raddr
}
func (conn *wrapConn) SetDeadline(t time.Time) error {
	return conn.conn.SetDeadline(t)
}
func (conn *wrapConn) SetReadDeadline(t time.Time) error {
	return conn.conn.SetReadDeadline(t)
}
func (conn *wrapConn) SetWriteDeadline(t time.Time) error {
	return conn.conn.SetWriteDeadline(t)
}

// Dial .
func Dial(ctx context.Context, raddr multiaddr.Multiaddr, options ...Option) (Conn, error) {

	configWriter := newConfigWriter()

	for _, option := range options {
		if err := option(configWriter); err != nil {
			return nil, err
		}
	}

	if err := configWriter.config.Load(configWriter.readerWriter); err != nil {
		return nil, err
	}

	addrs, nativeTransport, tunnelTransports, err := lookupTransports(raddr)

	if err != nil {
		return nil, err
	}

	conn, err := nativeTransport.Dial(ctx, addrs[0], configWriter.config)

	if err != nil {
		return nil, errors.Wrap(err, "call native transport %s Dial error", nativeTransport.Name())
	}

	for i, tunnel := range tunnelTransports {
		conn, err = tunnel.Client(conn, addrs[i+1], configWriter.config)

		if err != nil {
			return nil, errors.Wrap(err, "call tunnel transport %s Client error", tunnel.Name())
		}
	}

	return conn, nil
}
