package privacy

import (
	"net/http"
)

type HttpTransport struct {
	transport http.RoundTripper
	fakeMode  bool
}

func newHttpTransport() *HttpTransport {
	t := &HttpTransport{transport: http.DefaultTransport}
	return t
}

func (t *HttpTransport) SetFakeMode(mode bool) {
	t.fakeMode = mode
}

func (t *HttpTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.fakeMode {
		rsp := new(http.Response)
		rsp.StatusCode = 200
		rsp.Body = http.NoBody
		return rsp, nil
	}
	return t.transport.RoundTrip(req)
}
