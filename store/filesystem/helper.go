package filesystem

import (
	"crypto/sha1"
	"encoding/hex"
	"io"
	"os"
)

func Sha1HexDigest(content string) (string, error) {
	h := sha1.New()
	_, err := io.WriteString(h, content)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func Sha1HexDigestForBytes(content []byte) (string, error) {
	h := sha1.New()
	_, err := h.Write(content)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func overrideFile(path string, content string) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	if _, err := f.WriteString(content); err != nil {
		return err
	}
	return f.Close()
}

func writeIfNotExist(path string, content []byte) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		if os.IsExist(err) {
			return nil
		}
		return err
	}
	if _, err := f.Write(content); err != nil {
		return err
	}
	return f.Close()
}
