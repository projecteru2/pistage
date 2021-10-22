package eru

import (
	"context"
)


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
