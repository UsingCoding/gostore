package maybe

import (
	"encoding/json"
)

const (
	null = "null"
)

func (m *Maybe[T]) UnmarshalJSON(bytes []byte) error {
	if string(bytes) == null {
		// reset state when null passed
		m.valid = false
		var t T
		m.v = t
		return nil
	}

	var v T
	err := json.Unmarshal(bytes, &v)
	if err != nil {
		return err
	}

	m.v = v
	m.valid = true

	return nil
}

func (m Maybe[T]) MarshalJSON() ([]byte, error) {
	if !m.valid {
		return []byte("null"), nil
	}
	return json.Marshal(m.v)
}
