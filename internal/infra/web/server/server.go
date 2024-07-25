package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type handler struct {
	method string
	path   string
	fn     http.HandlerFunc
}

type WebServer struct {
	Router        chi.Router
	Handlers      []handler
	WebServerPort string
}

func NewWebServer(serverPort string) *WebServer {
	return &WebServer{
		Router:        chi.NewRouter(),
		Handlers:      []handler{},
		WebServerPort: serverPort,
	}
}

func (s *WebServer) AddHandler(method string, path string, fn http.HandlerFunc) {
	s.Handlers = append(s.Handlers, handler{
		method: method,
		path:   path,
		fn:     fn,
	})
}

func (s *WebServer) Start() {
	s.Router.Use(middleware.Logger)

	for _, handler := range s.Handlers {
		s.Router.Method(handler.method, handler.path, handler.fn)
	}

	http.ListenAndServe(":"+s.WebServerPort, s.Router)
}
