package update

import (
	"context"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"io"
	"log/slog"
	"net/http"
	resp "tt-demo/lib/api/response"
)

type Request struct {
	Value string `json:"value" validate:"required"`
}

type ValueUpdater interface {
	UpdateValue(ctx context.Context, key, value string) error
}

func New(log *slog.Logger, valueUpdater ValueUpdater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.UpdateValue.New"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req Request

		w.Header().Set("Content-Type", "application/json")
		err := render.DecodeJSON(r.Body, &req)
		if errors.Is(err, io.EOF) {
			w.WriteHeader(http.StatusBadRequest)
			log.Error("request body is empty")

			render.JSON(w, r, resp.Error("empty request"))

			return
		}
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			log.Error("failed to decode request body", err)

			render.JSON(w, r, resp.Error("failed to decode request"))

			return
		}

		if err := validator.New().Struct(req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			validateErr := err.(validator.ValidationErrors)

			log.Error("invalid request", err)

			render.JSON(w, r, resp.ValidationError(validateErr))

			return
		}

		key := chi.URLParam(r, "key")

		if key == "" {
			log.Info("key is empty")

			render.JSON(w, r, resp.Error("key is empty"))

			return
		}

		if err := valueUpdater.UpdateValue(r.Context(), key, req.Value); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			log.Error("failed to update", err)

			render.JSON(w, r, resp.Error(err.Error()))

			return
		}

		render.JSON(w, r, resp.OK())
	}
}
