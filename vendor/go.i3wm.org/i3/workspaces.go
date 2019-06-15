package i3

import "encoding/json"

// Workspace describes an i3 workspace.
//
// See https://i3wm.org/docs/ipc.html#_workspaces_reply for more details.
type Workspace struct {
	Num     int64  `json:"num"`
	Name    string `json:"name"`
	Visible bool   `json:"visible"`
	Focused bool   `json:"focused"`
	Urgent  bool   `json:"urgent"`
	Rect    Rect   `json:"rect"`
	Output  string `json:"output"`
}

// GetWorkspaces returns i3’s current workspaces.
//
// GetWorkspaces is supported in i3 ≥ v4.0 (2011-07-31).
func GetWorkspaces() ([]Workspace, error) {
	reply, err := roundTrip(messageTypeGetWorkspaces, nil)
	if err != nil {
		return nil, err
	}

	var ws []Workspace
	err = json.Unmarshal(reply.Payload, &ws)
	return ws, err
}
