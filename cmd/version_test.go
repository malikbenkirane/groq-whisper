package cmd

import "testing"

func TestVersion(t *testing.T) {
	expected := "gs://groq-whisper/groq-v0.4.0.exe"
	if bucketFile("gs://groq-whisper", "v0.4.0") != expected {
		t.Logf("epected %q", expected)
		t.Fail()
	}
	expected = remoteBucket + "/" + "groq-v0.4.0.exe"
	if bucketFile(remoteBucket, "v0.4.0") != expected {
		t.Logf("expected %q", expected)
		t.Fail()
	}
}
