package mysql

import "github.com/projecteru2/pistage/common"

func testingPistage() *common.Pistage {
	return &common.Pistage{
		WorkflowNamespace: "test-namespace",
		WorkflowIdentifier: "test-identifier",
		Jobs: map[string]*common.Job{
			"job1": {
				Name: "job1",
				Steps: []*common.Step{
					{
						Name: "step1",
					},
				},
			},
		},
		Environment: map[string]string{},
		Executor: "eru",
	}
}

func (s *MySQLStoreTestSuite) TestPistageSnapshot() {
	id, err := s.ms.CreatePistageSnapshot(testingPistage())
	s.NoError(err)
	s.NotEqual("", id)

	id2, err := s.ms.CreatePistageSnapshot(testingPistage())
	s.NoError(err)
	s.Equal(id, id2)

	p2 := testingPistage()
	p2.Jobs["job2"] = &common.Job{
		Name: "job2",
		Steps: []*common.Step{
			{
				Name: "step1",
			},
		},
	}

	id3, err := s.ms.CreatePistageSnapshot(p2)
	s.NoError(err)
	s.NotEqual("", id3)
	s.NotEqual(id, id3)

	pr, err := s.ms.GetPistageBySnapshotID(id)
	s.NoError(err)
	s.Equal("test-namespace", pr.WorkflowNamespace)
	_, ok := pr.Jobs["job1"]
	s.True(ok)
	_, ok = pr.Jobs["job2"]
	s.False(ok)

	pr, err = s.ms.GetPistageBySnapshotID(id3)
	s.NoError(err)
	s.Equal("test-namespace", pr.WorkflowNamespace)
	_, ok = pr.Jobs["job1"]
	s.True(ok)
	_, ok = pr.Jobs["job2"]
	s.True(ok)
}

func (s *MySQLStoreTestSuite) TestPistageRun() {
	id, err := s.ms.CreatePistageRun(testingPistage(), "1")
	s.NoError(err)
	s.NotEqual("", id)

	run, err := s.ms.GetPistageRun(id)
	s.NoError(err)
	s.Equal(id, run.ID)
	s.Equal("test-namespace", run.WorkflowNamespace)
	s.Equal(common.RunStatusPending, run.Status)

	run.Status = common.RunStatusRunning
	run.Start = common.EpochMillis()
	s.NoError(s.ms.UpdatePistageRun(run))

	run, err = s.ms.GetPistageRun(id)
	s.NoError(err)
	s.Equal(id, run.ID)
	s.Equal("test-namespace", run.WorkflowNamespace)
	s.Equal(common.RunStatusRunning, run.Status)
	s.Greater(run.Start, int64(0))
}
