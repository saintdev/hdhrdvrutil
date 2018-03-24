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
