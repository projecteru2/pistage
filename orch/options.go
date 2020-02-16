package orch

import (
	"time"

	pb "github.com/projecteru2/core/rpc/gen"

	"github.com/projecteru2/aa/config"
)

// ExecuteOptions .
type ExecuteOptions struct {
	*pb.ExecuteContainerOptions
}

// NewExecuteOptions .
func NewExecuteOptions(containerID, workdir string, commands, envs []string) ExecuteOptions {
	return ExecuteOptions{
		ExecuteContainerOptions: &pb.ExecuteContainerOptions{
			ContainerId: containerID,
			Commands:    commands,
			Envs:        envs,
			Workdir:     workdir,
		},
	}
}

// LambdaOptions .
type LambdaOptions struct {
	ResourceMetadata
	Appname string
	Command string
	Data    map[string][]byte
	Timeout int
	Labels  map[string]string
}

// NewLambdaOptions .
func NewLambdaOptions(image, command string, timeout time.Duration) LambdaOptions {
	return LambdaOptions{
		ResourceMetadata: ResourceMetadata{
			CPU:     float64(config.Conf.LambdaCPU),
			CPUBind: config.Conf.LambdaCPUBind,
			Memory:  config.Conf.LambdaMemory,
			Image:   image,
			Podname: config.Conf.LambdaPodname,
			Count:   1,
			Network: config.Conf.LambdaNetwork,
		},
		Appname: config.Conf.LambdaAppname,
		Command: command,
		Timeout: int(timeout.Seconds()),
		Labels:  map[string]string{},
	}
}
