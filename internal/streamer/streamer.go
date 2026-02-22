package streamer

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	GetUser(ctx context.Context, streamerName string) (uuid.UUID, error)
}
