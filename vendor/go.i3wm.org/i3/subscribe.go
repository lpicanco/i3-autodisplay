package i3

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"time"
)

// Event is an event received from i3.
//
// Type-assert or type-switch on Event to obtain a more specific type.
type Event interface{}

// WorkspaceEvent contains details about various workspace-related changes.
//
// See https://i3wm.org/docs/ipc.html#_workspace_event for more details.
type WorkspaceEvent struct {
	Change  string `json:"change"`
	Current Node   `json:"current"`
	Old     Node   `json:"old"`
}

// OutputEvent contains details about various output-related changes.
//
// See https://i3wm.org/docs/ipc.html#_output_event for more details.
type OutputEvent struct {
	Change string `json:"change"`
}

// ModeEvent contains details about various mode-related changes.
//
// See https://i3wm.org/docs/ipc.html#_mode_event for more details.
type ModeEvent struct {
	Change      string `json:"change"`
	PangoMarkup bool   `json:"pango_markup"`
}

// WindowEvent contains details about various window-related changes.
//
// See https://i3wm.org/docs/ipc.html#_window_event for more details.
type WindowEvent struct {
	Change    string `json:"change"`
	Container Node   `json:"container"`
}

// BarconfigUpdateEvent contains details about various bar config-related changes.
//
// See https://i3wm.org/docs/ipc.html#_barconfig_update_event for more details.
type BarconfigUpdateEvent BarConfig

// BindingEvent contains details about various binding-related changes.
//
// See https://i3wm.org/docs/ipc.html#_binding_event for more details.
type BindingEvent struct {
	Change  string `json:"change"`
	Binding struct {
		Command        string   `json:"command"`
		EventStateMask []string `json:"event_state_mask"`
		InputCode      int64    `json:"input_code"`
		Symbol         string   `json:"symbol"`
		InputType      string   `json:"input_type"`
	} `json:"binding"`
}

// ShutdownEvent contains the reason for which the IPC connection is about to be
// shut down.
//
// See https://i3wm.org/docs/ipc.html#_shutdown_event for more details.
type ShutdownEvent struct {
	Change string `json:"change"`
}

// TickEvent contains the payload of the last tick command.
//
// See https://i3wm.org/docs/ipc.html#_tick_event for more details.
type TickEvent struct {
	First   bool   `json:"first"`
	Payload string `json:"payload"`
}

type eventReplyType int

const (
	eventReplyTypeWorkspace eventReplyType = iota
	eventReplyTypeOutput
	eventReplyTypeMode
	eventReplyTypeWindow
	eventReplyTypeBarconfigUpdate
	eventReplyTypeBinding
	eventReplyTypeShutdown
	eventReplyTypeTick
)

const (
	eventFlagMask = uint32(0x80000000)
	eventTypeMask = ^eventFlagMask
)

// EventReceiver is not safe for concurrent use.
type EventReceiver struct {
	types     []EventType // for re-subscribing on io.EOF
	sock      *socket
	conn      net.Conn
	ev        Event
	err       error
	reconnect bool
}

// Event returns the most recent event received from i3 by a call to Next.
func (r *EventReceiver) Event() Event {
	return r.ev
}

func (r *EventReceiver) subscribe() error {
	var err error
	if r.conn != nil {
		r.conn.Close()
	}
	r.sock, r.conn, err = getIPCSocket(r.reconnect)
	r.reconnect = true
	if err != nil {
		return err
	}
	payload, err := json.Marshal(r.types)
	if err != nil {
		return err
	}
	b, err := r.sock.roundTrip(messageTypeSubscribe, payload)
	if err != nil {
		return err
	}
	var reply struct {
		Success bool `json:"success"`
	}
	if err := json.Unmarshal(b.Payload, &reply); err != nil {
		return err
	}
	if !reply.Success {
		return fmt.Errorf("could not subscribe, check the i3 log")
	}
	r.err = nil
	return nil
}

func (r *EventReceiver) next() (Event, error) {
	reply, err := r.sock.recvMsg()
	if err != nil {
		return nil, err
	}
	if (uint32(reply.Type) & eventFlagMask) == 0 {
		return nil, fmt.Errorf("unexpectedly did not receive an event")
	}
	t := uint32(reply.Type) & eventTypeMask
	switch eventReplyType(t) {
	case eventReplyTypeWorkspace:
		var e WorkspaceEvent
		return &e, json.Unmarshal(reply.Payload, &e)

	case eventReplyTypeOutput:
		var e OutputEvent
		return &e, json.Unmarshal(reply.Payload, &e)

	case eventReplyTypeMode:
		var e ModeEvent
		return &e, json.Unmarshal(reply.Payload, &e)

	case eventReplyTypeWindow:
		var e WindowEvent
		return &e, json.Unmarshal(reply.Payload, &e)

	case eventReplyTypeBarconfigUpdate:
		var e BarconfigUpdateEvent
		return &e, json.Unmarshal(reply.Payload, &e)

	case eventReplyTypeBinding:
		var e BindingEvent
		return &e, json.Unmarshal(reply.Payload, &e)

	case eventReplyTypeShutdown:
		var e ShutdownEvent
		return &e, json.Unmarshal(reply.Payload, &e)

	case eventReplyTypeTick:
		var e TickEvent
		return &e, json.Unmarshal(reply.Payload, &e)
	}
	return nil, fmt.Errorf("BUG: event reply type %d not implemented yet", t)
}

// Next advances the EventReceiver to the next event, which will then be
// available through the Event method. It returns false when reaching an
// error. After Next returns false, the Close method will return the first
// error.
//
// Until you call Close, you must call Next in a loop for every EventReceiver
// (usually in a separate goroutine), otherwise i3 will deadlock as soon as the
// UNIX socket buffer is full of unprocessed events.
func (r *EventReceiver) Next() bool {
Outer:
	for r.err == nil {
		r.ev, r.err = r.next()
		if r.err == nil {
			return true // happy path
		}

		// reconnect
		start := time.Now()
		for time.Since(start) < reconnectTimeout && (r.sock == nil || i3Running()) {
			if err := r.subscribe(); err == nil {
				continue Outer
			} else {
				r.err = err
			}

			// Reconnect within [10, 20) ms to prevent CPU-starving i3.
			time.Sleep(time.Duration(10+rand.Int63n(10)) * time.Millisecond)
		}
	}
	return r.err == nil
}

// Close closes the connection to i3. If you don’t ever call Close, you must
// consume events via Next to prevent i3 from deadlocking.
func (r *EventReceiver) Close() error {
	if r.conn != nil {
		if r.err == nil {
			r.err = r.conn.Close()
		} else {
			// Retain the original error.
			r.conn.Close()
		}
		r.conn = nil
		r.sock = nil
	}
	return r.err
}

// EventType indicates the specific kind of event to subscribe to.
type EventType string

// i3 currently implements the following event types:
const (
	WorkspaceEventType       EventType = "workspace"        // since 4.0
	OutputEventType          EventType = "output"           // since 4.0
	ModeEventType            EventType = "mode"             // since 4.4
	WindowEventType          EventType = "window"           // since 4.5
	BarconfigUpdateEventType EventType = "barconfig_update" // since 4.6
	BindingEventType         EventType = "binding"          // since 4.9
	ShutdownEventType        EventType = "shutdown"         // since 4.14
	TickEventType            EventType = "tick"             // since 4.15
)

type majorMinor struct {
	major int64
	minor int64
}

var eventAtLeast = map[EventType]majorMinor{
	WorkspaceEventType:       {4, 0},
	OutputEventType:          {4, 0},
	ModeEventType:            {4, 4},
	WindowEventType:          {4, 5},
	BarconfigUpdateEventType: {4, 6},
	BindingEventType:         {4, 9},
	ShutdownEventType:        {4, 14},
	TickEventType:            {4, 15},
}

// Subscribe returns an EventReceiver for receiving events of the specified
// types from i3.
//
// Unless the ordering of events matters to your use-case, you are encouraged to
// call Subscribe once per event type, so that you can use type assertions
// instead of type switches.
//
// Subscribe is supported in i3 ≥ v4.0 (2011-07-31).
func Subscribe(eventTypes ...EventType) *EventReceiver {
	// Error out early in case any requested event type is not yet supported by
	// the running i3 version.
	for _, t := range eventTypes {
		if err := AtLeast(eventAtLeast[t].major, eventAtLeast[t].minor); err != nil {
			return &EventReceiver{err: err}
		}
	}
	return &EventReceiver{types: eventTypes}
}

// restart runs the restart i3 command without entering an infinite loop: as
// RUN_COMMAND with payload "restart" does not result in a reply, we subscribe
// to the shutdown event beforehand (on a dedicated connection), which we can
// receive instead of a reply.
func restart(firstAttempt bool) error {
	sock, conn, err := getIPCSocket(!firstAttempt)
	if err != nil {
		return err
	}
	defer conn.Close()
	payload, err := json.Marshal([]EventType{ShutdownEventType})
	if err != nil {
		return err
	}
	b, err := sock.roundTrip(messageTypeSubscribe, payload)
	if err != nil {
		return err
	}
	var sreply struct {
		Success bool `json:"success"`
	}
	if err := json.Unmarshal(b.Payload, &sreply); err != nil {
		return err
	}
	if !sreply.Success {
		return fmt.Errorf("could not subscribe, check the i3 log")
	}
	rreply, err := sock.roundTrip(messageTypeRunCommand, []byte("restart"))
	if err != nil {
		return err
	}
	if (uint32(rreply.Type) & eventFlagMask) == 0 {
		var crs []CommandResult
		err = json.Unmarshal(rreply.Payload, &crs)
		if err == nil {
			for _, cr := range crs {
				if !cr.Success {
					return &CommandUnsuccessfulError{
						command: "restart",
						cr:      cr,
					}
				}
			}
		}
		return nil // restart command successful
	}
	t := uint32(rreply.Type) & eventTypeMask
	if got, want := eventReplyType(t), eventReplyTypeShutdown; got != want {
		return fmt.Errorf("unexpected reply type: got %d, want %d", got, want)
	}
	return nil // shutdown event received
}

// Restart sends the restart command to i3. Sending restart via RunCommand will
// result in a deadlock: since i3 restarts before it sends the reply to the
// restart command, RunCommand will retry the command indefinitely.
//
// Restart is supported in i3 ≥ v4.14 (2017-09-04).
func Restart() error {
	if err := AtLeast(eventAtLeast[ShutdownEventType].major, eventAtLeast[ShutdownEventType].minor); err != nil {
		return err
	}

	// TODO: send a PR which makes restarts lazy (executed after parse_command
	// returns), generating a reply. Can version-switch using AtLeast here.

	var (
		firstAttempt = true
		start        = time.Now()
		lastErr      error
	)
	for time.Since(start) < reconnectTimeout && (firstAttempt || i3Running()) {
		lastErr = restart(firstAttempt)
		if lastErr == nil {
			return nil // success
		}
		firstAttempt = false
	}
	return lastErr
}
