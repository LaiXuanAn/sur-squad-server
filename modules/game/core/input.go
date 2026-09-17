package core

import "encoding/json"

type MovementInput struct {
	X        float64 `json:"x"`
	Y        float64 `json:"y"`
	Sequence uint64  `json:"sequence"`
}

func DecodeMovementInput(data []byte) (MovementInput, error) {
	var input MovementInput
	err := json.Unmarshal(data, &input)
	return input, err
}
