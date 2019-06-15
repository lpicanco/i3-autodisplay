package i3

import "encoding/json"

// Output describes an i3 output.
//
// See https://i3wm.org/docs/ipc.html#_outputs_reply for more details.
type Output struct {
	Name             string `json:"name"`
	Active           bool   `json:"active"`
	Primary          bool   `json:"primary"`
	CurrentWorkspace string `json:"current_workspace"`
	Rect             Rect   `json:"rect"`
}

// GetOutputs returns i3’s current outputs.
//
// GetOutputs is supported in i3 ≥ v4.0 (2011-07-31).
func GetOutputs() ([]Output, error) {
	reply, err := roundTrip(messageTypeGetOutputs, nil)
	if err != nil {
		return nil, err
	}

	var outputs []Output
	err = json.Unmarshal(reply.Payload, &outputs)
	return outputs, err
}
