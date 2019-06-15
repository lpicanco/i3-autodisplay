package i3

import "encoding/json"

// SyncRequest represents the payload of a Sync request.
type SyncRequest struct {
	Window uint32 `json:"window"` // X11 window id
	Rnd    uint32 `json:"rnd"`    // Random value for distinguishing requests
}

// SyncResult attests the sync command was successful.
type SyncResult struct {
	Success bool `json:"success"`
}

// Sync sends a tick event with the provided payload.
//
// Sync is supported in i3 ≥ v4.16 (2018-11-04).
func Sync(req SyncRequest) (SyncResult, error) {
	b, err := json.Marshal(req)
	if err != nil {
		return SyncResult{}, err
	}
	reply, err := roundTrip(messageTypeSync, b)
	if err != nil {
		return SyncResult{}, err
	}

	var tr SyncResult
	err = json.Unmarshal(reply.Payload, &tr)
	return tr, err
}
