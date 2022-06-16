//go:build !(linux || freebsd)
// +build !linux,!freebsd

package run

import (
	"github.com/sirupsen/logrus"
)

func AddSyslogHooks() {
	logrus.Warn("Not on linux, cannot add syslog hooks")
}
