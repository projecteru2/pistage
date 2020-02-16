package os

import linux "os"

const hnEnv = "HOSTNAME"

// Hostname .
func Hostname() (string, error) {
	if hn := linux.Getenv(hnEnv); len(hn) > 0 {
		return hn, nil
	}
	return linux.Hostname()
}
