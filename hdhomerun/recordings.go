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
	"bufio"
	"bytes"
	"encoding/json"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/ziutek/dvb/ts"
)

type RecordingService struct {
	client *Client
}

type Recording struct {
	Category      *string
	CmdURL        *string
	EpisodeTitle  *string
	EpisodeString *string `json:"EpisodeNumber"`
	ImageURL      *string
	ProgramID     *string
	SeriesID      *string
	Synopsis      *string
	Title         *string
	Filename      *string
	Season        int
	Episode       int
}

type RecordingFile Recording

func (r *RecordingFile) Parse() error {
	file, err := os.Open(*r.Filename)
	if err != nil {
		log.Printf("Error: Unable to open file %q: %v\n", r.Filename, err)
		return err
	}
	defer file.Close()

	tsfile := ts.NewPktStreamReader(bufio.NewReader(file))

	var buf [ts.PktLen]byte
	var jsonBuf []byte

	for i := 0; i < 64; i++ {
		pkt := ts.AsPkt(buf[:])
		if err := tsfile.ReadPkt(pkt); err != nil {
			log.Printf("Error: Unable to read TS packet: %v\n", err)
			continue
		}
		if pkt.Pid() != 0x1FFA {
			break
		}
		payload := pkt.Payload()
		jsonBuf = append(jsonBuf, payload...)
	}

	jsonBuf = bytes.Trim(jsonBuf, "\xFF")

	if err = json.Unmarshal(jsonBuf, &r); err != nil {
		log.Printf("Error parsing TS packet JSON: %v\n", err)
		return err
	}

	return nil
}

//FIXME: This needs a better name
func (s *RecordingService) ScanRecordingsDir(dir string, recordings []*Recording) error {
	episodeMap := map[string]*Recording{}

	dir, err := filepath.Abs(dir)
	if err != nil {
		return err
	}

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return err
	}

	for i := range recordings {
		episodeMap[*recordings[i].ProgramID] = recordings[i]
	}

	err = filepath.Walk(dir, func(path string, finfo os.FileInfo, err error) error {
		if err != nil {
			log.Printf("Failed to access path %q: %v\n", dir, err)
			return err
		}

		if finfo.IsDir() && strings.HasPrefix(finfo.Name(), ".") {
			return filepath.SkipDir
		}

		if !finfo.IsDir() && filepath.Ext(finfo.Name()) == ".mpg" {
			file := &RecordingFile{Filename: &path}
			if err := file.Parse(); err != nil {
				return err
			}

			r, ok := episodeMap[*file.ProgramID]
			if !ok {
				return nil
			}
			r.Filename = &path
		}

		return nil
	})

	return nil
}

func (s *RecordingService) Delete(recording *Recording, rerecord bool) error {
	u, err := url.Parse(*recording.CmdURL)
	if err != nil {
		return err
	}
	q := u.Query()
	q.Set("cmd", "delete")
	if rerecord {
		q.Set("rerecord", "1")
	}
	u.RawQuery = q.Encode()

	_, err = s.client.Get(u, nil)

	return err
}
