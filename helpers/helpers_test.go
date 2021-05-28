package helpers

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSha1HexDigest(t *testing.T) {
	assert := assert.New(t)

	s1, err := Sha1HexDigest("tonic")
	assert.NoError(err)
	assert.Equal(s1, "dfe953579d49b555adf16d1823a71a8e463351c2")

	s2, err := Sha1HexDigest([]byte{'t', 'o', 'n', 'i', 'c'})
	assert.NoError(err)
	assert.Equal(s2, "dfe953579d49b555adf16d1823a71a8e463351c2")
}

func TestFileOperations(t *testing.T) {
	assert := assert.New(t)

	root, err := ioutil.TempDir("", "phistage-*")
	assert.NoError(err)
	defer os.RemoveAll(root)

	f1 := filepath.Join(root, "逍遥派")

	assert.NoError(OverWriteFile(f1, "北冥神功"))
	c1, err := ioutil.ReadFile(f1)
	assert.NoError(err)
	assert.Equal(string(c1), "北冥神功")

	assert.NoError(OverWriteFile(f1, "北冥神功"))
	c2, err := ioutil.ReadFile(f1)
	assert.NoError(err)
	assert.Equal(string(c2), "北冥神功")

	assert.NoError(OverWriteFile(f1, "天山六阳掌"))
	c3, err := ioutil.ReadFile(f1)
	assert.NoError(err)
	assert.Equal(string(c3), "天山六阳掌")

	f2 := filepath.Join(root, "明教")

	assert.NoError(WriteIfNotExist(f2, []byte("乾坤大挪移")))
	c4, err := ioutil.ReadFile(f2)
	assert.NoError(err)
	assert.Equal(string(c4), "乾坤大挪移")

	assert.NoError(WriteIfNotExist(f2, []byte("圣火令法")))
	c5, err := ioutil.ReadFile(f2)
	assert.NoError(err)
	assert.Equal(string(c5), "乾坤大挪移")
}
