package mysql

import (
	"context"
	"sync"

	"github.com/bwmarrin/snowflake"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/projecteru2/pistage/common"
	"github.com/projecteru2/pistage/store"
)

type MySQLStore struct {
	root           string
	mutex          sync.Mutex
	snowflake      *snowflake.Node
	khoriumManager *store.KhoriumManager
	db             *gorm.DB
}

func NewMySQLStore(c *common.SQLDataSourceConfig, khoriumManager *store.KhoriumManager) (*MySQLStore, error) {
	gormDB, err := gorm.Open(mysql.Open(c.DSN()), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	sqlDB, err := gormDB.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(c.MaxConns)
	sqlDB.SetMaxIdleConns(c.MaxIdleConns)

	sn, err := store.NewSnowflake()
	if err != nil {
		return nil, err
	}
	return &MySQLStore{
		snowflake:      sn,
		khoriumManager: khoriumManager,
		db:             gormDB,
	}, nil
}

func (ms *MySQLStore) Close() error {
	sqlDB, err := ms.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func (ms *MySQLStore) GetRegisteredKhoriumStep(ctx context.Context, name string) (*common.KhoriumStep, error) {
	return ms.khoriumManager.GetKhoriumStep(ctx, name)
}
