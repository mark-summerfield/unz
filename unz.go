// Copyright © 2023 Mark Summerfield. All rights reserved.
// License: Apache-2.0

package main

import (
	_ "embed"
	"fmt"
	"github.com/mark-summerfield/clip"
	"log"
	"strings"
)

//go:embed Version.dat
var Version string

func main() {
	verbose, list, archives := getConfig()
	for _, archive := range archives {
		if list {
			listContents(archive, verbose)
		} else {
			unpack(archive, verbose)
		}
	}
}

func getConfig() (bool, bool, []string) {
	parser := clip.NewParserVersion(Version)
	parser.LongDesc = `Unpacks each archive (tar or zip). For each archive
	at most one file or folder is created in the current folder. If the
	archive contains one file or folder, that file or folder is unpacked
	into the current folder. If the archive contains more than one member,
	then a new subfolder is created based on the archive's name, and all the
	archive's contents are unpacked into the subfolder.`
	parser.PositionalCount = clip.OneOrMorePositionals
	_ = parser.SetPositionalVarName("ARCHIVE")
	verboseOpt := parser.Flag("verbose", "Show actions.")
	listOpt := parser.Flag("list",
		"List each archive's contents (don't unpack).")
	err := parser.Parse()
	if err != nil {
		log.Fatal(err)
	}
	return verboseOpt.Value(), listOpt.Value(), parser.Positionals
}

func listContents(archive string, verbose bool) {
	if isTarball(archive) {
		listTarball(archive, verbose)
	} else {
		listZip(archive, verbose)
	}
}

func listTarball(archive string, verbose bool) {
	// NOTE if verbose use bold for archive names
	fmt.Println("listTarball", archive, verbose) // TODO
}

func listZip(archive string, verbose bool) {
	// NOTE if verbose use bold for archive names
	fmt.Println("listZip", archive, verbose) // TODO
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
		strings.Contains(name, ".TAR.")
}
