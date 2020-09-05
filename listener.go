package stf4go

import (
	"net"

	"github.com/libs4go/errors"
	"github.com/libs4go/scf4go"
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
	config           scf4go.Config
	nativeTransport  NativeTransport
	tunnelTransports []TunnelTransport
	nativeListener   Listener
	tunnelAddrs      []multiaddr.Multiaddr
}

// Listen .
func Listen(laddr multiaddr.Multiaddr, options ...Option) (Listener, error) {

	configWriter := newConfigWriter()

	for _, option := range options {
		if err := option(configWriter); err != nil {
			return nil, err
		}
	}

	if err := configWriter.config.Load(configWriter.readerWriter); err != nil {
		return nil, err
	}

	addrs, nativeTransport, tunnelTransports, err := lookupTransports(laddr)

	if err != nil {
		return nil, err
	}

	listener, err := nativeTransport.Listen(addrs[0], configWriter.config)

	if err != nil {
		return nil, errors.Wrap(err, "call native transport %s Listen error", nativeTransport.Name())
	}

	return &chainListener{
		laddr:            laddr,
		config:           configWriter.config,
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
		return nil, errors.Wrap(err, "call native transport %s listener#Accept error", listener.nativeTransport.Name())
	}

	for i, tunnel := range listener.tunnelTransports {
		conn, err = tunnel.Server(conn, listener.tunnelAddrs[i], listener.config)

		if err != nil {
			return nil, errors.Wrap(err, "call tunnel transport %s Server error", tunnel.Name())
		}
	}

	return conn, nil
}

func (listener *chainListener) Addr() multiaddr.Multiaddr {
	return listener.laddr
}
