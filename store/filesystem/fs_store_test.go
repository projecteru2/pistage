package filesystem

import (
	"context"
	"io/ioutil"
	"os"
	"testing"

	"github.com/projecteru2/phistage/common"
	"github.com/stretchr/testify/assert"
)

func TestFileSystemStorePhistage(t *testing.T) {
	assert := assert.New(t)
	rootfs, err := ioutil.TempDir("", "phistagetest-*")
	assert.NoError(err)
	defer os.RemoveAll(rootfs)

	fs, err := NewFileSystemStore(rootfs, nil)
	assert.NoError(err)
	content, err := ioutil.ReadFile("./pistage.test.yml")
	assert.NoError(err)

	p1, err := common.FromSpec(content)
	assert.NoError(err)

	assert.NoError(fs.CreatePhistage(context.TODO(), p1))

	p2, err := fs.GetPhistage(context.TODO(), p1.Name)
	assert.NoError(err)
	assert.Equal(p2.Name, p1.Name)
	assert.Equal(len(p2.Jobs), len(p1.Jobs))
	assert.Equal(len(p2.Environment), len(p1.Environment))
	assert.Equal(p2.Executor, p1.Executor)

	assert.NoError(fs.CreatePhistage(context.TODO(), p2))
	p3, err := fs.GetPhistage(context.TODO(), p1.Name)
	assert.NoError(err)
	assert.Equal(p3.Name, p1.Name)
	assert.Equal(len(p3.Jobs), len(p1.Jobs))
	assert.Equal(len(p3.Environment), len(p1.Environment))
	assert.Equal(p3.Executor, p1.Executor)

	assert.NoError(fs.DeletePhistage(context.TODO(), p1.Name))
	p4, err := fs.GetPhistage(context.TODO(), p1.Name)
	assert.Error(err)
	assert.Nil(p4)
}

func TestFileSystemRegister(t *testing.T) {
	assert := assert.New(t)
	rootfs, err := ioutil.TempDir("", "phistagetest-*")
	assert.NoError(err)
	defer os.RemoveAll(rootfs)

	fs, err := NewFileSystemStore(rootfs, nil)
	assert.NoError(err)
	content, err := ioutil.ReadFile("./pistage.test.yml")
	assert.NoError(err)

	p, err := common.FromSpec(content)
	assert.NoError(err)

	for _, job := range p.Jobs {
		assert.NoError(fs.RegisterJob(context.TODO(), job))
	}

	job1, err := fs.GetRegisteredJob(context.TODO(), "job1")
	assert.NoError(err)
	assert.Equal(job1.Name, "job1")
	assert.Equal(len(job1.Steps), 2)
	assert.Equal(job1.Files, []string{"job1file"})
	assert.Equal(job1.Timeout, 120)

	job2, err := fs.GetRegisteredJob(context.TODO(), "job2")
	assert.NoError(err)
	assert.Equal(job2.Name, "job2")
	assert.Equal(len(job2.Steps), 3)
	assert.Equal(job2.Files, []string{"file1", "file2"})
	assert.Equal(job2.DependsOn, []string{"job1"})
	assert.Equal(job2.Timeout, 120)

	job3, err := fs.GetRegisteredJob(context.TODO(), "job3")
	assert.NoError(err)
	assert.Equal(job3.Name, "job3")
	assert.Equal(len(job3.Steps), 2)
	assert.Equal(len(job3.Files), 0)
	assert.Equal(job3.DependsOn, []string{"job1", "job2"})
	assert.Equal(job3.Timeout, 120)

	for _, step := range job2.Steps {
		assert.NoError(fs.RegisterStep(context.TODO(), step))
	}

	step1, err := fs.GetRegisteredStep(context.TODO(), "build binary")
	assert.NoError(err)
	assert.Equal(step1.Run, []string{"env", "ls", "cat {{ $env.GOOS }}"})
	assert.Equal(len(step1.OnError), 1)
	assert.Equal(len(step1.Environment), 2)
	assert.Equal(len(step1.With), 1)
}
