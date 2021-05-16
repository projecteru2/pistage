package store

import (
	"hash/crc32"
	"os"

	"github.com/bwmarrin/snowflake"
)

func NewSnowflake() (*snowflake.Node, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}

	checksum := crc32.ChecksumIEEE([]byte(hostname))
	nodeID := checksum % 1024
	return snowflake.NewNode(int64(nodeID))
}
