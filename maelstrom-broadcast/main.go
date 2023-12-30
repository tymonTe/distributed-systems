package main

import (
	"encoding/json"
	"fmt"
	"log"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

var messagesReceived []int64 = []int64{}
var topology map[string][]string = make(map[string][]string)

func handleBroadcast(node *maelstrom.Node, msg *maelstrom.Message, newMessage int64) error {
	messagesReceived = append(messagesReceived, newMessage)
	return node.Reply(*msg, map[string]any{
		"type": "broadcast_ok",
	})
}

func main() {
	n := maelstrom.NewNode()

	n.Handle("broadcast", func(msg maelstrom.Message) error {
		var requestBody map[string]any
		if err := json.Unmarshal(msg.Body, &requestBody); err != nil {
			return err
		}

		messageValue, ok := requestBody["message"].(float64)
		if !ok {
			return fmt.Errorf("invalid message type: expected float64, got %T", requestBody["message"])
		}

		return handleBroadcast(n, &msg, int64(messageValue))
	})

	n.Handle("read", func(msg maelstrom.Message) error {
		var requestBody map[string]any
		if err := json.Unmarshal(msg.Body, &requestBody); err != nil {
			return err
		}

		requestBody["type"] = "read_ok"
		requestBody["messages"] = messagesReceived
		return n.Reply(msg, requestBody)
	})

	n.Handle("topology", func(msg maelstrom.Message) error {
		var requestBody map[string]any
		if err := json.Unmarshal(msg.Body, &requestBody); err != nil {
			return err
		}

		receivedTopology, ok := requestBody["topology"].(map[string][]string)
		if !ok {
			return fmt.Errorf("invalid topology type: expected map[string][]string, got %T", requestBody["topology"])
		}

		topology = receivedTopology
		return n.Reply(msg, map[string]any{
			"type": "topology_ok",
		})
	})

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}
