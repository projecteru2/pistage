package mysql

import (
	"context"
	"sync"

	"github.com/bwmarrin/snowflake"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

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
	gormDB, err := gorm.Open(mysql.Open(c.DSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
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

func (ms *MySQLStore) findWithPagination(conn *gorm.DB, dst interface{}, pageSize int, pageNum int) (cnt int64, err error) {
	// The more straightforward way is to use conn.Count(&cnt)
	// But, this doesn't work if conn contains a .Distinct(table.*) statement because COUNT(DISTINCT table.*) is invalid SQL
	// Workaround is to use the conditions in conn as a subquery
	if err := conn.Session(&gorm.Session{NewDB: true}).Table("(?) as foo", conn).Count(&cnt).Error; err != nil {
		return 0, err
	}

	if err := conn.Limit(pageSize).Offset((pageNum - 1) * pageSize).Find(dst).Error; err != nil {
		return 0, err
	}

	return
}
