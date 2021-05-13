package stager

import (
	"context"
	"fmt"
	"sync"

	"github.com/projecteru2/phistage/common"
	"github.com/projecteru2/phistage/executors"
)

type Stager struct {
	config  *common.Config
	stageCh chan *common.Phistage
	stopCh  chan struct{}
	errorCh chan error
	wg      sync.WaitGroup
}

func NewStager(config *common.Config) *Stager {
	return &Stager{
		config:  config,
		stageCh: make(chan *common.Phistage),
		stopCh:  make(chan struct{}),
		errorCh: make(chan error),
		wg:      sync.WaitGroup{},
	}
}

func (s *Stager) Start() {
	for i := 0; i < s.config.StagerWorkers; i++ {
		s.wg.Add(1)
		go s.worker()
	}
}

func (s *Stager) Stop() {
	close(s.stopCh)
	s.wg.Wait()
}

func (s *Stager) Add(phistage *common.Phistage) {
	s.stageCh <- phistage
}

func (s *Stager) worker() error {
	defer s.wg.Done()
	for {
		select {
		case <-s.stopCh:
			return nil
		case phistage := <-s.stageCh:
			if err := s.run(phistage); err != nil {
				fmt.Println(err)
			}
		}
	}
}

func (s *Stager) run(phistage *common.Phistage) error {
	jobs, err := phistage.JobDependencies()
	if err != nil {
		return err
	}

	for _, js := range jobs {
		wg := sync.WaitGroup{}
		for _, j := range js {
			wg.Add(1)
			go func(j *common.Job) {
				defer wg.Done()
				executorProvider := executors.GetExecutorProvider(phistage.Executor)
				if executorProvider == nil {
					fmt.Println("error getting provider")
					return
				}

				executor, err := executorProvider.GetJobExecutor(j, phistage)
				if err != nil {
					fmt.Printf("error creating executor, %v\n", err)
					return
				}

				defer func() {
					if err := executor.Cleanup(context.TODO()); err != nil {
						fmt.Printf("error cleaning, %v\n", err)
						return
					}
				}()

				if err := executor.Prepare(context.TODO()); err != nil {
					fmt.Printf("error preparing, %v\n", err)
					return
				}

				if err := executor.Execute(context.TODO()); err != nil {
					fmt.Printf("error executing, %v\n", err)
					return
				}

			}(j)
		}
		wg.Wait()
	}
	return nil
}
