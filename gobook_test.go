package gobook

import "testing"

func TestExample(t *testing.T) {
	err := makeBook("./examples/xorm_book", "./examples/xorm")
	if err != nil {
		t.Error(err)
	}
}
