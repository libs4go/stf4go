package stf4go

import (
	"net"

	"github.com/libs4go/errors"
	"github.com/multiformats/go-multiaddr"
	mnet "github.com/multiformats/go-multiaddr/net"
)

type wrapListener struct {
	listener Listener
}

// WrapListener wrap stf4go Listener to net.Listener
func WrapListener(listener Listener) net.Listener {
	return nil
}

func (wrap *wrapListener) Accept() (net.Conn, error) {
	conn, err := wrap.listener.Accept()

	if err != nil {
		return nil, err
	}

	return WrapConn(conn)
}

func (wrap *wrapListener) Close() error {
	return wrap.listener.Close()
}

func (wrap *wrapListener) Addr() net.Addr {

	addr, err := mnet.ToNetAddr(wrap.listener.Addr())

	if err != nil {
		return &net.TCPAddr{
			IP:   net.ParseIP("127.0.0.1"),
			Port: 0,
		}
	}

	return addr
}

type chainListener struct {
	laddr            multiaddr.Multiaddr
	config           *Options
	nativeTransport  NativeTransport
	tunnelTransports []TunnelTransport
	nativeListener   Listener
	tunnelAddrs      []multiaddr.Multiaddr
}

// Listen .
func Listen(laddr multiaddr.Multiaddr, options ...Option) (Listener, error) {

	configWriter := newOptions()

	for _, option := range options {
		if err := option(configWriter); err != nil {
			return nil, err
		}
	}

	if err := configWriter.Load(); err != nil {
		return nil, err
	}

	addrs, nativeTransport, tunnelTransports, err := lookupTransports(laddr)

	if err != nil {
		return nil, err
	}

	listener, err := nativeTransport.Listen(addrs[0], configWriter)

	if err != nil {
		return nil, errors.Wrap(err, "call native transport %s Listen error", nativeTransport)
	}

	return &chainListener{
		laddr:            laddr,
		config:           configWriter,
		nativeTransport:  nativeTransport,
		tunnelTransports: tunnelTransports,
		nativeListener:   listener,
		tunnelAddrs:      addrs[1:],
	}, nil
}

func (listener *chainListener) Close() error {
	return nil
}

func (listener *chainListener) Accept() (Conn, error) {
	conn, err := listener.nativeListener.Accept()

	if err != nil {
		return nil, errors.Wrap(err, "call native transport %s listener#Accept error", listener.nativeTransport)
	}

	for i, tunnel := range listener.tunnelTransports {
		conn, err = tunnel.Server(conn, listener.tunnelAddrs[i], listener.config)

		if err != nil {
			return nil, errors.Wrap(err, "call tunnel transport %s Server error", tunnel)
		}
	}

	return conn, nil
}

func (listener *chainListener) Addr() multiaddr.Multiaddr {
	return listener.laddr
}
