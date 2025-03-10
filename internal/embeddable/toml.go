package embeddable

import (
	"github.com/pelletier/go-toml"
)

func MustTOML[T any](d []byte) T {
	t, err := TOML[T](d)
	if err != nil {
		panic(err)
	}

	return t
}

func TOML[T any](data []byte) (T, error) {
	var nop T

	if err := toml.Unmarshal(data, &nop); err != nil {
		return nop, err
	}

	return nop, nil
}
