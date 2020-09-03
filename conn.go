package stf4go

import (
	"context"

	"github.com/libs4go/errors"
	"github.com/multiformats/go-multiaddr"
)

type chainConn struct {
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
