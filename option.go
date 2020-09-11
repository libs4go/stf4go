package stf4go

import (
	"strings"

	"github.com/libs4go/scf4go"
	"github.com/libs4go/scf4go/reader/memory"
)

// Options .
type Options struct {
	Config       scf4go.Config
	readerWriter memory.ReaderWriter
	objs         map[string]interface{}
}

func newOptions() *Options {
	return &Options{
		Config:       scf4go.New(),
		readerWriter: memory.New(),
		objs:         make(map[string]interface{}),
	}
}

// SetConfig set config value
func (cw *Options) SetConfig(value interface{}, path ...string) {
	cw.readerWriter.Write(value, path...)
}

// SetObject .
func (cw *Options) SetObject(value interface{}, path ...string) {
	cw.objs[strings.Join(path, ".")] = value
}

// Load load config
func (cw *Options) Load() error {
	err := cw.Config.Load(cw.readerWriter)

	return err
}

// GetObj .
func (cw *Options) GetObj(path ...string) (interface{}, bool) {
	v, ok := cw.objs[strings.Join(path, ".")]

	return v, ok
}

// Option stf4go function option arg
type Option func(*Options) error

// Config create config Option
func Config(config scf4go.Config) Option {
	return func(cw *Options) error {
		cw.Config = config
		return nil
	}
}
