// Copyright © 2018 Nathan Caldwell <saintdev@gmail.com>
//
// This program is free software; you can redistribute it and/or
// modify it under the terms of the GNU General Public License
// as published by the Free Software Foundation; either version 2
// of the License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

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
	log.Printf("%s\t%s\n", method, u.String())

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

	if val != nil {
		err = json.NewDecoder(response.Body).Decode(val)
	}
	return response, err
}
