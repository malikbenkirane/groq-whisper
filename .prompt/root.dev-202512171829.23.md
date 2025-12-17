**Input:**

**Scope:** *None*

**Git Log:** *None*

**Diff:**
```diff
diff --git a/cmd/record.go b/cmd/record.go
index 685e1ef..81a4be3 100644
--- a/cmd/record.go
+++ b/cmd/record.go
@@ -10,15 +10,16 @@ import (
 	"groq/internal/sampler"
 
 	"github.com/spf13/cobra"
-	"go.uber.org/zap"
 )
 
-func newCommandRecord(log *zap.Logger) *cobra.Command {
+func newCommandRecord() *cobra.Command {
 	var freq *int
+	var debug *bool
 	cmd := &cobra.Command{
 		Use:     "record",
 		Aliases: []string{"rec", "r"},
 		RunE: func(cmd *cobra.Command, args []string) (err error) {
+			log := newLogger(*debug)
 			quit := make(chan os.Signal, 1)
 			signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
 
@@ -33,5 +34,6 @@ func newCommandRecord(log *zap.Logger) *cobra.Command {
 		},
 	}
 	freq = cmd.Flags().IntP("freq", "f", 16000, "sample rate")
+	debug = cmd.Flags().Bool("debug", false, "set log level at debug")
 	return cmd
 }
diff --git a/cmd/root.go b/cmd/root.go
index 34e2bef..174d862 100644
--- a/cmd/root.go
+++ b/cmd/root.go
@@ -12,10 +12,23 @@ func NewCLI() *cobra.Command {
 			return cmd.Help()
 		},
 	}
-	log, err := zap.NewDevelopment()
+	cmd.AddCommand(
+		newCommandRecord(),
+		newCommandSidecar())
+	return cmd
+}
+
+func newLogger(debug bool) *zap.Logger {
+	lvl := zap.InfoLevel
+	if debug {
+		lvl = zap.DebugLevel
+	}
+	config := zap.Config{
+		Level: zap.NewAtomicLevelAt(lvl),
+	}
+	log, err := config.Build()
 	if err != nil {
 		panic(err)
 	}
-	cmd.AddCommand(newCommandRecord(log), newCommandSidecar(log))
-	return cmd
+	return log
 }
diff --git a/cmd/sidecar.go b/cmd/sidecar.go
index a94ff23..3a04199 100644
--- a/cmd/sidecar.go
+++ b/cmd/sidecar.go
@@ -133,12 +133,14 @@ func (gc *groqClient) post(audio, model string) (err error) {
 	return nil
 }
 
-func newCommandSidecar(log *zap.Logger) *cobra.Command {
-	var dry *bool
+func newCommandSidecar() *cobra.Command {
+	var dry, debug *bool
 	cmd := &cobra.Command{
 		Use:     "sidecar",
 		Aliases: []string{"watch", "w", "s"},
 		RunE: func(cmd *cobra.Command, args []string) (err error) {
+			log := newLogger(*debug)
+
 			quit := make(chan os.Signal, 1)
 			signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
 
@@ -232,5 +234,6 @@ func newCommandSidecar(log *zap.Logger) *cobra.Command {
 		},
 	}
 	dry = cmd.Flags().Bool("dry", false, "don't post to groq")
+	debug = cmd.Flags().Bool("debug", false, "set log level at debug")
 	return cmd
 }


----------------

git status -s

M  cmd/record.go
M  cmd/root.go
M  cmd/sidecar.go

```


-------------------------------------------------
-------- NEXT STEPS AND WORK IN PROGRESS --------
-------------------------------------------------


**WIP Context:**

*Technical Debt Workarounds:*
- newLogger panic
