package main

import (
	"io/ioutil"

	"github.com/projecteru2/phistage/common"
	"github.com/projecteru2/phistage/executors"
	"github.com/projecteru2/phistage/executors/eru"
	"github.com/projecteru2/phistage/stager"
	"github.com/sethvargo/go-signalcontext"
)

func main() {
	config := &common.Config{
		DefaultJobExecutor:       "eru",
		DefaultJobExecuteTimeout: 1200,
		EruAddress:               "10.22.12.87:5001",
		StagerWorkers:            5,
	}

	eruProvider, err := eru.NewEruJobExecutorProvider(config)
	if err != nil {
		return
	}
	executors.RegisterExecutorProvider(eruProvider.GetName(), eruProvider)

	s := stager.NewStager(config)
	if err != nil {
		return
	}

	s.Start()

	go func() {
		content, err := ioutil.ReadFile("./pistage.yml")
		if err != nil {
			return
		}

		phistage, err := common.FromSpec(content)
		if err != nil {
			return
		}

		s.Add(phistage)
	}()

	ctx, cancel := signalcontext.OnInterrupt()
	defer cancel()

	select {
	case <-ctx.Done():
		s.Stop()
	}
}
