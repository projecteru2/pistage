package eru

import (
	"context"
)

// Rollback is a function can execute rollback_steps commands which are defined in yaml file
func (e *EruJobExecutor) Rollback(ctx context.Context, jobName string) error {
	for _, step := range e.job.RollBackSteps {
		if step.Name != jobName {
			continue
		}
		var err error
		switch step.Uses {
		case "":
			err = e.executeStep(ctx, step)
		default:
			err = e.executeKhoriumStep(ctx, step)
		}
		if err != nil {
			return err
		}
	}
	return nil
}
