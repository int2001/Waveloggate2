package qsy

import (
	"crypto/tls"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/soheilhy/cmux"
)

// Server is the QSY HTTP server.
type Server struct {
	setFreqMode func(hz int64, mode string) error
}

// New creates a new QSY server.
func New(fn func(hz int64, mode string) error) *Server {
	return &Server{setFreqMode: fn}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")

	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Max-Age", "86400")
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Path: /{freq} or /{freq}/{mode}
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		http.NotFound(w, r)
		return
	}

	hz, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		http.Error(w, "Invalid frequency", http.StatusBadRequest)
		return
	}

	mode := ""
	if len(parts) > 1 {
		mode = parts[1]
	}

	if err := s.setFreqMode(hz, mode); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK")) //nolint:errcheck
}

// ListenAndServe starts the QSY server on the given address.
func (s *Server) ListenAndServe(addr string) error {
	return http.ListenAndServe(addr, s)
}

// ListenAndServePolyglot accepts both plain HTTP and TLS on the same port.
func (s *Server) ListenAndServePolyglot(addr, certFile, keyFile string) error {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	tlsCert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return err
	}
	tlsCfg := &tls.Config{Certificates: []tls.Certificate{tlsCert}}

	m := cmux.New(ln)
	tlsL := m.Match(cmux.TLS())
	httpL := m.Match(cmux.HTTP1Fast(), cmux.HTTP2())

	go http.Serve(httpL, s)            //nolint:errcheck
	go http.Serve(tls.NewListener(tlsL, tlsCfg), s) //nolint:errcheck

	return m.Serve()
}
