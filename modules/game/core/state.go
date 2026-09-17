package core

import "encoding/json"

type StateSnapshot struct {
	Tick        int64 `json:"tick"`
	PlayerCount int   `json:"player_count"`
}

func EncodeStateSnapshot(tick int64, playerCount int) ([]byte, error) {
	return json.Marshal(StateSnapshot{Tick: tick, PlayerCount: playerCount})
}
