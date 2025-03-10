package embeddable

import (
	"gopkg.in/yaml.v3"
)

func MustYAML[T any](d []byte) T {
	t, err := YAML[T](d)
	if err != nil {
		panic(err)
	}

	return t
}

func YAML[T any](data []byte) (T, error) {
	var nop T

	if err := yaml.Unmarshal(data, &nop); err != nil {
		return nop, err
	}

	return nop, nil
}
