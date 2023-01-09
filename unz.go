// Copyright © 2023 Mark Summerfield. All rights reserved.
// License: Apache-2.0

package main

import (
	"archive/tar"
	"archive/zip"
	"compress/bzip2"
	"compress/gzip"
	_ "embed"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/mark-summerfield/clip"
	"github.com/mark-summerfield/gong"
	"github.com/ulikunitz/xz"
)

//go:embed Version.dat
var Version string

func main() {
	log.SetFlags(0)
	verbose, unpack, archives := getConfig()
	for _, archive := range archives {
		if unpack {
			unpackArchive(archive, verbose)
		} else {
			listArchive(archive, verbose)
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
	verboseOpt := parser.Flag("verbose", "Show actions.")
	listOpt := parser.Flag("list",
		"List each archive's contents (don't unpack).")
	err := parser.Parse()
	if err != nil {
		log.Fatal(gong.Underline(fmt.Sprintf("%s\n", err)))
	}
	return verboseOpt.Value(), !listOpt.Value(), parser.Positionals
}

func listArchive(archive string, verbose bool) {
	if isTarball(archive) {
		listTarball(archive, verbose)
	} else {
		listZip(archive, verbose)
	}
}

func listTarball(archive string, verbose bool) {
	names := tarballNames(archive)
	if verbose {
		fmt.Print(gong.Bold(archive))
		n := len(names)
		fmt.Printf(" (%s member%s)\n", commas(n), s(n))
	} else {
		fmt.Println(archive)
	}
	for _, name := range names {
		fmt.Println(name)
	}
}

func tarballNames(archive string) []string {
	names := []string{}
	reader, closer := openTarball(archive)
	if reader == nil {
		return names
	}
	defer closer()
	for {
		header, err := reader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Println(gong.Underline(fmt.Sprintf(
				"failed to read from %s: %s", archive, err)))
		} else {
			names = append(names, header.Name)
		}
	}
	return names
}

func listZip(archive string, verbose bool) {
	names := zipNames(archive)
	if verbose {
		fmt.Print(gong.Bold(archive))
		n := len(names)
		fmt.Printf(" (%s member%s)\n", commas(n), s(n))
	} else {
		fmt.Println(archive)
	}
	for _, name := range names {
		fmt.Println(name)
	}
}

func zipNames(archive string) []string {
	names := []string{}
	reader, err := zip.OpenReader(archive)
	if err != nil {
		log.Println(gong.Underline(fmt.Sprintf(
			"failed to open from %s: %s", archive, err)))
		return names
	}
	defer reader.Close()
	for _, member := range reader.File {
		names = append(names, member.Name)
	}
	return names
}

func unpackArchive(archive string, verbose bool) {
	fmt.Println("TODO unpackArchive", archive, verbose)
}

func isTarball(name string) bool {
	name = strings.ToUpper(name)
	return strings.HasSuffix(name, ".TAR") ||
		strings.HasSuffix(name, ".TGZ") || strings.Contains(name, ".TAR.")
}

type closer func()

func openTarball(archive string) (*tar.Reader, closer) {
	file, err := os.Open(archive)
	if err != nil {
		log.Println(gong.Underline(fmt.Sprintf("failed to open %s: %s",
			archive, err)))
		return nil, nil
	}
	var reader *tar.Reader
	var closer closer
	uarchive := strings.ToUpper(archive)
	if strings.HasSuffix(uarchive, ".GZ") || strings.HasSuffix(uarchive,
		".TGZ") {
		ufile, err := gzip.NewReader(file)
		if err != nil {
			log.Println(gong.Underline(fmt.Sprintf("failed to open %s: %s",
				archive, err)))
			return nil, nil
		}
		closer = func() {
			ufile.Close()
			file.Close()
		}
		reader = tar.NewReader(ufile)
	} else if strings.HasSuffix(uarchive, ".BZ2") {
		ufile := bzip2.NewReader(file)
		reader = tar.NewReader(ufile)
	} else if strings.HasSuffix(uarchive, ".XZ") {
		ufile, err := xz.NewReader(file)
		if err != nil {
			log.Println(gong.Underline(fmt.Sprintf("failed to open %s: %s",
				archive, err)))
			return nil, nil
		}
		reader = tar.NewReader(ufile)
	} else {
		reader = tar.NewReader(file)
	}
	if closer == nil {
		closer = func() { file.Close() }
	}
	return reader, closer
}

func s(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}

func commas(i int) string {
	pos := true
	s := strconv.Itoa(i)
	if s[0] == '-' {
		pos = false
		s = s[1:]
	}
	n := len(s) - 3
	for n >= 0 {
		s = s[:n] + "," + s[n:]
		n -= 3
	}
	if s[0] == ',' {
		s = s[1:]
	}
	if !pos {
		s = "-" + s
	}
	return s
}
