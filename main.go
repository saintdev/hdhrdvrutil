package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"

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
	var recordingfiles []hdhomerun.RecordingFile
	var recordings []hdhomerun.Recording
	a := App{}

	recordingfileidmap := map[string]int{}

	validateOptions(&a)

	files, err := filepath.Glob(filepath.Join(a.srcDir, "*.mpg"))
	if err != nil {
		log.Fatalln("Bad glob pattern")
	}

	for i, filename := range files {
		r := hdhomerun.RecordingFile{Filename: &filename}

		fmt.Println(r.Filename)

		if err := r.Parse(); err != nil {
			log.Println("Failed to parse", r.Filename)
			continue
		}

		recordingfiles = append(recordingfiles, r)
		recordingfileidmap[*r.ProgramID] = i
	}

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

	for _, r := range recordings {
		i, ok := recordingfileidmap[*r.ProgramID]
		if !ok {
			continue
		}
		file := &recordingfiles[i]

		file.CmdURL = r.CmdURL

		copyToMkv(file, a.destDir)

		if a.delete {
			if err = hdhomerunClient.Recordings.Delete((*hdhomerun.Recording)(file), false); err != nil {
				log.Println("Failed to delete recording:", file.Filename)
			}
		}
	}
}

func copyToMkv(f *hdhomerun.RecordingFile, destdir string) {
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

	mkvcmd.Quiet = false

	if err := mkvcmd.Exec(); err != nil {
		log.Fatalln("Failed to exec mkvmerge cmd", err)
	}
	defer mkvcmd.Close()
}
