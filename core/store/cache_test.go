package store

import "testing"

type cache struct {
	a map[any]any
}

func TestM(t *testing.T) {
	m := map[any]any{}
	k := []byte("000")
	m[[]byte("000")] = 1
	v := m[k]
	t.Log(v)
}
