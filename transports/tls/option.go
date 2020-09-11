package tls

import (
	"github.com/libs4go/bcf4go/key"
	"github.com/libs4go/errors"
	"github.com/libs4go/stf4go"
)

func getKey(options *stf4go.Options) (key.Key, error) {
	obj, ok := options.GetObj("tls", "key")

	if !ok {
		return nil, errors.Wrap(stf4go.ErrResource, "expect key")
	}

	k, ok := obj.(key.Key)

	if !ok {
		return nil, errors.Wrap(stf4go.ErrResource, "expect key")
	}

	return k, nil
}

// WithKey .
func WithKey(k key.Key) stf4go.Option {
	return func(options *stf4go.Options) error {
		options.SetObject(k, "tls", "key")

		return nil
	}
}
