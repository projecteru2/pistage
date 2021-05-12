package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"sync"

	"github.com/projecteru2/phistage/common"
	"github.com/projecteru2/phistage/executors/eru"
)

func main() {
	b, err := ioutil.ReadFile("./pistage.yml")
	if err != nil {
		fmt.Printf("error reading file, %v\n", err)
		return
	}

	p, err := common.FromSpec(b)
	if err != nil {
		fmt.Printf("error when loading, %v\n", err)
		return
	}

	jobs, err := p.JobDependencies()
	if err != nil {
		fmt.Printf("error resolving job dependencies, %v\n", err)
		return
	}

	for _, js := range jobs {
		wg := sync.WaitGroup{}
		for _, j := range js {
			wg.Add(1)
			go func(j *common.Job) {
				defer wg.Done()
				executor, err := eru.NewEruJobExecutor(j, p)
				if err != nil {
					fmt.Printf("error creating executor, %v\n", err)
					return
				}
				if err := executor.Prepare(context.TODO()); err != nil {
					fmt.Printf("error preparing, %v\n", err)
				}
				if err := executor.Execute(context.TODO()); err != nil {
					fmt.Printf("error executing, %v\n", err)
				}
				if err := executor.Cleanup(context.TODO()); err != nil {
					fmt.Printf("error cleaning, %v\n", err)
				}
			}(j)
		}
		wg.Wait()
	}
}
