package mysql

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type MySQLStoreTestSuite struct {
	suite.Suite

	ms *MySQLStore
}

func TestMySQLStore(t *testing.T) {
	ms, err := NewTestingMySQLStore()
	if err != nil {
		t.Error(err)
		return
	}
	defer ms.Close()

	suite.Run(t, &MySQLStoreTestSuite{ms: ms})
}

func (s *MySQLStoreTestSuite) SetupTest() {
	TruncateTables(s.ms.db)
}

func (s *MySQLStoreTestSuite) TearDownTest() {
	TruncateTables(s.ms.db)
}
