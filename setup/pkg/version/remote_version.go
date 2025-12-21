package version

import (
	"bufio"
	"fmt"
	"net/http"
	"strings"
)

func Get(url string) (string, error) {
	var upstreamVersion string
	err := func() (err error) {
		resp, err := http.Get(url)
		if err != nil {
			return fmt.Errorf("get %q: %w", url, err)
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
		return fmt.Errorf("unable to retrieve version from %q", url)
	}()
	if err != nil {
		return "", err
	}
	return upstreamVersion, nil
}
