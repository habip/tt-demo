package get

import (
	"context"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
	"tt-demo/internal/storage"
	resp "tt-demo/lib/api/response"
)

type Response struct {
	resp.Response
	Value string `json:"value"`
}

type ValueGetter interface {
	GetValue(ctx context.Context, key string) (string, error)
}

func New(log *slog.Logger, valueGetter ValueGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.GetValue.New"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		w.Header().Set("Content-Type", "application/json")

		key := chi.URLParam(r, "key")

		if key == "" {
			log.Info("key is empty")

			render.JSON(w, r, resp.Error("key is empty"))

			return
		}

		value, err := valueGetter.GetValue(r.Context(), key)

		if errors.Is(err, storage.KeyNotFound) {
			w.WriteHeader(http.StatusNotFound)
			log.Error("key not found", err)

			render.JSON(w, r, resp.Error("key not found"))

			return
		}

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Error("failed to get value", err)

			render.JSON(w, r, resp.Error("failed to get value"))

			return
		}

		render.JSON(w, r, Response{Response: resp.OK(), Value: value})

		log.Info("get value for key", key)
	}
}
