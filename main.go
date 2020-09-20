package main

import (
	"flag"
	"fmt"
	"ltt-tui/tui"
)

func help() {
	w := flag.CommandLine.Output()
	fmt.Fprintln(w, "ltt-ui v0.1.0")
	fmt.Fprintln(w, "Joakim Hamren <joakim.hamren@gmail.com>")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Terminal-UI for ltt")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "Navigate using vi/vim binds or the arrow keys")
	flag.PrintDefaults()
}

func main() {
	var uri string
	var updateInterval int
	flag.StringVar(&uri, "uri", "http://localhost:4141", "URI to the ltt server")
	flag.IntVar(&updateInterval, "update-interval", 1, "Update interval in seconds")
	flag.Usage = help
	flag.Parse()

	client := tui.NewLTTClient(uri)
	ui := tui.NewUIView(client, updateInterval)
	ui.Run()
}
