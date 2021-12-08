package stageserver

import (
	"context"
	"io"
	"sort"
	"sync"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/projecteru2/pistage/common"
	"github.com/projecteru2/pistage/executors"
	"github.com/projecteru2/pistage/store"
)

// PistageRunner runs a complete workflow.
type PistageRunner struct {
	sync.Mutex
	ctx context.Context

	// Pistage holds the pistage to execute.
	p *common.Pistage

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
		p:       pt.Pistage,
		store:   store,
		o:       pt.Output,
		jobRuns: map[string]*common.JobRun{},
	}
}

func (r *PistageRunner) runWithStream(ctx context.Context) error {
	p := r.p
	logger := logrus.WithField("pistage", p.Name())

	if err := p.GenerateHash(); err != nil {
		logger.WithError(err).Error("[Stager runWithStream] gen hash failed")
		return err
	}

	version, err := r.store.CreatePistageSnapshot(p)
	if err != nil {
		logger.WithError(err).Error("[Stager runWithStream] fail to create Pistage")
		return err
	}

	runID, err := r.store.CreatePistageRun(p, version)
	if err != nil {
		logger.WithError(err).Error("[Stager runWithStream] fail to create Pistage")
		return err
	}

	r.run = &common.Run{
		ID:                 runID,
		WorkflowNamespace:  p.WorkflowNamespace,
		WorkflowIdentifier: p.WorkflowIdentifier,
		Start:              common.EpochMillis(),
		Status:             common.RunStatusRunning,
	}

	defer func() {
		r.run.End = common.EpochMillis()
		if r.run.Status == common.RunStatusRunning {
			r.run.Status = common.RunStatusFinished
		}
		if err := r.store.UpdatePistageRun(r.run); err != nil {
			logger.WithError(err).Errorf("[Stager runWithStream] error update Run")
		}
	}()

	if err := r.store.UpdatePistageRun(r.run); err != nil {
		logger.WithError(err).Error("[Stager runWithStream] fail to update run")
		return err
	}

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
			if err = r.runOneJob(ctx, job); err != nil {
				r.Lock()
				defer r.Unlock()
				r.run.Status = common.RunStatusFailed
				logger.WithError(err).Errorf("[Stager runWithStream] error occurred, skip following jobs")
				once.Do(finish)
				return
			}
			finished <- job.Name
		}(job)
	}
	return nil
}

func (r *PistageRunner) runOneJob(ctx context.Context, job *common.Job) error {
	p := r.p
	logger := logrus.WithFields(logrus.Fields{"pistage": p.Name(), "executor": p.Executor, "job": job.Name})

	jobRun := &common.JobRun{
		WorkflowNamespace:  p.WorkflowNamespace,
		WorkflowIdentifier: p.WorkflowIdentifier,
		JobName:            job.Name,
		Status:             common.RunStatusPending,
	}
	if err := r.store.CreateJobRun(r.run, jobRun); err != nil {
		logger.WithError(err).Error("[Stager runOneJob] fail to create JobRun")
		return err
	}
	r.jobRuns[job.Name] = jobRun

	defer func() {
		if jobRun.Status == common.RunStatusRunning {
			jobRun.Status = common.RunStatusFinished
		}
		jobRun.End = common.EpochMillis()
		if err := r.store.UpdateJobRun(jobRun); err != nil {
			logger.WithError(err).Errorf("[Stager runOneJob] error updating JobRun")
		}

		if err := jobRun.LogTracer.Close(); err != nil {
			logger.WithError(err).Errorf("[Stager runOneJob] error closing logtracer")
		}
	}()

	// start JobRun
	jobRun.Start = common.EpochMillis()
	jobRun.Status = common.RunStatusRunning
	jobRun.LogTracer = common.NewLogTracer(r.run.ID, r.o)
	if err := r.store.UpdateJobRun(jobRun); err != nil {
		logger.WithError(err).Errorf("[Stager runOneJob] error update JobRun")
		return err
	}

	executorProvider := executors.GetExecutorProvider(p.Executor)
	if executorProvider == nil {
		logger.Errorf("[Stager runOneJob] fail to get a provider")
		return errors.WithMessage(executors.ErrorExecuteProviderNotFound, p.Name())
	}

	executor, err := executorProvider.GetJobExecutor(job, p, jobRun.LogTracer)
	if err != nil {
		logger.WithError(err).Errorf("[Stager runOneJob] fail to get a job executor")
		return err
	}

	defer func() {
		if err := executor.Cleanup(ctx); err != nil {
			logger.WithError(err).Errorf("[Stager runOneJob] error when CLEANUP")
			return
		}
	}()

	if err := executor.Prepare(ctx); err != nil {
		jobRun.Status = common.RunStatusFailed
		logger.WithError(err).Errorf("[Stager runOneJob] error when PREPARE")
		return err
	}

	if err := executor.Execute(ctx); err != nil {
		jobRun.Status = common.RunStatusFailed
		logger.WithError(err).Errorf("[Stager runOneJob] error when EXECUTE")
		return err
	}

	return nil
}

func (r *PistageRunner) rollbackWithStream(ctx context.Context) error {
	p := r.p
	logger := logrus.WithFields(logrus.Fields{"pistage": p.Name(), "executor": p.Executor, "function": "rollback"})

	if err := p.GenerateHash(); err != nil {
		logger.WithError(err).Error("[Stager runWithStream] gen hash failed")
		return err
	}

	pistageRun, err := r.store.GetLatestPistageRunByNamespaceAndFlowIdentifier(p.WorkflowNamespace, p.WorkflowIdentifier)
	if err != nil {
		logger.WithError(err).Errorf("[Stager rollback] error when GetLatestPistageRunByNamespaceAndFlowIdentifier")
		return err
	}

	id := pistageRun.ID

	jobRuns, err := r.store.GetJobRunsByPistageRunId(id)
	if err != nil {
		logger.WithError(err).Errorf("[Stager rollback] error when GetJobRunsByPistageRunId")
		return err
	}

	finishedJobRuns := make([]*common.JobRun, 0)
	for _, jobRun := range jobRuns {
		if jobRun.Status == common.RunStatusFinished {
			jobRun.LogTracer = common.NewLogTracer(id, r.o)
			finishedJobRuns = append(finishedJobRuns, jobRun)
		}
	}

	if len(finishedJobRuns) == 0 {
		return nil
	}

	// descending sort by start time, start firstly will roll back finally
	sort.Slice(finishedJobRuns, func(i, j int) bool {
		return finishedJobRuns[i].Start > finishedJobRuns[j].Start
	})

	err = r.rollbackJobs(ctx, finishedJobRuns, id)

	if err != nil {
		logger.WithError(err).Errorf("[Stager rollback] error when rollbackJobs")
		return err
	}
	return nil
}

func (r *PistageRunner) rollbackJobs(ctx context.Context, jobRuns []*common.JobRun, pistageRunId string) error {
	p := r.p
	logger := logrus.WithFields(logrus.Fields{"pistage": p.Name(), "executor": p.Executor, "function": "rollback"})
	for _, jobRun := range jobRuns {
		if job, ok := p.Jobs[jobRun.JobName]; ok {
			err := r.rollbackOneJob(ctx, job, pistageRunId)
			if err != nil {
				logger.WithError(err).Errorf("[Stager rollback] fail to rollback")
				return err
			}
		}
	}
	return nil
}

func (r *PistageRunner) rollbackOneJob(ctx context.Context, job *common.Job, pistageRunId string) error {
	p := r.p
	logger := logrus.WithFields(logrus.Fields{"pistage": p.Name(), "executor": p.Executor, "function": "rollback"})
	executorProvider := executors.GetExecutorProvider(p.Executor)
	if executorProvider == nil {
		logger.Errorf("[Stager rollbackOneJob] fail to get a provider")
		return errors.WithMessage(executors.ErrorExecuteProviderNotFound, p.Name())
	}

	executor, err := executorProvider.GetJobExecutor(job, p, common.NewLogTracer(pistageRunId, r.o))
	if err != nil {
		logger.WithError(err).Errorf("[Stager rollback] fail to get a job executor")
		return err
	}

	if err = executor.Prepare(ctx); err != nil {
		logger.WithError(err).Errorf("[Stager rollback] error when PREPARE")
		return err
	}

	if err = executor.Rollback(ctx); err != nil {
		logger.WithError(err).Errorf("[Stager rollback] error when EXECUTE")
		return err
	}

	if err = executor.Cleanup(ctx); err != nil {
		logger.WithError(err).Errorf("[Stager rollback] error when CLEANUP")
		return err
	}
	return nil
}
