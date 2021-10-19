package shell

import (
	"context"
	"io"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/projecteru2/pistage/common"
	"github.com/projecteru2/pistage/helpers/command"
	"github.com/projecteru2/pistage/helpers/variable"
	"github.com/projecteru2/pistage/store"
)

type ShellJobExecutor struct {
	store  store.Store
	config *common.Config

	job     *common.Job
	pistage *common.Pistage

	output         io.Writer
	workingDir     string
	jobEnvironment map[string]string
}

// NewShellJobExecutor creates an Shell executor for this job.
// Since job needs to know its context, pistage is assigned too.
func NewShellJobExecutor(job *common.Job, pistage *common.Pistage, output io.Writer, store store.Store, config *common.Config) (*ShellJobExecutor, error) {
	return &ShellJobExecutor{
		store:          store,
		config:         config,
		job:            job,
		pistage:        pistage,
		output:         output,
		jobEnvironment: pistage.Environment,
	}, nil
}

// Prepare does all the preparations before actually running a job
func (sje *ShellJobExecutor) Prepare(ctx context.Context) error {
	preparations := []func(context.Context) error{
		sje.prepareJobRuntime,
		sje.prepareFileContext,
	}
	for _, f := range preparations {
		if err := f(ctx); err != nil {
			return err
		}
	}
	return nil
}

// Prepare creates a temp working dir for this job.
func (sje *ShellJobExecutor) prepareJobRuntime(ctx context.Context) error {
	var err error
	sje.workingDir, err = ioutil.TempDir("", "pistage-*")
	return err
}

func (sje *ShellJobExecutor) prepareFileContext(ctx context.Context) error {
	dependentJobs := sje.pistage.GetJobs(sje.job.DependsOn)
	for _, job := range dependentJobs {
		fc := job.GetFileCollector()
		if fc == nil {
			continue
		}
		if err := fc.CopyTo(ctx, sje.workingDir, nil); err != nil {
			return err
		}
	}
	return nil
}

// defaultEnvironmentVariables sets some useful information into environment variables.
// This will be set to the whole running context within the workload.
func (sje *ShellJobExecutor) defaultEnvironmentVariables() map[string]string {
	return map[string]string{
		"PHISTAGE_WORKING_DIR": sje.workingDir,
		"PHISTAGE_JOB_NAME":    sje.job.Name,
	}
}

// Execute will execute all steps within this job one by one
func (sje *ShellJobExecutor) Execute(ctx context.Context) error {
	for _, step := range sje.job.Steps {
		var err error
		switch step.Uses {
		case "":
			err = sje.executeStep(ctx, step)
		default:
			// step, err = e.replaceStepWithUses(ctx, step)
			// if err != nil {
			// 	return err
			// }
			err = sje.executeKhoriumStep(ctx, step)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

// executeStep executes a step.
// It first replaces the step with uses if uses is given,
// then prepare the arguments and environments to the command.
// Then execute the command, retrieve the output, the execution will stop if any error occurs.
// It then retries to execute the OnError commands, also with the arguments and environments.
func (sje *ShellJobExecutor) executeStep(ctx context.Context, step *common.Step) error {
	var (
		err  error
		vars map[string]string
	)

	environment := command.MergeVariables(sje.jobEnvironment, step.Environment)

	defer func() {
		if !errors.Is(err, common.ErrExecutionError) {
			return
		}
		if err := sje.executeCommands(ctx, step.OnError, step.With, environment, vars); err != nil {
			logrus.WithField("step", step.Name).WithError(err).Errorf("[EruJobExecutor] error when executing on_error")
		}
	}()

	err = sje.executeCommands(ctx, step.Run, step.With, environment, vars)
	return err
}

// executeKhoriumStep executes a KhoriumStep defined by step.Uses.
func (sje *ShellJobExecutor) executeKhoriumStep(ctx context.Context, step *common.Step) error {
	ks, err := sje.store.GetRegisteredKhoriumStep(ctx, step.Uses)
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
	envs := command.MergeVariables(sje.defaultEnvironmentVariables(), step.Environment)
	envs = command.MergeVariables(envs, ksEnv)

	khoriumStepWorkingDir, err := ioutil.TempDir("", "pistage-khoriumstep-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(khoriumStepWorkingDir)

	fc := NewShellFileCollector()
	fc.SetFiles(ks.Files)
	if err := fc.CopyTo(ctx, khoriumStepWorkingDir, nil); err != nil {
		return err
	}

	cmd := exec.CommandContext(ctx, "/bin/sh", "-c", ks.Run.Main)
	cmd.Dir = khoriumStepWorkingDir
	cmd.Env = command.ToEnvironmentList(envs)
	cmd.Stdout = sje.output
	cmd.Stderr = sje.output
	if err := cmd.Run(); err != nil {
		return errors.WithMessagef(common.ErrExecutionError, "exec error: %v", err)
	}
	return nil
}

// executeCommands executes cmd with given arguments, environments and variables.
// use args, envs, and reserved vars to build the cmd.
func (sje *ShellJobExecutor) executeCommands(ctx context.Context, cmds []string, args, env, vars map[string]string) error {
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

	for _, c := range commands {
		cmd := exec.CommandContext(ctx, "/bin/sh", "-c", c)
		cmd.Dir = sje.workingDir
		cmd.Env = command.ToEnvironmentList(command.MergeVariables(sje.defaultEnvironmentVariables(), env))
		cmd.Stdout = sje.output
		cmd.Stderr = sje.output
		if err := cmd.Run(); err != nil {
			return errors.WithMessagef(common.ErrExecutionError, "exec error: %v", err)
		}
	}
	return nil
}

// beforeCleanup collects files
func (sje *ShellJobExecutor) beforeCleanup(ctx context.Context) error {
	if len(sje.job.Files) == 0 {
		return nil
	}

	fc := NewShellFileCollector()
	if err := fc.Collect(ctx, sje.workingDir, sje.job.Files); err != nil {
		return err
	}

	sje.job.SetFileCollector(fc)
	return nil
}

// cleanup removes the temp working dir.
func (sje *ShellJobExecutor) cleanup(ctx context.Context) error {
	return nil // return os.RemoveAll(sje.workingDir)
}

// Cleanup does all the cleanup work
func (sje *ShellJobExecutor) Cleanup(ctx context.Context) error {
	cleanups := []func(context.Context) error{
		sje.beforeCleanup,
		sje.cleanup,
	}
	for _, f := range cleanups {
		if err := f(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (sje *ShellJobExecutor) Rollback (ctx context.Context) error {
	return nil
}



func (sje *ShellJobExecutor) RollbackOneJob(ctx context.Context, jobName string) error {
	return nil
}