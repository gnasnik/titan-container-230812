package handler

import (
	"context"
	"github.com/filecoin-project/go-jsonrpc/auth"
	logging "github.com/ipfs/go-log/v2"
	"net/http"
	"strings"
)

var log = logging.Logger("handler")

type (
	// RemoteAddr client address
	RemoteAddr struct{}
)

// Handler represents an HTTP handler that also adds remote client address and node ID to the request context
type Handler struct {
	handler *auth.Handler
}

// GetRemoteAddr returns the remote address of the client
func GetRemoteAddr(ctx context.Context) string {
	v, ok := ctx.Value(RemoteAddr{}).(string)
	if !ok {
		return ""
	}
	return v
}

// New returns a new HTTP handler with the given auth handler and additional request context fields
func New(handler *auth.Handler) http.Handler {
	return &Handler{handler: handler}
}

// ServeHTTP serves an HTTP request with the added client remote address and node ID in the request context
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	remoteAddr := getClientIP(r)
	if remoteAddr == "" {
		remoteAddr = r.RemoteAddr
	}

	ctx := r.Context()
	ctx = context.WithValue(ctx, RemoteAddr{}, remoteAddr)

	h.handler.Next(w, r.WithContext(ctx))
}

func getClientIP(req *http.Request) string {
	clientIP := req.Header.Get("X-Forwarded-For")
	clientIP = strings.TrimSpace(strings.Split(clientIP, ",")[0])
	if clientIP == "" {
		clientIP = strings.TrimSpace(req.Header.Get("X-Real-Ip"))
	}
	if clientIP != "" {
		return clientIP
	}
	return ""
}
