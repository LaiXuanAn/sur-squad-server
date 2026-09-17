package core

import "testing"

func TestDecodeMovementInputRejectsMalformedJSON(t *testing.T) {
	if _, err := DecodeMovementInput([]byte(`{"x":`)); err == nil {
		t.Fatal("expected malformed JSON to fail")
	}
}
