package main

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

func captureStdout() (*os.File, *os.File, *os.File) {
	var originalStdout *os.File = os.Stdout

	r, w, _ := os.Pipe()
	os.Stdout = w
	return originalStdout, r, w
}

func getOutputString(r *os.File, w *os.File) string {
	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

func TestConfirmsBroadcast(t *testing.T) {
	originalStdout, r, w := captureStdout()
	defer func() { w.Close(); os.Stdout = originalStdout }()

	node := maelstrom.NewNode()
	msg := maelstrom.Message{Body: []byte(`{"message": 123}`)}
	err := handleBroadcast(node, &msg, 123)
	if err != nil {
		t.Errorf("handleBroadcast returned an error: %v", err)
	}

	output := getOutputString(r, w)
	if output == "" {
		t.Errorf("Expected pipe to contain output, but it was empty")
	}
	if !strings.Contains(output, "broadcast_ok") {
		t.Errorf("Expected output to contain 'broadcast_ok', but it did not")
	}
	os.Stdout = originalStdout
}

func TestDoesntBroadcastRepeatMessages(t *testing.T) {
	originalStdout, r, w := captureStdout()
	defer func() { w.Close(); os.Stdout = originalStdout }()

	messagesReceived[123] = true // already received

	node := maelstrom.NewNode()
	msg := maelstrom.Message{Body: []byte(`{"message": 123}`)}
	err := handleBroadcast(node, &msg, 123)
	if err != nil {
		t.Errorf("handleBroadcast returned an error: %v", err)
	}

	output := getOutputString(r, w)
	if output == "" {
		t.Errorf("Expected pipe to contain output, but it was empty")
	}
	if strings.Contains(output, "\"type\":\"broadcast\"") {
		t.Errorf("Must not broadcast messages which were already received.")
	}
	os.Stdout = originalStdout
}

func TestGossipsNewMessages(t *testing.T) {
	originalStdout, r, w := captureStdout()
	defer func() { w.Close(); os.Stdout = originalStdout }()

	topology["n1"] = []string{"n2"}
	messagesReceived[123] = true // already received
	newMessage := int64(456)

	node := maelstrom.NewNode()
	msg := maelstrom.Message{Body: []byte(`{"message": 123}`)}
	err := handleBroadcast(node, &msg, newMessage)
	if err != nil {
		t.Errorf("handleBroadcast returned an error: %v", err)
	}

	output := getOutputString(r, w)
	if output == "" {
		t.Errorf("Expected pipe to contain output, but it was empty")
	}
	firstLine := strings.Split(output, "\n")[0]
	var parsedMap struct {
		Dest string
		Body struct {
			Messages []int64
			Type     string
		}
	}
	err = json.Unmarshal([]byte(firstLine), &parsedMap)
	if err != nil {
		t.Errorf("Expected output to be valid JSON, but it was not: %v", err)
	}
	if parsedMap.Dest != "n2" {
		t.Errorf("Expected output to be sent to n2, but it was sent to %v", parsedMap.Dest)
	}

	found := false
	for _, msg := range parsedMap.Body.Messages {
		if msg == 456 {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected message 456 to be in the broadcast, but it was not: %v", parsedMap.Body.Messages)
	}

	os.Stdout = originalStdout
}
