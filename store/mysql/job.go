package mysql

import (
	"errors"
	"gorm.io/gorm"
	"strconv"

	"github.com/projecteru2/pistage/common"
)

type JobRunModel struct {
	ID         int64 `gorm:"primaryKey"`
	CreateTime int64 `gorm:"column:create_time;autoCreateTime:milli"`
	UpdateTime int64 `gorm:"column:update_time;autoUpdateTime:milli"`
	StartTime  int64 `gorm:"column:start_time"`
	EndTime    int64 `gorm:"column:end_time"`

	WorkflowNamespace  string `gorm:"workflow_namespace"`
	WorkflowIdentifier string `gorm:"workflow_identifier"`
	PistageRunID       int64  `gorm:"pistage_run_id"`
	JobName            string `gorm:"job_name"`
	RunStatus          string `gorm:"run_status"`
}

func (JobRunModel) TableName() string {
	return "job_run_tab"
}

func (ms *MySQLStore) CreateJobRun(run *common.Run, jobRun *common.JobRun) error {
	pistageRunID, _ := strconv.ParseInt(run.ID, 10, 64)
	jobRunModel := &JobRunModel{
		WorkflowNamespace:  run.WorkflowNamespace,
		WorkflowIdentifier: run.WorkflowIdentifier,
		PistageRunID:       pistageRunID,
		JobName:            jobRun.JobName,
		RunStatus:          string(common.RunStatusPending),
	}
	err := ms.db.Create(jobRunModel).Error
	if err != nil {
		return err
	}
	jobRun.ID = strconv.FormatInt(jobRunModel.ID, 10)
	return nil
}

func (ms *MySQLStore) GetJobRun(id string) (run *common.JobRun, err error) {
	var runModel JobRunModel
	if err = ms.db.First(&runModel, id).Error; err != nil {
		return
	}
	run = &common.JobRun{
		ID:                 strconv.FormatInt(runModel.ID, 10),
		WorkflowNamespace:  runModel.WorkflowNamespace,
		WorkflowIdentifier: runModel.WorkflowIdentifier,
		JobName:            runModel.JobName,
		Status:             common.RunStatus(runModel.RunStatus),
		Start:              runModel.StartTime,
		End:                runModel.EndTime,
	}
	return
}

func (ms *MySQLStore) UpdateJobRun(jobRun *common.JobRun) error {
	return ms.db.Model(&JobRunModel{}).Where("id = ?", jobRun.ID).Updates(map[string]interface{}{
		"start_time": jobRun.Start,
		"end_time":   jobRun.End,
		"run_status": string(jobRun.Status),
	}).Error
}

func (ms *MySQLStore) GetJobRunsByPistageRunId(pistageRunId string) (jobRuns []*common.JobRun, err error) {
	var result []*common.JobRun
	err = ms.db.First(&result).Where("pistage_run_id = ? ", pistageRunId).Error
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	return result, nil
}
