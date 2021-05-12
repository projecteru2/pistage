package eru

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	coreclient "github.com/projecteru2/core/client"
	corecluster "github.com/projecteru2/core/cluster"
	corepb "github.com/projecteru2/core/rpc/gen"
	coretypes "github.com/projecteru2/core/types"
	"github.com/projecteru2/phistage/common"
	"github.com/projecteru2/phistage/helpers/command"
)

const (
	// stupid eru-core doesn't export this
	exitMessagePrefix = "[exitcode]"
	workingDir        = "/phistage"
)

type EruJobExecutor struct {
	eru      corepb.CoreRPCClient
	job      *common.Job
	phistage *common.Phistage

	workloadID     string
	jobEnvironment map[string]string
}

// NewEruJobExecutor creates an ERU executor for this job.
// Since job needs to know its context, phistage is assigned too.
func NewEruJobExecutor(job *common.Job, phistage *common.Phistage) (*EruJobExecutor, error) {
	c, err := coreclient.NewClient(context.TODO(), "10.22.12.87:5001", coretypes.AuthConfig{
		Username: "",
		Password: "",
	})
	if err != nil {
		return nil, err
	}

	return &EruJobExecutor{
		eru:            c.GetRPCClient(),
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

func (e *EruJobExecutor) executeStep(ctx context.Context, step *common.Step) error {
	if step.Uses != "" && false {
		// get the uses context
		// use that registered step's run and replace with current args
	}

	environment := command.MergeEnvironments(e.jobEnvironment, step.Environment)

	for _, run := range step.Run {
		// use args, envs, and reserved vars to build the cmd
		// currently reserved vars is empty
		cmd, err := command.RenderCommand(run, step.With, environment, nil)
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
			Envs:       command.ToEnvironmentList(environment),
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
			} else {
				fmt.Print(data)
			}
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
