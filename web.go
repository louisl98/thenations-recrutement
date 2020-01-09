// Static file server.
// Serves static files from the given directory.

package main

import (
	"expvar"
	"flag"
	"log"

	"github.com/valyala/fasthttp"
)

var (
	addr      = flag.String("addr", "fiche-recrutement.herokuapp.com", "TCP address to listen to")
	byteRange = flag.Bool("byteRange", false, "Enables byte range requests if set to true")
	dir       = flag.String("dir", "./vendor/", "Directory to serve static files from")
)

func main() {
	// Parse command-line flags
	flag.Parse()

	// Setup FS handler
	fs := &fasthttp.FS{
		Root:               *dir,
		IndexNames:         []string{"index.html"},
		GenerateIndexPages: true,
		Compress:           false,
		AcceptByteRange:    *byteRange,
	}
	fsHandler := fs.NewRequestHandler()

	// Create RequestHandler serving server stats on /stats and files
	// on other requested paths
	requestHandler := func(ctx *fasthttp.RequestCtx) {
		fsHandler(ctx)
		updateFSCounters(ctx)
	}

	// Start HTTP server
	if len(*addr) > 0 {
		log.Printf("Starting HTTP server on %q", *addr)
		go func() {
			if err := fasthttp.ListenAndServe(*addr, requestHandler); err != nil {
				log.Fatalf("error in ListenAndServe: %s", err)
			}
		}()
	}

	log.Printf("Serving files from directory %q", *dir)
	select {}
}

func updateFSCounters(ctx *fasthttp.RequestCtx) {
	// Increment the number of fsHandler calls
	fsCalls.Add(1)

	// Update other stats counters
	resp := &ctx.Response
	switch resp.StatusCode() {
	case fasthttp.StatusOK:
		fsOKResponses.Add(1)
		fsResponseBodyBytes.Add(int64(resp.Header.ContentLength()))
	case fasthttp.StatusNotModified:
		fsNotModifiedResponses.Add(1)
	case fasthttp.StatusNotFound:
		fsNotFoundResponses.Add(1)
	default:
		fsOtherResponses.Add(1)
	}
}

// Various counters - see https://golang.org/pkg/expvar/ for details
var (
	// Counter for total number of fs calls
	fsCalls = expvar.NewInt("fsCalls")

	// Counters for various response status codes
	fsOKResponses          = expvar.NewInt("fsOKResponses")
	fsNotModifiedResponses = expvar.NewInt("fsNotModifiedResponses")
	fsNotFoundResponses    = expvar.NewInt("fsNotFoundResponses")
	fsOtherResponses       = expvar.NewInt("fsOtherResponses")

	// Total size in bytes for OK response bodies served
	fsResponseBodyBytes = expvar.NewInt("fsResponseBodyBytes")
)
