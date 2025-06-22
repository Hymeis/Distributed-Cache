package singleflight

import "testing"

func TestDo(t *testing.T) {
	var g Group
	v, err := g.Do("key", func() (interface{}, error) {
		return "test", nil
	})

	if v != "test" || err != nil {
		t.Errorf("v = %v, eeror = %v", v, err)
	}
}
