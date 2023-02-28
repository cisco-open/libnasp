package nasp

import (
	"testing"
)

func TestSelectedKey(t *testing.T) {
	sk := NewSelectedKey(OP_ACCEPT, 12345)

	if sk.Operation() != OP_ACCEPT {
		t.Fatalf("sk.Operation() -> %d != operation -> %d", sk.Operation(), OP_ACCEPT)
	}

	if sk.SelectedKeyId() != 12345 {
		t.Fatalf("sk.SelectedKeyId() -> %d != selectedKeyId -> %d", sk.SelectedKeyId(), 12345)
	}
}
