// Copyright © 2023 Mark Summerfield. All rights reserved.
// License: Apache-2.0

package main

import (
	"archive/tar"
	"compress/bzip2"
	"compress/gzip"
	_ "embed"
	"fmt"
	"github.com/mark-summerfield/clip"
	"github.com/mark-summerfield/gong"
	"github.com/ulikunitz/xz"
	"io"
	"log"
	"os"
	"strings"
)

//go:embed Version.dat
var Version string

func main() {
	log.SetFlags(0)
	verbose, list, archives := getConfig()
	for _, archive := range archives {
		if list {
			listContents(archive)
		} else {
			unpack(archive, verbose)
		}
	}
}

func getConfig() (bool, bool, []string) {
	parser := clip.NewParserUser("unz", Version)
	parser.LongDesc = `Unpacks (or lists) each archive (.tar, .tar.gz,
	.tar.bz2, .tar.xz, .tgz, or .zip).

	When unpacking (the default behavior), for each archive at most one file
	or folder is created in the current folder. If the archive contains one
	file or folder, that file or folder is unpacked into the current folder.
	If the archive contains more than one member, then a new subfolder is
	created based on the archive's name, and all the archive's contents are
	unpacked into the subfolder.`
	parser.PositionalCount = clip.OneOrMorePositionals
	_ = parser.SetPositionalVarName("ARCHIVE")
	verboseOpt := parser.Flag("verbose", "Show actions if unpacking.")
	listOpt := parser.Flag("list",
		"List each archive's contents (don't unpack).")
	err := parser.Parse()
	if err != nil {
		log.Fatal(gong.Underline(fmt.Sprintf("%s\n", err)))
	}
	return verboseOpt.Value(), listOpt.Value(), parser.Positionals
}

func listContents(archive string) {
	if isTarball(archive) {
		listTarball(archive)
	} else {
		listZip(archive)
	}
}

func listTarball(archive string) {
	file, err := os.Open(archive)
	if err != nil {
		log.Fatal(gong.Underline(fmt.Sprintf("failed to open %s: %s",
			archive, err)))
	}
	defer file.Close()
	var reader *tar.Reader
	uarchive := strings.ToUpper(archive)
	if strings.HasSuffix(uarchive, ".GZ") || strings.HasSuffix(uarchive,
		".TGZ") {
		ufile, err := gzip.NewReader(file)
		if err != nil {
			log.Println(gong.Underline(fmt.Sprintf("failed to open %s: %s",
				archive, err)))
			return
		}
		defer ufile.Close()
		reader = tar.NewReader(ufile)
	} else if strings.HasSuffix(uarchive, ".BZ2") {
		ufile := bzip2.NewReader(file)
		reader = tar.NewReader(ufile)
	} else if strings.HasSuffix(uarchive, ".XZ") {
		ufile, err := xz.NewReader(file)
		if err != nil {
			log.Println(gong.Underline(fmt.Sprintf("failed to open %s: %s",
				archive, err)))
			return
		}
		reader = tar.NewReader(ufile)
	} else {
		reader = tar.NewReader(file)
	}
	fmt.Println(gong.Bold(archive))
	for {
		header, err := reader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Println(gong.Underline(fmt.Sprintf(
				"failed to read from %s: %s", archive, err)))
		}
		fmt.Println(header.Name)
	}
}

func listZip(archive string) {
	fmt.Println("listZip", archive) // TODO
}

func unpack(archive string, verbose bool) {
	if isTarball(archive) {
		unpackTarball(archive, verbose)
	} else {
		unpackZip(archive, verbose)
	}
}

func unpackTarball(archive string, verbose bool) {
	fmt.Println("unpackTarball", archive, verbose) // TODO
}

func unpackZip(archive string, verbose bool) {
	fmt.Println("unpackZip", archive, verbose) // TODO
}

func isTarball(name string) bool {
	name = strings.ToUpper(name)
	return strings.HasSuffix(name, ".TAR") ||
		strings.HasSuffix(name, ".TGZ") || strings.Contains(name, ".TAR.")
}
