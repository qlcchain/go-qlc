package privacy

import (
	"net/http"
)

func newHttpTransport() http.RoundTripper {
	return http.DefaultTransport
}
