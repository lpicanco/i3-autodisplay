package i3

import (
	"encoding/json"
	"fmt"
)

// CommandResult always contains Success, and command-specific fields where
// appropriate.
type CommandResult struct {
	// Success indicates whether the command was run without any errors.
	Success bool `json:"success"`

	// Error is a human-readable error message, non-empty for unsuccessful
	// commands.
	Error string `json:"error"`
}

// IsUnsuccessful is a convenience function which can be used to check if an
// error is a CommandUnsuccessfulError.
func IsUnsuccessful(err error) bool {
	_, ok := err.(*CommandUnsuccessfulError)
	return ok
}

// CommandUnsuccessfulError is returned by RunCommand for unsuccessful
// commands. This type is exported so that you can ignore this error if you
// expect your command(s) to fail.
type CommandUnsuccessfulError struct {
	command string
	cr      CommandResult
}

// Error implements error.
func (e *CommandUnsuccessfulError) Error() string {
	return fmt.Sprintf("command %q unsuccessful: %v", e.command, e.cr.Error)
}

// RunCommand makes i3 run the specified command.
//
// Error is non-nil if any CommandResult.Success is not true. See IsUnsuccessful
// if you send commands which are expected to fail.
//
// RunCommand is supported in i3 ≥ v4.0 (2011-07-31).
func RunCommand(command string) ([]CommandResult, error) {
	reply, err := roundTrip(messageTypeRunCommand, []byte(command))
	if err != nil {
		return []CommandResult{}, err
	}

	var crs []CommandResult
	err = json.Unmarshal(reply.Payload, &crs)
	if err == nil {
		for _, cr := range crs {
			if !cr.Success {
				return crs, &CommandUnsuccessfulError{
					command: command,
					cr:      cr,
				}
			}
		}
	}
	return crs, err
}
