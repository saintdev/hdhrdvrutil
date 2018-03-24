package hdhomerun

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

type Client struct {
	httpClient *http.Client

	Devices    *DeviceService
	Recordings *RecordingService
}

func NewClient(httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	c := &Client{httpClient: httpClient}
	c.Devices = &DeviceService{client: c}
	c.Recordings = &RecordingService{client: c}

	return c
}

func (c *Client) Get(u *url.URL, val interface{}) (*http.Response, error) {
	request, err := c.newRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}

	response, err := c.do(request, val)
	return response, err
}

func (c *Client) newRequest(method string, u *url.URL, body interface{}) (*http.Request, error) {
	log.Println(u)

	var buf io.ReadWriter
	if body != nil {
		buf = new(bytes.Buffer)

		if err := json.NewEncoder(buf).Encode(body); err != nil {
			return nil, err
		}
	}

	request, err := http.NewRequest(method, u.String(), buf)
	if err != nil {
		return nil, err
	}

	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	//	request.Header.Set("Accept", "application/json")
	request.Header.Set("User-Agent", "hdhrdvrutil v0.0.1") //FIXME: Use real version here

	return request, nil
}

func (c *Client) do(request *http.Request, val interface{}) (*http.Response, error) {
	response, err := c.httpClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer func() {
		io.CopyN(ioutil.Discard, response.Body, 512)
		response.Body.Close()
	}()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return response, fmt.Errorf("Bad HTTP Response: %v", response.StatusCode)
	}

	// buf := new(bytes.Buffer)
	// io.Copy(buf, response.Body)
	// log.Println(buf.String())
	if val != nil {
		err = json.NewDecoder(response.Body).Decode(val)
		// err = json.NewDecoder(buf).Decode(val)
	}
	return response, err
}
