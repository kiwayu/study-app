//go:build !windows

package main

import "log"

func runDesktopMode() {
	log.Fatal("desktop mode needs Windows and WebView2; run without --desktop for web mode")
}
