package stf4go

import (
	"github.com/libs4go/scf4go"
	"github.com/libs4go/scf4go/reader/memory"
)

// ConfigWriter .
type ConfigWriter struct {
	config       scf4go.Config
	readerWriter memory.ReaderWriter
}

func newConfigWriter() *ConfigWriter {
	return &ConfigWriter{
		config:       scf4go.New(),
		readerWriter: memory.New(),
	}
}

// Set set config value
func (cw *ConfigWriter) Set(value interface{}, path ...string) {
	cw.readerWriter.Write(value, path...)
}

// Load load config
func (cw *ConfigWriter) Load() (scf4go.Config, error) {
	err := cw.config.Load(cw.readerWriter)

	return cw.config, err
}

// Option stf4go function option arg
type Option func(*ConfigWriter) error

// Config create config Option
func Config(config scf4go.Config) Option {
	return func(cw *ConfigWriter) error {
		cw.config = config
		return nil
	}
}
