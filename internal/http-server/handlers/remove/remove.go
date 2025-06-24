package remove

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

type KeyDelete interface {
	Delete(ctx context.Context, key string) error
}

func New(log *slog.Logger, keyDelete KeyDelete) http.HandlerFunc {
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

		err := keyDelete.Delete(r.Context(), key)

		if errors.Is(err, storage.KeyNotFound) {
			w.WriteHeader(http.StatusNotFound)
			log.Error("key not found", err)

			render.JSON(w, r, resp.Error("key not found"))

			return
		}

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Error("failed to delete key", err)

			render.JSON(w, r, resp.Error("failed to delete key"))

			return
		}

		render.JSON(w, r, resp.OK())

		log.Info("key deleted", key)
	}
}
