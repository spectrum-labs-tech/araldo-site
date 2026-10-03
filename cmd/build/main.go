// SPDX-License-Identifier: AGPL-3.0-or-later

// Command build renders araldo.dev into a directory: the landing page, and
// the app repo's documentation as HTML, with links between documents
// rewritten to the site's pages and links into the code sent to GitHub.
//
//	go run ./cmd/build -app ../araldo -out dist
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

func main() {
	app := flag.String("app", "../araldo", "a checkout of the araldo app repo")
	out := flag.String("out", "dist", "where to write the site")
	flag.Parse()
	n, err := Build(*app, *out)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("wrote %d pages to %s\n", n, *out)
	os.Exit(0)
}
