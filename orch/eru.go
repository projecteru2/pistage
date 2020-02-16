package orch

import (
	"context"
	"io"
	"sync"
	"time"

	coreclient "github.com/projecteru2/core/client"
	corecluster "github.com/projecteru2/core/cluster"
	pb "github.com/projecteru2/core/rpc/gen"
	coretypes "github.com/projecteru2/core/types"

	"github.com/projecteru2/aa/config"
	"github.com/projecteru2/aa/errors"
	"github.com/projecteru2/aa/log"
	"github.com/projecteru2/aa/metrics"
)

// Eru .
type Eru struct {
	cli pb.CoreRPCClient
}

// NewEru .
func NewEru() *Eru {
	cc := coreclient.NewClient(
		config.Conf.EruAddr,
		coretypes.AuthConfig{
			Username: config.Conf.EruUsername,
			Password: config.Conf.EruPassword,
		},
	)
	return &Eru{cli: cc.GetRPCClient()}
}

// GetContainerID queries container ID via combination of appname, entrypoint and labels.
func (e Eru) GetContainerID(ctx context.Context, app, entry string, labels []string) (string, error) {
	return "cid", nil
}

// Execute .
func (e Eru) Execute(ctx context.Context, eopts ExecuteOptions) (<-chan Message, error) {
	exec, err := e.cli.ExecuteContainer(ctx)
	if err != nil {
		return nil, errors.Trace(err)
	}

	if err := exec.Send(eopts.ExecuteContainerOptions); err != nil {
		return nil, errors.Trace(err)
	}

	exit := newExitCh()
	noti := make(chan Message)

	go func() {
		defer close(noti)
		defer exit.close()

		if err := e.notify(ctx, exec, noti, exit); err != nil {
			log.ErrorStack(err)
			metrics.IncrError()
		}
	}()

	return noti, err
}

// Lambda .
func (e Eru) Lambda(ctx context.Context, lopts LambdaOptions) (string, <-chan Message, error) {
	dopts := &pb.DeployOptions{
		Name: config.Conf.EruDeployName,
		Entrypoint: &pb.EntrypointOptions{
			Name:    lopts.Appname,
			Command: lopts.Command,
			Dir:     "/",
		},
		CpuQuota:     lopts.CPU,
		CpuBind:      lopts.CPUBind,
		Memory:       lopts.Memory,
		Storage:      lopts.Storage,
		Podname:      lopts.Podname,
		Image:        lopts.Image,
		Count:        1,
		Env:          lopts.Env,
		Dns:          lopts.DNS,
		Volumes:      lopts.Volumes,
		Networkmode:  lopts.Network,
		User:         config.Conf.EruDeployUser,
		Data:         lopts.Data,
		IgnoreHook:   false,
		DeployMethod: corecluster.DeployAuto,
		Labels:       lopts.Labels,
	}

	opts := &pb.RunAndWaitOptions{
		DeployOptions: dopts,
		Cmd:           []byte(lopts.Command),
		Async:         false,
	}

	id, noti, err := e.lambda(ctx, opts)
	if err != nil {
		return "", nil, errors.Trace(err)
	}

	return id, noti, nil
}

func (e Eru) lambda(ctx context.Context, opts *pb.RunAndWaitOptions) (string, <-chan Message, error) {
	lamb, err := e.cli.RunAndWait(ctx)
	if err != nil {
		return "", nil, errors.Trace(err)
	}

	if err := lamb.Send(opts); err != nil {
		return "", nil, errors.Trace(err)
	}

	exit := newExitCh()
	noti := make(chan Message)

	go func() {
		defer close(noti)
		defer exit.close()

		if err := e.notify(ctx, lamb, noti, exit); err != nil {
			log.ErrorStack(err)
			metrics.IncrError()
		}
	}()

	id, err := e.getContainerID(ctx, opts.DeployOptions, exit)
	if err != nil {
		exit.close()
		return "", nil, errors.Trace(err)
	}

	return id, noti, nil
}

func (e Eru) notify(ctx context.Context, recv receiver, noti chan<- Message, exit exitCh) error {
	buf := make([]byte, 1024)
	next := 0
	start := -8

	for {
		msg := e.recv(recv)
		if msg.EOF || msg.Error != nil {
			if next > 0 {
				noti <- Message{Data: buf[:next]}
			}

			noti <- msg

			return msg.Error
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-exit.C:
			return nil
		default:
		}

		// TODO: process exitCode
		if start < 0 {
			start++
			continue
		}

		switch n := next + len(msg.Data[start:]); {
		case n == 1024:
			copy(buf[next:], msg.Data[start:])
			fallthrough

		case n > 1024:
			noti <- Message{Data: buf[:next]}
			next = 0

		default:
			next += copy(buf[next:], msg.Data[start:])
		}

		// skips the first byte due to https://stackoverflow.com/questions/52774830/docker-exec-command-from-golang-api
		start = 0
	}
}

func (e Eru) getContainerID(ctx context.Context, opts *pb.DeployOptions, exit exitCh) (string, error) {
	for i := 1; i <= 10; i = i % 10 {
		id, err := e.doGetContainerID(ctx, opts)
		if err == nil {
			return id, nil
		}

		if !errors.Contain(err, errors.ErrNoSuchContainer) {
			return "", err
		}

		select {
		case <-exit.C:
			break

		default:
			time.Sleep(time.Second * time.Duration(i))
			i++
		}
	}

	return "", errors.Annotatef(errors.ErrInvalidValue, "cannot fetch container ID, as exitCh had been closed")
}

func (e Eru) doGetContainerID(ctx context.Context, dopts *pb.DeployOptions) (string, error) {
	lopts := &pb.ListContainersOptions{
		Appname:    dopts.Name,
		Entrypoint: dopts.Entrypoint.Name,
		Labels:     dopts.Labels,
		Limit:      2,
	}
	conts, err := e.listContainers(ctx, lopts)
	if err != nil {
		return "", errors.Trace(err)
	}

	switch {
	case len(conts) < 1:
		return "", errors.Annotatef(errors.ErrNoSuchContainer, "for %s/%s with labels %s",
			lopts.Appname, lopts.Entrypoint, lopts.Labels)
	case len(conts) > 1:
		return "", errors.Annotatef(errors.ErrInvalidValue, "there are more than one container for %s/%s with labels %s",
			lopts.Appname, lopts.Entrypoint, lopts.Labels)
	}

	return conts[0].Id, nil
}

func (e Eru) listContainers(ctx context.Context, opts *pb.ListContainersOptions) ([]*pb.Container, error) {
	conts := []*pb.Container{}
	resp, err := e.cli.ListContainers(ctx, opts)
	if err != nil {
		return nil, errors.Trace(err)
	}

	for {
		con, err := resp.Recv()
		switch {
		case err == io.EOF:
			return conts, nil

		case err != nil:
			return nil, errors.Trace(err)

		default:
			conts = append(conts, con)
		}
	}
}

func (e Eru) recv(recv receiver) Message {
	switch msg, err := recv.Recv(); {
	case err == io.EOF:
		return Message{EOF: true}
	case err != nil:
		return Message{Error: err}
	default:
		return Message{ID: msg.ContainerId, Data: msg.Data}
	}
}

type receiver interface {
	Recv() (*pb.AttachContainerMessage, error)
}

type exitCh struct {
	*sync.Once
	C chan struct{}
}

func newExitCh() exitCh {
	return exitCh{
		Once: &sync.Once{},
		C:    make(chan struct{}),
	}
}

func (c exitCh) close() {
	c.Once.Do(func() {
		close(c.C)
	})
}
