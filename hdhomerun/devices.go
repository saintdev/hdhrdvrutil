package hdhomerun

/*
[
    {
        "DeviceID": "13105323",
        "LocalIP": "10.20.20.187",
        "ConditionalAccess": 1,
        "BaseURL": "http://10.20.20.187:80",
        "DiscoverURL": "http://10.20.20.187:80/discover.json",
        "LineupURL": "http://10.20.20.187:80/lineup.json"
    },
    {
        "DeviceID": "1315F9E1",
        "LocalIP": "10.20.20.197",
        "ConditionalAccess": 1,
        "BaseURL": "http://10.20.20.197:80",
        "DiscoverURL": "http://10.20.20.197:80/discover.json",
        "LineupURL": "http://10.20.20.197:80/lineup.json"
    },
    {
        "StorageID": "a7bf100b-e3c1-76bb-47fa-340d3319bc72",
        "LocalIP": "10.20.20.68:34561",
        "BaseURL": "http://10.20.20.68:34561",
        "DiscoverURL": "http://10.20.20.68:34561/discover.json",
        "StorageURL": "http://10.20.20.68:34561/recorded_files.json"
    }
]
*/

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
		log.Println("Error parsing discover JSON: ", err)
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
				log.Print("Error: Parsing EpisodeString", err)
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
