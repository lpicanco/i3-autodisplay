package i3

import "encoding/json"

// TickResult attests the tick command was successful.
type TickResult struct {
	Success bool `json:"success"`
}

// SendTick sends a tick event with the provided payload.
//
// SendTick is supported in i3 ≥ v4.15 (2018-03-10).
func SendTick(command string) (TickResult, error) {
	reply, err := roundTrip(messageTypeSendTick, []byte(command))
	if err != nil {
		return TickResult{}, err
	}

	var tr TickResult
	err = json.Unmarshal(reply.Payload, &tr)
	return tr, err
}
