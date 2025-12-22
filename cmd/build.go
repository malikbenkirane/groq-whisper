package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"

	"github.com/malikbenkirane/groq-whisper/setup/pkg/build"
	"github.com/malikbenkirane/groq-whisper/setup/pkg/gcloud"
	"github.com/malikbenkirane/groq-whisper/setup/pkg/version"
)

func newCommandDev() *cobra.Command {
	cmd := &cobra.Command{
		Use: "dev",
	}
	cmd.AddCommand(
		newCommandBuild(),
	)
	return cmd
}

func newCommandBuild() *cobra.Command {
	var push, gcloudLogin *bool
	var gitRemote, gcpBucket, gcpProject, outputGroq, outputSetup *string
	cmd := &cobra.Command{
		Use: "build",
		Long: `
Read docs/install.md first to setup you build environment.

If you are an expert in cross-compilation you should be able to make it on any
platform, but a the moment what was tested is documented in [docs/install.md].

When using --push option, make sure either [gcloud] is available through your
PATH environment variable (e.g. export PATH=$PATH:$HOME/google-cloud-sdk/bin).
		`,
		Args: func(cmd *cobra.Command, args []string) (err error) {
			if *push {
				for _, opt := range []*string{
					gcpBucket,
					gcpProject,
					gitRemote,
					outputGroq,
					outputSetup,
				} {
					*opt = strings.TrimSpace(*opt)
				}
				if len(*gcpProject) == 0 {
					*gcpProject, err = gcloud.DefaultGcloudProject()
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
				if len(*outputSetup) == 0 {
					return errors.New("--output-setup is expected")
				}
				if len(*outputGroq) == 0 {
					return errors.New("--output-groq is expected")
				}
				return nil
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			nameGroq, _ := strings.CutSuffix(*outputGroq, ".exe")
			nameSetup, _ := strings.CutSuffix(*outputSetup, ".exe")
			exeGroq, exeSetup :=
				version.Executable(nameGroq, version.Version),
				version.Executable(nameSetup, version.Version)
			if err = build.Build(".", exeGroq); err != nil {
				return fmt.Errorf("go build: %w", err)
			}
			if err = build.Build("./setup", exeSetup); err != nil {
				return fmt.Errorf("go build: %w", err)
			}
			fmt.Println("built", exeGroq, "and", exeSetup)
			if !*push {
				return nil
			}
			opts := []gcloud.BucketOption{
				gcloud.OptionBucketProject(*gcpProject),
			}
			if *gcloudLogin {
				opts = append(opts, gcloud.BucketWithLogin())
			}
			bucket := gcloud.NewBucket(*gcpBucket, opts...)
			for _, dst := range []string{exeGroq, exeSetup} {
				fmt.Println("pushing", dst)
				if err := bucket.Push(dst, dst); err != nil {
					return fmt.Errorf("push %q: %w", dst, err)
				}
				fmt.Println("pushed", dst)
			}
			return nil
		},
	}
	push = cmd.Flags().Bool("push", false, "push built version to upstream main and gcloud storage")
	gitRemote = cmd.Flags().String("git-remote", "", "use first git remote if not set")
	gcpBucket = cmd.Flags().String("gcp-bucket", "groq-whisper", "GCP storage bucket name")
	gcpProject = cmd.Flags().String("gcp-project", "", "GCP project name (use default project if not set)")
	gcloudLogin = cmd.Flags().Bool("gcloud-login", false, "prompt gcloud auth login")
	outputGroq = cmd.Flags().String("output-groq", "groq.exe", "go build output")
	outputSetup = cmd.Flags().String("output-setup", "groq-setup.exe", "go build setup output")
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
