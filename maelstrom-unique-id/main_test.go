package main

import (
	"fmt"
	"strconv"
	"testing"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

func TestGeneratedUniqueId(t *testing.T) {
	var n maelstrom.Node
	n.Init("n2", []string{})

	result, err := generatedUniqueId(&n, 2)
	if err != nil {
		t.Error(err)
	}

	inBinary := strconv.FormatInt(result, 2)
	fmt.Printf(`Result: %v`, inBinary)
}
