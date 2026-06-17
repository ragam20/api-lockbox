package idgen

import (
	"errors"

	"lockbox/internal/config"

	"github.com/bwmarrin/snowflake"
	"github.com/google/uuid"
)

type IDGenerator interface {
	NewID() int64
	NewUUIDV7() (uuid.UUID, error)
}

type idGen struct {
	snow *snowflake.Node
}

func NewIDGen(cfg *config.AppConfig) (IDGenerator, error) {
	snow, err := snowflake.NewNode(cfg.NodeID)
	if err != nil {
		return nil, errors.New("error init snowflake gen")
	}

	return &idGen{
		snow: snow,
	}, nil
}

func (s *idGen) NewID() int64 {
	return s.snow.Generate().Int64()
}

func (s *idGen) NewUUIDV7() (uuid.UUID, error) {
	id, err := uuid.NewV7()
	return id, err
}
