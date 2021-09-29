package mysql

import "github.com/projecteru2/pistage/common"

func testingRun() *common.Run {
	return &common.Run{
		ID: "1",
		WorkflowNamespace: "testing-namespace",
		WorkflowIdentifier: "testing-identifier",
		Status: common.RunStatusRunning,
		Start: common.EpochMillis(),
	}
}

func testingJobRun(jobName string) *common.JobRun {
	return &common.JobRun{
		ID: "1",
		WorkflowNamespace: "testing-namespace",
		WorkflowIdentifier: "testing-identifier",
		JobName: jobName,
		Status: common.RunStatusRunning,
		Start: common.EpochMillis(),
	}
}

func (s *MySQLStoreTestSuite) TestJobRun() {
	jobRun := testingJobRun("job1")
	s.NoError(s.ms.CreateJobRun(testingRun(), jobRun))
	s.NotEqual("", jobRun.ID)

	jobRun2 := testingJobRun("job2")
	s.NoError(s.ms.CreateJobRun(testingRun(), jobRun2))
	s.NotEqual("", jobRun2.ID)
	s.NotEqual(jobRun.ID, jobRun2.ID)

	jobRun2.Status = common.RunStatusFailed
	jobRun2.End = common.EpochMillis() + 1
	s.NoError(s.ms.UpdateJobRun(jobRun2))

	jobRun2, err := s.ms.GetJobRun(jobRun2.ID)
	s.NoError(err)
	s.Equal("testing-namespace", jobRun2.WorkflowNamespace)
	s.Equal(common.RunStatusFailed, jobRun2.Status)
	s.Greater(jobRun2.End, jobRun2.Start)
}
