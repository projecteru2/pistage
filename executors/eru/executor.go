package eru

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/projecteru2/phistage/common"
	"github.com/projecteru2/phistage/helpers/command"
	"github.com/projecteru2/phistage/store"

	"github.com/pkg/errors"
	corecluster "github.com/projecteru2/core/cluster"
	corepb "github.com/projecteru2/core/rpc/gen"
)

const (
	// stupid eru-core doesn't export this
	// BTW I really didn't think this can be a string with a space as suffix...
	exitMessagePrefix = "[exitcode] "
	workingDir        = "/phistage"
)

var ErrExecutionError = errors.New("Execution error")

type EruJobExecutor struct {
	eru      corepb.CoreRPCClient
	job      *common.Job
	phistage *common.Phistage
	store    store.Store

	workloadID     string
	jobEnvironment map[string]string
}

// NewEruJobExecutor creates an ERU executor for this job.
// Since job needs to know its context, phistage is assigned too.
func NewEruJobExecutor(job *common.Job, phistage *common.Phistage, eru corepb.CoreRPCClient, store store.Store) (*EruJobExecutor, error) {
	return &EruJobExecutor{
		eru:            eru,
		store:          store,
		job:            job,
		phistage:       phistage,
		jobEnvironment: phistage.Environment,
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
			if err == io.EOF {
				break
			}
			if err != nil {
				fmt.Println(err)
			}
		}
	}()
	return nil
}

func (e *EruJobExecutor) buildEruLambdaOptions() *corepb.RunAndWaitOptions {
	return &corepb.RunAndWaitOptions{
		DeployOptions: &corepb.DeployOptions{
			Name: e.job.Name,
			Entrypoint: &corepb.EntrypointOptions{
				Name:       e.job.Name,
				Commands:   command.EmptyWorkloadCommand(e.job.Timeout),
				Privileged: true,
				Dir:        workingDir,
			},
			Podname:        "ci",
			Image:          e.job.Image,
			Count:          1,
			Env:            command.ToEnvironmentList(e.jobEnvironment),
			Networks:       map[string]string{"host": ""},
			DeployStrategy: corepb.DeployOptions_AUTO,
			ResourceOpts:   &corepb.ResourceOptions{},
			User:           "root",
		},
		Async: false,
	}
}

func (e *EruJobExecutor) prepareFileContext(ctx context.Context) error {
	dependentJobs := e.phistage.GetJobs(e.job.DependsOn)
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
	fmt.Printf("========= %s =========\n", e.job.Name)
	defer fmt.Printf("========= %s =========\n", e.job.Name)
	for _, step := range e.job.Steps {
		if err := e.executeStep(ctx, step); err != nil {
			return err
		}
	}
	return nil
}

// replaceStepWithUses replaces the step with the actual step identified by uses.
// Name will not be replaced, the commands to execute aka Run and OnError will be replaced,
// also Environment and With will be merged, uses' environment and with has a lower priority,
// will be overridden by step's environment and with,
// If uses is not given, directly return the original step.
func (e *EruJobExecutor) replaceStepWithUses(ctx context.Context, step *common.Step) (*common.Step, error) {
	if step.Uses == "" {
		return step, nil
	}

	uses, err := e.store.GetRegisteredStep(ctx, step.Uses)
	if err != nil {
		return nil, err
	}
	s := &common.Step{
		Name:        step.Name,
		Run:         uses.Run,
		OnError:     uses.OnError,
		Environment: command.MergeVariables(uses.Environment, step.Environment),
		With:        command.MergeVariables(uses.With, step.With),
	}
	return s, nil
}

// executeStep executes a step.
// It first replace the step with uses if uses is given,
// then prepare the arguments and environments to the command.
// Then execute the command, retrieve the output, the execution will stop if any error occurs.
// It then retries to execute the OnError commands, also with the arguments and environments.
func (e *EruJobExecutor) executeStep(ctx context.Context, step *common.Step) error {
	var err error

	step, err = e.replaceStepWithUses(ctx, step)
	if err != nil {
		return err
	}

	environment := command.MergeVariables(e.jobEnvironment, step.Environment)

	defer func() {
		if !errors.Is(err, ErrExecutionError) {
			return
		}
		for _, onError := range step.OnError {
			e.executeCommand(ctx, onError, step.With, environment, nil)
		}
	}()

	for _, run := range step.Run {
		err = e.executeCommand(ctx, run, step.With, environment, nil)
		if err != nil {
			return err
		}

	}
	return nil
}

// executeCommand executes cmd with given arguments, environments and variables.
// use args, envs, and reserved vars to build the cmd, currently reserved vars is empty.
// This method should be sync.
func (e *EruJobExecutor) executeCommand(ctx context.Context, cmd string, args, env, vars map[string]string) error {
	cmd, err := command.RenderCommand(cmd, args, env, vars)
	if err != nil {
		return err
	}

	exec, err := e.eru.ExecuteWorkload(ctx)
	if err != nil {
		return err
	}

	if err := exec.Send(&corepb.ExecuteWorkloadOptions{
		WorkloadId: e.workloadID,
		Commands:   []string{"/bin/sh", "-c", cmd},
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
			fmt.Println("[step finished]")
			exitcode, err := strconv.Atoi(strings.TrimPrefix(data, exitMessagePrefix))
			if err != nil {
				return err
			}
			if exitcode != 0 {
				return errors.WithMessagef(ErrExecutionError, "exitcode: %d", exitcode)
			}
		} else {
			// log trace
			fmt.Print(data)
		}
	}
	return nil
}

// beforeCleanup collects files if any
func (e *EruJobExecutor) beforeCleanup(ctx context.Context) error {
	if len(e.job.Files) == 0 {
		return nil
	}

	var files []string
	for _, file := range e.job.Files {
		files = append(files, filepath.Join(workingDir, file))
	}

	fc := NewEruFileCollector(e.eru)
	if err := fc.Collect(ctx, e.workloadID, files); err != nil {
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
