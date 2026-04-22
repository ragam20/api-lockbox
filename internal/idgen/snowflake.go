package idgen

import (
	"errors"

	"lockbox/internal/config"

	"github.com/bwmarrin/snowflake"
)

type IDGenerator interface {
	NewID() int64
}

type SnowflakeGen struct {
	node *snowflake.Node
}

func NewSnowflakeGen(cfg *config.AppConfig) (IDGenerator, error) {
	node, err := snowflake.NewNode(cfg.NodeID)
	if err != nil {
		return nil, errors.New("error init snowflake gen")
	}

	return &SnowflakeGen{
		node: node,
	}, nil
}

func (s *SnowflakeGen) NewID() int64 {
	return s.node.Generate().Int64()
}
