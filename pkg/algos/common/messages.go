package common

// DirectiveMessage represents an action that a given replica should perform
type DirectiveMessage struct {
	Action string `json:"action"`
}

var (
	RestartDirective = DirectiveMessage{Action: "RESTART"}
	IsReadyDirective = DirectiveMessage{Action: "ISREADY"}
)
