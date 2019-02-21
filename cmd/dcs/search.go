package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"regexp/syntax"

	"github.com/Debian/dcs/internal/index"
	"github.com/google/codesearch/regexp"
)

const searchHelp = `search - list the filename[:pos] matches for the specified search query

Example:
  % dcs search -idx=/srv/dcs/shard4/full -query=i3Font
  i3-wm_4.16.1-1/i3-config-wizard/main.c
  i3-wm_4.16.1-1/i3-input/main.c
  i3-wm_4.16.1-1/i3-nagbar/main.c
  […]
`

func search(args []string) error {
	fset := flag.NewFlagSet("search", flag.ExitOnError)
	fset.Usage = usage(fset, searchHelp)
	var idx string
	fset.StringVar(&idx, "idx", "", "path to the index file to work with")
	var query string
	fset.StringVar(&query, "query", "", "search query")
	var pos bool
	fset.BoolVar(&pos, "pos", false, "do a positional query for identifier searches")
	if err := fset.Parse(args); err != nil {
		return err
	}
	if idx == "" || query == "" {
		fset.Usage()
		os.Exit(1)
	}

	log.Printf("search for %q", query)
	re, err := regexp.Compile(query)
	if err != nil {
		return fmt.Errorf("regexp.Compile(%q): %v", query, err)
	}
	s := re.Syntax.Simplify()
	queryPos := pos && s.Op == syntax.OpLiteral

	ix, err := index.Open(idx)
	if err != nil {
		return fmt.Errorf("Could not open index: %v", err)
	}
	defer ix.Close()

	if queryPos {
		matches, err := ix.QueryPositional(string(s.Rune))
		if err != nil {
			return err
		}
		for _, match := range matches {
			fn, err := ix.DocidMap.Lookup(match.Docid)
			if err != nil {
				return fmt.Errorf("DocidMap.Lookup(%v): %v", match.Docid, err)
			}
			fmt.Printf("%s\n", fn)
			// TODO: actually verify the search term occurs at match.Position
		}
	} else {
		q := index.RegexpQuery(re.Syntax)
		log.Printf("q = %v", q)
		docids := ix.PostingQuery(q)
		for _, docid := range docids {
			fn, err := ix.DocidMap.Lookup(docid)
			if err != nil {
				return err
			}
			fmt.Printf("%s\n", fn)
			// TODO: actually grep the file to find a match
		}
	}
	return nil
}
