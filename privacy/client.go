package privacy

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/tv42/httpunix"
)

func unixTransport(socketPath string) *httpunix.Transport {
	t := &httpunix.Transport{
		DialTimeout:           1 * time.Second,
		RequestTimeout:        5 * time.Second,
		ResponseHeaderTimeout: 5 * time.Second,
	}
	t.RegisterLocation("c", socketPath)
	return t
}

func unixClient(socketPath string) *http.Client {
	return &http.Client{
		Transport: unixTransport(socketPath),
	}
}

type Client struct {
	httpClient *http.Client
}

func NewClient(socketPath string) *Client {
	return &Client{
		httpClient: unixClient(socketPath),
	}
}

func (c *Client) doJson(path string, apiReq interface{}) (*http.Response, error) {
	buf := new(bytes.Buffer)
	err := json.NewEncoder(buf).Encode(apiReq)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", "http+unix://c/"+path, buf)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := c.httpClient.Do(req)
	if err == nil && res.StatusCode != 200 {
		return nil, fmt.Errorf("Non-200 status code: %+v", res)
	}
	return res, err
}

func (c *Client) Upcheck() (bool, error) {
	res, err := c.httpClient.Get("http+unix://c/upcheck")
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
	req, err := http.NewRequest("POST", "http+unix://c/sendraw", buf)
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
		return nil, fmt.Errorf("Non-200 status code: %+v", res)
	}

	return ioutil.ReadAll(base64.NewDecoder(base64.StdEncoding, res.Body))
}

func (c *Client) SendSignedPayload(signedPayload []byte, b64To []string) ([]byte, error) {
	buf := bytes.NewBuffer(signedPayload)
	req, err := http.NewRequest("POST", "http+unix://c/sendsignedtx", buf)
	if err != nil {
		return nil, err
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
		return nil, fmt.Errorf("Non-200 status code: %+v", res)
	}

	return ioutil.ReadAll(base64.NewDecoder(base64.StdEncoding, res.Body))
}

func (c *Client) ReceivePayload(key []byte) ([]byte, error) {
	req, err := http.NewRequest("GET", "http+unix://c/receiveraw", nil)
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
