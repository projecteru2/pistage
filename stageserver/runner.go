package stageserver

import (
	"context"
	"io"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/projecteru2/pistage/common"
	"github.com/projecteru2/pistage/executors"
	"github.com/projecteru2/pistage/store"
)

// PistageTask contains a pistage and an output tracing stream.
// Tracing stream is used to trace this process.
type PistageRunner struct {
	sync.Mutex

	// Pistage holds the pistage to execute.
	p *common.Pistage

	status common.RunStatus

	store store.Store
	// Output is the tracing stream for logs.
	// It's an io.WriteCloser, closing this output indicates that
	// all logs have been written into this stream, the pistage
	// has finished.
	// Do remember to close the Output, or find some other methods to
	// control the halt of the process.
	o io.WriteCloser

	jobRuns map[string]*common.JobRun
	run     *common.Run
}

func NewRunner(pt *common.PistageTask, store store.Store) *PistageRunner {
	return &PistageRunner{
		p:     pt.Pistage,
		store: store,
		o:     pt.Output,
	}
}

func (r *PistageRunner) runWithStream() error {
	p := r.p
	logger := logrus.WithField("pistage", p.Name)

	if err := r.store.CreatePistage(context.TODO(), p); err != nil {
		logger.WithError(err).Error("[Stager runWithStream] fail to create Pistage")
		return err
	}

	r.run = &common.Run{
		Pistage: p.Name,
		Start:   time.Now(),
	}
	if err := r.store.CreateRun(context.TODO(), r.run); err != nil {
		logger.WithError(err).Error("[Stager runWithStream] fail to create Run")
		return err
	}

	defer func() {
		r.run.End = time.Now()
		if r.status == common.RunStatusRunning {
			r.status = common.RunStatusFinished
		}
		if err := r.store.UpdateRun(context.TODO(), r.run); err != nil {
			logger.WithError(err).Errorf("[Stager runWithStream] error update Run")
		}
	}()

	once := sync.Once{}
	jobs, finished, finish := p.JobStream()
	defer once.Do(finish)

	wg := sync.WaitGroup{}
	defer wg.Wait()

	for jobName := range jobs {
		job, err := p.GetJob(jobName)
		if err != nil {
			logger.WithError(err).Error("[Stager runWithStream] fail to get Job")
			return err
		}

		wg.Add(1)
		go func(job *common.Job) {
			defer wg.Done()
			if err = r.runOneJob(job); err != nil {
				once.Do(finish)
				r.status = common.RunStatusFailed
				logger.WithError(err).Errorf("[Stager runWithStream] error occurred, skip following jobs")
				return
			}
			finished <- job.Name
		}(job)
	}
	return nil
}

func (r *PistageRunner) runOneJob(job *common.Job) error {
	p := r.p
	logger := logrus.WithFields(logrus.Fields{"pistage": p.Name, "executor": p.Executor, "job": job.Name})

	jobRun := &common.JobRun{
		Pistage: p.Name,
		Job:     job.Name,
		Status:  common.RunStatusPending,
	}
	if err := r.store.CreateJobRun(context.TODO(), r.run, jobRun); err != nil {
		logger.WithError(err).Error("[Stager runOneJob] fail to create JobRun")
		return err
	}
	r.jobRuns[job.Name] = jobRun

	defer func() {
		if err := r.store.FinishJobRun(context.TODO(), r.run, jobRun); err != nil {
			logger.WithError(err).Errorf("[Stager runOneJob] error updating JobRun")
		}

		if err := jobRun.LogTracer.Close(); err != nil {
			logger.WithError(err).Errorf("[Stager runOneJob] error closing logtracer")
		}
	}()

	// start JobRun
	jobRun.Start = time.Now()
	jobRun.Status = common.RunStatusRunning
	jobRun.LogTracer = common.NewLogTracer(r.run.ID, r.o)
	if err := r.store.UpdateJobRun(context.TODO(), r.run, jobRun); err != nil {
		logger.WithError(err).Errorf("[Stager runOneJob] error update JobRun")
		return err
	}

	executorProvider := executors.GetExecutorProvider(p.Executor)
	if executorProvider == nil {
		logger.Errorf("[Stager runOneJob] fail to get a provider")
		return errors.WithMessage(executors.ErrorExecuteProviderNotFound, p.Name)
	}

	executor, err := executorProvider.GetJobExecutor(job, p, jobRun.LogTracer)
	if err != nil {
		logger.WithError(err).Errorf("[Stager runOneJob] fail to get a job executor")
		return err
	}

	defer func() {
		if err := executor.Cleanup(context.TODO()); err != nil {
			logger.WithError(err).Errorf("[Stager runOneJob] error when CLEANUP")
			return
		}
	}()

	if err := executor.Prepare(context.TODO()); err != nil {
		jobRun.Status = common.RunStatusFailed
		logger.WithError(err).Errorf("[Stager runOneJob] error when PREPARE")
		return err
	}

	if err := executor.Execute(context.TODO()); err != nil {
		jobRun.Status = common.RunStatusFailed
		logger.WithError(err).Errorf("[Stager runOneJob] error when EXECUTE")
		return err
	}

	return nil
}
