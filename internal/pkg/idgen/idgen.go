package idgen

import "github.com/bwmarrin/snowflake"

var defaultNode = func() *snowflake.Node {
	node, err := snowflake.NewNode(1)
	if err != nil {
		panic(err)
	}
	return node
}()

func GenerateUniqueID() string {
	return defaultNode.Generate().String()
}
