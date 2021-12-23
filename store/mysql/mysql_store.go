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

// findWithPagination calls conn.Find(dst) while also returning the total number of results in the query
// This method will not work if conn contains a .Distinct(table.*) statement because COUNT(DISTINCT table.*) is invalid SQL
// If this is required, a workaround is to instead use:
//     conn.Session(&gorm.Session{NewDB: true}).Table("(?) as foo", conn).Count(&cnt).Error
func (s *MySQLStore) findWithPagination(conn *gorm.DB, dst interface{}, pageSize int, pageNum int) (cnt int64, err error) {
	if err := conn.Count(&cnt).Error; err != nil {
		return 0, err
	}

	if err := conn.Limit(pageSize).Offset((pageNum - 1) * pageSize).Find(dst).Error; err != nil {
		return 0, err
	}

	return
}
