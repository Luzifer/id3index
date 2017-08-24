package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/Luzifer/rconfig"
	"github.com/bogem/id3v2"
	log "github.com/sirupsen/logrus"
)

var (
	cfg struct {
		LogLevel       string `flag:"log-level" default:"info" description:"Set log level (debug, info, warning, error)"`
		Output         string `flag:"output,o" default:"-" description:"Output to write to (Filename or '-' for StdOut)"`
		VersionAndExit bool   `flag:"version" default:"false" description:"Print version information and exit"`
	}

	product = "id3index"
	version = "dev"

	output io.Writer
)

func init() {
	if err := rconfig.ParseAndValidate(&cfg); err != nil {
		log.Fatalf("Error parsing CLI arguments: %s", err)
	}

	if l, err := log.ParseLevel(cfg.LogLevel); err == nil {
		log.SetLevel(l)
	} else {
		log.Fatalf("Invalid log level: %s", err)
	}

	if cfg.VersionAndExit {
		fmt.Printf("%s %s\n", product, version)
		os.Exit(0)
	}

}

func main() {
	directory := "."
	if len(rconfig.Args()) > 1 {
		directory = rconfig.Args()[1]
	}

	output = os.Stdout
	if cfg.Output != "-" {
		f, err := os.Create(cfg.Output)
		if err != nil {
			log.Fatalf("Could not open output file %q: %s", cfg.Output, err)
		}
		defer f.Close()
		output = f
	}

	log.Infof("Starting file scan of directory %q", directory)

	fmt.Fprintln(output, "File path,Artist,Title,Album,Error")

	if err := filepath.Walk(directory, walk); err != nil {
		log.Fatalf("Experienced error while scanning directory: %s", err)
	}

	log.Infof("File scan finished.")
}

func walk(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}

	if !strings.HasSuffix(path, ".mp3") {
		return nil
	}

	var errorMsg string

	mp3File, err := id3v2.Open(path, id3v2.Options{Parse: true})
	switch {
	case err == nil:
	case strings.HasPrefix(err.Error(), ""):
		errorMsg = err.Error()
	default:
		return err
	}
	defer mp3File.Close()

	fmt.Fprintf(output, "\"%s\"\n", strings.Join([]string{
		path,
		mp3File.Artist(),
		mp3File.Title(),
		mp3File.Album(),
		errorMsg,
	}, "\",\""))

	return nil
}
