// SPDX-License-Identifier: AGPL-3.0-or-later

// Command build renders araldo.dev into a directory: the landing page, and
// the app repo's documentation as HTML, with links between documents
// rewritten to the site's pages and links into the code sent to GitHub.
//
//	go run ./cmd/build -app ../araldo -out dist [-plausible https://…/js/pa-….js]
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
	analytics := flag.String("plausible", "", "a Plausible site's tracker script address; none by default")
	flag.Parse()
	n, err := Build(*app, *out, *analytics)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("wrote %d pages to %s\n", n, *out)
	os.Exit(0)
}
