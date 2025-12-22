package build

import (
	"fmt"
	"os"
	"os/exec"
	path_util "path"
)

func Build(path, output string) error {
	gopath, err := exec.LookPath("go")
	if err != nil {
		gopath = path_util.Join("mingw64", "lib", "go", "bin", "go")
	}
	if err := os.Chdir(path); err != nil {
		return fmt.Errorf("chdir %q: %w", path, err)
	}
	cmd := exec.Command(gopath, "build", "-o", output)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("run go build with %q: %w", gopath, err)
	}
	return nil
}
