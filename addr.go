package stf4go

import (
	"net"

	"github.com/libs4go/errors"
	"github.com/multiformats/go-multiaddr"
	mnet "github.com/multiformats/go-multiaddr/net"
)

// ToNetAddr .
func ToNetAddr(addr multiaddr.Multiaddr) (net.Addr, error) {
	addrs := multiaddr.Split(addr)

	if len(addrs) < 2 {
		return nil, errors.Wrap(ErrMultiAddr, "multiaddr stack must > 2, but %s", addr.String())
	}

	netAddr, err := mnet.ToNetAddr(addrs[0].Encapsulate(addrs[1]))

	if err != nil {
		return nil, errors.Wrap(err, "")
	}

	return netAddr, nil
}

// FromNetAddr .
func FromNetAddr(addr net.Addr) (multiaddr.Multiaddr, error) {
	maddr, err := mnet.FromNetAddr(addr)

	if err != nil {
		return nil, errors.Wrap(err, "")
	}

	return maddr, nil
}
