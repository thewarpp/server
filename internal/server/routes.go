package server

import (
	"encoding/json"
	"log"
	"net/http"
	"server/handler"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (s *Server) RegisterRoutes() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Get("/", s.HelloWorldHandler)

	r.Get("/health", s.healthHandler)

	r.Route("/api/v1", func(r chi.Router) {
		r.Mount("/workspace", s.workspaceRouter())
	})

	return r
}

func (s *Server) workspaceRouter() http.Handler {
	r := chi.NewRouter()

	r.Get("/", handler.HandleGetWorkspaces)
	r.Post("/", handler.HandlePostWorkspace)
	r.Get("/{workspaceID}", handler.HandleGetWorkspace)
	r.Delete("/{workspaceID}", handler.HandleDeleteWorkspace)

	return r
}

func (s *Server) HelloWorldHandler(w http.ResponseWriter, r *http.Request) {
	resp := make(map[string]string)
	resp["message"] = "Hello World"

	jsonResp, err := json.Marshal(resp)
	if err != nil {
		log.Fatalf("error handling JSON marshal. Err: %v", err)
	}

	_, _ = w.Write(jsonResp)
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	jsonResp, _ := json.Marshal(s.db.Health())
	_, _ = w.Write(jsonResp)
}
