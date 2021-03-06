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

package cmd

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"syscall"

	"github.com/saintdev/hdhrdvrutil/hdhomerun"
	"github.com/saintdev/hdhrdvrutil/mkvmerge"

	"github.com/gosimple/slug"
	"github.com/spf13/cobra"
)

var archiveCmd = &cobra.Command{
	Use:   "archive SRC DEST",
	Short: "Archive recordings",
	Long:  `Remux and archive recording files from srcdir into destdir.`,
	Args:  cobra.ExactArgs(2),
	Run:   archiveMain,
}

var (
	delete  = false
	srcDir  = ""
	destDir = ""
)

func init() {
	rootCmd.AddCommand(archiveCmd)

	archiveCmd.Flags().BoolVarP(&delete, "delete", "", false, "Delete recordings after archiving")
}

func validateDirs(args []string) {
	var err error
	srcDir, err = filepath.Abs(args[0])
	if err != nil {
		log.Fatalf("Unable to construct absolute srcdir path %q: %v", args[0], err)
	}
	if _, err := os.Stat(srcDir); os.IsNotExist(err) {
		log.Fatalf("Path does not exist %q", srcDir)
	}

	destDir, err = filepath.Abs(args[1])
	if err != nil {
		log.Fatalf("Unable to construct absolute destdir path %q: %v", args[1], err)
	}
	if _, err := os.Stat(destDir); os.IsNotExist(err) {
		log.Fatalf("Path does not exist %q", destDir)
	}
}

func archiveMain(cmd *cobra.Command, args []string) {
	var recordings []*hdhomerun.Recording

	validateDirs(args)

	log.Printf("Source: %q\n", srcDir)
	log.Printf("Destination: %q\n", destDir)

	dvrClient := hdhomerun.NewClient(nil)

	devices, err := dvrClient.Devices.Discover()
	if err != nil {
		log.Fatalf("Unable to discover devices: %v\n", err)
	}

	for _, device := range devices {
		if !device.IsRecordEngine() {
			continue
		}

		recordings, err = dvrClient.Devices.RecordedFiles(device)
		if err != nil {
			log.Printf("Failed to parse `recorded_files.json` for device at %q: %v\n", *device.BaseURL, err)
			continue
		}
	}

	if len(recordings) == 0 {
		log.Fatalln("No recordings found!")
	}

	if err = dvrClient.Recordings.ScanRecordingsDir(srcDir, recordings); err != nil {
		log.Fatalf("Error scanning recordings in %q: %v\n", srcDir, err)
	}

	for _, r := range recordings {
		if r.Filename == nil {
			continue
		}

		copyToMkv(r, destDir)

		if delete {
			if err = dvrClient.Recordings.Delete(r, false); err != nil {
				log.Printf("Failed to delete recording %q: %v\n", r.Filename, err)
			}
		}
	}
}

func copyToMkv(f *hdhomerun.Recording, destdir string) {
	var filename string

	mkvcmd := mkvmerge.New()
	mkvcmd.SetInput(*f.Filename)
	if f.EpisodeTitle == nil {
		filename = fmt.Sprintf("%s", *f.Title)
	} else if f.EpisodeString == nil {
		filename = fmt.Sprintf("%s", *f.EpisodeTitle)
	} else {
		filename = fmt.Sprintf("%02d%02d-%s", f.Season, f.Episode, *f.EpisodeTitle)
	}
	mkvcmd.SetOutput(path.Join(destdir, fmt.Sprintf("%s.mkv", slug.Make(filename))))

	if f.EpisodeString != nil {
		mkvcmd.SetEpisodeTag(f.Episode)
		mkvcmd.SetSeasonTag(f.Season)
	}
	if f.EpisodeTitle != nil {
		mkvcmd.SetSubTitleTag(*f.EpisodeTitle)
	}
	if f.Synopsis != nil {
		mkvcmd.SetSynopsisTag(*f.Synopsis)
	}
	mkvcmd.SetTitleTag(*f.Title)

	mkvcmd.Quiet = true

	if err := mkvcmd.Exec(); err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			waitStatus := exitError.Sys().(syscall.WaitStatus)
			if waitStatus.ExitStatus() != 1 {
				log.Fatalf("Failed to exec mkvmerge: %v", err)
			}
		}
	}
	defer mkvcmd.Close()
}
