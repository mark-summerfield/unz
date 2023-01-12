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
	"path/filepath"
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

func unpackArchive(archive string, verbose bool) {
	if isTarball(archive) {
		unpackTarball(archive, verbose)
	} else {
		unpackZip(archive, verbose)
	}
}

func unpackTarball(archive string, verbose bool) {
	names := tarballNames(archive)
	switch len(names) {
	case 0:
		if verbose {
			fmt.Println("no members to unpack")
		}
		return
	case 1:
		reader, closer := openTarball(archive)
		if reader == nil {
			return
		}
		defer closer()
		unpackOneTarMember(archive, reader, cwd(), verbose)
	default:
		folder := gong.LongestCommonPath(names)
		//if folder == ""
		// TODO
		fmt.Println("TODO unpackTarball", archive, verbose, folder)
	}
}

func unpackOneTarMember(archive string, reader *tar.Reader, folder string,
	verbose bool) bool {
	header, err := reader.Next()
	if err == io.EOF {
		return false // no more to do
	}
	if err != nil {
		log.Println(gong.Underline(fmt.Sprintf("failed to read %s: %s",
			archive, err)))
		return false // don't go further
	}
	name := filepath.Clean(header.Name)
	if filepath.IsAbs(name) {
		log.Printf("skipping risky absolute path member %s\n", name)
		return true // try next one
	}
	name = filepath.Join(folder, name)
	switch header.Typeflag {
	case tar.TypeDir:
		log.Printf("TODO create folder %s\n", name)
		// TODO make dir name in given folder
		if verbose {
			fmt.Printf("created folder %s\n", name)
		}
	case tar.TypeReg:
		log.Printf("TODO create file %s\n", name)
		// TODO write file name in given folder
		if verbose {
			fmt.Printf("created file %s\n", name)
		}
	case tar.TypeSymlink:
		log.Printf("TODO create soft link %s\n", name)
		// TODO create soft link
		log.Printf("skipping unsupported soft link %s\n", name)
	case tar.TypeLink:
		log.Printf("skipping unsupported hard link %s\n", name)
	default:
		log.Printf("skipping unsupported member type (device or FIFO) %s\n",
			name)
	}
	return true
}

func unpackZip(archive string, verbose bool) {
	// TODO
	fmt.Println("TODO unpackZip", archive, verbose)
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

func cwd() string {
	dir, err := os.Getwd()
	if err == nil {
		return dir
	}
	dir, err = filepath.Abs(".")
	if err == nil {
		return dir
	}
	return "."
}
