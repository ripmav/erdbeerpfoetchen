package handle

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/ripmav/erdbeerpfoetchen/internal/lamimi"
	"github.com/ripmav/erdbeerpfoetchen/internal/middleware"
)

type LamimiService interface {
	WriteLamimiCollections(ctx context.Context, key string, user string, collection *lamimi.Collection) error
	ReadLamimiCollections(ctx context.Context, key string, user string) (*lamimi.Collection, error)
}

type LamimiHandler struct {
	service LamimiService
}

func NewLamimiHandler(service LamimiService) *LamimiHandler {
	return &LamimiHandler{
		service: service,
	}
}

func (h *LamimiHandler) receive(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	storageKey := r.Header.Get("X-STORAGE-KEY")

	if storageKey == "" {
		slog.Error("empty storage key")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	userKey := r.Header.Get("X-USER-KEY")

	if userKey == "" {
		slog.Error("empty user key")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	collection, err := decode(r.Body)

	if err != nil {
		slog.Error("cannot decode payload", "error", err)
		w.WriteHeader(http.StatusBadRequest)
	}

	if err := h.service.WriteLamimiCollections(ctx, storageKey, userKey, collection); err != nil {
		slog.Error("cannot write collections", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	slog.Debug("lamimi collection successful written")
	w.WriteHeader(http.StatusAccepted)
}

func (h *LamimiHandler) send(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	storageKey := r.Header.Get("X-STORAGE-KEY")

	if storageKey == "" {
		slog.Error("empty storage key")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	userKey := r.Header.Get("X-USER-KEY")

	if userKey == "" {
		slog.Error("empty user key")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	collection, err := h.service.ReadLamimiCollections(ctx, storageKey, userKey)
	if err != nil {
		slog.Error("cannot read collections", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	slog.Debug("lamimi collection successful read")
	_, err = w.Write(collection.RawMessage)

	if err != nil {
		slog.Error("cannot write response", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func (h *LamimiHandler) Install(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/lamimi", middleware.Auth(h.receive))
	mux.HandleFunc("GET /api/v1/lamimi", middleware.Auth(h.send))
}

func decode(r io.Reader) (*lamimi.Collection, error) {
	buf, err := io.ReadAll(io.LimitReader(r, 12<<20)) // 12MiB
	if err != nil {
		return nil, fmt.Errorf("error read bytes: %w", err)
	}

	var collection lamimi.Collection
	err = json.Unmarshal(buf, &collection)
	if err != nil {
		return nil, fmt.Errorf("error unmarshal json: %w", err)
	}

	return &collection, nil
}
