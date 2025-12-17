# Installation

## Windows 11 (64bit)

**Requirements:**
- soundcard and driver installed
- microphone (integrated or wired)
- wifi card and driver installed

**Portaudio**

- Download `libportaudio64bit.dll` from [PortAudio Binaries](
    https://github.com/spatialaudio/portaudio-binaries
): Click on "<> Code" then "Download Zip".
- Exctract and copy to `C:\Windows\System32\libportaudio.dll`.

**Ffmpeg**

- Download [ffmpeg-release-full.7z](
https://www.gyan.dev/ffmpeg/builds/ffmpeg-release-full.7z
) from [gyan builds](
https://www.gyan.dev/ffmpeg/builds/
)
- Alternatively you can download from [groq-whisper cloud storage](
https://storage.googleapis.com/groq-whisper/ffmpeg-8.0.1-full_build.7z
)
- Exctract to `C:\Program Files (x86)\ffmpeg'

**Groq**

- Download [groq.exe](
https://storage.googleapis.com/groq-whisper/groq.exe
)
- Copy to `C:\Program Files (x86)\groq-whisper\bin`

**Set paths**

Open `cmd`

```
setx path "%PATH%;C:\Program Files (x86)\ffmpeg\bin"
setx path "%PATH%;C:\Program Files (x86)\groq-whisper\bin"
```

**Runtime**

Get API key (or contact your API support) and write to `C:\Users\user\key.txt` (for example use Notepad).

Open new `cmd` session
```
groq w
```

Open new `cmd` session
```
groq r
```
