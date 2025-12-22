package dep

const portAudioDllUrl = "https://raw.githubusercontent.com/spatialaudio/portaudio-binaries/refs/heads/master/libportaudio64bit.dll"

type PortAudioSrc struct {
	Dll64Url string
}

func DefaultPortAudioSrc() PortAudioSrc {
	return PortAudioSrc{Dll64Url: portAudioDllUrl}
}
