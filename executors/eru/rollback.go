package eru

import (
	"context"
)

// Rollback will rollback all steps which are defined in rollback_steps
func (e *EruJobExecutor) Rollback(ctx context.Context) error {
	for _, step := range e.job.RollBackSteps {
		var err error
		switch step.Uses {
		case "":
			err = e.executeStep(ctx, step)
		default:
			// step, err = e.replaceStepWithUses(ctx, step)
			// if err != nil {
			// 	return err
			// }
			err = e.executeKhoriumStep(ctx, step)
		}
		if err != nil {
			return err
		}
	}
	return nil
}


func (e *EruJobExecutor) RollbackOneJob(ctx context.Context, jobName string) error {
	for _, step := range e.job.RollBackSteps {
		if step.Name != jobName {
			continue
		}
		var err error
		switch step.Uses {
		case "":
			err = e.executeStep(ctx, step)
		default:
			// step, err = e.replaceStepWithUses(ctx, step)
			// if err != nil {
			// 	return err
			// }
			err = e.executeKhoriumStep(ctx, step)
		}
		if err != nil {
			return err
		}
	}
	return nil
}
