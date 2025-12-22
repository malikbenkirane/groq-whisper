# Installation

## Windows 11 (64bit)

**Requirements:**
- soundcard and driver installed
- microphone (integrated or wired)
- wifi card and driver installed

You should be able to download [groq-setup-v0.7.0.zip](
https://storage.googleapis.com/groq-whisper/groq-setup-v0.7.3.zip
) to take care of all the steps mentioned below.

1. extract the archive to `C:\Windows\System32`

2. Open `cmd` as an administrator.
   ```bash
   cd groq
   ```

3. Install dependencies
   ```bash
   groq-setup-0.7.3.exe d
   ```

4. Install upgrades
   ```
   groq-setup-0.7.3.exe i
   ```

Each time you want to upgrade you can either run
`groq-setup.exe install` or `groq.exe upgrade`.

---

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
- Alternatively you can download from [cloud storage](
https://storage.googleapis.com/groq-whisper/ffmpeg-8.0.1-full_build.7z
) (much faster!)
- Exctract to `C:\Program Files (x86)\ffmpeg`

**Groq**

- Download [groq.exe](
https://storage.googleapis.com/groq-whisper/groq-0.2.2.exe
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
