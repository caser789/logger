package env

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"time"
)

func GetEnv() string {
	val, ok := os.LookupEnv("ENV")
	if !ok {
		val = "dev"
	}
	return val
}

func IsLive() bool {
	return GetEnv() == "live"
}

func IsSzK8S() bool {
	e, ok := os.LookupEnv("ORCHESTRATOR")
	if ok && e == "sz-kubernetes" {
		return true
	}
	return false
}

func IsSplitLog() bool {
	_, ok := os.LookupEnv("SPLIT_LOG")
	return ok
}

func GetFilePath(logDir, filename string) string {
	if logDir == "" {
		logDir = "./log"
	}
	path := filepath.Join(logDir, fmt.Sprintf("%s.log", filename))
	if IsSzK8S() {
		dir := ""
		podName, ok := os.LookupEnv("POD_NAME")
		if ok {
			dir = podName
		}
		if dir == "" {
			rand.Seed(time.Now().UnixNano() ^ int64(os.Getpid()))
			dir = fmt.Sprintf("%s-%d", time.Now().Format("20060102150405"), rand.Int()%1000)
		}
		path = filepath.Join(logDir, dir, fmt.Sprintf("%s.log", filename))
	}
	return path
}
