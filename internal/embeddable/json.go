package embeddable

import (
	"encoding/json"
)

func MustJSON[T any](d []byte) T {
	t, err := JSON[T](d)
	if err != nil {
		panic(err)
	}

	return t
}

func JSON[T any](data []byte) (T, error) {
	var nop T

	if err := json.Unmarshal(data, &nop); err != nil {
		return nop, err
	}

	return nop, nil
}
