package stageserver

import (
	"context"
	"io"
	"runtime"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/projecteru2/pistage/common"
	"github.com/projecteru2/pistage/executors"
	"github.com/projecteru2/pistage/store"
)

type StageServer struct {
	config *common.Config
	stages chan *common.PistageTask
	stop   chan struct{}
	store  store.Store
	wg     sync.WaitGroup
}

func NewStageServer(config *common.Config, store store.Store) *StageServer {
	return &StageServer{
		config: config,
		stages: make(chan *common.PistageTask),
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

func (s *StageServer) Add(pt *common.PistageTask) {
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
			// 	logrus.WithField("pistage", pt.Pistage.Name).WithError(err).Errorf("[Stager runner] error when running a pistage")
			// }
			if err := s.runWithStream(pt); err != nil {
				logrus.WithField("pistage", pt.Pistage.Name).WithError(err).Errorf("[Stager runner] error when running a pistage")
			}

			// We need to close the Output here, indicating the pistage is finished,
			// all logs are written into this Output.
			if err := pt.Output.Close(); err != nil {
				logrus.WithField("pistage", pt.Pistage.Name).WithError(err).Errorf("[Stager runner] error when closing the output writer")
			}
			runtime.GC()
		}
	}
}

func (s *StageServer) runWithGraph(pt *common.PistageTask) error {
	pistage := pt.Pistage
	logger := logrus.WithField("pistage", pistage.Name)

	if err := s.store.CreatePistage(context.TODO(), pistage); err != nil {
		logger.WithError(err).Error("[Stager runWithGraph] fail to create Pistage")
		return err
	}

	jobGraph, err := pistage.JobDependencies()
	if err != nil {
		logger.WithError(err).Errorf("[Stager runWithGraph] error getting job graph")
		return err
	}

	run := &common.Run{
		Pistage: pistage.Name,
		Start:   time.Now(),
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
				err = s.runOneJob(pistage, job, run, pt.Output)
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

func (s *StageServer) runWithStream(pt *common.PistageTask) error {
	pistage := pt.Pistage
	logger := logrus.WithField("pistage", pistage.Name)

	if err := s.store.CreatePistage(context.TODO(), pistage); err != nil {
		logger.WithError(err).Error("[Stager runWithStream] fail to create Pistage")
		return err
	}

	run := &common.Run{
		Pistage: pistage.Name,
		Start:   time.Now(),
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

	jobs, finished, finish := pistage.JobStream()
	defer once.Do(finish)

	wg := sync.WaitGroup{}
	defer wg.Wait()

	for jobName := range jobs {
		job, err := pistage.GetJob(jobName)
		if err != nil {
			logger.WithError(err).Error("[Stager runWithStream] fail to get Job")
			return err
		}

		wg.Add(1)
		go func(job *common.Job) {
			defer wg.Done()
			if err = s.runOneJob(pistage, job, run, pt.Output); err != nil {
				logger.WithError(err).Errorf("[Stager runWithStream] error occurred, skip following jobs")
				once.Do(finish)
				return
			}
			finished <- job.Name
		}(job)
	}
	return nil
}

func (s *StageServer) runOneJob(pistage *common.Pistage, job *common.Job, run *common.Run, logCollector io.Writer) error {
	logger := logrus.WithFields(logrus.Fields{"pistage": pistage.Name, "executor": pistage.Executor, "job": job.Name})

	jobRun := &common.JobRun{
		Pistage: pistage.Name,
		Job:     job.Name,
		Status:  common.JobRunStatusPending,
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

	executorProvider := executors.GetExecutorProvider(pistage.Executor)
	if executorProvider == nil {
		logger.Errorf("[Stager runOneJob] fail to get a provider")
		return errors.WithMessage(executors.ErrorExecuteProviderNotFound, pistage.Name)
	}

	executor, err := executorProvider.GetJobExecutor(job, pistage, jobRun.LogTracer)
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
