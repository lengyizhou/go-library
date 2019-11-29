package configer

import (
	yaml "gopkg.in/yaml.v2"
)

type yamlConfig struct {
}

type yamlConfiger struct {
	data map[string]interface{}
}

func (ys *yamlConfig) Configer(filename string) Configer {
	file := configFile(filename)
	defer file.Close()

	yc := &yamlConfiger{
		data: make(map[string]interface{}),
	}

	if err := yaml.NewDecoder(file).Decode(&(yc.data)); err != nil {
		panic(err)
	}
	return yc
}

func (yc *yamlConfiger) Set(key string, value interface{}) {
	yc.data[key] = value
}

func (yc *yamlConfiger) Get(key string) interface{} {
	return get(yc.data, key)
}

func (yc *yamlConfiger) Int(key string) (int, error) {
	return Int(yc.data, key)
}

func (yc *yamlConfiger) String(key string) (string, error) {
	return String(yc.data, key)
}

func (yc *yamlConfiger) DefaultInt(key string, defaultValue int) int {
	v, err := Int(yc.data, key)
	if err != nil {
		return defaultValue
	}

	return v
}
func (yc *yamlConfiger) DefaultString(key string, defaultValue string) string {
	v, err := String(yc.data, key)
	if err != nil {
		return defaultValue
	}

	return v
}

func init() {
	regAdapter("yaml", &yamlConfig{})
}
