package helpers

import (
	"os"

	"github.com/projecteru2/pistage/common"
	"github.com/projecteru2/pistage/store"
	"github.com/projecteru2/pistage/store/mysql"

	"github.com/sirupsen/logrus"
)

// InitStorage initiates storage, only one storage can be used.
func InitStorage(config *common.Config) (store.Store, error) {
	return mysql.NewMySQLStore(&config.Storage, store.NewKhoriumManager(config.Khorium))
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
