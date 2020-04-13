package privacy

import (
	"context"
	"net/http"
	"os"
	"testing"
)

func TestUnixTransport_RoundTrip(t *testing.T) {
	ipcFile, _ := os.Create("/tmp/httpunix-test.ipc")
	defer func() {
		_ = ipcFile.Close()
	}()
	_ = os.Remove("/tmp/httpunix-test.ipc")

	u := newUnixTransport("t", "/tmp/httpunix-test.ipc")
	u.SetFakeMode(true)

	_, _ = u.dialContext(context.Background(), "tcp", "c:80")

	req, _ := http.NewRequest("GET", "http+unix://t/send", nil)
	_, _ = u.RoundTrip(req)
}
