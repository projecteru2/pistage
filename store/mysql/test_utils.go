package mysql

import (
	"os"
	"strings"

	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/projecteru2/pistage/common"
)

func TruncateTables(db *gorm.DB) error {
	var (
		err  error
		sqls = `TRUNCATE TABLE pistage_snapshot_tab
TRUNCATE TABLE pistage_run_tab
TRUNCATE TABLE job_run_tab`
	)
	for _, sql := range strings.Split(sqls, "\n") {
		if terr := db.Exec(sql).Error; terr != nil {
			err = errors.Wrap(err, terr.Error())
		}
	}
	return err
}

func getEnvDefault(key, dft string) string {
	r, ok := os.LookupEnv(key)
	if !ok {
		return dft
	}
	return r
}

func NewTestingMySQLStore() (*MySQLStore, error) {
	return NewMySQLStore(&common.SQLDataSourceConfig{
		Username:     getEnvDefault("PISTAGE_MYSQL_USERNAME", "root"),
		Password:     getEnvDefault("PISTAGE_MYSQL_PASSWORD", ""),
		Host:         getEnvDefault("PISTAGE_MYSQL_HOST", "localhost"),
		Port:         3306,
		Database:     getEnvDefault("PISTAGE_MYSQL_DATABASE", "pistagetest"),
		MaxConns:     10,
		MaxIdleConns: 5,
	}, nil)
}
