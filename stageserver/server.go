package stageserver

import (
	"context"
	"io"
	"runtime"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/projecteru2/phistage/common"
	"github.com/projecteru2/phistage/executors"
	"github.com/projecteru2/phistage/store"
)

type StageServer struct {
	config *common.Config
	stages chan *common.PhistageTask
	stop   chan struct{}
	store  store.Store
	wg     sync.WaitGroup
}

func NewStager(config *common.Config, store store.Store) *StageServer {
	return &StageServer{
		config: config,
		stages: make(chan *common.PhistageTask),
		stop:   make(chan struct{}),
		store:  store,
		wg:     sync.WaitGroup{},
	}
}

func (s *StageServer) Start() {
	for id := 0; id < s.config.StageServerWorkers; id++ {
		s.wg.Add(1)
		go func(id int) {
			defer s.wg.Done()
			s.runner(id)
		}(id)
	}
}

func (s *StageServer) Stop() {
	logrus.Info("[Stager] exiting...")
	close(s.stop)
	s.wg.Wait()
	logrus.Info("[Stager] gracefully stopped")
}

func (s *StageServer) Add(pt *common.PhistageTask) {
	s.stages <- pt
}

func (s *StageServer) runner(id int) {
	logrus.WithField("runner id", id).Info("[Stager] runner started")
	for {
		select {
		case <-s.stop:
			logrus.WithField("runner id", id).Info("[Stager] runner stopped")
			return
		case pt := <-s.stages:
			// if err := s.runWithGraph(pt); err != nil {
			// 	logrus.WithField("phistage", pt.Phistage.Name).WithError(err).Errorf("[Stager runner] error when running a phistage")
			// }
			if err := s.runWithStream(pt); err != nil {
				logrus.WithField("phistage", pt.Phistage.Name).WithError(err).Errorf("[Stager runner] error when running a phistage")
			}

			// We need to close the Output here, indicating the phistage is finished,
			// all logs are written into this Output.
			if err := pt.Output.Close(); err != nil {
				logrus.WithField("phistage", pt.Phistage.Name).WithError(err).Errorf("[Stager runner] error when closing the output writer")
			}
			runtime.GC()
		}
	}
}

func (s *StageServer) runWithGraph(pt *common.PhistageTask) error {
	phistage := pt.Phistage
	logger := logrus.WithField("phistage", phistage.Name)

	if err := s.store.CreatePhistage(context.TODO(), phistage); err != nil {
		logger.WithError(err).Error("[Stager runWithGraph] fail to create Phistage")
		return err
	}

	jobGraph, err := phistage.JobDependencies()
	if err != nil {
		logger.WithError(err).Errorf("[Stager runWithGraph] error getting job graph")
		return err
	}

	run := &common.Run{
		Phistage: phistage.Name,
		Start:    time.Now(),
	}
	if err := s.store.CreateRun(context.TODO(), run); err != nil {
		logger.WithError(err).Error("[Stager runWithGraph] fail to create Run")
		return err
	}
	defer func() {
		run.End = time.Now()
		if err := s.store.UpdateRun(context.TODO(), run); err != nil {
			logger.WithError(err).Errorf("[Stager runWithGraph] error update Run")
		}
	}()

	for _, jobs := range jobGraph {
		wg := sync.WaitGroup{}
		for _, job := range jobs {
			wg.Add(1)
			go func(job *common.Job) {
				defer wg.Done()
				err = s.runOneJob(phistage, job, run, pt.Output)
			}(job)
		}
		wg.Wait()

		if err != nil {
			logger.WithError(err).Errorf("[Stager runWithGraph] error occurred, skip following jobs")
			return err
		}
	}
	return nil
}

func (s *StageServer) runWithStream(pt *common.PhistageTask) error {
	phistage := pt.Phistage
	logger := logrus.WithField("phistage", phistage.Name)

	if err := s.store.CreatePhistage(context.TODO(), phistage); err != nil {
		logger.WithError(err).Error("[Stager runWithStream] fail to create Phistage")
		return err
	}

	run := &common.Run{
		Phistage: phistage.Name,
		Start:    time.Now(),
	}
	if err := s.store.CreateRun(context.TODO(), run); err != nil {
		logger.WithError(err).Error("[Stager runWithStream] fail to create Run")
		return err
	}

	defer func() {
		run.End = time.Now()
		if err := s.store.UpdateRun(context.TODO(), run); err != nil {
			logger.WithError(err).Errorf("[Stager runWithStream] error update Run")
		}
	}()

	once := sync.Once{}

	jobs, finished, finish := phistage.JobStream()
	defer once.Do(finish)

	wg := sync.WaitGroup{}
	defer wg.Wait()

	for jobName := range jobs {
		job, err := phistage.GetJob(jobName)
		if err != nil {
			logger.WithError(err).Error("[Stager runWithStream] fail to get Job")
			return err
		}

		wg.Add(1)
		go func(job *common.Job) {
			defer wg.Done()
			if err = s.runOneJob(phistage, job, run, pt.Output); err != nil {
				logger.WithError(err).Errorf("[Stager runWithStream] error occurred, skip following jobs")
				once.Do(finish)
				return
			}
			finished <- job.Name
		}(job)
	}
	return nil
}

func (s *StageServer) runOneJob(phistage *common.Phistage, job *common.Job, run *common.Run, logCollector io.Writer) error {
	logger := logrus.WithFields(logrus.Fields{"phistage": phistage.Name, "executor": phistage.Executor, "job": job.Name})

	jobRun := &common.JobRun{
		Phistage: phistage.Name,
		Job:      job.Name,
		Status:   common.JobRunStatusPending,
	}
	if err := s.store.CreateJobRun(context.TODO(), run, jobRun); err != nil {
		logger.WithError(err).Error("[Stager runOneJob] fail to create JobRun")
		return err
	}

	defer func() {
		if err := s.store.FinishJobRun(context.TODO(), run, jobRun); err != nil {
			logger.WithError(err).Errorf("[Stager runOneJob] error updating JobRun")
		}

		if err := jobRun.LogTracer.Close(); err != nil {
			logger.WithError(err).Errorf("[Stager runOneJob] error closing logtracer")
		}
	}()

	// start JobRun
	jobRun.Start = time.Now()
	jobRun.Status = common.JobRunStatusRunning
	jobRun.LogTracer = common.NewLogTracer(run.ID, logCollector)
	if err := s.store.UpdateJobRun(context.TODO(), run, jobRun); err != nil {
		logger.WithError(err).Errorf("[Stager runOneJob] error update JobRun")
		return err
	}

	executorProvider := executors.GetExecutorProvider(phistage.Executor)
	if executorProvider == nil {
		logger.Errorf("[Stager runOneJob] fail to get a provider")
		return errors.WithMessage(executors.ErrorExecuteProviderNotFound, phistage.Name)
	}

	executor, err := executorProvider.GetJobExecutor(job, phistage, jobRun.LogTracer)
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
		logger.WithError(err).Errorf("[Stager runOneJob] error when PREPARE")
		return err
	}

	if err := executor.Execute(context.TODO()); err != nil {
		logger.WithError(err).Errorf("[Stager runOneJob] error when EXECUTE")
		return err
	}

	return nil
}
