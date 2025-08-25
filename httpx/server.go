package httpx

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/vloryan/go-libs/stringx"
)

type Server struct {
	http.Server
	onStartUp      func()
	middlewareFunc func(req *http.Request) *http.Request
	Router         http.Handler
	status         Status
}
type Status int

const (
	StatusStopped Status = iota
	StatusStarting
	StatusStarted
)

func NewServer(router http.Handler) *Server {
	s := &Server{
		Router: router,
		status: StatusStopped,
	}
	s.Server = http.Server{
		Handler:      s,
		Addr:         ":8080",
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}
	return s
}

func (s *Server) WithMiddleware(middlewareFunc func(req *http.Request) *http.Request) *Server {
	s.middlewareFunc = middlewareFunc
	return s
}

func (s *Server) WithOnStartUp(onStartUp func()) *Server {
	s.onStartUp = onStartUp
	return s
}

func (s *Server) Start() {
	s.status = StatusStarting
	go func() {
		if err := s.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()
}

func (s *Server) URL() *url.URL {
	parts := strings.Split(s.Addr, ":")
	hostName := parts[0]
	port := parts[1]
	if hostName == "" {
		hostName = "localhost"
	}
	protocol := "https://"
	if s.TLSConfig == nil {
		protocol = "http://"
	}
	u, err := url.Parse(protocol + hostName + ":" + port)
	if err != nil {
		return nil
	}
	return u
}

func (s *Server) ListenAndServe() (err error) {
	startTime := time.Now().UTC()
	if s.onStartUp != nil {
		s.onStartUp()
	}
	log.Printf("Server started within %s", time.Since(startTime))
	s.status = StatusStarted
	return s.Server.ListenAndServe()
}

func (s *Server) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown error: %v", err)
		return
	}
	s.status = StatusStopped
	log.Printf("Server stopped")
}

func (s *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	start := time.Now().UTC()
	sw := &StatusAwareResponseWriter{ResponseWriter: w}
	defer s.logResponse(sw, req, start)
	if s.middlewareFunc != nil {
		req = s.middlewareFunc(req)
	}
	if s.Router == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	s.Router.ServeHTTP(sw, req)
}

func (s *Server) logResponse(writer *StatusAwareResponseWriter, req *http.Request, start time.Time) {
	var statusText string
	if writer.Status() >= http.StatusOK && writer.Status() < http.StatusMultipleChoices {
		statusText = stringx.FormatColored(stringx.ConsoleColorBgGreen, strconv.Itoa(writer.Status()))
	} else if writer.Status() >= http.StatusMultipleChoices && writer.Status() < http.StatusBadRequest {
		statusText = stringx.FormatColored(stringx.ConsoleColorBgGray, strconv.Itoa(writer.Status()))
	} else if writer.Status() >= http.StatusBadRequest && writer.Status() < http.StatusInternalServerError {
		statusText = stringx.FormatColored(stringx.ConsoleColorBgYellow, strconv.Itoa(writer.Status()))
	} else {
		statusText = stringx.FormatColored(stringx.ConsoleColorBgRed, strconv.Itoa(writer.Status()))
	}
	duration := time.Since(start)
	path := req.URL.Path
	if req.URL.RawQuery != "" {
		path += "?" + req.URL.RawQuery
	}
	msg := fmt.Sprintf("%s %s %s %s",
		statusText,
		stringx.FormatColoredRight(stringx.ConsoleColorReset, duration.String(), 15),
		stringx.FormatColoredCenter(stringx.ConsoleColorBgGray, req.Method, 5),
		path)
	log.Print(msg)
}
