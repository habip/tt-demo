package save

import (
	"context"
	"errors"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"io"
	"log/slog"
	"net/http"
	resp "tt-demo/lib/api/response"
)

type Request struct {
	Key   string `json:"key" validate:"required,alphanum"`
	Value string `json:"value" validate:"required"`
}

type ValueSetter interface {
	SetValue(ctx context.Context, key, value string) error
}

func New(log *slog.Logger, valueSetter ValueSetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.SetValue.New"

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
			log.Error("failed to decode request body", err.Error())

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

		if err := valueSetter.SetValue(r.Context(), req.Key, req.Value); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			log.Error("failed to create", err)

			render.JSON(w, r, resp.Error(err.Error()))

			return
		}

		render.JSON(w, r, resp.OK())
	}
}
