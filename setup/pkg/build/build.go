package build

import (
	"fmt"
	"os"
	"os/exec"
	path_util "path"
)

func build(output string) error {
	path, err := exec.LookPath("go")
	if err != nil {
		path = path_util.Join("mingw64", "lib", "go", "bin", "go")
	}
	cmd := exec.Command(path, "build", "-o", output)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("run go build with %q: %w", path, err)
	}
	return nil
}
