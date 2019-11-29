package configer

import (
	"fmt"
	"os"
)

type Configer interface {
	Set(key string, value interface{})
	Get(key string) interface{}
	Int(key string) (int, error)
	String(key string) (string, error)
	DefaultInt(key string, defaultValue int) int
	DefaultString(key string, defaultValue string) string
}

type config interface {
	Configer(filename string) Configer
}

var adapters = make(map[string]config)

func adapter(name, filename string) Configer {
	adapter, ok := adapters[name]
	if !ok {
		panic(fmt.Sprintf("unknown adapter: %s", name))
	}

	return adapter.Configer(filename)
}

func New(name, filename string) Configer {
	return adapter(name, filename)
}

func regAdapter(name string, adapter config) {
	if adapter == nil {
		panic("config adapter is nil")
	}

	if _, ok := adapters[name]; ok {
		panic(fmt.Sprintf("adapter %s had registered", name))
	}

	adapters[name] = adapter
}

func configFile(filename string) *os.File {
	if filename == "" {
		filename = os.Getenv("APP_CONFIG_FILE")
	}
	if filename == "" {
		panic("config file is not exist")
	}
	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	return file
}
