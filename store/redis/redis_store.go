package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/snowflake"
	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"

	"github.com/projecteru2/pistage/common"
	"github.com/projecteru2/pistage/helpers"
	"github.com/projecteru2/pistage/store"
)

type RedisStore struct {
	root      string
	mutex     sync.Mutex
	snowflake *snowflake.Node
	client    *redis.Client
}

func NewRedisStore(root, address, password string, db int) (*RedisStore, error) {
	sn, err := store.NewSnowflake()
	if err != nil {
		return nil, err
	}

	client := redis.NewClient(&redis.Options{
		Addr:     address,
		Password: password,
		DB:       db,
	})
	return &RedisStore{
		root:      root,
		snowflake: sn,
		client:    client,
	}, nil
}

// makeKey use ":" to join all components.
func makeKey(components ...string) string {
	return strings.Join(components, ":")
}

func currentTimestamp() float64 {
	return float64(time.Now().UnixNano())
}

// string ${root}:pistage:${sha1 of pistage name}:meta:${sha1 of file content}
// sorted set ${root}:pistage:${sha1 of pistage name}:meta:records, score: timestamp, member: sha1 of file content
func (rs *RedisStore) CreatePistage(ctx context.Context, pistage *common.Pistage) error {
	rs.mutex.Lock()
	defer rs.mutex.Unlock()

	sha1OfName, err := helpers.Sha1HexDigest(pistage.Name)
	if err != nil {
		return err
	}

	content, err := json.Marshal(pistage)
	if err != nil {
		return err
	}

	sha1OfFile, err := helpers.Sha1HexDigest(content)
	if err != nil {
		return err
	}

	contentKey := makeKey(rs.root, "pistage", sha1OfName, "meta", sha1OfFile)
	recordsKey := makeKey(rs.root, "pistage", sha1OfName, "meta:records")

	pipe := rs.client.Pipeline()
	pipe.Set(ctx, contentKey, content, 0)
	pipe.ZAdd(ctx, recordsKey, &redis.Z{Score: currentTimestamp(), Member: sha1OfFile})
	_, err = pipe.Exec(ctx)
	return err
}

// sorted set ${root}:pistage:${sha1 of pistage name}:meta:records, score: timestamp, member: sha1 of file content
func (rs *RedisStore) GetPistage(ctx context.Context, name string) (*common.Pistage, error) {
	rs.mutex.Lock()
	defer rs.mutex.Unlock()

	sha1OfName, err := helpers.Sha1HexDigest(name)
	if err != nil {
		return nil, err
	}

	key := makeKey(rs.root, "pistage", sha1OfName, "meta:records")
	keys, err := rs.client.ZRevRangeByScore(ctx, key, &redis.ZRangeBy{Min: "-inf", Max: "+inf", Offset: 0, Count: 1}).Result()
	if err != nil {
		return nil, err
	}
	if len(keys) != 1 {
		return nil, errors.New("pistage not found")
	}

	contentKey := makeKey(rs.root, "pistage", sha1OfName, "meta", keys[0])
	content, err := rs.client.Get(ctx, contentKey).Result()
	if err != nil {
		return nil, err
	}

	p := &common.Pistage{}
	if err := json.Unmarshal([]byte(content), p); err != nil {
		return nil, err
	}
	return p, nil
}

func (rs *RedisStore) DeletePistage(ctx context.Context, name string) error {
	rs.mutex.Lock()
	defer rs.mutex.Unlock()

	sha1OfName, err := helpers.Sha1HexDigest(name)
	if err != nil {
		return err
	}

	pattern := makeKey(rs.root, "pistage", sha1OfName, "*")
	keys, err := rs.client.Keys(ctx, pattern).Result()
	if err != nil {
		return err
	}
	_, err = rs.client.Del(ctx, keys...).Result()
	return err
}

// string ${root}:run:${run id}:run
// sorted set ${root}:pistage:${sha1 of pistage name}:run:records, score: timestamp, member: run id
func (rs *RedisStore) CreateRun(ctx context.Context, run *common.Run) error {
	rs.mutex.Lock()
	defer rs.mutex.Unlock()

	run.ID = rs.snowflake.Generate().String()

	// run
	runKey := makeKey(rs.root, "run", run.ID, "run")
	content, err := json.Marshal(run)
	if err != nil {
		return err
	}
	created, err := rs.client.SetNX(ctx, runKey, content, 0).Result()
	switch {
	case err != nil:
		return err
	case !created:
		return nil
	}

	// pistage run
	sha1OfPistageName, err := helpers.Sha1HexDigest(run.Pistage)
	if err != nil {
		return err
	}

	pistageRunRecordsKey := makeKey(rs.root, "pistage", sha1OfPistageName, "run:records")
	_, err = rs.client.ZAdd(ctx, pistageRunRecordsKey, &redis.Z{Score: currentTimestamp(), Member: run.ID}).Result()
	return err
}

func (rs *RedisStore) GetRun(ctx context.Context, id string) (*common.Run, error) {
	rs.mutex.Lock()
	defer rs.mutex.Unlock()

	runKey := makeKey(rs.root, "run", id, "run")
	content, err := rs.client.Get(ctx, runKey).Result()
	if err != nil {
		return nil, err
	}

	run := &common.Run{}
	if err := json.Unmarshal([]byte(content), run); err != nil {
		return nil, err
	}
	return run, nil
}

func (rs *RedisStore) UpdateRun(ctx context.Context, run *common.Run) error {
	rs.mutex.Lock()
	defer rs.mutex.Unlock()

	runKey := makeKey(rs.root, "run", run.ID, "run")

	content, err := json.Marshal(run)
	if err != nil {
		return err
	}
	_, err = rs.client.Set(ctx, runKey, content, 0).Result()
	return err
}

// sorted set ${root}:pistage:${sha1 of pistage name}:run:records, score: timestamp, member: run id
func (rs *RedisStore) GetRunsByPistage(ctx context.Context, name string) ([]*common.Run, error) {
	rs.mutex.Lock()
	defer rs.mutex.Unlock()

	sha1OfName, err := helpers.Sha1HexDigest(name)
	if err != nil {
		return nil, err
	}

	key := makeKey(rs.root, "pistage", sha1OfName, "run:records")
	ids, err := rs.client.ZRevRangeByScore(ctx, key, &redis.ZRangeBy{Min: "-inf", Max: "+inf"}).Result()
	if err != nil {
		return nil, err
	}

	var runs []*common.Run
	for _, id := range ids {
		run, err := rs.GetRun(ctx, id)
		if err != nil {
			return nil, err
		}
		runs = append(runs, run)
	}
	return runs, nil
}

// string ${root}:run:${id}:jobrun:${jobrun id}
// sorted set ${root}:run:${id}:jobrun:records, score: timestamp, member: jobrun id
func (rs *RedisStore) CreateJobRun(ctx context.Context, run *common.Run, jobRun *common.JobRun) error {
	rs.mutex.Lock()
	defer rs.mutex.Unlock()

	jobRun.ID = rs.snowflake.Generate().String()

	content, err := json.Marshal(jobRun)
	if err != nil {
		return err
	}

	jobRunKey := makeKey(rs.root, "run", run.ID, "jobrun", jobRun.ID)
	created, err := rs.client.SetNX(ctx, jobRunKey, content, 0).Result()
	switch {
	case err != nil:
		return err
	case !created:
		return nil
	}

	jobRunRecordsKey := makeKey(rs.root, "run", run.ID, "jobrun:records")
	_, err = rs.client.ZAdd(ctx, jobRunRecordsKey, &redis.Z{Score: currentTimestamp(), Member: jobRun.ID}).Result()
	return err
}

// string ${root}:run:${id}:jobrun:${jobrun id}
func (rs *RedisStore) GetJobRun(ctx context.Context, runID, jobRunID string) (*common.JobRun, error) {
	rs.mutex.Lock()
	defer rs.mutex.Unlock()

	jobRunKey := makeKey(rs.root, "run", runID, "jobrun", jobRunID)
	content, err := rs.client.Get(ctx, jobRunKey).Result()
	if err != nil {
		return nil, err
	}

	jobRun := &common.JobRun{}
	if err := json.Unmarshal([]byte(content), jobRun); err != nil {
		return nil, err
	}
	return jobRun, nil
}

// string ${root}:run:${id}:jobrun:${jobrun id}
func (rs *RedisStore) UpdateJobRun(ctx context.Context, run *common.Run, jobRun *common.JobRun) error {
	rs.mutex.Lock()
	defer rs.mutex.Unlock()

	content, err := json.Marshal(jobRun)
	if err != nil {
		return err
	}

	jobRunKey := makeKey(rs.root, "run", run.ID, "jobrun", jobRun.ID)
	_, err = rs.client.Set(ctx, jobRunKey, content, 0).Result()
	return err
}

// string ${root}:run:${id}:jobrun:${jobrun id}
// string ${root}:run:${id}:jobrun:${jobrun id}.log
func (rs *RedisStore) FinishJobRun(ctx context.Context, run *common.Run, jobRun *common.JobRun) error {
	rs.mutex.Lock()
	defer rs.mutex.Unlock()

	jobRun.End = time.Now()
	jobRun.Status = common.JobRunStatusFinished
	content, err := json.Marshal(jobRun)
	if err != nil {
		return err
	}

	jobRunKey := makeKey(rs.root, "run", run.ID, "jobrun", jobRun.ID)
	if _, err = rs.client.Set(ctx, jobRunKey, content, 0).Result(); err != nil {
		return err
	}

	jobRunLogKey := makeKey(rs.root, "run", run.ID, "jobrun", fmt.Sprintf("%s.log", jobRun.ID))
	content, err = ioutil.ReadAll(jobRun.LogTracer)
	if err != nil {
		return err
	}
	_, err = rs.client.Set(ctx, jobRunLogKey, content, 0).Result()
	return err
}

// string ${root}:run:${id}:jobrun:${jobrun id}
// sorted set ${root}:run:${id}:jobrun:records, score: timestamp, member: jobrun id
func (rs *RedisStore) GetJobRuns(ctx context.Context, runID string) ([]*common.JobRun, error) {
	rs.mutex.Lock()
	defer rs.mutex.Unlock()

	key := makeKey(rs.root, "run", runID, "jobrun:records")
	ids, err := rs.client.ZRevRangeByScore(ctx, key, &redis.ZRangeBy{Min: "-inf", Max: "+inf"}).Result()
	if err != nil {
		return nil, err
	}

	var jobRuns []*common.JobRun
	for _, id := range ids {
		jobRun, err := rs.GetJobRun(ctx, runID, id)
		if err != nil {
			return nil, err
		}
		jobRuns = append(jobRuns, jobRun)
	}
	return jobRuns, nil
}

// string ${root}:registered:job:${sha1 of job name}
func (rs *RedisStore) RegisterJob(ctx context.Context, job *common.Job) error {
	rs.mutex.Lock()
	defer rs.mutex.Unlock()

	sha1OfName, err := helpers.Sha1HexDigest(job.Name)
	if err != nil {
		return err
	}

	content, err := json.Marshal(job)
	if err != nil {
		return err
	}

	key := makeKey(rs.root, "registered:job", sha1OfName)
	_, err = rs.client.SetNX(ctx, key, content, 0).Result()
	return err
}

// string ${root}:registered:job:${sha1 of job name}
func (rs *RedisStore) GetRegisteredJob(ctx context.Context, name string) (*common.Job, error) {
	rs.mutex.Lock()
	defer rs.mutex.Unlock()

	sha1OfName, err := helpers.Sha1HexDigest(name)
	if err != nil {
		return nil, err
	}

	key := makeKey(rs.root, "registered:job", sha1OfName)
	content, err := rs.client.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	job := &common.Job{}
	if err := json.Unmarshal([]byte(content), job); err != nil {
		return nil, err
	}
	return job, nil
}

// string ${root}:registered:step:${sha1 of step name}
func (rs *RedisStore) RegisterStep(ctx context.Context, step *common.Step) error {
	rs.mutex.Lock()
	defer rs.mutex.Unlock()

	sha1OfName, err := helpers.Sha1HexDigest(step.Name)
	if err != nil {
		return err
	}

	content, err := json.Marshal(step)
	if err != nil {
		return err
	}

	key := makeKey(rs.root, "registered:step", sha1OfName)
	_, err = rs.client.SetNX(ctx, key, content, 0).Result()
	return err
}

// string ${root}:registered:step:${sha1 of step name}
func (rs *RedisStore) GetRegisteredStep(ctx context.Context, name string) (*common.Step, error) {
	rs.mutex.Lock()
	defer rs.mutex.Unlock()

	sha1OfName, err := helpers.Sha1HexDigest(name)
	if err != nil {
		return nil, err
	}

	key := makeKey(rs.root, "registered:step", sha1OfName)
	content, err := rs.client.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	step := &common.Step{}
	if err := json.Unmarshal([]byte(content), step); err != nil {
		return nil, err
	}
	return step, nil
}

// string ${root}:pistage:${sha1 of pistage name}:vars
func (rs *RedisStore) SetVariablesForPistage(ctx context.Context, name string, vars map[string]string) error {
	rs.mutex.Lock()
	defer rs.mutex.Unlock()

	sha1OfName, err := helpers.Sha1HexDigest(name)
	if err != nil {
		return err
	}

	content, err := json.Marshal(vars)
	if err != nil {
		return err
	}

	key := makeKey(rs.root, "pistage", sha1OfName, "vars")
	_, err = rs.client.Set(ctx, key, content, 0).Result()
	return err
}

func isRedisNoKeyError(e error) bool {
	return strings.Contains(e.Error(), "redis:nil")
}

// string ${root}:pistage:${sha1 of pistage name}:vars
func (rs *RedisStore) GetVariablesForPistage(ctx context.Context, name string) (map[string]string, error) {
	rs.mutex.Lock()
	defer rs.mutex.Unlock()

	sha1OfName, err := helpers.Sha1HexDigest(name)
	if err != nil {
		return nil, err
	}

	key := makeKey(rs.root, "pistage", sha1OfName, "vars")
	content, err := rs.client.Get(ctx, key).Result()
	if err != nil {
		if isRedisNoKeyError(err) {
			return nil, nil
		}
		return nil, err
	}

	vars := map[string]string{}
	if err := json.Unmarshal([]byte(content), &vars); err != nil {
		return nil, err
	}
	return vars, nil
}

// string ${root}:registered:khoriumstep:${sha1 of step name}:khoriumstep.yml
func (rs *RedisStore) RegisterKhoriumStep(ctx context.Context, ks *common.KhoriumStep) error {
	rs.mutex.Lock()
	defer rs.mutex.Unlock()

	sha1OfName, err := helpers.Sha1HexDigest(ks.Name)
	if err != nil {
		return err
	}

	content, err := yaml.Marshal(ks)
	if err != nil {
		return err
	}
	key := makeKey(rs.root, "registered:khoriumstep", sha1OfName, "khoriumstep.yml")
	_, err = rs.client.SetNX(ctx, key, content, 0).Result()
	return err
}

// string ${root}:registered:khoriumstep:${sha1 of step name}:khoriumstep.yml
func (rs *RedisStore) GetRegisteredKhoriumStep(ctx context.Context, name string) (*common.KhoriumStep, error) {
	rs.mutex.Lock()
	defer rs.mutex.Unlock()

	sha1OfName, err := helpers.Sha1HexDigest(name)
	if err != nil {
		return nil, err
	}

	key := makeKey(rs.root, "registered", "khoriumstep", sha1OfName, "khoriumstep.yml")
	content, err := rs.client.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	ks, err := common.LoadKhoriumStep([]byte(content))
	if err != nil {
		return nil, err
	}
	return ks, nil
}
