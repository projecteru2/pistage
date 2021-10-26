package eru

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	corecluster "github.com/projecteru2/core/cluster"
	corepb "github.com/projecteru2/core/rpc/gen"
	"github.com/sirupsen/logrus"

	"github.com/projecteru2/pistage/common"
	"github.com/projecteru2/pistage/helpers/command"
	"github.com/projecteru2/pistage/helpers/variable"
	"github.com/projecteru2/pistage/store"
)

const (
	// stupid eru-core doesn't export this
	// BTW I really didn't think this can be a string with a space as suffix...
	exitMessagePrefix = "[exitcode] "

	// working dir for KhoriumStep.
	// will be added after DefaultWorkingDir
	khoriumStepWorkingDir = "/_khoriumstep/"
)

type EruJobExecutor struct {
	eru    corepb.CoreRPCClient
	store  store.Store
	config *common.Config

	job     *common.Job
	pistage *common.Pistage

	output         io.Writer
	workloadID     string
	jobEnvironment map[string]string
	workingDir     string
}

// NewEruJobExecutor creates an ERU executor for this job.
// Since job needs to know its context, pistage is assigned too.
func NewEruJobExecutor(job *common.Job, pistage *common.Pistage, output io.Writer, eru corepb.CoreRPCClient, store store.Store, config *common.Config) (*EruJobExecutor, error) {
	return &EruJobExecutor{
		eru:            eru,
		store:          store,
		config:         config,
		job:            job,
		pistage:        pistage,
		output:         output,
		jobEnvironment: pistage.Environment,
		workingDir:     config.Eru.DefaultWorkingDir,
	}, nil
}

// Prepare does all the preparations before actually running a job
func (e *EruJobExecutor) Prepare(ctx context.Context) error {
	preparations := []func(context.Context) error{
		e.prepareJobRuntime,
		e.prepareFileContext,
	}
	for _, f := range preparations {
		if err := f(ctx); err != nil {
			return err
		}
	}
	return nil
}

// prepareJobRuntime currently creates an empty lambda workload.
// The empty lambda workload is actually a sleep process which lasts timeout seconds.
func (e *EruJobExecutor) prepareJobRuntime(ctx context.Context) error {
	lambda, err := e.eru.RunAndWait(ctx)
	if err != nil {
		return err
	}

	if err := lambda.Send(e.buildEruLambdaOptions()); err != nil {
		return err
	}

	message, err := lambda.Recv()
	if err != nil {
		return err
	}

	e.workloadID = message.WorkloadId

	// eat all the remaing messages
	go func() {
		for {
			_, err := lambda.Recv()
			if err != nil {
				// since we don't really care the ouput,
				// just break when hit and error.
				break
			}
		}
	}()
	return nil
}

// defaultEnvironmentVariables sets some useful information into environment variables.
// This will be set to the whole running context within the workload.
func (e *EruJobExecutor) defaultEnvironmentVariables() map[string]string {
	return map[string]string{
		"PISTAGE_WORKING_DIR": e.workingDir,
		"PISTAGE_JOB_NAME":    e.job.Name,
		"PISTAGE_DEPENDS_ON":  strings.Join(e.job.DependsOn, ","),
	}
}

// buildEruLambdaOptions builds the options for ERU lambda workload.
// Currently only container supports, it's just because I never tried virtual machines
// or systemd engine...
func (e *EruJobExecutor) buildEruLambdaOptions() *corepb.RunAndWaitOptions {
	jobImage := e.job.Image
	if jobImage == "" {
		jobImage = e.config.Eru.DefaultJobImage
	}

	return &corepb.RunAndWaitOptions{
		DeployOptions: &corepb.DeployOptions{
			Name: e.job.Name,
			Entrypoint: &corepb.EntrypointOptions{
				Name:       e.job.Name,
				Commands:   command.EmptyWorkloadCommand(e.job.Timeout),
				Privileged: e.config.Eru.DefaultPrivileged,
				Dir:        e.workingDir,
			},
			Podname:        e.config.Eru.DefaultPodname,
			Image:          jobImage,
			Count:          1,
			Env:            command.ToEnvironmentList(command.MergeVariables(command.PreparePistageEnvs(e.jobEnvironment), e.defaultEnvironmentVariables())),
			Networks:       map[string]string{e.config.Eru.DefaultNetwork: ""},
			DeployStrategy: corepb.DeployOptions_AUTO,
			ResourceOpts:   &corepb.ResourceOptions{},
			User:           e.config.Eru.DefaultUser,
		},
		Async: false,
	}
}

func (e *EruJobExecutor) prepareFileContext(ctx context.Context) error {
	dependentJobs := e.pistage.GetJobs(e.job.DependsOn)
	for _, job := range dependentJobs {
		fc := job.GetFileCollector()
		if fc == nil {
			continue
		}
		if err := fc.CopyTo(ctx, e.workloadID, nil); err != nil {
			return err
		}
	}
	return nil
}

// Execute will execute all steps within this job one by one
func (e *EruJobExecutor) Execute(ctx context.Context) error {
	return e.executeSteps(ctx, e.job.Steps)
}

// executeDifferentJob dispatch executor
func (e *EruJobExecutor) executeSteps(ctx context.Context, steps []*common.Step) error {
	for _, step := range steps {
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

// executeStep executes a step.
// It first replace the step with uses if uses is given,
// then prepare the arguments and environments to the command.
// Then execute the command, retrieve the output, the execution will stop if any error occurs.
// It then retries to execute the OnError commands, also with the arguments and environments.
func (e *EruJobExecutor) executeStep(ctx context.Context, step *common.Step) error {
	var (
		err  error
		vars map[string]string
	)

	environment := command.MergeVariables(e.jobEnvironment, step.Environment)

	defer func() {
		if !errors.Is(err, common.ErrExecutionError) {
			return
		}
		if err := e.executeCommands(ctx, step.OnError, step.With, environment, vars); err != nil {
			logrus.WithField("step", step.Name).WithError(err).Errorf("[EruJobExecutor] error when executing on_error")
		}
	}()

	err = e.executeCommands(ctx, step.Run, step.With, environment, vars)
	return err
}

// executeKhoriumStep executes a KhoriumStep defined by step.Uses.
func (e *EruJobExecutor) executeKhoriumStep(ctx context.Context, step *common.Step) error {
	ks, err := e.store.GetRegisteredKhoriumStep(ctx, step.Uses)
	if err != nil {
		return err
	}

	arguments, err := variable.RenderArguments(step.With, step.Environment, map[string]string{})
	if err != nil {
		return err
	}

	ksEnv, err := ks.BuildEnvironmentVariables(arguments)
	if err != nil {
		return err
	}
	envs := command.MergeVariables(step.Environment, ksEnv)

	fc := NewEruFileCollector(e.eru, khoriumStepWorkingDir, e.job)
	fc.SetFiles(ks.Files)
	if err := fc.CopyTo(ctx, e.workloadID, nil); err != nil {
		return err
	}

	exec, err := e.eru.ExecuteWorkload(ctx)
	if err != nil {
		return err
	}

	if err := exec.Send(&corepb.ExecuteWorkloadOptions{
		WorkloadId: e.workloadID,
		Commands:   []string{"/bin/sh", "-c", ks.Run.Main},
		Envs:       command.ToEnvironmentList(envs),
		Workdir:    khoriumStepWorkingDir,
	}); err != nil {
		return err
	}

	for {
		message, err := exec.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		data := string(message.Data)
		if strings.HasPrefix(data, exitMessagePrefix) {
			exitcode, err := strconv.Atoi(strings.TrimPrefix(data, exitMessagePrefix))
			if err != nil {
				return err
			}
			if exitcode != 0 {
				return errors.WithMessagef(common.ErrExecutionError, "exitcode: %d", exitcode)
			}
		} else {
			if _, err := io.WriteString(e.output, data); err != nil {
				return err
			}
		}
	}
	return exec.CloseSend()
}

// executeCommands executes cmd with given arguments, environments and variables.
// use args, envs, and reserved vars to build the cmd.
// This method should be sync.
func (e *EruJobExecutor) executeCommands(ctx context.Context, cmds []string, args, env, vars map[string]string) error {
	if len(cmds) == 0 {
		return nil
	}

	var commands []string
	for _, cmd := range cmds {
		c, err := command.RenderCommand(cmd, args, env, vars)
		if err != nil {
			return err
		}
		commands = append(commands, c)
	}

	shell, err := command.RenderShell(commands)
	if err != nil {
		return err
	}

	exec, err := e.eru.ExecuteWorkload(ctx)
	if err != nil {
		return err
	}

	if err := exec.Send(&corepb.ExecuteWorkloadOptions{
		WorkloadId: e.workloadID,
		Commands:   []string{"/bin/sh", "-c", shell},
		Envs:       command.ToEnvironmentList(env),
	}); err != nil {
		return err
	}

	for {
		message, err := exec.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		data := string(message.Data)
		if strings.HasPrefix(data, exitMessagePrefix) {
			exitcode, err := strconv.Atoi(strings.TrimPrefix(data, exitMessagePrefix))
			if err != nil {
				return err
			}
			if exitcode != 0 {
				return errors.WithMessagef(common.ErrExecutionError, "exitcode: %d", exitcode)
			}
		} else {
			if _, err := io.WriteString(e.output, data); err != nil {
				return err
			}
		}
	}
	return exec.CloseSend()
}

// beforeCleanup collects files if any
func (e *EruJobExecutor) beforeCleanup(ctx context.Context) error {
	fc := NewEruFileCollector(e.eru, e.workingDir, e.job)
	if err := fc.Collect(ctx, e.workloadID, e.job.Files); err != nil {
		return err
	}

	e.job.SetFileCollector(fc)
	return nil
}

// cleanup currently just stops the workload.
// On ERU side, the stopped lambda workload will be removed automatically,
// so just leave the cleanup work to ERU.
func (e *EruJobExecutor) cleanup(ctx context.Context) error {
	opts := &corepb.ControlWorkloadOptions{
		Ids:   []string{e.workloadID},
		Type:  corecluster.WorkloadStop,
		Force: true,
	}
	control, err := e.eru.ControlWorkload(ctx, opts)
	if err != nil {
		return err
	}

	for {
		message, err := control.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if message.Error != "" {
			return fmt.Errorf(message.Error)
		}
	}
	return nil
}

// Cleanup does all the cleanup work
func (e *EruJobExecutor) Cleanup(ctx context.Context) error {
	cleanups := []func(context.Context) error{
		e.beforeCleanup,
		e.cleanup,
	}
	for _, f := range cleanups {
		if err := f(ctx); err != nil {
			return err
		}
	}
	return nil
}

// Rollback is a function can execute rollback_steps commands which are defined in yaml file
func (e *EruJobExecutor) Rollback(ctx context.Context) error {
	return e.executeSteps(ctx, e.job.RollbackSteps)
}
