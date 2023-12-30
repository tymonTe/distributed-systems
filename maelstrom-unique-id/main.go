package main

import (
	"encoding/json"
	"log"
	"strconv"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

func generatedUniqueId(node *maelstrom.Node, counter int64) (int64, error) {
	substr := node.ID()[1:]
	nodeId, err := strconv.ParseInt(substr, 10, 64)
	if err != nil {
		return 0, err
	}

	return (nodeId << 32) + counter, nil
}

func main() {

	n := maelstrom.NewNode()
	var counter int64 = 0

	n.Handle("generate", func(msg maelstrom.Message) error {
		var requestBody map[string]any
		if err := json.Unmarshal(msg.Body, &requestBody); err != nil {
			return err
		}

		requestBody["type"] = "generate_ok"

		counter += 1
		generatedUniqueId, err := generatedUniqueId(n, counter)
		if err != nil {
			return err
		}
		requestBody["id"] = generatedUniqueId

		return n.Reply(msg, requestBody)
	})

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}

}
