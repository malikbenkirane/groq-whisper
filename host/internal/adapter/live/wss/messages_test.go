package wss

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestJoinMarshaling(t *testing.T) {
	{
		var msg eventMsg
		if err := json.Unmarshal([]byte(`
	{"member": "marcus", "event": "eventError"}
	`), &msg); err != nil {
			t.Fatalf("unmarshal: %s", err.Error())
		}
		if msg.Event == nil {
			t.Fatalf("nil event expected eventError")
		}
		if msg.Member != "marcus" {
			t.Fatalf("expected member marcus")
		}
		var b bytes.Buffer
		enc := json.NewEncoder(&b)
		enc.SetIndent("", "  ")
		if err := enc.Encode(msg); err != nil {
			t.Fatalf("encode msg: %s", err)
		}
		t.Log(b.String())
	}

	{
		var msg eventMsg
		if err := json.Unmarshal([]byte(`
{"member":"Bertrand","event":"eventJoin","payload":{"join":{"key":"age1m6nhvlatjjhwr65te8nruz708ymlx8ww77gtresf0qd9vdz3danqu452cg"}}}
	`), &msg); err != nil {
			t.Fatalf("unmarshal: %s", err.Error())
		}
		if *msg.Event != eventJoin {
			t.Fatalf("expected event joint got %q", *msg.Event)
		}
		if msg.Body == nil {
			t.Fatalf("expected body")
		}
		var b bytes.Buffer
		enc := json.NewEncoder(&b)
		enc.SetIndent("", "  ")
		if err := enc.Encode(msg); err != nil {
			t.Fatalf("encode msg: %s", err)
		}
		t.Log(b.String())
	}
}
