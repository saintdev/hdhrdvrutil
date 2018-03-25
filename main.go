package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"syscall"

	"saintdev/hdhrdvrutil/hdhomerun"
	"saintdev/hdhrdvrutil/mkvmerge"

	"github.com/gosimple/slug"
)

type App struct {
	srcDir  string
	destDir string
	delete  bool
}

func init() {

}

func usage() {
	fmt.Printf("%s srcdir_dir destdir_dir\n", filepath.Base(os.Args[0]))
}

func validateOptions(a *App) {
	var err error

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s [OPTION] SOURCE DEST\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.BoolVar(&a.delete, "delete", false, "Delete recordings after archiving")

	flag.Parse()

	if flag.NArg() != 2 {
		flag.Usage()
		os.Exit(1)
	}

	a.srcDir, err = filepath.Abs(flag.Arg(0))
	if err != nil {
		log.Fatalf("Unable to construct absolute srcdir path: %v", err)
	}
	if _, err := os.Stat(a.srcDir); os.IsNotExist(err) {
		log.Fatalf("Path does not exist: %v", a.srcDir)
	}

	a.destDir, err = filepath.Abs(flag.Arg(1))
	if err != nil {
		log.Fatalf("Unable to construct absolute destdir path: %v", err)
	}
	if _, err := os.Stat(a.destDir); os.IsNotExist(err) {
		log.Fatalf("Path does not exist %v", a.destDir)
	}

}

func main() {
	var recordings []hdhomerun.Recording
	a := App{}

	validateOptions(&a)

	hdhomerunClient := hdhomerun.NewClient(nil)

	devices, err := hdhomerunClient.Devices.Discover()
	if err != nil {
		log.Fatalln("Unable to discover devices: ", err)
	}

	for _, device := range devices {
		if !device.IsRecordEngine() {
			continue
		}

		recordings, err = hdhomerunClient.Devices.RecordedFiles(&device)
		if err != nil {
			log.Println("Failed to parse recorded files for device: ", err)
			continue
		}
	}

	if len(recordings) == 0 {
		log.Fatalln("No recordings found!")
	}

	recPtrArr := []*hdhomerun.Recording{}
	for i := range recordings {
		recPtrArr = append(recPtrArr, &recordings[i])
	}

	if err = hdhomerunClient.Recordings.ScanRecordingsDir(a.srcDir, recPtrArr); err != nil {
		log.Fatalf("Error scanning recordings: %v\n", err)
	}

	for i := range recordings {
		r := &recordings[i]
		if r.Filename == nil {
			continue
		}

		copyToMkv(r, a.destDir)

		if a.delete {
			if err = hdhomerunClient.Recordings.Delete(r, false); err != nil {
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
