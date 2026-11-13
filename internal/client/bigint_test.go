package client

import (
	"encoding/json"
	"testing"
)

func TestBigIntUnmarshal(t *testing.T) {
	for in, want := range map[string]int64{`123`: 123, `"9007199254740993"`: 9007199254740993, `0`: 0} {
		var b BigInt
		if err := json.Unmarshal([]byte(in), &b); err != nil || int64(b) != want {
			t.Errorf("%s: got %d, %v; want %d", in, b, err, want)
		}
	}
	var b BigInt
	if err := json.Unmarshal([]byte(`"abc"`), &b); err == nil {
		t.Error("expected error for non-numeric BigInt")
	}
}
