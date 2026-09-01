// verify-i18n checks internal/ajean/ui/src/js/00a-i18n-data.js for completeness:
// every language block must have exactly the same set of keys as fr (the
// source language). Run it after editing a translation, before opening a PR.
//
//	go run ./tools/verify-i18n
//
// Exit code 0 = all languages complete. Non-zero = something's missing;
// details printed to stdout.
package main

import (
	"fmt"
	"os"
	"regexp"
	"sort"
)

const dataFile = "internal/ajean/ui/src/js/00a-i18n-data.js"

// Matches "lang: {" at the start of a language block, and a top-level "key": "value" line.
var reLangStart = regexp.MustCompile(`(?m)^(\w+):\s*\{\s*$`)
var reEntry = regexp.MustCompile(`(?m)^\s*"((?:[^"\\]|\\.)*)":\s*"((?:[^"\\]|\\.)*)",?\s*$`)

func main() {
	b, err := os.ReadFile(dataFile)
	if err != nil {
		fmt.Println("could not read", dataFile, ":", err)
		fmt.Println("(run this from the repository root)")
		os.Exit(2)
	}
	src := string(b)

	starts := reLangStart.FindAllStringSubmatchIndex(src, -1)
	if len(starts) == 0 {
		fmt.Println("no language blocks found — is the file format still var I18N = { lang: { ... }, ... } ?")
		os.Exit(2)
	}

	langs := map[string]map[string]bool{}
	order := []string{}
	for i, m := range starts {
		name := src[m[2]:m[3]]
		blockStart := m[1]
		blockEnd := len(src)
		if i+1 < len(starts) {
			blockEnd = starts[i+1][0]
		}
		block := src[blockStart:blockEnd]
		keys := map[string]bool{}
		for _, e := range reEntry.FindAllStringSubmatch(block, -1) {
			keys[e[1]] = true
		}
		langs[name] = keys
		order = append(order, name)
	}

	fr, ok := langs["fr"]
	if !ok {
		fmt.Println("no 'fr' block found — fr is the reference language, every other block is checked against it")
		os.Exit(2)
	}
	fmt.Printf("Reference (fr): %d keys\n\n", len(fr))

	ok = true
	for _, lang := range order {
		if lang == "fr" {
			continue
		}
		keys := langs[lang]
		var missing, extra []string
		for k := range fr {
			if !keys[k] {
				missing = append(missing, k)
			}
		}
		for k := range keys {
			if !fr[k] {
				extra = append(extra, k)
			}
		}
		sort.Strings(missing)
		sort.Strings(extra)
		pct := 100.0
		if len(fr) > 0 {
			pct = 100.0 * float64(len(fr)-len(missing)) / float64(len(fr))
		}
		fmt.Printf("%s: %d/%d keys (%.0f%%)\n", lang, len(keys)-len(extra), len(fr), pct)
		if len(missing) > 0 {
			ok = false
			fmt.Printf("  missing %d key(s):\n", len(missing))
			for _, k := range missing {
				fmt.Printf("    %s\n", k)
			}
		}
		if len(extra) > 0 {
			ok = false
			fmt.Printf("  %d key(s) not in fr (typo in the key name?):\n", len(extra))
			for _, k := range extra {
				fmt.Printf("    %s\n", k)
			}
		}
	}
	if !ok {
		fmt.Println("\nFAIL — see missing/extra keys above.")
		os.Exit(1)
	}
	fmt.Println("\nOK — every language has exactly the same keys as fr.")
}
