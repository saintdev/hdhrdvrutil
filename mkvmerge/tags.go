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

package mkvmerge

import (
	"encoding/xml"
	"fmt"
	"io"
)

type TargetTypeValue uint

const (
	Collection TargetTypeValue = 70
	Season     TargetTypeValue = 60
	Episode    TargetTypeValue = 50
	Chapter    TargetTypeValue = 30
)

type SimpleTag struct {
	Name       string
	String     string      `xml:",omitempty"`
	Binary     []byte      `xml:",omitempty"`
	SimpleTags []SimpleTag `xml:"Simple,omitempty"`
}

type Target struct {
	TrackUIDs       []uint `xml:"TrackUID,omitempty"`
	ChapterUIDs     []uint `xml:"ChapterUID,omitempty"`
	TargetTypeValue `xml:",omitempty"`
	TargetType      string `xml:",omitempty"`
}

type Tag struct {
	Target     *Target     `xml:"Targets,omitempty"`
	SimpleTags []SimpleTag `xml:"Simple"`
}

type Tags struct {
	XMLName xml.Name `xml:"Tags"`
	Tags    []Tag    `xml:"Tag"`
	tagMap  map[TargetTypeValue]int
}

func (t *Tags) encode(w io.Writer) error {
	xmlheader := []byte(xml.Header +
		"<!DOCTYPE Tags SYSTEM \"matroskatags.dtd\">\n")
	if _, err := w.Write(xmlheader); err != nil {
		return err
	}

	encoder := xml.NewEncoder(w)

	return encoder.Encode(t)
}

func newTags() *Tags {
	t := &Tags{
		tagMap: map[TargetTypeValue]int{},
	}

	t.Tags = append(t.Tags, Tag{})
	t.tagMap[Episode] = 0

	return t
}

func (t *Tags) setTitle(title string) {
	i := t.tagMap[Episode]
	tag := &t.Tags[i]
	tag.SimpleTags = append(tag.SimpleTags, SimpleTag{Name: "TITLE", String: title})
}

func (t *Tags) setSubTitle(subtitle string) {
	i := t.tagMap[Episode]
	tag := &t.Tags[i]
	tag.SimpleTags = append(tag.SimpleTags, SimpleTag{Name: "SUBTITLE", String: subtitle})
}

func (t *Tags) setSynopsis(synopsis string) {
	i := t.tagMap[Episode]
	tag := &t.Tags[i]
	tag.SimpleTags = append(tag.SimpleTags, SimpleTag{Name: "SYNOPSIS", String: synopsis})
}

func (t *Tags) setEpisode(episode int) {
	i := t.tagMap[Episode]
	tag := &t.Tags[i]
	tag.SimpleTags = append(tag.SimpleTags, SimpleTag{Name: "PART_NUMBER", String: fmt.Sprint(episode)})
}

func (t *Tags) setSeason(season int) {
	i, ok := t.tagMap[Season]
	if !ok {
		tag := Tag{}
		tag.Target = new(Target)
		tag.Target.TargetTypeValue = Season
		t.Tags = append(t.Tags, tag)
		i = len(t.Tags) - 1
		t.tagMap[Season] = i
	}
	tag := &t.Tags[i]

	tag.SimpleTags = append(tag.SimpleTags, SimpleTag{Name: "PART_NUMBER", String: fmt.Sprint(season)})
}
