package handle

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"github.com/ripmav/erdbeerpfoetchen/internal/collection"
	"github.com/ripmav/erdbeerpfoetchen/internal/database/model"
	"github.com/ripmav/erdbeerpfoetchen/internal/middleware"
)

type CollectionService interface {
	WriteCollection(ctx context.Context, collectionKey, user string, streamerId uuid.UUID, collection *collection.Collection) error
	ReadCollection(ctx context.Context, collectionKey, user string, streamerId uuid.UUID) (*collection.Collection, error)
}

type UserService interface {
	GetUserApiToken(ctx context.Context, streamerName string) (uuid.UUID, error)
	GetUserByUserName(ctx context.Context, userName string) (*model.StreamingUser, error)
}

type CollectionHandler struct {
	service CollectionService
	user    UserService
}

func NewCollectionHandler(service CollectionService, user UserService) *CollectionHandler {
	return &CollectionHandler{
		service: service,
		user:    user,
	}
}

func (h *CollectionHandler) write(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	userKey := r.Header.Get("X-USER-KEY")

	if userKey == "" {
		slog.ErrorContext(ctx, "empty user key")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	collectionKey := r.Header.Get("X-COLLECTION-KEY")

	if collectionKey == "" {
		slog.ErrorContext(ctx, "empty collection key")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	c, err := decode(r.Body)

	if err != nil {
		slog.ErrorContext(ctx, "cannot decode payload", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	streamerName := r.PathValue("streamer_name")

	userId, err := h.user.GetUserApiToken(ctx, streamerName)
	if err != nil {
		slog.ErrorContext(ctx, "cannot get user", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := h.service.WriteCollection(ctx, collectionKey, userKey, userId, c); err != nil {
		slog.ErrorContext(ctx, "cannot write collections", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	slog.DebugContext(ctx, "collection collection successful written")
	w.WriteHeader(http.StatusAccepted)
}

func (h *CollectionHandler) read(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	userKey := r.Header.Get("X-USER-KEY")

	if userKey == "" {
		slog.ErrorContext(ctx, "empty user key")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	collectionKey := r.Header.Get("X-COLLECTION-KEY")

	if collectionKey == "" {
		slog.ErrorContext(ctx, "empty collection key")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	streamerName := r.PathValue("streamer_name")

	u, err := h.user.GetUserByUserName(ctx, streamerName)
	if err != nil {
		slog.ErrorContext(ctx, "cannot get user", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	c, err := h.service.ReadCollection(ctx, collectionKey, userKey, u.ID)
	if err != nil {
		slog.ErrorContext(ctx, "cannot read collections", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	slog.DebugContext(ctx, "collection collection successful read")

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(c.RawMessage)))
	w.WriteHeader(http.StatusOK)

	_, err = w.Write(c.RawMessage)

	if err != nil {
		slog.ErrorContext(ctx, "cannot write response", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (h *CollectionHandler) Install(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/collection/{streamer_name}", middleware.Auth(h.write))
	mux.HandleFunc("GET /api/v1/collection/{streamer_name}", middleware.Auth(h.read))
}

func decode(r io.Reader) (*collection.Collection, error) {
	buf, err := io.ReadAll(io.LimitReader(r, 12<<20)) // 12MiB
	if err != nil {
		return nil, fmt.Errorf("error read bytes: %w", err)
	}

	var c collection.Collection
	err = json.Unmarshal(buf, &c)
	if err != nil {
		return nil, fmt.Errorf("error unmarshal json: %w", err)
	}

	return &c, nil
}
