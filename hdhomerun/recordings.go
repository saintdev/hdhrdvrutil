package hdhomerun

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"

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
		log.Print("Error: Unable to open file", r.Filename, err)
		return err
	}
	defer file.Close()

	tsfile := ts.NewPktStreamReader(bufio.NewReader(file))

	var buf [ts.PktLen]byte
	var jsonBuf []byte

	for i := 0; i < 64; i++ {
		pkt := ts.AsPkt(buf[:])
		if err := tsfile.ReadPkt(pkt); err != nil {
			log.Print("Error: Unable to read TS packet", err)
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
		log.Print("Error: Parsing JSON", err)
		return err
	}

	if r.EpisodeString != nil {
		if _, err = fmt.Sscanf(*r.EpisodeString, "S%dE%d", &r.Season, &r.Episode); err != nil {
			log.Print("Error: Parsing EpisodeString", err)
			return err
		}
	}

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
