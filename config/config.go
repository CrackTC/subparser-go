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
		switch value.(type) {
		case Config:
			if dest[key] == nil {
				dest[key] = make(Config)
			}
			dest[key].(Config).Merge(value.(Config))
		case []any:
			if dest[key] == nil {
				dest[key] = make([]any, 0)
			}
			var tmp []any = make([]any, 0, len(dest[key].([]any))+len(value.([]any)))
			for _, v := range value.([]any) {
				if _, ok := v.(Config); ok {
					var cp Config = make(Config)
					cp.Merge(v.(Config))
					tmp = append(tmp, cp)
				} else {
					tmp = append(tmp, v)
				}
			}
			dest[key] = append(tmp, dest[key].([]any)...)
		default:
			dest[key] = value
		}
	}
}

func (c Config) String() (string, error) {
	bytes, err := yaml.Marshal(&c)

	if err != nil {
		return "", err
	}

	return string(bytes), nil
}
