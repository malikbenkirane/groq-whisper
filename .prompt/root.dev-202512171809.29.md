**Input:**

**Scope:** `docs`

**Git Log:** *None*

**Diff:**
```diff
diff --git a/README.md b/README.md
new file mode 100644
index 0000000..5dad835
--- /dev/null
+++ b/README.md
@@ -0,0 +1,7 @@
+# Groq
+
+Brainstorming Whisperer
+
+## Installation
+
+See **[docs/install.md]**
diff --git a/docs/install.md b/docs/install.md
new file mode 100644
index 0000000..5904748
--- /dev/null
+++ b/docs/install.md
@@ -0,0 +1,57 @@
+# Installation
+
+## Windows 11 (64bit)
+
+**Requirements:**
+- soundcard and driver installed
+- microphone (integrated or wired)
+- wifi card and driver installed
+
+**Portaudio**
+
+- Download `libportaudio64bit.dll` from [PortAudio Binaries](
+    https://github.com/spatialaudio/portaudio-binaries
+): Click on "<> Code" then "Download Zip".
+- Exctract and copy to `C:\Windows\System32\libportaudio.dll`.
+
+**Ffmpeg**
+
+- Download [ffmpeg-release-full.7z](
+https://www.gyan.dev/ffmpeg/builds/ffmpeg-release-full.7z
+) from [gyan builds](
+https://www.gyan.dev/ffmpeg/builds/
+)
+- Alternatively you can download from [groq-whisper cloud storage](
+https://storage.googleapis.com/groq-whisper/ffmpeg-8.0.1-full_build.7z
+)
+- Exctract to `C:\Program Files (x86)\ffmpeg'
+
+**Groq**
+
+- Download [groq.exe](
+https://storage.googleapis.com/groq-whisper/groq.exe
+)
+- Copy to `C:\Program Files (x86)\groq-whisper\bin`
+
+**Set paths**
+
+Open `cmd`
+
+```
+setx path "%PATH%;C:\Program Files (x86)\ffmpeg\bin"
+setx path "%PATH%;C:\Program Files (x86)\groq-whisper\bin"
+```
+
+**Runtime**
+
+Get API key (or contact your API support) and write to `C:\Users\user\key.txt` (for example use Notepad).
+
+Open new `cmd` session
+```
+groq w
+```
+
+Open new `cmd` session
+```
+groq r
+```


----------------

git status -s

A  README.md
A  docs/install.md

```


-------------------------------------------------
-------- NEXT STEPS AND WORK IN PROGRESS --------
-------------------------------------------------


**WIP Context:**

*Needs Decisions:*
- is automatisation necessary
- groq upgrade automatisation

*Ideas:*
- improve details for low level IT technicians
- use LLM to split the steps for any regular windows user without any knowldge of system files, e.g say extract and rename file/folder instead of just saying extract to

*Known Issues:*
- security wise key.txt is not protected so we need to be careful about key management
