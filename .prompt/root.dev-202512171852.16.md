**Input:**

**Scope:** *None*

**Git Log:** *None*

**Diff:**
```diff
diff --git a/.prompt/flags.json b/.prompt/flags.json
index 5e5c2b8..e647449 100644
--- a/.prompt/flags.json
+++ b/.prompt/flags.json
@@ -18,23 +18,15 @@
     "DocumentationUserDoc": [],
     "DoucmentationBuildGuide": [],
     "FixAttempt": [],
-    "Idea": [
-	    "improve details for low level IT technicians",
-	    "use LLM to split the steps for any regular windows user without any knowldge of system files, e.g say extract and rename file/folder instead of just saying extract to"
-    ],
+    "Idea": [],
     "ImplementationNeedsReview": [],
     "ImplementationParial": [],
     "ImplementationTodo": [],
     "InProgress": [],
-    "KnownIssue": [
-	    "security wise key.txt is not protected so we need to be careful about key management"
-    ],
+    "KnownIssue": [],
     "KnownIssueError": [],
     "KnownIssueUpdate": [],
-    "NeedsDecision": [
-	    "is automatisation necessary",
-	    "groq upgrade automatisation"
-    ],
+    "NeedsDecision": [],
     "Other": [],
     "Question": [],
     "TechincalDebt": [],
@@ -53,7 +45,7 @@
     "TestingUserInterface": []
   },
   "Logs": [],
-  "Scope": "Docs",
+  "Scope": "",
   "_available_scopes": [
     "Other",
     "Accessibility",
diff --git a/cmd/root.go b/cmd/root.go
index 174d862..2cb3bbb 100644
--- a/cmd/root.go
+++ b/cmd/root.go
@@ -19,12 +19,9 @@ func NewCLI() *cobra.Command {
 }
 
 func newLogger(debug bool) *zap.Logger {
-	lvl := zap.InfoLevel
+	config := zap.NewProductionConfig()
 	if debug {
-		lvl = zap.DebugLevel
-	}
-	config := zap.Config{
-		Level: zap.NewAtomicLevelAt(lvl),
+		config = zap.NewDevelopmentConfig()
 	}
 	log, err := config.Build()
 	if err != nil {


----------------

git status -s

M  .prompt/flags.json
M  cmd/root.go

```


-------------------------------------------------
-------- NEXT STEPS AND WORK IN PROGRESS --------
-------------------------------------------------


**WIP Context:** *None*