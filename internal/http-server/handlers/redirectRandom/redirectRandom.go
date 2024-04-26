package redirectrandom

import (
	"fmt"
	"log/slog"
	"net/http"

	resp "github.com/Gonnekone/rest-url-shortener/internal/lib/api/response"
	"github.com/Gonnekone/rest-url-shortener/internal/lib/logger/sl"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

//go:generate go run github.com/vektra/mockery/v2@v2.42.1 --name=AliasGetter
type AliasGetter interface {
	GetRandomAlias() (string, error)
}

func New(log *slog.Logger, aliasGetter AliasGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.redirectRandom.New"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		alias, err := aliasGetter.GetRandomAlias()
		if err != nil {
			log.Error("failed to get alias", sl.Err(err))

			render.JSON(w, r, resp.Error("internal error"))

			return
		}

		log.Info("got alias", slog.String("alias", alias))

		fmt.Println("redirecting to", fmt.Sprint("redirect/", alias))

		http.Redirect(w, r, fmt.Sprint("redirect/", alias), http.StatusFound)
	}
}
