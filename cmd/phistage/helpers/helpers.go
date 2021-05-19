package helpers

import (
	"os"

	"github.com/projecteru2/phistage/common"
	"github.com/projecteru2/phistage/executors"
	"github.com/projecteru2/phistage/executors/eru"
	"github.com/projecteru2/phistage/store"
	"github.com/projecteru2/phistage/store/filesystem"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// ErrorStorageNotSpecified indicates error when setting the storage type.
var ErrorStorageNotSpecified = errors.New("Storage not specified")

// InitStorage initiates storage, only one storage can be used.
func InitStorage(config *common.Config) (store.Store, error) {
	switch config.Storage.Type {
	case "file":
		return filesystem.NewFileSystemStore(config.Storage.FileSystemStoreRoot)
	default:
		return nil, ErrorStorageNotSpecified
	}
}

// InitExecutorProvider initiates and registers executor providers.
func InitExecutorProvider(config *common.Config, store store.Store) error {
	// init eru executor provider.
	// this is the only kind currently.
	eruProvider, err := eru.NewEruJobExecutorProvider(config, store)
	if err != nil {
		return err
	}
	executors.RegisterExecutorProvider(eruProvider)
	return nil
}

// SetupLog initiates logrus default logger.
func SetupLog(levelName string) error {
	level, err := logrus.ParseLevel(levelName)
	if err != nil {
		return err
	}
	logrus.SetLevel(level)

	formatter := &logrus.TextFormatter{
		ForceColors:     true,
		TimestampFormat: "2006-01-02 15:04:05",
		FullTimestamp:   true,
	}
	logrus.SetFormatter(formatter)
	logrus.SetOutput(os.Stdout)
	return nil
}
