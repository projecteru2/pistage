package mysql

import (
	"strconv"

	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/projecteru2/pistage/common"
)

type PistageSnapshotModel struct {
	ID                 int64  `gorm:"primaryKey"`
	CreateTime         int64  `gorm:"column:create_time;autoCreateTime:milli"`
	UpdateTime         int64  `gorm:"column:update_time;autoUpdateTime:milli"`
	WorkflowType       string `gorm:"workflow_type"`
	WorkflowIdentifier string `gorm:"workflow_identifier"`
	ContentHash        string `gorm:"content_hash"`
	Content            []byte `gorm:"content"`
}

func (PistageSnapshotModel) TableName() string {
	return "pistage_snapshot_tab"
}

func (ms *MySQLStore) CreatePistageSnapshot(pistage *common.Pistage) (id string, err error) {
	if pistage.Content == nil || pistage.ContentHash == "" {
		if err = pistage.GenerateHash(); err != nil {
			return
		}
	}
	var snapshot PistageSnapshotModel
	err = ms.db.FirstOrCreate(&snapshot, PistageSnapshotModel{
		WorkflowType:       pistage.WorkflowType,
		WorkflowIdentifier: pistage.WorkflowIdentifier,
		Content:            pistage.Content,
		ContentHash:        pistage.ContentHash,
	}).Error
	id = strconv.FormatInt(snapshot.ID, 10)
	return
}

func (ms *MySQLStore) GetPistageBySnapshotID(id string) (*common.Pistage, error) {
	var snapshot PistageSnapshotModel
	if err := ms.db.First(&snapshot, id).Error; err != nil {
		return nil, err
	}

	return common.UnmarshalPistage(snapshot.Content)
}

type PistageRunModel struct {
	ID int64 `gorm:"primaryKey"`
	UUIDMixin

	CreateTime int64 `gorm:"column:create_time;autoCreateTime:milli"`
	UpdateTime int64 `gorm:"column:update_time;autoUpdateTime:milli"`
	StartTime  int64 `gorm:"column:start_time"`
	EndTime    int64 `gorm:"column:end_time"`

	WorkflowType       string `gorm:"workflow_type"`
	WorkflowIdentifier string `gorm:"workflow_identifier"`
	SnapshotVersion    int64  `gorm:"snapshot_version"`
	RunStatus          string `gorm:"run_status"`
}

func (PistageRunModel) TableName() string {
	return "pistage_run_tab"
}

func (m *PistageRunModel) BeforeCreate(db *gorm.DB) error {
	uuid, err := GenerateUUID(db)
	m.UUID = uuid
	return err
}

func (ms *MySQLStore) CreatePistageRun(pistage *common.Pistage, version string) (id string, err error) {
	snapshotVersion, _ := strconv.ParseInt(version, 10, 64)
	run := PistageRunModel{
		WorkflowType:       pistage.WorkflowType,
		WorkflowIdentifier: pistage.WorkflowIdentifier,
		SnapshotVersion:    snapshotVersion,
		RunStatus:          string(common.RunStatusPending),
	}
	if err = ms.db.Create(&run).Error; err == nil {
		id = strconv.FormatInt(run.ID, 10)
	}
	return
}

func (ms *MySQLStore) GetLatestPistageRunByWorkflowIdentifier(workflowIdentifier string) (pistageRun *common.Run, err error) {
	var pistageRunModel PistageRunModel
	err = ms.db.
		Where("workflow_identifier = ?", workflowIdentifier).Last(&pistageRunModel).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	return pistageRunModel.toDTO(), nil
}

func (ms *MySQLStore) GetPistageRun(id string) (run *common.Run, err error) {
	var pistageRun PistageRunModel
	if err = ms.db.First(&pistageRun, id).Error; err != nil {
		return
	}
	return pistageRun.toDTO(), nil
}

func (ms *MySQLStore) UpdatePistageRun(run *common.Run) error {
	return ms.db.Model(&PistageRunModel{}).Where("id = ?", run.ID).Updates(map[string]interface{}{
		"start_time": run.Start,
		"end_time":   run.End,
		"run_status": run.Status,
	}).Error
}

type PistageRunModels []*PistageRunModel

func (ms *MySQLStore) GetPaginatedPistageRunsByWorkflowIdentifier(workflowIdentifier string, pageSize int, pageNum int) (pistageRuns []*common.Run, cnt int64, err error) {
	conn := ms.db.Model(&PistageRunModel{}).Where("workflow_identifier = ?", workflowIdentifier).Order("id")

	var pistageRunModels PistageRunModels

	cnt, err = ms.findWithPagination(conn, &pistageRunModels, pageSize, pageNum)
	return pistageRunModels.toDTOs(), cnt, nil
}

func (m *PistageRunModel) toDTO() *common.Run {
	return &common.Run{
		ID:                 strconv.FormatInt(m.ID, 10),
		UUID:               m.UUID,
		WorkflowType:       m.WorkflowType,
		WorkflowIdentifier: m.WorkflowIdentifier,
		Status:             common.RunStatus(m.RunStatus),
		Start:              m.StartTime,
		End:                m.EndTime,
	}
}

func (ms PistageRunModels) toDTOs() []*common.Run {
	dtos := make([]*common.Run, 0, len(ms))
	for _, m := range ms {
		dtos = append(dtos, m.toDTO())
	}
	return dtos
}
