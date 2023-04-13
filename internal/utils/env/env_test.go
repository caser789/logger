package env

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// test dir nil
func TestGetFilePath(t *testing.T) {
	var dir string
	path := GetFilePath(dir, "server")
	fmt.Println(path)
	assert.Equal(t, "log/server.log", path)
}

// test dir nil and sz k8s
func TestGetFilePathWithSzK8s(t *testing.T) {
	var dir string
	os.Setenv("ORCHESTRATOR", "sz-kubernetes")

	path := GetFilePath(dir, "server")
	fmt.Println(path)
	ss := strings.Split(path, "/")
	assert.Equal(t, "log", ss[0])
	assert.Equal(t, "server.log", ss[2])

	os.Setenv("POD_NAME", "test-podname")
	path = GetFilePath(dir, "server")
	fmt.Println(path)
	assert.Equal(t, "log/test-podname/server.log", path)
}

func TestGetFilePathWithDir(t *testing.T) {
	dir := "./test"
	path := GetFilePath(dir, "server")
	fmt.Println(path)
	assert.Equal(t, "test/server.log", path)
}

func TestGetFilePathWithDirSzK8s(t *testing.T) {
	dir := "./test"
	os.Setenv("ORCHESTRATOR", "sz-kubernetes")

	path := GetFilePath(dir, "server")
	fmt.Println(path)
	ss := strings.Split(path, "/")
	assert.Equal(t, "test", ss[0])
	assert.Equal(t, "server.log", ss[2])

	os.Setenv("POD_NAME", "test-podname")
	path = GetFilePath(dir, "server")
	fmt.Println(path)
	assert.Equal(t, "test/test-podname/server.log", path)
}
