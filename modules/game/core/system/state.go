package system

import "google.golang.org/protobuf/proto"

func EncodeStateSnapshot(tick int64, playerCount int) ([]byte, error) {
	return proto.Marshal(&StateSnapshot{Tick: tick, PlayerCount: int32(playerCount)})
}
