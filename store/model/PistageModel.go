package model

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