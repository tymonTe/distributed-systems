package main

import (
	"encoding/json"
	"fmt"
	"log"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

var (
	messagesReceived map[int64]bool      = make(map[int64]bool)
	topology         map[string][]string = make(map[string][]string)
)

func handleBroadcast(node *maelstrom.Node, msg *maelstrom.Message, newMessage int64) error {
	if isNewMessage := !messagesReceived[newMessage]; isNewMessage {

		messagesReceived[newMessage] = true

		listOfMessages := []int64{}
		for message := range messagesReceived {
			listOfMessages = append(listOfMessages, message)
		}

		for _, peer := range topology[node.ID()] {
			if err := node.Send(peer, map[string]any{
				"type":     "gossip",
				"messages": listOfMessages,
			}); err != nil {
				return err
			}
		}
	}

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

	n.Handle("gossip", func(msg maelstrom.Message) error {
		var requestBody struct {
			Messages []int64
			Type     string
		}
		if err := json.Unmarshal(msg.Body, &requestBody); err != nil {
			return err
		}

		for _, message := range requestBody.Messages {
			messagesReceived[message] = true
		}

		return nil
	})

	n.Handle("read", func(msg maelstrom.Message) error {
		var requestBody map[string]any
		if err := json.Unmarshal(msg.Body, &requestBody); err != nil {
			return err
		}
		listOfMessages := []int64{}
		for message := range messagesReceived {
			listOfMessages = append(listOfMessages, message)
		}

		return n.Reply(msg, map[string]any{
			"type":     "read_ok",
			"messages": listOfMessages,
		})
	})

	n.Handle("topology", func(msg maelstrom.Message) error {
		var requestBody struct {
			Type     string
			Topology map[string][]string
		}
		if err := json.Unmarshal(msg.Body, &requestBody); err != nil {
			return err
		}

		topology = requestBody.Topology
		return n.Reply(msg, map[string]any{
			"type": "topology_ok",
		})
	})

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}
