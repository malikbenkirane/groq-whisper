package gcloud

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

func gcloud(out io.Writer, args ...string) error {
	path, err := exec.LookPath("gcloud")
	if err != nil {
		return fmt.Errorf("look path gcloud: %w", err)
	}
	cmd := exec.Command(path, args...)
	cmd.Stdout = os.Stdout
	if out != nil {
		cmd.Stdout = out
	}
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("run %q: %w", path, err)
	}
	return nil
}

func DefaultGcloudProject() (project string, err error) {
	var b bytes.Buffer
	if err = gcloud(&b, "config", "get", "project"); err != nil {
		return "", fmt.Errorf("gcloud config get project: %w", err)
	}
	return strings.TrimSpace(b.String()), nil
}
