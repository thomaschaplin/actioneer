package executor

import (
	"fmt"
	"thomaschaplin/github-actions-poc/actions"
	"thomaschaplin/github-actions-poc/types"
)

func ExecuteJob(jobName string, job types.Job) error {
	for _, step := range job.Steps {
		switch step.Actions {
		case "checkout":
			err := actions.Checkout(jobName, step.Repo, step.Branch, "./repo")
			if err != nil {
				return fmt.Errorf("checkout failed: %w", err)
			}
		case "run":
			err := actions.RunCommand(jobName, step.Command)
			if err != nil {
				return fmt.Errorf("command failed: %w", err)
			}
		default:
			return fmt.Errorf("unknown action: %s", step.Actions)
		}
	}
	return nil
}

func ExecuteWorkflow(workflow *types.Workflow) {
	// Iterate over the jobs in the order they are defined in the workflow
	for name, job := range workflow.Jobs {
		err := ExecuteJob(name, job)
		if err != nil {
			fmt.Printf("Job %s failed: %v\n", name, err)
			return // Stop execution if a job fails
		}
	}
}
