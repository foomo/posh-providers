package task

type Task struct {
	// Prompt string to confirm execution
	Prompt string `json:"prompt" yaml:"prompt"`
	// Dependencies to run
	Deps []string `json:"deps" yaml:"deps"`
	// Commands to execute
	Cmds []string `json:"cmds" yaml:"cmds"`
	// Don't show in the completion list
	Hidden bool `json:"hidden" yaml:"hidden"`
}
