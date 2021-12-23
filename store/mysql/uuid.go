package mysql

import (
	"fmt"
	"math/rand"
	"time"

	"gorm.io/gorm"
)

type UUIDMixin struct {
	UUID string `gorm:"column:uuid;unique"`
}

type UUIDModel struct {
	ID         int64 `gorm:"primaryKey"`
	CreateTime int64 `gorm:"column:create_time;autoCreateTime:milli"`

	UUIDMixin
}

func (UUIDModel) TableName() string {
	return "uuid_tab"
}

func generateRandomUUIDWithTimestamp() string {
	// generate random number with timestamp
	// [0 ... 35](36 bits) random number, [36...63](28 bit) current timestamp
	// (a full timestamp takes 32 bits, we emit the first 4 bits because they hardly change)
	uuidUint64 := (rand.Uint64() << 28) | (uint64(time.Now().Unix()) & 0xfffffff) //nolint:gosec, gomnd

	// UUID in HEX form will be a 16-char string
	return fmt.Sprintf("%016x", uuidUint64)
}

// GemerateUUID creates a new UUID string and removes the hyphens
// Then, it checks uuid_tab for uniqueness. If not unique, there are 5 retries.
func GenerateUUID(db *gorm.DB) (uuid string, err error) {
	maxRetries := 5

	for numRetries := 0; numRetries < maxRetries; numRetries++ {
		uuidModel := &UUIDModel{}
		uuidModel.UUID = generateRandomUUIDWithTimestamp()

		if err := db.Create(uuidModel).Error; err == nil {
			return uuidModel.UUID, nil
		}
	}

	return "", fmt.Errorf("cannot get a valid UUID String after %d tries", maxRetries)
}
