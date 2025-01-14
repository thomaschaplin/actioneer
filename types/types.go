package types

type Step struct {
	Actions string `yaml:"action"`
	Command string `yaml:"command,omitempty"`
	Repo string `yaml:"repo,omitempty"`
	Branch string `yaml:"branch,omitempty"`
}

type Job struct {
	Steps []Step `yaml:"steps"`
}

type Workflow struct {
	Jobs map[string]Job `yaml:"jobs"`
}
