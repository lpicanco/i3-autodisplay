package i3

import "encoding/json"

// GetBindingModes returns the names of all currently configured binding modes.
//
// GetBindingModes is supported in i3 ≥ v4.13 (2016-11-08).
func GetBindingModes() ([]string, error) {
	reply, err := roundTrip(messageTypeGetBindingModes, nil)
	if err != nil {
		return nil, err
	}

	var bm []string
	err = json.Unmarshal(reply.Payload, &bm)
	return bm, err
}
