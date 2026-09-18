package system

import (
	"encoding/json"
	"testing"
)

func TestStateSnapshotContainsPlayerCount(t *testing.T) {
	data, err := EncodeStateSnapshot(7, 2)
	if err != nil {
		t.Fatal(err)
	}

	var snapshot StateSnapshot
	if err = json.Unmarshal(data, &snapshot); err != nil {
		t.Fatal(err)
	}
	if snapshot.Tick != 7 || snapshot.PlayerCount != 2 {
		t.Fatalf("unexpected snapshot: %+v", snapshot)
	}
}
