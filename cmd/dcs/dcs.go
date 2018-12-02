// Binary dcs is the swiss-army knife for Debian Code Search. It displays index
// files in a variety of ways.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
	"runtime/trace"

	"net/http"
	_ "net/http/pprof"
)

const globalHelp = `dcs - Debian Code Search swiss-army knife

Syntax: dcs [global flags] <command> [flags] [args]

Index query commands:
	du       — shows disk usage of the specified index files
	docids   - list the documents covered by this index
	trigram  - display metadata of the specified trigram
	raw      - print raw (encoded) index data for the specified trigram
	posting  - list the (decoded) posting list for the specified trigram
	matches  - list the filename[:pos] matches for the specified trigram
	search   - list the filename[:pos] matches for the specified search query
	replay   — replay a query log

Index manipulation commands:
	create   - create an index
	merge    - merge multiple index files into one
`

func help(topic string) {
	var err error
	switch topic {
	case "du":
		fmt.Fprintf(os.Stdout, "%s", duHelp)
		err = du([]string{"-help"})
	case "raw":
		fmt.Fprintf(os.Stdout, "%s", rawHelp)
		err = raw([]string{"-help"})
	case "trigram":
		fmt.Fprintf(os.Stdout, "%s", trigramHelp)
		trigram([]string{"-help"})
	case "docids":
		fmt.Fprintf(os.Stdout, "%s", docidsHelp)
		docids([]string{"-help"})
	case "posting":
		fmt.Fprintf(os.Stdout, "%s", postingHelp)
		posting([]string{"-help"})
	case "matches":
		fmt.Fprintf(os.Stdout, "%s", matchesHelp)
		matches([]string{"-help"})
	case "create":
		fmt.Fprintf(os.Stdout, "%s", createHelp)
		create([]string{"-help"})
	case "merge":
		fmt.Fprintf(os.Stdout, "%s", mergeHelp)
		merge([]string{"-help"})
	case "search":
		fmt.Fprintf(os.Stdout, "%s", searchHelp)
		search([]string{"-help"})
	case "replay":
		fmt.Fprintf(os.Stdout, "%s", replayHelp)
		replay([]string{"-help"})
	case "":
		flag.Usage()
	default:
		fmt.Fprintf(os.Stderr, "unknown help topic %q", topic)
	}
	if err != nil {
		log.Fatal(err)
	}
}

// Global flags (not command-specific)
var cpuprofile, memprofile, listen, traceFn string

func init() {
	// TODO: remove in favor of running as a test
	flag.StringVar(&cpuprofile, "cpuprofile", "", "")
	flag.StringVar(&memprofile, "memprofile", "", "write memory profile to this file")
	flag.StringVar(&listen, "listen", "", "speak HTTP on this [host]:port if non-empty")
	flag.StringVar(&traceFn, "trace", "", "create runtime/trace file")
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "%s", globalHelp)
		//fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	if listen != "" {
		go func() {
			if err := http.ListenAndServe(listen, nil); err != nil {
				log.Fatal(err)
			}
		}()
	}

	if cpuprofile != "" {
		f, err := os.Create(cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	if memprofile != "" {
		defer func() {
			f, err := os.Create(memprofile)
			if err != nil {
				log.Fatal(err)
			}
			runtime.GC()
			pprof.WriteHeapProfile(f)
			f.Close()
		}()
	}

	if traceFn != "" {
		f, err := os.Create(traceFn)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		if err := trace.Start(f); err != nil {
			log.Fatal(err)
		}
		defer trace.Stop()
	}

	args := flag.Args()
	if len(args) == 0 {
		flag.Usage()
		return
	}
	cmd, args := args[0], args[1:]
	var err error
	switch cmd {
	case "du":
		err = du(args)
	case "raw":
		err = raw(args)
	case "trigram":
		trigram(args)
	case "docids":
		docids(args)
	case "posting":
		posting(args)
	case "matches":
		matches(args)
	case "create":
		create(args)
	case "merge":
		merge(args)
	case "search":
		search(args)
	case "replay":
		replay(args)
	case "help":
		if len(args) > 0 {
			help(args[0])
		} else {
			help("")
		}
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n", cmd)
		flag.Usage()
		os.Exit(1)
	}
	if err != nil {
		log.Fatal(err)
	}
}