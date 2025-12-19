package cmd

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

const versionUrl = "https://raw.githubusercontent.com/malikbenkirane/groq-whisper/refs/heads/main/version.go"
const remoteBucket = "https://storage.googleapis.com/groq-whisper/"

func remoteVersion() (string, error) {
	var upstreamVersion string
	err := func() (err error) {
		resp, err := http.Get(versionUrl)
		if err != nil {
			return fmt.Errorf("get %q: %w", versionUrl, err)
		}
		defer func() {
			errClose := resp.Body.Close()
			if errClose != nil {
				err = fmt.Errorf("close response body: %w", err)
			}
		}()
		scan := bufio.NewScanner(resp.Body)
		for scan.Scan() {
			after, found := strings.CutPrefix(scan.Text(), "const version = \"")
			if !found {
				continue
			}
			version, found := strings.CutSuffix(after, "\"")
			if !found {
				continue
			}
			upstreamVersion = version
			return nil
		}
		return fmt.Errorf("unable to retrieve version from %q", versionUrl)
	}()
	if err != nil {
		return "", err
	}
	return upstreamVersion, nil
}

func upgrade(current string) (bool, string, error) {
	upstream, err := remoteVersion()
	if err != nil {
		return false, "", fmt.Errorf("read remote version: %w", err)
	}
	if current == upstream {
		return false, current, nil
	}
	if err := func() (err error) {
		var b bytes.Buffer
		upgrade := fmt.Sprintf("%sgroq-%s.exe", remoteBucket, upstream)
		resp, err := http.Get(upgrade)
		if err != nil {
			return fmt.Errorf("http get %q: %w", upgrade, err)
		}
		defer func() {
			err = deferCheck(resp.Body.Close(), err, "http response body close")
		}()
		_, err = io.Copy(&b, resp.Body)
		if err != nil {
			return fmt.Errorf("copy response body: %w", err)
		}
		exe, err := os.Executable()
		if err != nil {
			return fmt.Errorf("os executable: %w", err)
		}
		var oldExe string
		{
			noExt, found := strings.CutSuffix(exe, ".exe")
			if !found {
				return fmt.Errorf("expected exe file, found %q", exe)
			}
			oldExe = fmt.Sprintf("%s-%s.exe", noExt, current)
		}
		if err = os.Rename(exe, oldExe); err != nil {
			return fmt.Errorf("os rename %q to %q: %w", exe, oldExe, err)
		}
		f, err := os.Create(exe)
		if err != nil {
			return fmt.Errorf("os create %q: %w", exe, err)
		}
		defer func() {
			err = deferCheck(f.Close(), err, "close new %q", exe)
		}()
		_, err = f.Write(b.Bytes())
		if err != nil {
			return fmt.Errorf("write upgrade to %q: %w", exe, err)
		}
		return nil
	}(); err != nil {
		return false, "", err
	}
	return true, upstream, nil
}

func deferCheck(errDefer, err error, msg string, args ...any) error {
	if errDefer != nil && err != nil {
		return fmt.Errorf("%w then %s: %w",
			err, fmt.Sprintf(msg, args...), errDefer)
	}
	if errDefer != nil {
		return fmt.Errorf("%s: %w", msg, errDefer)
	}
	return err
}

func newCommandUpgrade(currentVersion string) *cobra.Command {
	return &cobra.Command{
		Use: "upgrade",
		RunE: func(cmd *cobra.Command, args []string) error {
			upgraded, newVersion, err := upgrade(currentVersion)
			if err != nil {
				return fmt.Errorf("upgrade: %w", err)
			}
			if upgraded {
				fmt.Println(currentVersion, "->", newVersion)
			} else {
				fmt.Print("Already up-to-date.")
			}
			return nil
		},
	}
}
