package main

import (
	"bytes"
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
