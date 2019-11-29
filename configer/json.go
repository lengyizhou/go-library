package configer

import (
	"encoding/json"
)

type jsonConfig struct {
}

type jsonConfiger struct {
	data map[string]interface{}
}

func (js *jsonConfig) Configer(filename string) Configer {
	file := configFile(filename)
	defer file.Close()

	jc := &jsonConfiger{
		data: make(map[string]interface{}),
	}

	if err := json.NewDecoder(file).Decode(&(jc.data)); err != nil {
		panic(err)
	}
	return jc
}

func (jc *jsonConfiger) Set(key string, value interface{}) {
	jc.data[key] = value
}

func (jc *jsonConfiger) Get(key string) interface{} {
	return get(jc.data, key)
}

func (jc *jsonConfiger) Int(key string) (int, error) {
	return Int(jc.data, key)
}

func (jc *jsonConfiger) String(key string) (string, error) {
	return String(jc.data, key)
}

func (jc *jsonConfiger) DefaultInt(key string, defaultValue int) int {
	v, err := Int(jc.data, key)
	if err != nil {
		return defaultValue
	}

	return v
}
func (jc *jsonConfiger) DefaultString(key string, defaultValue string) string {
	v, err := String(jc.data, key)
	if err != nil {
		return defaultValue
	}

	return v
}

func init() {
	regAdapter("json", &jsonConfig{})
}
