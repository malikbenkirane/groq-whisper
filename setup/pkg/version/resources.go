package version

import "fmt"

const Url = "https://raw.githubusercontent.com/malikbenkirane/groq-whisper/refs/heads/main/setup/pkg/version/version.go"

const (
	ExecutableGroq        = "groq-whisper"
	ExecutableGroqInstall = "groq-whisper-install"
)

func Executable(name, version string) string {
	if version == "" {
		return name + ".exe"
	}
	return fmt.Sprintf("%s-%s.exe", name, version)
}
