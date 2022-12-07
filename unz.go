// Copyright © 2022 Mark Summerfield. All rights reserved.
// License: Apache-2.0

package main

import (
    "fmt"
    _ "embed"
    )

//go:embed Version.dat
var Version string

func main() {
    fmt.Printf("Hello unz v%s\n", Version)
}
