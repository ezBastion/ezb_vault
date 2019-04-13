package tools_test

import (
	"testing"

	"github.com/ezbastion/ezb_vault/tools"
)

func TestStrIsInt(t *testing.T) {

	tables := []struct {
		x string
		y bool
	}{
		{"", false},
		{"-1", false},
		{"65536", true},
		{"string", false},
	}
	for _, table := range tables {
		ret := tools.StrIsInt(table.x)
		if ret != table.y {
			t.Errorf("test of %s was incorrect, got: %t, want: %t.", table.x, ret, table.y)
		}
	}
}
