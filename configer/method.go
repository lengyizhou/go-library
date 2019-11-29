package configer

import (
	"fmt"
	"strconv"
	"strings"
)

func get(data map[string]interface{}, key string) interface{} {
	if key == "" {
		panic("key is empty")
	}

	keys := strings.Split(key, ".")
	if len(keys) == 0 {
		panic("key is empty")
	}
	length := len(keys)
	if length == 1 {
		return data[keys[0]]
	}

	var value interface{}
	for i := 0; i < length; i++ {
		if i == 0 {
			value = data[keys[i]]
			if value == nil {
				return nil
			}
		} else {
			if v, ok := value.(map[interface{}]interface{}); ok {
				value = v[keys[i]]
			} else {
				if v, ok := value.(map[string]interface{}); ok {
					value = v[keys[i]]
				} else {
					if i == length-1 {
						return value
					}
					return nil
				}
				if i == length-1 {
					return value
				}
				return nil
			}
		}
	}
	return value
}

func Int(data map[string]interface{}, key string) (int, error) {
	value := get(data, key)
	if value == nil {
		return 0, fmt.Errorf("invalid key")
	}
	switch value := value.(type) {
	case int:
		return value, nil
	case float64:
		if t := int(value); fmt.Sprint(value) == fmt.Sprint(t) {
			return t, nil
		} else {
			return 0, fmt.Errorf("value(%v) can't be converted to int", value)
		}
	case string:
		if v, err := strconv.ParseInt(value, 10, 0); err == nil {
			return int(v), nil
		} else {
			return 0, err
		}
	}
	return 0, fmt.Errorf("unknow value")
}

func String(data map[string]interface{}, key string) (string, error) {
	value := get(data, key)
	if value == nil {
		return "", fmt.Errorf("invalid key")
	}

	switch value := value.(type) {
	case bool, float64, int:
		return fmt.Sprint(value), nil
	case string:
		return value, nil
	}
	return "", fmt.Errorf("unknow value")
}
