package tls

import (
	"bytes"
	"encoding/json"

	"github.com/libs4go/bcf4go/key"
	"github.com/libs4go/errors"
	"github.com/libs4go/scf4go"
	"github.com/libs4go/stf4go"
)

func getKey(config scf4go.Config) (key.Key, error) {

	var data interface{}

	if err := config.Get("tls", "key", "store").Scan(&data); err != nil {
		return nil, err
	}

	buff, err := json.Marshal(data)

	if err != nil {
		return nil, errors.Wrap(err, "marshal key error")
	}

	provider := config.Get("tls", "key", "provider").String("")

	if provider == "" {
		return nil, errors.Wrap(stf4go.ErrPassword, "tls key need key provider")
	}

	password := config.Get("tls", "key", "password").String("")

	buff, err = key.Decode("web3.standard", key.Property{
		"password": password,
	}, bytes.NewBuffer(buff))

	if err != nil {
		return nil, err
	}

	return key.FromPriKey(provider, buff)
}

// KeyProvider tls key provider
func KeyProvider(name string) stf4go.Option {
	return func(cw *stf4go.ConfigWriter) error {
		cw.Set(name, "tls", "key", "provider")
		return nil
	}
}

// KeyPassword tls key protection password
func KeyPassword(password string) stf4go.Option {
	return func(cw *stf4go.ConfigWriter) error {
		cw.Set(password, "tls", "key", "password")
		return nil
	}
}

// KeyWeb3 tls key web3 encoding
func KeyWeb3(buff []byte) stf4go.Option {
	return func(cw *stf4go.ConfigWriter) error {
		var data interface{}
		err := json.Unmarshal(buff, &data)

		if err != nil {
			return errors.Wrap(err, "unmarshal web3 keystore error")
		}

		cw.Set(data, "tls", "key", "store")
		return nil
	}
}
