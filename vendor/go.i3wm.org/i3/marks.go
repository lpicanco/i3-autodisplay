package i3

import "encoding/json"

// GetMarks returns the names of all currently set marks.
//
// GetMarks is supported in i3 ≥ v4.1 (2011-11-11).
func GetMarks() ([]string, error) {
	reply, err := roundTrip(messageTypeGetMarks, nil)
	if err != nil {
		return nil, err
	}

	var marks []string
	err = json.Unmarshal(reply.Payload, &marks)
	return marks, err
}
