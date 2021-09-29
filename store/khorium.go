package store

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/projecteru2/pistage/common"
)

// KhoriumManager manages khorium steps.
// Currently it can download (git clone)
// repo identified by khorium step name from
// GitLab and GitHub.
type KhoriumManager struct {
	config common.KhoriumConfig
}

func NewKhoriumManager(config common.KhoriumConfig) *KhoriumManager {
	return &KhoriumManager{config: config}
}

// GetKhoriumStep gets a khorium step from code repository.
// name can be like:
//   - github.com/test/checkout
//   - github.com/test/checkout@v2.1.1
//   - git.selfhostedgitlab.com/test/checkout
//   - git.selfhostedgitlab.com/test/checkout@v2.1.1
// If name has the prefix "github.com", it's considered as a GitHub repo,
// it can be cloned directly since GitHub is opened to all anonymous visit.
// Other scenarios are treated as self hosted GitLab, will need to use
// the access token and username to clone repository.
// The "@" symbol represents at which tag / commit / branch, will be used
// like "git checkout @symbol".
func (k *KhoriumManager) GetKhoriumStep(ctx context.Context, name string) (*common.KhoriumStep, error) {
	var (
		khoriumName    string
		khoriumVersion string
	)

	ps := strings.SplitN(name, "@", 2)
	if len(ps) == 2 {
		khoriumName = ps[0]
		khoriumVersion = ps[1]
	} else {
		khoriumName = name
		khoriumVersion = "master"
	}

	dir, err := ioutil.TempDir("", "pistage-khoriumstep-*")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(dir)

	if err := k.cloneRepository(ctx, khoriumName, khoriumVersion, dir); err != nil {
		return nil, err
	}
	return k.loadKhoriumStepFromFilesystem(dir)
}

func (k *KhoriumManager) cloneRepository(ctx context.Context, name, version, path string) error {
	var url string
	if strings.HasPrefix(name, "github.com") {
		url = fmt.Sprintf("https://%s.git", name)
	} else {
		url = fmt.Sprintf("https://%s:%s@%s.git", k.config.GitLabUsername, k.config.GitLabAccessToken, name)
	}

	// clone
	if err := exec.CommandContext(ctx, "git", "clone", url, path).Run(); err != nil {
		return err
	}

	// checkout
	checkout := exec.CommandContext(ctx, "git", "checkout", version)
	checkout.Dir = path
	if err := checkout.Run(); err != nil {
		return err
	}

	cleanup := exec.CommandContext(ctx, "rm", "-rf", ".git")
	cleanup.Dir = path
	return cleanup.Run()
}

func (k *KhoriumManager) loadKhoriumStepFromFilesystem(path string) (*common.KhoriumStep, error) {
	content, err := ioutil.ReadFile(filepath.Join(path, "khoriumstep.yml"))
	if err != nil {
		return nil, err
	}

	ks, err := common.LoadKhoriumStep(content)
	if err != nil {
		return nil, err
	}

	traverse := func(file string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		c, err := ioutil.ReadFile(file)
		if err != nil {
			return err
		}
		ks.Files[strings.TrimPrefix(file, path+"/")] = c
		return nil
	}

	if err := filepath.Walk(path, traverse); err != nil {
		return nil, err
	}
	return ks, nil
}
