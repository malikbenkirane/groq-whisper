package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	path_util "path"
	"strings"

	"github.com/spf13/cobra"
)

func newCommandDev(version string) *cobra.Command {
	cmd := &cobra.Command{
		Use: "dev",
	}
	cmd.AddCommand(
		newCommandBuild(version),
	)
	return cmd
}

func newCommandBuild(version string) *cobra.Command {
	var push *bool
	var gitRemote, gcpBucket, gcpProject, output *string
	cmd := &cobra.Command{
		Use: "build",
		Long: `
Read docs/install.md first to setup you build environment.

If you are an expert in cross-compilation you should be able to make it on any
platform, but a the moment what was tested is documented in [docs/install.md].

When using --push option, make sure either [gcloud] is available through your
PATH setting or google-cloud-sdk is in your home directory. 
		`,
		Args: func(cmd *cobra.Command, args []string) (err error) {
			if *push {
				for _, opt := range []*string{gcpBucket, gcpProject, gitRemote, output} {
					*opt = strings.TrimSpace(*opt)
				}
				if len(*gcpProject) == 0 {
					*gcpProject, err = defaultGcloudProject()
					if err != nil {
						return fmt.Errorf("gcloud config get project: %w", err)
					}
				}
				if len(*gitRemote) == 0 {
					*gitRemote, err = defaultGitRemote()
					if err != nil {
						return fmt.Errorf("git remote: %w", err)
					}
				}
				if len(*gcpBucket) == 0 {
					return errors.New("--gcp-bucket is expected with --push")
				}
				if len(*output) == 0 {
					return errors.New("--output is expected")
				}
				return nil
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			if err = build(*output); err != nil {
				return fmt.Errorf("go build: %w", err)
			}
			fmt.Println("built", version)
			if !*push {
				return nil
			}
			if err = gcloud(nil, "auth", "login"); err != nil {
				return fmt.Errorf("gcloud auth login: %w", err)
			}
			dst := bucketFile("gs://"+*gcpBucket, version)
			if err = gcloud(nil, "storage", "cp", "groq.exe", dst); err != nil {
				return fmt.Errorf("copy %q to %q: %w", *output, dst, err)
			}
			fmt.Println("new version available at", bucketFile(
				strings.Replace(remoteBucket, "groq-whisper", *gcpBucket, 1), version))
			return nil

		},
	}
	push = cmd.Flags().Bool("push", false, "push built version to upstream main and gcloud storage")
	gitRemote = cmd.Flags().String("git-remote", "", "use first git remote if not set")
	gcpBucket = cmd.Flags().String("gcp-bucket", "groq-whisper", "GCP storage bucket name")
	gcpProject = cmd.Flags().String("gcp-project", "", "GCP project name (use default project if not set)")
	output = cmd.Flags().String("output", "groq.exe", "go build output")
	return cmd
}

func git(out io.Writer, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Stdout = out
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("run git %v: %w", args, err)
	}
	return nil
}

func defaultGitRemote() (remote string, err error) {
	var b bytes.Buffer
	if err = git(&b, "remote"); err != nil {
		return "", fmt.Errorf("git remote: %w", err)
	}
	return strings.TrimSpace(b.String()), nil
}

func defaultGcloudProject() (project string, err error) {
	var b bytes.Buffer
	if err = gcloud(&b, "config", "get", "project"); err != nil {
		return "", fmt.Errorf("gcloud config get project: %w", err)
	}
	return strings.TrimSpace(b.String()), nil
}

func gcloud(out io.Writer, args ...string) error {
	path, err := exec.LookPath("gcloud")
	if err != nil {
		home := "/" + path_util.Join("home", os.Getenv("USER"))
		path = path_util.Join(home, "google-cloud-sdk", "bin", "gcloud")
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
