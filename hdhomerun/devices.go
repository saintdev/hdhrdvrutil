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
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
)

const (
	defaultBaseURL = "http://ipv4-api.hdhomerun.com"
)

type DeviceService struct {
	client *Client
}

type Device struct {
	StorageID   *string
	LocalIP     *string
	BaseURL     *string
	DiscoverURL *string
	StorageURL  *string
}

func (s *DeviceService) Discover() ([]*Device, error) {
	var devices []*Device

	u, _ := url.Parse(defaultBaseURL)
	u, err := u.Parse("/discover")
	if err != nil {
		return nil, err
	}

	_, err = s.client.Get(u, &devices)
	if err != nil {
		log.Printf("Error parsing discover JSON: %v\n", err)
		return nil, err
	}

	return devices, nil
}

func (s *DeviceService) RecordedFiles(device *Device) ([]*Recording, error) {
	var recordings []*Recording
	if device.IsRecordEngine() == false {
		return nil, errors.New("Not a RECORD device")
	}

	u, err := url.Parse(*device.BaseURL)
	if err != nil {
		return nil, err
	}
	if u, err = u.Parse("/recorded_files.json"); err != nil {
		return nil, err
	}

	response, err := s.client.Get(u, &recordings)
	if err != nil {
		return nil, err
	} else if response.StatusCode != http.StatusOK {
		return nil, errors.New(response.Status)
	}

	for _, r := range recordings {
		if r.EpisodeString != nil {
			if _, err = fmt.Sscanf(*r.EpisodeString, "S%dE%d", &r.Season, &r.Episode); err != nil {
				log.Printf("Error parsing EpisodeString %q: %v\n", *r.EpisodeString, err)
				return nil, err
			}
		}
	}

	return recordings, nil
}

func (d *Device) IsRecordEngine() bool {
	// FIXME: Come up with a better test
	if d.StorageID != nil && d.StorageURL != nil {
		return true
	}

	return false
}
