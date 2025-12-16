**Input:**

**Scope:** `feat`

**Git Log:**
```
commit bcfbdb1dc6f9294f0c49e8945fa3a3830e68b571
Author: Malik Benkirane <mbenkirane@internetworks.blog>
Date:   Tue Dec 16 17:49:01 2025 +0100

    root.dev-202512161746.53
    
    feat: add CLI for audio recording and file watching
    
    Implements a CLI tool with two primary commands for audio
    processing workflows: continuous audio recording with automatic
    chunking, and filesystem watching for new audio files.
    
    The `record` command captures audio using PortAudio at
    configurable sample rates (default 16kHz), automatically splits
    recordings into 10-second chunks, and converts raw PCM data to
    MP3 format using ffmpeg. Audio data flows through channels from
    the streaming goroutine to a consumer goroutine that handles
    file I/O and conversion asynchronously.
    
    The `sidecar` command uses fsnotify to watch the current
    directory for newly created MP3 files, enabling automated
    processing pipelines that can trigger on recording completion.
    
    Both commands support graceful shutdown via SIGINT/SIGTERM,
    properly canceling contexts and closing resources. Structured
    logging with zap provides observability for debugging audio
    capture and file conversion issues.
    
    Dependencies added:
    - gordonklaus/portaudio for cross-platform audio capture
    - fsnotify for filesystem event monitoring
    - spf13/cobra for CLI framework
    - uber/zap for structured logging
    
    Known limitations:
    - Sample rate flag uses integer type but converts to float64
    - Channel count hardcoded to stereo (2) in ffmpeg conversion
    - No error recovery for failed chunk conversions
    - ffmpeg output suppressed (stdout/stderr commented out)
    - Consumer goroutine uses polling with 1s sleep instead of
      blocking channel reads
    - No cleanup of partial files on shutdown
    - Missing input validation for negative sample rates
```

**Diff:**
```diff
diff --git a/.gitignore b/.gitignore
new file mode 100644
index 0000000..4dcd5a6
--- /dev/null
+++ b/.gitignore
@@ -0,0 +1 @@
+key.txt
diff --git a/cmd/sidecar.go b/cmd/sidecar.go
index f7504b4..a94ff23 100644
--- a/cmd/sidecar.go
+++ b/cmd/sidecar.go
@@ -1,10 +1,16 @@
 package cmd
 
 import (
+	"bytes"
 	"context"
+	"encoding/json"
 	"fmt"
+	"io"
+	"mime/multipart"
+	"net/http"
 	"os"
 	"os/signal"
+	"path/filepath"
 	"strings"
 	"syscall"
 
@@ -13,10 +19,125 @@ import (
 	"go.uber.org/zap"
 )
 
+func groqKey(file string) (string, error) {
+	var b bytes.Buffer
+	f, err := os.Open(file)
+	if err != nil {
+		return "", fmt.Errorf("open %q: %w", file, err)
+	}
+	if _, err = io.Copy(&b, f); err != nil {
+		return "", fmt.Errorf("copy key: %w", err)
+	}
+	return strings.TrimSpace(b.String()), nil
+}
+
+type groqClient struct {
+	key  string
+	log  *zap.Logger
+	tx   groqTx
+	lang string // iso-693-1
+}
+
+type groqTx struct {
+	Text  string `json:"text"`
+	XGroq groqX  `json:"x_groq"`
+}
+
+type groqX struct {
+	Id string `json:"id"`
+}
+
+const gtxUrl = "https://api.groq.com/openai/v1/audio/transcriptions"
+
+func (gc groqClient) newRequest(audio, model string) (req *http.Request, err error) {
+	gc.log = gc.log.Named("gc request builder")
+	defer func() {
+		if err != nil {
+			gc.log.Error("groq client failed to build request", zap.Error(err))
+		}
+	}()
+	var body bytes.Buffer
+	writer := multipart.NewWriter(&body)
+	base := filepath.Base(audio)
+	gc.log.Debug("create form file", zap.String("base", base))
+	part, err := writer.CreateFormFile("file", base)
+	if err != nil {
+		return nil, fmt.Errorf("create form file: %w", err)
+	}
+	gc.log.Debug("opening audio", zap.String("file", audio))
+	stat, err := os.Stat(audio)
+	if err != nil {
+		return nil, fmt.Errorf("os stat %q: %w", audio, err)
+	}
+	f, err := os.Open(audio)
+	if err != nil {
+		return nil, fmt.Errorf("open %q: %w", audio, err)
+	}
+	defer func() {
+		err = f.Close()
+	}()
+	n, err := io.Copy(part, f)
+	if err != nil {
+		return nil, fmt.Errorf("copy file part: %w", err)
+	}
+	gc.log.Debug("form file written", zap.Int64("copied", n), zap.Int64("stated", stat.Size()))
+	gc.log.Debug("write model field", zap.String("model", model))
+	for k, v := range map[string]string{
+		"language": gc.lang,
+		"model":    model,
+	} {
+		if err = writer.WriteField(k, v); err != nil {
+			return nil, fmt.Errorf("write field %q: %w", k, err)
+		}
+	}
+	if err := writer.Close(); err != nil {
+		return nil, fmt.Errorf("multipart writer close: %w", err)
+	}
+	gc.log.Debug("prepare post request", zap.String("url", gtxUrl))
+	req, err = http.NewRequest(http.MethodPost, gtxUrl, &body)
+	if err != nil {
+		return nil, fmt.Errorf("http new request: %w", err)
+	}
+	req.Header.Set("Content-Type", writer.FormDataContentType())
+	req.Header.Add("Authorization", "bearer "+gc.key)
+	return
+}
+
+func (gc *groqClient) post(audio, model string) (err error) {
+	defer func() {
+		if err != nil {
+			gc.log.Error("groq client failed to post", zap.Error(err))
+		}
+	}()
+	gc.log.Debug("new request",
+		zap.String("audio", audio), zap.String("model", model))
+	req, err := gc.newRequest(audio, model)
+	if err != nil {
+		return fmt.Errorf("gc new request: %w", err)
+	}
+	resp, err := http.DefaultClient.Do(req)
+	if err != nil {
+		return fmt.Errorf("do request: %w", err)
+	}
+	if resp.StatusCode != http.StatusOK {
+		return fmt.Errorf("response: %q", resp.Status)
+	}
+	defer func() {
+		err = resp.Body.Close()
+	}()
+	var tx groqTx
+	if err = json.NewDecoder(resp.Body).Decode(&tx); err != nil {
+		return fmt.Errorf("decode response body: %w", err)
+	}
+	gc.tx = tx
+	return nil
+}
+
 func newCommandSidecar(log *zap.Logger) *cobra.Command {
-	return &cobra.Command{
+	var dry *bool
+	cmd := &cobra.Command{
 		Use:     "sidecar",
-		Aliases: []string{"watch", "w"},
+		Aliases: []string{"watch", "w", "s"},
 		RunE: func(cmd *cobra.Command, args []string) (err error) {
 			quit := make(chan os.Signal, 1)
 			signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
@@ -27,6 +148,22 @@ func newCommandSidecar(log *zap.Logger) *cobra.Command {
 				}
 			}()
 
+			var gc groqClient
+			{
+				key, err := groqKey("key.txt")
+				if err != nil {
+					return fmt.Errorf("groq key: %w", err)
+				}
+
+				gc = groqClient{
+					key:  key,
+					log:  log,
+					lang: "fr",
+				}
+				log.Debug("new client",
+					zap.String("lang", gc.lang), zap.String("url", gtxUrl))
+			}
+
 			w, err := fsnotify.NewWatcher()
 			if err != nil {
 				return fmt.Errorf("fsnotify new watcher: %w", err)
@@ -37,15 +174,42 @@ func newCommandSidecar(log *zap.Logger) *cobra.Command {
 
 			ctx, cancel := context.WithCancel(cmd.Context())
 			go func(ctx context.Context) {
+				counter := make(map[string]struct{})
+			loop:
 				for {
 					select {
 					case event, ok := <-w.Events:
 						if !ok {
 							return
 						}
-						if event.Has(fsnotify.Create) &&
-							strings.HasSuffix(event.Name, ".mp3") {
-							log.Info("notify write", zap.String("file", event.Name))
+						log.Debug("fsnotify event",
+							zap.String("event", event.Op.String()),
+							zap.String("name", event.Name),
+						)
+						if event.Has(fsnotify.Write) &&
+							strings.HasSuffix(event.Name, ".flac") {
+							if _, ok := counter[event.Name]; ok {
+								continue loop
+							}
+							counter[event.Name] = struct{}{}
+							stat, err := os.Stat(event.Name)
+							if err != nil {
+								log.Error("unable to stat sample", zap.String("file", event.Name))
+								continue loop
+							}
+							log.Info("found new sample",
+								zap.String("file", event.Name), zap.Int64("size", stat.Size()))
+							if !*dry {
+								if err = gc.post(event.Name, "whisper-large-v3"); err != nil {
+									log.Error("gc post failed", zap.Error(err))
+									continue loop
+								}
+								e := json.NewEncoder(os.Stdout)
+								e.SetIndent("", "  ")
+								if err := e.Encode(gc.tx); err != nil {
+									log.Error("unable to encode gc response", zap.Error(err))
+								}
+							}
 						}
 					case err, ok := <-w.Errors:
 						if !ok {
@@ -67,4 +231,6 @@ func newCommandSidecar(log *zap.Logger) *cobra.Command {
 			return nil
 		},
 	}
+	dry = cmd.Flags().Bool("dry", false, "don't post to groq")
+	return cmd
 }
diff --git a/internal/sampler/sampler.go b/internal/sampler/sampler.go
index cbc51b7..9bea3ec 100644
--- a/internal/sampler/sampler.go
+++ b/internal/sampler/sampler.go
@@ -79,9 +79,7 @@ func (s sampler) consume(ctx context.Context) (err error) {
 			if err = f.Close(); err != nil {
 				return fmt.Errorf("close chunk %q: %w", pr, err)
 			}
-			p3 := outPath(c.ts, outMp3)
-			s.log.Info("converting new chunk", zap.String("path", p3))
-			err = convert(pr, p3, strconv.Itoa(int(s.sample)/2))
+			err = convert(c.ts, strconv.Itoa(int(s.sample)))
 			if err != nil {
 				return fmt.Errorf("convert: %w", err)
 			}
@@ -156,16 +154,23 @@ loop:
 type outType int
 
 const (
-	outMp3 outType = iota
+	outFlac outType = iota
+	outMp3
 	outRaw
 )
 
-func outPath(ts time.Time, t outType) string {
-	ext := "mp3"
-	if t == outRaw {
-		ext = "raw"
+func (t outType) String() string {
+	if t == outMp3 {
+		return "mp3"
+	}
+	if t == outFlac {
+		return "flac"
 	}
-	return fmt.Sprintf("%s.%s", ts.Format("20060102150405"), ext)
+	return "raw"
+}
+
+func outPath(ts time.Time, t outType) string {
+	return fmt.Sprintf("%s.%s", ts.Format("20060102150405"), t)
 }
 
 type chunk struct {
@@ -173,26 +178,48 @@ type chunk struct {
 	ts  time.Time
 }
 
-func convert(raw, mp3, freq string) (err error) {
-	_, err = os.Stat(mp3)
-	if err == nil {
-		if err = os.Remove(mp3); err != nil {
-			return fmt.Errorf("remove %q: %w", mp3, err)
-		}
-	}
+func convert(ts time.Time, freq string) (err error) {
+	raw, mp3 := outPath(ts, outRaw), outPath(ts, outMp3)
 	args := []string{
 		"-f", "s16le",
 		"-ar", freq,
-		"-ac", "2",
+		"-ac", "1",
 		"-i", raw, mp3,
 	}
-	fmt.Println(append([]string{"ffmpeg"}, args...))
-	cmd := exec.Command("ffmpeg", args...)
-	// cmd.Stdout = os.Stdout
-	// cmd.Stderr = os.Stderr
-	err = cmd.Run()
-	if err != nil {
-		return fmt.Errorf("exec ffmpeg: %w", err)
+	if err = ffmpeg(mp3, args...); err != nil {
+		return fmt.Errorf("ffmpeg %q->%q: %w", raw, mp3, err)
+	}
+	flac := outPath(ts, outFlac)
+	args = []string{
+		"-i", mp3,
+		"-ar", freq,
+		"-ac", "1",
+		"-map", "0:a",
+		"-c:a", "flac",
+		flac,
+	}
+	if err = ffmpeg(flac, args...); err != nil {
+		return fmt.Errorf("ffmpeg %q->%q: %w", mp3, flac, err)
 	}
 	return nil
 }
+
+func ffmpeg(rm string, args ...string) (err error) {
+	_, err = os.Stat(rm)
+	if err == nil {
+		if err = os.Remove(rm); err != nil {
+			return fmt.Errorf("remove %q: %w", rm, err)
+		}
+	}
+	fmt.Println(append([]string{"ffmpeg"}, args...))
+	{
+		cmd := exec.Command("ffmpeg", args...)
+		//cmd.Stdout = os.Stdout
+		//cmd.Stderr = os.Stderr
+		err = cmd.Run()
+		if err != nil {
+			return fmt.Errorf("exec ffmpeg: %w", err)
+		}
+		return nil
+	}
+}


----------------

git status -s

A  .gitignore
M  cmd/sidecar.go
M  internal/sampler/sampler.go

```


-------------------------------------------------
-------- NEXT STEPS AND WORK IN PROGRESS --------
-------------------------------------------------


**WIP Context:**

*WIP Known Issues:*
- using a single 10s frame may not produce enough context for the transcription

*WIP Planned Improvements:*
- obtain keys from a server to improve security
- retry logic when groq client post fails
- working on multiple time frames e.g. 10s and 30s may help reconciliation of the speech
- save the text to test chunks

*WIP Planned Fixes:*
- remove raw and flac files after consumed
- first consumer is ffmpeg to flag => remove mp3 at this stage (in sampler covert func)
- second consumer is the sidecar
