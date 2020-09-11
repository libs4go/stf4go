package tls

import (
	"crypto/tls"
	"net"

	_ "github.com/libs4go/bcf4go/key/encoding" //
	_ "github.com/libs4go/bcf4go/key/provider" //
	"github.com/libs4go/errors"
	"github.com/libs4go/slf4go"
	"github.com/libs4go/stf4go"
	"github.com/multiformats/go-multiaddr"
)

const protocolTLSID = 483

var protoTLS = multiaddr.Protocol{
	Name:  "tls",
	Code:  protocolTLSID,
	VCode: multiaddr.CodeToVarint(protocolTLSID),
}

var tlsMultiAddr multiaddr.Multiaddr

func init() {

	if err := multiaddr.AddProtocol(protoTLS); err != nil {
		panic(err)
	}

	var err error
	tlsMultiAddr, err = multiaddr.NewMultiaddr("/tls")
	if err != nil {
		panic(err)
	}
}

type tlsTransport struct {
	slf4go.Logger
}

func newTLSTransport() *tlsTransport {
	return &tlsTransport{
		Logger: slf4go.Get("stf4go-transport-tls"),
	}
}

func (transport *tlsTransport) String() string {
	return "stf4go-transport-tls"
}

func (transport *tlsTransport) Protocols() []multiaddr.Protocol {
	return []multiaddr.Protocol{
		protoTLS,
	}
}

func (transport *tlsTransport) Client(conn stf4go.Conn, raddr multiaddr.Multiaddr, options *stf4go.Options) (stf4go.Conn, error) {

	wrapConn, err := stf4go.WrapConn(conn)

	if err != nil {
		return nil, err
	}

	key, err := getKey(options)

	if err != nil {
		return nil, err
	}

	tlsConfig, remoteKey, err := newTLSConfig(key)

	if err != nil {
		return nil, err
	}

	session := tls.Client(wrapConn, tlsConfig)

	if err := session.Handshake(); err != nil {
		return nil, errors.Wrap(err, "tls handshake error")
	}

	return newTLSConn(session, conn, key.PubKey(), remoteKey)
}

func (transport *tlsTransport) Server(conn stf4go.Conn, laddr multiaddr.Multiaddr, options *stf4go.Options) (stf4go.Conn, error) {

	wrapConn, err := stf4go.WrapConn(conn)

	if err != nil {
		return nil, err
	}

	key, err := getKey(options)

	if err != nil {
		return nil, err
	}

	tlsConfig, remoteKey, err := newTLSConfig(key)

	if err != nil {
		return nil, err
	}

	session := tls.Server(wrapConn, tlsConfig)

	if err := session.Handshake(); err != nil {
		return nil, errors.Wrap(err, "tls handshake error")
	}

	return newTLSConn(session, conn, key.PubKey(), remoteKey)
}

type tlsConn struct {
	net.Conn
	laddr      multiaddr.Multiaddr
	raddr      multiaddr.Multiaddr
	remoteKey  chan []byte
	underlying stf4go.Conn
	localKey   []byte
}

func newTLSConn(conn net.Conn, underlying stf4go.Conn, localKey []byte, remoteKey chan []byte) (*tlsConn, error) {

	return &tlsConn{
		Conn:       conn,
		laddr:      underlying.LocalAddr().Encapsulate(tlsMultiAddr),
		raddr:      underlying.RemoteAddr().Encapsulate(tlsMultiAddr),
		remoteKey:  remoteKey,
		underlying: underlying,
		localKey:   localKey,
	}, nil
}

func (conn *tlsConn) LocalAddr() multiaddr.Multiaddr {
	return conn.laddr
}

func (conn *tlsConn) RemoteAddr() multiaddr.Multiaddr {
	return conn.raddr
}

func (conn *tlsConn) Underlying() stf4go.Conn {
	return conn.underlying
}

func (conn *tlsConn) Close() error {
	close(conn.remoteKey)
	return conn.Conn.Close()
}

func (conn *tlsConn) RemoteKey() <-chan []byte {
	return conn.remoteKey
}

func (conn *tlsConn) LocalKey() []byte {
	return conn.localKey
}

func init() {
	stf4go.RegisterTransport(newTLSTransport())
}

// Conn .
type Conn interface {
	stf4go.Conn
	RemoteKey() <-chan []byte
	LocalKey() []byte
}
