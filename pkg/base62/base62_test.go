package base62

import "testing"

func TestIntToBase62(t *testing.T) {
	tests := []struct {
		originalNum uint64
		want        string
	}{
		{0, "0"},
		{1, "1"},
		{61, "Z"},
		{62, "10"},
	}

	for _, test := range tests {
		if got := IntToBase62(test.originalNum); got != test.want {
			t.Errorf("IntToBase62(%v): %v", test.originalNum, got)
		}
	}
}
