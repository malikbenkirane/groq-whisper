**Input:**

**Scope:** *None*

**Git Log:** *None*

**Diff:**
```diff
diff --git a/cmd/record.go b/cmd/record.go
new file mode 100644
index 0000000..685e1ef
--- /dev/null
+++ b/cmd/record.go
@@ -0,0 +1,37 @@
+package cmd
+
+import (
+	"context"
+	"os"
+	"os/signal"
+	"syscall"
+	"time"
+
+	"groq/internal/sampler"
+
+	"github.com/spf13/cobra"
+	"go.uber.org/zap"
+)
+
+func newCommandRecord(log *zap.Logger) *cobra.Command {
+	var freq *int
+	cmd := &cobra.Command{
+		Use:     "record",
+		Aliases: []string{"rec", "r"},
+		RunE: func(cmd *cobra.Command, args []string) (err error) {
+			quit := make(chan os.Signal, 1)
+			signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
+
+			ctx, cancel := context.WithCancel(cmd.Context())
+			s := sampler.New(log, float64(*freq), time.Duration(time.Second*10))
+			go s.Sample(ctx)
+
+			<-quit
+			cancel()
+
+			return nil
+		},
+	}
+	freq = cmd.Flags().IntP("freq", "f", 16000, "sample rate")
+	return cmd
+}
diff --git a/cmd/root.go b/cmd/root.go
new file mode 100644
index 0000000..34e2bef
--- /dev/null
+++ b/cmd/root.go
@@ -0,0 +1,21 @@
+package cmd
+
+import (
+	"github.com/spf13/cobra"
+	"go.uber.org/zap"
+)
+
+func NewCLI() *cobra.Command {
+	cmd := &cobra.Command{
+		Use: "groq",
+		RunE: func(cmd *cobra.Command, args []string) error {
+			return cmd.Help()
+		},
+	}
+	log, err := zap.NewDevelopment()
+	if err != nil {
+		panic(err)
+	}
+	cmd.AddCommand(newCommandRecord(log), newCommandSidecar(log))
+	return cmd
+}
diff --git a/cmd/sidecar.go b/cmd/sidecar.go
new file mode 100644
index 0000000..f7504b4
--- /dev/null
+++ b/cmd/sidecar.go
@@ -0,0 +1,70 @@
+package cmd
+
+import (
+	"context"
+	"fmt"
+	"os"
+	"os/signal"
+	"strings"
+	"syscall"
+
+	"github.com/fsnotify/fsnotify"
+	"github.com/spf13/cobra"
+	"go.uber.org/zap"
+)
+
+func newCommandSidecar(log *zap.Logger) *cobra.Command {
+	return &cobra.Command{
+		Use:     "sidecar",
+		Aliases: []string{"watch", "w"},
+		RunE: func(cmd *cobra.Command, args []string) (err error) {
+			quit := make(chan os.Signal, 1)
+			signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
+
+			defer func() {
+				if err != nil {
+					log.Error("sidecar failed", zap.Error(err))
+				}
+			}()
+
+			w, err := fsnotify.NewWatcher()
+			if err != nil {
+				return fmt.Errorf("fsnotify new watcher: %w", err)
+			}
+			defer func() {
+				err = w.Close()
+			}()
+
+			ctx, cancel := context.WithCancel(cmd.Context())
+			go func(ctx context.Context) {
+				for {
+					select {
+					case event, ok := <-w.Events:
+						if !ok {
+							return
+						}
+						if event.Has(fsnotify.Create) &&
+							strings.HasSuffix(event.Name, ".mp3") {
+							log.Info("notify write", zap.String("file", event.Name))
+						}
+					case err, ok := <-w.Errors:
+						if !ok {
+							return
+						}
+						log.Error("fsnotify", zap.Error(err))
+					}
+				}
+			}(ctx)
+
+			if err = w.Add("."); err != nil {
+				cancel()
+				return fmt.Errorf("fsnotify add cwd: %w", err)
+			}
+
+			<-quit
+			cancel()
+
+			return nil
+		},
+	}
+}
diff --git a/go.mod b/go.mod
new file mode 100644
index 0000000..e366420
--- /dev/null
+++ b/go.mod
@@ -0,0 +1,17 @@
+module groq
+
+go 1.25.4
+
+require (
+	github.com/fsnotify/fsnotify v1.9.0
+	github.com/gordonklaus/portaudio v0.0.0-20250206071425-98a94950218b
+	github.com/spf13/cobra v1.10.2
+	go.uber.org/zap v1.27.1
+)
+
+require (
+	github.com/inconshreveable/mousetrap v1.1.0 // indirect
+	github.com/spf13/pflag v1.0.9 // indirect
+	go.uber.org/multierr v1.10.0 // indirect
+	golang.org/x/sys v0.13.0 // indirect
+)
diff --git a/go.sum b/go.sum
new file mode 100644
index 0000000..ffd6b55
--- /dev/null
+++ b/go.sum
@@ -0,0 +1,30 @@
+github.com/cpuguy83/go-md2man/v2 v2.0.6/go.mod h1:oOW0eioCTA6cOiMLiUPZOpcVxMig6NIQQ7OS05n1F4g=
+github.com/davecgh/go-spew v1.1.1 h1:vj9j/u1bqnvCEfJOwUhtlOARqs3+rkHYY13jYWTU97c=
+github.com/davecgh/go-spew v1.1.1/go.mod h1:J7Y8YcW2NihsgmVo/mv3lAwl/skON4iLHjSsI+c5H38=
+github.com/fsnotify/fsnotify v1.9.0 h1:2Ml+OJNzbYCTzsxtv8vKSFD9PbJjmhYF14k/jKC7S9k=
+github.com/fsnotify/fsnotify v1.9.0/go.mod h1:8jBTzvmWwFyi3Pb8djgCCO5IBqzKJ/Jwo8TRcHyHii0=
+github.com/gordonklaus/portaudio v0.0.0-20250206071425-98a94950218b h1:WEuQWBxelOGHA6z9lABqaMLMrfwVyMdN3UgRLT+YUPo=
+github.com/gordonklaus/portaudio v0.0.0-20250206071425-98a94950218b/go.mod h1:esZFQEUwqC+l76f2R8bIWSwXMaPbp79PppwZ1eJhFco=
+github.com/inconshreveable/mousetrap v1.1.0 h1:wN+x4NVGpMsO7ErUn/mUI3vEoE6Jt13X2s0bqwp9tc8=
+github.com/inconshreveable/mousetrap v1.1.0/go.mod h1:vpF70FUmC8bwa3OWnCshd2FqLfsEA9PFc4w1p2J65bw=
+github.com/pmezard/go-difflib v1.0.0 h1:4DBwDE0NGyQoBHbLQYPwSUPoCMWR5BEzIk/f1lZbAQM=
+github.com/pmezard/go-difflib v1.0.0/go.mod h1:iKH77koFhYxTK1pcRnkKkqfTogsbg7gZNVY4sRDYZ/4=
+github.com/russross/blackfriday/v2 v2.1.0/go.mod h1:+Rmxgy9KzJVeS9/2gXHxylqXiyQDYRxCVz55jmeOWTM=
+github.com/spf13/cobra v1.10.2 h1:DMTTonx5m65Ic0GOoRY2c16WCbHxOOw6xxezuLaBpcU=
+github.com/spf13/cobra v1.10.2/go.mod h1:7C1pvHqHw5A4vrJfjNwvOdzYu0Gml16OCs2GRiTUUS4=
+github.com/spf13/pflag v1.0.9 h1:9exaQaMOCwffKiiiYk6/BndUBv+iRViNW+4lEMi0PvY=
+github.com/spf13/pflag v1.0.9/go.mod h1:McXfInJRrz4CZXVZOBLb0bTZqETkiAhM9Iw0y3An2Bg=
+github.com/stretchr/testify v1.8.1 h1:w7B6lhMri9wdJUVmEZPGGhZzrYTPvgJArz7wNPgYKsk=
+github.com/stretchr/testify v1.8.1/go.mod h1:w2LPCIKwWwSfY2zedu0+kehJoqGctiVI29o6fzry7u4=
+go.uber.org/goleak v1.3.0 h1:2K3zAYmnTNqV73imy9J1T3WC+gmCePx2hEGkimedGto=
+go.uber.org/goleak v1.3.0/go.mod h1:CoHD4mav9JJNrW/WLlf7HGZPjdw8EucARQHekz1X6bE=
+go.uber.org/multierr v1.10.0 h1:S0h4aNzvfcFsC3dRF1jLoaov7oRaKqRGC/pUEJ2yvPQ=
+go.uber.org/multierr v1.10.0/go.mod h1:20+QtiLqy0Nd6FdQB9TLXag12DsQkrbs3htMFfDN80Y=
+go.uber.org/zap v1.27.1 h1:08RqriUEv8+ArZRYSTXy1LeBScaMpVSTBhCeaZYfMYc=
+go.uber.org/zap v1.27.1/go.mod h1:GB2qFLM7cTU87MWRP2mPIjqfIDnGu+VIO4V/SdhGo2E=
+go.yaml.in/yaml/v3 v3.0.4/go.mod h1:DhzuOOF2ATzADvBadXxruRBLzYTpT36CKvDb3+aBEFg=
+golang.org/x/sys v0.13.0 h1:Af8nKPmuFypiUBjVoU9V20FiaFXOcuZI21p0ycVYYGE=
+golang.org/x/sys v0.13.0/go.mod h1:oPkhp1MJrh7nUepCBck5+mAzfO9JrbApNNgaTdGDITg=
+gopkg.in/check.v1 v0.0.0-20161208181325-20d25e280405/go.mod h1:Co6ibVJAznAaIkqp8huTwlJQCZ016jof/cbN4VW5Yz0=
+gopkg.in/yaml.v3 v3.0.1 h1:fxVm/GzAzEWqLHuvctI91KS9hhNmmWOoWu0XTYJS7CA=
+gopkg.in/yaml.v3 v3.0.1/go.mod h1:K4uyk7z7BCEPqu6E+C64Yfv1cQ7kz7rIZviUmN+EgEM=
diff --git a/internal/sampler/sampler.go b/internal/sampler/sampler.go
new file mode 100644
index 0000000..cbc51b7
--- /dev/null
+++ b/internal/sampler/sampler.go
@@ -0,0 +1,198 @@
+package sampler
+
+import (
+	"bytes"
+	"context"
+	"encoding/binary"
+	"fmt"
+	"os"
+	"os/exec"
+	"strconv"
+	"time"
+
+	"github.com/gordonklaus/portaudio"
+	"go.uber.org/zap"
+)
+
+func New(log *zap.Logger, sampleRate float64, splitPeriod time.Duration) Sampler {
+	s := &sampler{
+		log:       log,
+		sample:    sampleRate,
+		splitFreq: splitPeriod,
+		chunk:     make(chan chunk, 5),
+		err:       make(chan error),
+	}
+	return s
+}
+
+type Sampler interface {
+	Sample(ctx context.Context)
+}
+
+type sampler struct {
+	log       *zap.Logger
+	sample    float64
+	splitFreq time.Duration
+	chunk     chan chunk
+	err       chan error
+}
+
+func (s sampler) Sample(ctx context.Context) {
+	go func() {
+		for err := range s.err {
+			s.log.Error("sample failed", zap.Error(err))
+		}
+	}()
+	go func() {
+		if err := s.consume(ctx); err != nil {
+			s.err <- fmt.Errorf("consume: %w", err)
+		}
+	}()
+	if err := s.stream(ctx); err != nil {
+		s.err <- fmt.Errorf("stream: %w", err)
+	}
+}
+
+func (s sampler) consume(ctx context.Context) (err error) {
+	defer func() {
+		if err != nil {
+			s.log.Error("consume failed", zap.Error(err))
+		}
+	}()
+	for {
+		select {
+		case <-ctx.Done():
+			return
+		case c := <-s.chunk:
+			pr := outPath(c.ts, outRaw)
+			s.log.Info("writing new chunk", zap.String("path", pr))
+			f, err := os.Create(pr)
+			if err != nil {
+				return fmt.Errorf("os create %q: %w", pr, err)
+			}
+			defer func() {
+				err = f.Close()
+			}()
+			if _, err = f.Write(c.raw); err != nil {
+				return fmt.Errorf("write chunk %q: %w", pr, err)
+			}
+			if err = f.Close(); err != nil {
+				return fmt.Errorf("close chunk %q: %w", pr, err)
+			}
+			p3 := outPath(c.ts, outMp3)
+			s.log.Info("converting new chunk", zap.String("path", p3))
+			err = convert(pr, p3, strconv.Itoa(int(s.sample)/2))
+			if err != nil {
+				return fmt.Errorf("convert: %w", err)
+			}
+			if err := os.Remove(pr); err != nil {
+				return fmt.Errorf("remove %q: %w", pr, err)
+			}
+		default:
+			time.Sleep(time.Second)
+		}
+	}
+}
+
+func (s sampler) stream(ctx context.Context) (err error) {
+	defer func() {
+		if err != nil {
+			s.log.Error("sample failed", zap.Error(err))
+		}
+	}()
+
+	if err := portaudio.Initialize(); err != nil {
+		return fmt.Errorf("portaudio Initialize: %w", err)
+	}
+	defer func() {
+		err = portaudio.Terminate()
+	}()
+
+	in := make([]int16, 64)
+	stream, err := portaudio.OpenDefaultStream(1, 0, float64(s.sample), len(in), in)
+	if err != nil {
+		return fmt.Errorf("open stream: %w", err)
+	}
+	defer func() {
+		err = stream.Close()
+	}()
+
+	if err := stream.Start(); err != nil {
+		return fmt.Errorf("stream start: %w", err)
+	}
+
+	t := time.NewTicker(time.Duration(s.splitFreq))
+
+	var b bytes.Buffer
+
+loop:
+	for {
+		if err := stream.Read(); err != nil {
+			return fmt.Errorf("stream read: %w", err)
+		}
+		if err := binary.Write(&b, binary.LittleEndian, in); err != nil {
+			return fmt.Errorf("binary write: %w", err)
+		}
+		select {
+		case <-ctx.Done():
+			break loop
+		case <-t.C:
+			ts := time.Now()
+			s.chunk <- chunk{
+				raw: b.Bytes(),
+				ts:  ts,
+			}
+			b.Reset()
+		default:
+		}
+	}
+
+	if err := stream.Stop(); err != nil {
+		return fmt.Errorf("stream stop: %w", err)
+	}
+	return nil
+}
+
+type outType int
+
+const (
+	outMp3 outType = iota
+	outRaw
+)
+
+func outPath(ts time.Time, t outType) string {
+	ext := "mp3"
+	if t == outRaw {
+		ext = "raw"
+	}
+	return fmt.Sprintf("%s.%s", ts.Format("20060102150405"), ext)
+}
+
+type chunk struct {
+	raw []byte
+	ts  time.Time
+}
+
+func convert(raw, mp3, freq string) (err error) {
+	_, err = os.Stat(mp3)
+	if err == nil {
+		if err = os.Remove(mp3); err != nil {
+			return fmt.Errorf("remove %q: %w", mp3, err)
+		}
+	}
+	args := []string{
+		"-f", "s16le",
+		"-ar", freq,
+		"-ac", "2",
+		"-i", raw, mp3,
+	}
+	fmt.Println(append([]string{"ffmpeg"}, args...))
+	cmd := exec.Command("ffmpeg", args...)
+	// cmd.Stdout = os.Stdout
+	// cmd.Stderr = os.Stderr
+	err = cmd.Run()
+	if err != nil {
+		return fmt.Errorf("exec ffmpeg: %w", err)
+	}
+	return nil
+}
diff --git a/main.go b/main.go
new file mode 100644
index 0000000..cb3f555
--- /dev/null
+++ b/main.go
@@ -0,0 +1,11 @@
+package main
+
+import (
+	"groq/cmd"
+
+	"github.com/spf13/cobra"
+)
+
+func main() {
+	cobra.CheckErr(cmd.NewCLI().Execute())
+}


----------------

git status -s

A  cmd/record.go
A  cmd/root.go
A  cmd/sidecar.go
A  go.mod
A  go.sum
A  internal/sampler/sampler.go
A  main.go

```


-------------------------------------------------
-------- NEXT STEPS AND WORK IN PROGRESS --------
-------------------------------------------------


**WIP Context:** *None*