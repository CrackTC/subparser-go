package config

import (
	"io"

	"gopkg.in/yaml.v3"
)

type Config map[string]any

func Load(r io.Reader) (Config, error) {
	c := make(Config)

	if err := yaml.NewDecoder(r).Decode(&c); err != nil {
		return nil, err
	}

	return c, nil
}

func LoadString(s string) (Config, error) {
	c := make(Config)

	if err := yaml.Unmarshal([]byte(s), &c); err != nil {
		return nil, err
	}

	return c, nil
}

func (dest Config) Merge(src Config) {
	for key, value := range src {
		if _, ok := dest[key]; ok == false {
			dest[key] = value
		} else {
			switch value.(type) {
			case Config:
				dest[key].(Config).Merge(value.(Config))
			case []interface{}:
				dest[key] = append(value.([]interface{}), dest[key].([]interface{})...)
			default:
				dest[key] = value
			}
		}
	}
}

func (c Config) String() (string, error) {
	bytes, err := yaml.Marshal(c)

	if err != nil {
		return "", err
	}

	return string(bytes), nil
}
