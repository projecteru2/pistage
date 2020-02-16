package io

import (
	"io/ioutil"
)

// ReadFile reads all content of a specific file path.
func ReadFile(filepath string) ([]byte, error) {
	return ioutil.ReadFile(filepath)
}
