package configer

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

type textConfig struct {
}

type textConfiger struct {
	data map[string]interface{}
}

func parse(content string) ([]string, error) {
	kv := strings.Split(content, "=")
	if len(kv) != 2 {
		return nil, fmt.Errorf("invalid format: config should be {key}={value}")
	}
	return kv, nil
}

func (ts *textConfig) Configer(filename string) Configer {
	file := configFile(filename)
	defer file.Close()

	tc := &textConfiger{
		data: make(map[string]interface{}),
	}

	reader := bufio.NewReader(file)
	for {
		line, _, err := reader.ReadLine()
		if err != nil {
			if err == io.EOF {
				return tc
			}
			panic(err)
		}

		content := strings.TrimSpace(string(line))
		if content == "" {
			continue
		}

		kv, err := parse(content)
		if err != nil {
			panic(err)
		}
		tc.data[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
	}

	return tc
}

func (tc *textConfiger) Set(key string, value interface{}) {
	tc.data[key] = value
}

func (tc *textConfiger) Get(key string) interface{} {
	return get(tc.data, key)
}

func (tc *textConfiger) Int(key string) (int, error) {
	return Int(tc.data, key)
}

func (tc *textConfiger) String(key string) (string, error) {
	return String(tc.data, key)
}

func (tc *textConfiger) DefaultInt(key string, defaultValue int) int {
	v, err := Int(tc.data, key)
	if err != nil {
		return defaultValue
	}

	return v
}
func (tc *textConfiger) DefaultString(key string, defaultValue string) string {
	v, err := String(tc.data, key)
	if err != nil {
		return defaultValue
	}

	return v
}

func init() {
	regAdapter("text", &textConfig{})
}
