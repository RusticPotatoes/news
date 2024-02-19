package idgen

import (
	"context"
	"fmt"
	"strings"

	"github.com/bwmarrin/snowflake"
)

type Node struct {
    node *snowflake.Node
}

func (n *Node) Init(ctx context.Context) error {
	var err error
	n.node, err = snowflake.NewNode(137)
	return err
}

func (n *Node) New(prefix string) string {
    id := n.node.Generate()
    return fmt.Sprintf("%s_%s", prefix, id.Base58())
}

func Parse(id string) (snowflake.ID, error) {
	parts := strings.Split(id, "_")
	if len(parts) != 2 {
		return -1, fmt.Errorf("expected ID formatted as prefix_id, got %s", id)
	}
	return snowflake.ParseBase58([]byte(parts[1]))
}
