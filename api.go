package awesomeProject

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
)

func NewRouter(api portalAPI, api2 AppAPI) *chi.Mux {
	router := chi.NewRouter()
	router.Get("/hello", world)
	router.Route("/awesomeProject", func(r chi.Router) {
		r = api.router(r)
		r = api2.router(r)
	})
	return router
}

type portalAPI struct {
	db *sql.DB
}

func NewPortalAPI(dbConnStr string) (portalAPI, error) {
	conn, err := sql.Open("sqlserver", dbConnStr)
	if err != nil {
		return portalAPI{}, err
	}

	err = conn.Ping()
	if err != nil {
		return portalAPI{}, err
	}
	return portalAPI{
		db: conn,
	}, nil
}

func (p portalAPI) router(r chi.Router) chi.Router {
	return r.Route("/portal", func(r chi.Router) {
		r.Get("/", p.doSomething)
	})
}

func (p portalAPI) doSomething(w http.ResponseWriter, r *http.Request) {
	_, err := p.db.Exec(`select * from vipps.Users where name = "Rune Garborg"`)
	if err != nil {
		_, _ = w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
	}
}

type AppAPI struct{}

func (a AppAPI) router(r chi.Router) chi.Router {
	return r.Route("/app", func(r chi.Router) {
		r.Get("/", world)
	})
}

func world(_ http.ResponseWriter, _ *http.Request) {
	fmt.Println("world")
}
