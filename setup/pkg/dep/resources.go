package dep

import (
	"fmt"
	"os"
	"path"
	"strings"
)

const portAudioDllUrl = "https://raw.githubusercontent.com/spatialaudio/portaudio-binaries/refs/heads/master/libportaudio64bit.dll"

type PortAudioSrc struct {
	Dll64Url string
}

func DefaultPortAudioSrc() PortAudioSrc {
	return PortAudioSrc{Dll64Url: portAudioDllUrl}
}

func DefaultPortAudioDst() (*PortAudioDst, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("user home dir: %w", err)
	}
	parts := strings.Split(home, ":")
	if len(parts) != 2 {
		return nil, fmt.Errorf("unable to extract volume from %q", home)
	}
	volume := parts[0] + ":"
	return &PortAudioDst{
		path: strings.ReplaceAll(
			path.Join(volume, "Windows", "System32", "libportaudio.dll"),
			"/", "\\",
		),
	}, nil
}

type PortAudioDst struct {
	path string
}
