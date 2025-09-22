package task

type Task struct {
	// Prompt string to confirm execution
	Prompt string `json:"prompt" yaml:"prompt"`
	// Task environment variables
	Env []string `json:"env" yaml:"env"`
	// Dir to execute the task in
	Dir string `json:"dir" yaml:"dir"`
	// Description of the task
	Description string `json:"description" yaml:"description"`
	// Precondition to cancel the execution of a task and its dependencies
	Precondition []string `json:"precondition" yaml:"precondition"`
	// Dependencies to run
	Deps []string `json:"deps" yaml:"deps"`
	// Commands to execute
	Cmds []string `json:"cmds" yaml:"cmds"`
	// Don't show in the completion list
	Hidden bool `json:"hidden" yaml:"hidden"`
}
