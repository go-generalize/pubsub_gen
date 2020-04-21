// +build emulator

package gogen

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func execTest(t *testing.T) {
	t.Helper()

	b, err := exec.Command("go", "test", "./tests", "-v", "-tags", "internal").CombinedOutput()

	if err != nil {
		t.Fatalf("go test failed: %+v(%s)", err, string(b))
	}
}

func TestGenerator(t *testing.T) {
	root, err := os.Getwd()

	if err != nil {
		t.Fatalf("failed to getwd: %+v", err)
	}

	if err := os.Chdir(filepath.Join(root, "testfiles/a")); err != nil {
		t.Fatalf("chdir failed: %+v", err)
	}

	if err := runPubSub("Task", "task-topic"); err != nil {
		t.Fatalf("failed to generate for testfiles/a: %+v", err)
	}

	execTest(t)
}
