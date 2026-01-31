package handle

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/ripmav/erdbeerpfoetchen/internal/lamimi"
)

type LamimiService interface {
	WriteLamimiCollections(ctx context.Context, key string, lamimi *lamimi.Collection) error
}

type LamimiHandler struct {
	Service LamimiService
}

func NewLamimiHandler(service LamimiService) *LamimiHandler {
	return &LamimiHandler{Service: service}
}

func (h *LamimiHandler) Install(handle func(patter string, handler http.Handler)) {
	handle("POST /api/v1/lamimi", http.HandlerFunc(h.receive))
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

	collection, err := decode(r.Body)

	if err != nil {
		slog.Error("cannot decode payload", "error", err)
		w.WriteHeader(http.StatusBadRequest)
	}

	if err := h.Service.WriteLamimiCollections(ctx, storageKey, collection); err != nil {
		slog.Error("cannot write collections", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	slog.Debug("lamimi collection successful written")
	w.WriteHeader(http.StatusAccepted)
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
