package i3

import "encoding/json"

// BarConfigColors describes a serialized bar colors configuration block.
//
// See https://i3wm.org/docs/ipc.html#_bar_config_reply for more details.
type BarConfigColors struct {
	Background string `json:"background"`
	Statusline string `json:"statusline"`
	Separator  string `json:"separator"`

	FocusedBackground string `json:"focused_background"`
	FocusedStatusline string `json:"focused_statusline"`
	FocusedSeparator  string `json:"focused_separator"`

	FocusedWorkspaceText       string `json:"focused_workspace_text"`
	FocusedWorkspaceBackground string `json:"focused_workspace_bg"`
	FocusedWorkspaceBorder     string `json:"focused_workspace_border"`

	ActiveWorkspaceText       string `json:"active_workspace_text"`
	ActiveWorkspaceBackground string `json:"active_workspace_bg"`
	ActiveWorkspaceBorder     string `json:"active_workspace_border"`

	InactiveWorkspaceText       string `json:"inactive_workspace_text"`
	InactiveWorkspaceBackground string `json:"inactive_workspace_bg"`
	InactiveWorkspaceBorder     string `json:"inactive_workspace_border"`

	UrgentWorkspaceText       string `json:"urgent_workspace_text"`
	UrgentWorkspaceBackground string `json:"urgent_workspace_bg"`
	UrgentWorkspaceBorder     string `json:"urgent_workspace_border"`

	BindingModeText       string `json:"binding_mode_text"`
	BindingModeBackground string `json:"binding_mode_bg"`
	BindingModeBorder     string `json:"binding_mode_border"`
}

// BarConfig describes a serialized bar configuration block.
//
// See https://i3wm.org/docs/ipc.html#_bar_config_reply for more details.
type BarConfig struct {
	ID                   string          `json:"id"`
	Mode                 string          `json:"mode"`
	Position             string          `json:"position"`
	StatusCommand        string          `json:"status_command"`
	Font                 string          `json:"font"`
	WorkspaceButtons     bool            `json:"workspace_buttons"`
	BindingModeIndicator bool            `json:"binding_mode_indicator"`
	Verbose              bool            `json:"verbose"`
	Colors               BarConfigColors `json:"colors"`
}

// GetBarIDs returns an array of configured bar IDs.
//
// GetBarIDs is supported in i3 ≥ v4.1 (2011-11-11).
func GetBarIDs() ([]string, error) {
	reply, err := roundTrip(messageTypeGetBarConfig, nil)
	if err != nil {
		return nil, err
	}

	var ids []string
	err = json.Unmarshal(reply.Payload, &ids)
	return ids, err
}

// GetBarConfig returns the configuration for the bar with the specified barID.
//
// Obtain the barID from GetBarIDs.
//
// GetBarConfig is supported in i3 ≥ v4.1 (2011-11-11).
func GetBarConfig(barID string) (BarConfig, error) {
	reply, err := roundTrip(messageTypeGetBarConfig, []byte(barID))
	if err != nil {
		return BarConfig{}, err
	}

	cfg := BarConfig{
		Colors: BarConfigColors{
			Background: "#000000",
			Statusline: "#ffffff",
			Separator:  "#666666",

			FocusedBackground: "#000000",
			FocusedStatusline: "#ffffff",
			FocusedSeparator:  "#666666",

			FocusedWorkspaceText:       "#4c7899",
			FocusedWorkspaceBackground: "#285577",
			FocusedWorkspaceBorder:     "#ffffff",

			ActiveWorkspaceText:       "#333333",
			ActiveWorkspaceBackground: "#5f676a",
			ActiveWorkspaceBorder:     "#ffffff",

			InactiveWorkspaceText:       "#333333",
			InactiveWorkspaceBackground: "#222222",
			InactiveWorkspaceBorder:     "#888888",

			UrgentWorkspaceText:       "#2f343a",
			UrgentWorkspaceBackground: "#900000",
			UrgentWorkspaceBorder:     "#ffffff",

			BindingModeText:       "#2f343a",
			BindingModeBackground: "#900000",
			BindingModeBorder:     "#ffffff",
		},
	}
	err = json.Unmarshal(reply.Payload, &cfg)
	return cfg, err
}
