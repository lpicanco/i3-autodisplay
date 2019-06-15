package i3

import (
	"encoding/json"
)

// NodeType indicates the specific kind of Node.
type NodeType string

// i3 currently implements the following node types:
const (
	Root          NodeType = "root"
	OutputNode    NodeType = "output"
	Con           NodeType = "con"
	FloatingCon   NodeType = "floating_con"
	WorkspaceNode NodeType = "workspace"
	DockareaNode  NodeType = "dockarea"
)

// Layout indicates the layout of a Node.
type Layout string

// i3 currently implements the following layouts:
const (
	SplitH         Layout = "splith"
	SplitV         Layout = "splitv"
	Stacked        Layout = "stacked"
	Tabbed         Layout = "tabbed"
	DockareaLayout Layout = "dockarea"
	OutputLayout   Layout = "output"
)

// BorderStyle indicates the border style of a node.
type BorderStyle string

// i3 currently implements the following border styles:
const (
	NormalBorder BorderStyle = "normal"
	NoBorder     BorderStyle = "none"
	PixelBorder  BorderStyle = "pixel"
)

// Rect is a rectangle, used for various dimensions in Node, for example.
type Rect struct {
	X      int64 `json:"x"`
	Y      int64 `json:"y"`
	Width  int64 `json:"width"`
	Height int64 `json:"height"`
}

// WindowProperties correspond to X11 window properties
//
// See https://build.i3wm.org/docs/ipc.html#_tree_reply
type WindowProperties struct {
	Title     string `json:"title"`
	Instance  string `json:"instance"`
	Class     string `json:"class"`
	Role      string `json:"window_role"`
	Transient NodeID `json:"transient_for"`
}

// NodeID is an i3-internal ID for the node, which can be used to identify
// containers within the IPC interface.
type NodeID int64

// Node is a node in a Tree.
//
// See https://i3wm.org/docs/ipc.html#_tree_reply for more details.
type Node struct {
	ID                 NodeID           `json:"id"`
	Name               string           `json:"name"` // window: title, container: internal name
	Type               NodeType         `json:"type"`
	Border             BorderStyle      `json:"border"`
	CurrentBorderWidth int64            `json:"current_border_width"`
	Layout             Layout           `json:"layout"`
	Percent            float64          `json:"percent"`
	Rect               Rect             `json:"rect"`        // absolute (= relative to X11 display)
	WindowRect         Rect             `json:"window_rect"` // window, relative to Rect
	DecoRect           Rect             `json:"deco_rect"`   // decoration, relative to Rect
	Geometry           Rect             `json:"geometry"`    // original window geometry, absolute
	Window             int64            `json:"window"`      // X11 window ID of the client window
	WindowProperties   WindowProperties `json:"window_properties"`
	Urgent             bool             `json:"urgent"` // urgency hint set
	Focused            bool             `json:"focused"`
	Focus              []NodeID         `json:"focus"`
	Nodes              []*Node          `json:"nodes"`
	FloatingNodes      []*Node          `json:"floating_nodes"`
}

// FindChild returns the first Node matching predicate, using pre-order
// depth-first search.
func (n *Node) FindChild(predicate func(*Node) bool) *Node {
	if predicate(n) {
		return n
	}
	for _, c := range n.Nodes {
		if con := c.FindChild(predicate); con != nil {
			return con
		}
	}
	for _, c := range n.FloatingNodes {
		if con := c.FindChild(predicate); con != nil {
			return con
		}
	}
	return nil
}

// FindFocused returns the first Node matching predicate from the sub-tree of
// directly and indirectly focused containers.
//
// As an example, consider this layout tree (simplified):
//
//       root
//         │
//       HDMI2
//        ╱ ╲
//      …  workspace 1
//           ╱ ╲
//      XTerm   Firefox
//
// In this example, if Firefox is focused, FindFocused will return the first
// container matching predicate of root, HDMI2, workspace 1, Firefox (in this
// order).
func (n *Node) FindFocused(predicate func(*Node) bool) *Node {
	if predicate(n) {
		return n
	}
	if len(n.Focus) == 0 {
		return nil
	}
	first := n.Focus[0]
	for _, c := range n.Nodes {
		if c.ID == first {
			return c.FindFocused(predicate)
		}
	}
	for _, c := range n.FloatingNodes {
		if c.ID == first {
			return c.FindFocused(predicate)
		}
	}
	return nil
}

// Tree is an i3 layout tree, starting with Root.
type Tree struct {
	// Root is the root node of the layout tree.
	Root *Node
}

// GetTree returns i3’s layout tree.
//
// GetTree is supported in i3 ≥ v4.0 (2011-07-31).
func GetTree() (Tree, error) {
	reply, err := roundTrip(messageTypeGetTree, nil)
	if err != nil {
		return Tree{}, err
	}

	var root Node
	err = json.Unmarshal(reply.Payload, &root)
	return Tree{Root: &root}, err
}
