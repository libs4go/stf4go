package stf4go

import (
	"sync"

	"github.com/libs4go/errors"
	"github.com/multiformats/go-multiaddr"
)

type transportRegister struct {
	sync.RWMutex
	transports map[string]Transport
}

func newTransportRegister() *transportRegister {
	return &transportRegister{
		transports: make(map[string]Transport),
	}
}

func (register *transportRegister) add(transport Transport) error {
	for _, protocol := range transport.Protocols() {
		if _, ok := register.transports[protocol.Name]; ok {
			return errors.Wrap(ErrTransport, "transport %s protocol %s already register", transport.Name(), protocol.Name)
		}

		register.transports[protocol.Name] = transport

		if multiaddr.ProtocolWithName(protocol.Name).Code == 0 {
			if err := multiaddr.AddProtocol(protocol); err != nil {
				return errors.Wrap(err, "add protocol %s error", protocol.Name)
			}
		}
	}

	return nil
}

func (register *transportRegister) get(name string) (Transport, bool) {

	register.RLock()
	defer register.RUnlock()

	transport, ok := register.transports[name]

	return transport, ok
}

var globalRegister = newTransportRegister()

// RegisterTransport transport module init function call this function register transport
func RegisterTransport(transport Transport) {
	if err := globalRegister.add(transport); err != nil {
		panic(err)
	}
}
