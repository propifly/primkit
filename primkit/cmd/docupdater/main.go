// Command docupdater reads primkit command metadata JSON files and updates
// docs/agent-reference.md by replacing anchored command table sections.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/propifly/primkit/primkit/docgen"
)

func main() {
	check := flag.Bool("check", false, "check mode: fail if docs are out of date without writing")
	inputs := flag.String("inputs", "", "comma-separated paths to JSON metadata files")
	doc := flag.String("doc", "docs/agent-reference.md", "path to agent reference doc to update")
	flag.Parse()

	if *inputs == "" {
		fmt.Fprintln(os.Stderr, "error: --inputs is required")
		os.Exit(1)
	}

	var metas []docgen.PrimMeta
	for _, path := range strings.Split(*inputs, ",") {
		path = strings.TrimSpace(path)
		if path == "" {
			continue
		}
		f, err := os.Open(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error opening %s: %v\n", path, err)
			os.Exit(1)
		}
		var meta docgen.PrimMeta
		if err := json.NewDecoder(f).Decode(&meta); err != nil {
			fmt.Fprintf(os.Stderr, "error decoding %s: %v\n", path, err)
			f.Close()
			os.Exit(1)
		}
		f.Close()
		metas = append(metas, meta)
	}

	if err := docgen.UpdateDoc(*doc, metas, *check); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	if *check {
		fmt.Println("docs are up to date")
	} else {
		fmt.Printf("updated %s\n", *doc)
	}
}
