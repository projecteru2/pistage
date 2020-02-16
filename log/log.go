package log

import (
	"os"

	"github.com/sirupsen/logrus"

	"github.com/projecteru2/aa/errors"
)

// Setup .
func Setup(level, file string) error {
	if err := setupLevel(level); err != nil {
		return errors.Trace(err)
	}

	logrus.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})

	if err := setupOutput(file); err != nil {
		return errors.Trace(err)
	}

	return nil
}

func setupOutput(file string) error {
	if len(file) < 1 {
		return nil
	}

	var f, err = os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_RDWR, os.ModePerm)
	if err != nil {
		return errors.Trace(err)
	}

	logrus.SetOutput(f)

	return nil
}

func setupLevel(level string) error {
	if len(level) < 1 {
		return nil
	}

	var lv, err = logrus.ParseLevel(level)
	if err != nil {
		return errors.Trace(err)
	}

	logrus.SetLevel(lv)

	return nil
}

// WarnStackf .
func WarnStackf(err error, fmt string, args ...interface{}) {
	WarnStack(errors.Annotatef(err, fmt, args...))
}

// WarnStack .
func WarnStack(err error) {
	Warnf(errors.Stack(err))
}

// Warnf .
func Warnf(fmt string, args ...interface{}) {
	logrus.Warnf(fmt, args...)
}

// ErrorStackf .
func ErrorStackf(err error, fmt string, args ...interface{}) {
	ErrorStack(errors.Annotatef(err, fmt, args...))
}

// ErrorStack .
func ErrorStack(err error) {
	Errorf(errors.Stack(err))
}

// Errorf .
func Errorf(fmt string, args ...interface{}) {
	logrus.Errorf(fmt, args...)
}

// Infof .
func Infof(fmt string, args ...interface{}) {
	logrus.Infof(fmt, args...)
}

// Debugf .
func Debugf(fmt string, args ...interface{}) {
	logrus.Debugf(fmt, args...)
}

// FatalStack .
func FatalStack(err error) {
	Fatalf(errors.Stack(err))
}

// Fatalf .
func Fatalf(fmt string, args ...interface{}) {
	logrus.Fatalf(fmt, args...)
}
