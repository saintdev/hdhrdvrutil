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
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type MkvMerge struct {
	stdout   io.Writer
	stderr   io.Writer
	stdin    io.Reader
	input    string
	output   string
	tags     *Tags
	tempFile *os.File
	webm     bool
	Quiet    bool
	Verbose  bool
}

func New() *MkvMerge {
	m := new(MkvMerge)

	m.stdout = os.Stdout
	m.stderr = os.Stderr
	m.stdin = nil

	m.Quiet = true

	return m
}

func (m *MkvMerge) SetStdout(w io.Writer) {
	m.stdout = w
}

func (m *MkvMerge) SetStderr(w io.Writer) {
	m.stderr = w
}

func (m *MkvMerge) SetStdin(r io.Reader) {
	m.stdin = r
}

func (m *MkvMerge) SetInput(input string) error {
	//TODO: Verify input exists and is readable
	m.input = input

	return nil
}

func (m *MkvMerge) SetOutput(output string) error {
	//TODO: Verify dir(output) exists and is writable
	m.output = output

	return nil
}

func (m *MkvMerge) Exec() error {
	command, err := exec.LookPath("mkvmerge")
	if err != nil {
		return err
	}

	args := []string{
		"--output",
		m.output,
	}

	if m.Quiet {
		args = append(args, "--quiet")
	}

	if m.Verbose {
		args = append(args, "--verbose")
	}

	if m.tags != nil {
		m.tempFile, err = ioutil.TempFile("", filepath.Base(os.Args[0]))
		if err != nil {
			return err
		}

		m.tags.encode(m.tempFile)

		if err = m.tempFile.Close(); err != nil {
			return err
		}

		args = append(args, "--global-tags", m.tempFile.Name())
	}

	args = append(args, m.input)

	c := exec.Command(command, args...)

	c.Stderr = m.stderr
	c.Stdin = m.stdin
	c.Stdout = m.stdout

	log.Printf("%s %s", command, strings.Join(args, " "))

	return c.Run()
}

func (m *MkvMerge) Close() error {
	fileName := m.tempFile.Name()
	m.tempFile = nil
	return os.Remove(fileName)
}

func (m *MkvMerge) SetTitleTag(title string) {
	if m.tags == nil {
		m.tags = newTags()
	}
	m.tags.setTitle(title)
}

func (m *MkvMerge) SetSubTitleTag(subtitle string) {
	if m.tags == nil {
		m.tags = newTags()
	}
	m.tags.setSubTitle(subtitle)
}

func (m *MkvMerge) SetSeasonTag(season int) {
	if m.tags == nil {
		m.tags = newTags()
	}
	m.tags.setSeason(season)
}

func (m *MkvMerge) SetEpisodeTag(episode int) {
	if m.tags == nil {
		m.tags = newTags()
	}
	m.tags.setEpisode(episode)
}

func (m *MkvMerge) SetSynopsisTag(synopsis string) {
	if m.tags == nil {
		m.tags = newTags()
	}
	m.tags.setSynopsis(synopsis)
}
