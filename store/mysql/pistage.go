package mysql

import (
	"errors"
	"fmt"
	"gorm.io/gorm"
	"strconv"

	"github.com/projecteru2/pistage/common"
)

type PistageSnapshotModel struct {
	ID                 int64  `gorm:"primaryKey"`
	CreateTime         int64  `gorm:"column:create_time;autoCreateTime:milli"`
	UpdateTime         int64  `gorm:"column:update_time;autoUpdateTime:milli"`
	WorkflowNamespace  string `gorm:"workflow_namespace"`
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
		WorkflowNamespace:  pistage.WorkflowNamespace,
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
	ID         int64 `gorm:"primaryKey"`
	CreateTime int64 `gorm:"column:create_time;autoCreateTime:milli"`
	UpdateTime int64 `gorm:"column:update_time;autoUpdateTime:milli"`
	StartTime  int64 `gorm:"column:start_time"`
	EndTime    int64 `gorm:"column:end_time"`

	WorkflowNamespace  string `gorm:"workflow_namespace"`
	WorkflowIdentifier string `gorm:"workflow_identifier"`
	SnapshotVersion    int64  `gorm:"snapshot_version"`
	RunStatus          string `gorm:"run_status"`
}

func (PistageRunModel) TableName() string {
	return "pistage_run_tab"
}

func (ms *MySQLStore) CreatePistageRun(pistage *common.Pistage, version string) (id string, err error) {
	snapshotVersion, _ := strconv.ParseInt(version, 10, 64)
	run := PistageRunModel{
		WorkflowNamespace:  pistage.WorkflowNamespace,
		WorkflowIdentifier: pistage.WorkflowIdentifier,
		SnapshotVersion:    snapshotVersion,
		RunStatus:          string(common.RunStatusPending),
	}
	if err = ms.db.Create(&run).Error; err == nil {
		id = strconv.FormatInt(run.ID, 10)
	}
	return
}

func (ms *MySQLStore) GetPistageRunByNamespaceAndFlowIdentifier(workflowNamespace string,
	workflowIdentifier string) (pistageRun *common.Run, err error) {
	var pistageRunModel PistageRunModel
	fmt.Println("workflowNamespace is " + workflowNamespace + "workflowIdentifier is " + workflowIdentifier)
	err = ms.db.Debug().Where("workflow_namespace = ?", workflowNamespace).
		Where("workflow_identifier = ?", workflowIdentifier).Last(&pistageRunModel).Error
	fmt.Println("err is ", err)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	fmt.Println("93 ")
	pistageRun = &common.Run{
		ID:                 string(pistageRunModel.ID),
		WorkflowNamespace:  pistageRunModel.WorkflowNamespace,
		WorkflowIdentifier: pistageRunModel.WorkflowIdentifier,
		Status:             common.RunStatus(pistageRunModel.RunStatus),
		Start:              pistageRunModel.StartTime,
		End:                pistageRunModel.EndTime,
	}
	fmt.Println("102 pistageRun is ", pistageRun)
	return pistageRun, nil
}

func (ms *MySQLStore) GetPistageRun(id string) (run *common.Run, err error) {
	var pistageRun PistageRunModel
	err = ms.db.First(&pistageRun, id).Error
	if err != nil {
		return
	}
	run = &common.Run{
		ID:                 id,
		WorkflowNamespace:  pistageRun.WorkflowNamespace,
		WorkflowIdentifier: pistageRun.WorkflowIdentifier,
		Status:             common.RunStatus(pistageRun.RunStatus),
		Start:              pistageRun.StartTime,
		End:                pistageRun.EndTime,
	}
	return
}

func (ms *MySQLStore) UpdatePistageRun(run *common.Run) error {
	return ms.db.Model(&PistageRunModel{}).Where("id = ?", run.ID).Updates(map[string]interface{}{
		"start_time": run.Start,
		"end_time":   run.End,
		"run_status": run.Status,
	}).Error
}
