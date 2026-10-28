package idgenerator

import (
	"fmt"

	"github.com/gofrs/uuid"
	"github.com/samber/lo"
	"github.com/segmentio/ksuid"
)

type IdGenerator interface {
	GenUuidV4() uuid.UUID
	GenUuidV5(uid uuid.UUID, name string) uuid.UUID
	GenKsuid() ksuid.KSUID
	GenUID(id int64) string
}

type idGenerator struct{}

func (i idGenerator) GenUuidV4() uuid.UUID {
	return lo.Must(uuid.NewV4())
}

func (i idGenerator) GenUuidV5(uid uuid.UUID, name string) uuid.UUID {
	return uuid.NewV5(uid, name)
}

func (i idGenerator) GenKsuid() ksuid.KSUID {
	return lo.Must(ksuid.NewRandom())
}

func (i idGenerator) GenUID(id int64) string {
	return fmt.Sprintf("1%07d", id)
}

func New() IdGenerator {
	return idGenerator{}
}
