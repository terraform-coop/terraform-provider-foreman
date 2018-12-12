package api

import (
	"net/url"
)

// Server definition.  For this provider, the server represents the Foreman
// API handler.  The client will direct all API requests to this server based
// on the client and server configuration options.
type Server struct {
	// The URL of the API gateway
	URL url.URL
}
