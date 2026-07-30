//go:build js && wasm

package main

import (
	"net/url"
	"path"
	"sort"
	"strings"
	"syscall/js"

	errs "github.com/gomatic/go-error"

	tsvsheet "github.com/tsvsheet/go-tsvsheet"
)

// The seed store's refusals. The engine collapses every fetch failure to
// #IMPORT! in the grid; these strings surface only in an EXPLAIN trace, where
// they are the sole way to tell an unknown name from a forbidden reference.
const (
	errSeedUnknown  errs.Const = "no seed data with this name"
	errSeedAbsolute errs.Const = "absolute URLs are not available in the playground"
	errSeedEscape   errs.Const = "reference escapes the seed data"
)

// seed is the name→tabular-text store a relative IMPORT* resolves against —
// the browser's --data. The page calls seedData once at startup with the data
// it shipped; nothing inside a sheet can add to it or reach past it. The wasm
// bridge is single-threaded (the JS event loop), so a package map is safe for
// the same reason browserTick is.
var seed = map[string]string{}

// seedData replaces the seed store from a JS object of name → text
// (args: object). It returns the sorted names now available, so the page can
// show what a sheet may import.
func seedData(_ js.Value, args []js.Value) any {
	next := map[string]string{}
	names := []string{}
	obj := args[0]
	keys := js.Global().Get("Object").Call("keys", obj)
	for i := range keys.Length() {
		name := keys.Index(i).String()
		next[name] = obj.Get(name).String()
		names = append(names, name)
	}
	seed = next
	sort.Strings(names)
	return result(names, nil)
}

// seedFetcher is the playground's tsvsheet.Fetcher: a synchronous lookup in the
// seed store. Synchronous is the point — Fetch is called mid-compute on the JS
// event loop, where blocking on a browser fetch() promise would deadlock the
// page — and it is also the security posture: with no network behind it, a
// sheet arriving in a shared link cannot cause a request to anywhere, which
// keeps the operator-consent rule intact with the page as the operator.
type seedFetcher struct{}

// Fetch resolves a relative reference in the seed store. An absolute URL is
// refused outright rather than fetched: the playground's base is the seed the
// page shipped, and no sheet content can name a different one.
func (seedFetcher) Fetch(ref tsvsheet.ImportURL, _ tsvsheet.MediaType) (tsvsheet.FetchResult, error) {
	name, err := seedName(ref)
	if err != nil {
		return tsvsheet.FetchResult{}, err
	}
	text, ok := seed[name]
	if !ok {
		return tsvsheet.FetchResult{}, errSeedUnknown.With(nil, "name", name, "available", strings.Join(seedNames(), ", "))
	}
	return tsvsheet.FetchResult{
		ContentType: seedMedia(name),
		URL:         tsvsheet.ImportURL(name),
		Body:        []byte(text),
	}, nil
}

// seedName normalizes a reference to a store key: no scheme, no host, and no
// path that climbs above the store.
func seedName(ref tsvsheet.ImportURL) (string, error) {
	parsed, err := url.Parse(string(ref))
	if err != nil || parsed.IsAbs() || parsed.Host != "" {
		return "", errSeedAbsolute.With(err, "reference", string(ref))
	}
	name := path.Clean(string(ref))
	if name == "." || strings.HasPrefix(name, "../") || strings.HasPrefix(name, "/") {
		return "", errSeedEscape.With(nil, "reference", string(ref))
	}
	return name, nil
}

// seedMedia is the media type a stored name answers with — CSV by extension,
// TSV otherwise — mirroring how the data server maps extensions, so the same
// file imports identically here and under `tsv render --data`.
func seedMedia(name string) tsvsheet.MediaType {
	if strings.EqualFold(path.Ext(name), ".csv") {
		return "text/csv"
	}
	return "text/tab-separated-values"
}

// seedNames lists the store's names, sorted, for the unknown-name refusal.
func seedNames() []string {
	names := make([]string, 0, len(seed))
	for name := range seed {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
