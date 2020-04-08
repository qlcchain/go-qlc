package privacy

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

type Transport interface {
	http.RoundTripper
	SetFakeMode(mode bool)
}

type Client struct {
	scheme     string // http+unix, http
	rootPath   string
	transport  Transport
	httpClient *http.Client
}

func NewClient(ptmNode string) *Client {
	ptmNode = strings.TrimSpace(ptmNode)
	ptmNode = strings.TrimRight(ptmNode, "/")

	parts := strings.Split(ptmNode, ":")
	if len(parts) < 2 {
		return nil
	}
	scheme := parts[0]
	if scheme == "unix" {
		scheme = "http+unix"
	}
	rootPath := ptmNode[len(parts[0])+1:]

	c := &Client{scheme: scheme, rootPath: rootPath}
	c.httpClient = &http.Client{}

	if c.scheme == "http+unix" {
		c.transport = newUnixTransport("c", c.rootPath)
	} else if c.scheme == "http" {
		c.transport = newHttpTransport()
	} else {
		return nil
	}
	c.httpClient.Transport = c.transport

	return c
}

func (c *Client) formatPath(path string) string {
	if c.scheme == "http+unix" {
		return c.scheme + "://c" + path
	}
	return c.scheme + ":" + c.rootPath + path
}

func (c *Client) Upcheck() (bool, error) {
	res, err := c.httpClient.Get(c.formatPath("/upcheck"))
	if err != nil {
		return false, err
	}
	if res.StatusCode == 200 {
		return true, nil
	}

	return false, errors.New("PTM did not respond to upcheck request")
}

func (c *Client) SendPayload(pl []byte, b64From string, b64To []string) ([]byte, error) {
	buf := bytes.NewBuffer(pl)
	req, err := http.NewRequest("POST", c.formatPath("/sendraw"), buf)
	if err != nil {
		return nil, err
	}
	if b64From != "" {
		req.Header.Set("c11n-from", b64From)
	}
	req.Header.Set("c11n-to", strings.Join(b64To, ","))
	req.Header.Set("Content-Type", "application/octet-stream")
	res, err := c.httpClient.Do(req)

	if res != nil {
		defer res.Body.Close()
	}
	if err != nil {
		return nil, err
	}
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("non-200 status code: %d(%s)", res.StatusCode, res.Status)
	}

	return ioutil.ReadAll(base64.NewDecoder(base64.StdEncoding, res.Body))
}

func (c *Client) ReceivePayload(key []byte) ([]byte, error) {
	req, err := http.NewRequest("GET", c.formatPath("/receiveraw"), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("c11n-key", base64.StdEncoding.EncodeToString(key))
	res, err := c.httpClient.Do(req)

	if res != nil {
		defer res.Body.Close()
	}

	if err != nil {
		return nil, err
	}

	if res.StatusCode != 200 {
		if res.StatusCode == 404 {
			return nil, nil
		}
		return nil, fmt.Errorf("non-200 status code: %d(%s)", res.StatusCode, res.Status)
	}

	return ioutil.ReadAll(res.Body)
}
