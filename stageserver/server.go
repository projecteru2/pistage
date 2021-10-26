package stageserver

import (
	"runtime"
	"sync"

	"github.com/sirupsen/logrus"

	"github.com/projecteru2/pistage/common"
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
			r := NewRunner(pt, s.store)
			// if err := s.runWithGraph(pt); err != nil {
			// 	logrus.WithField("pistage", pt.Pistage.Name).WithError(err).Errorf("[Stager runner] error when running a pistage")
			// }

			switch pt.JobType {
			case common.Apply:
				if err := r.runWithStream(); err != nil {
					logrus.WithField("pistage", pt.Pistage.Name).WithError(err).Errorf("[Stager runner] error when running a pistage")
				}
			case common.Rollback:
				if err := r.rollbackWithStream(); err != nil {
					logrus.WithField("pistage", pt.Pistage.Name).WithError(err).Errorf("[Stager runner] error when rollback a pistage")
				}
			default:

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
