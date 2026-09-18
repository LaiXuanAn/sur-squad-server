package entity

import "testing"

func TestDecodeMovementInputRejectsMalformedProtobuf(t *testing.T) {
	if _, err := DecodeMovementInput([]byte{0x0a, 0x01}); err == nil {
		t.Fatal("expected malformed protobuf to fail")
	}
}
