//go:build windows

package main

import (
	"fmt"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"unsafe"

	webview "github.com/jchv/go-webview2"

)

// runDesktopMode starts the original JSON-backed desktop app with WebView2.
func runDesktopMode() {
	exe, err := os.Executable()
	if err != nil {
		log.Fatalf("os.Executable: %v", err)
	}
	dataDir = filepath.Join(filepath.Dir(exe), "data")
	stateFile = filepath.Join(dataDir, "state.json")

	if err := os.MkdirAll(dataDir, 0755); err != nil {
		log.Fatalf("create data dir: %v", err)
	}
	loadState()

	mux := http.NewServeMux()
	registerDesktopRoutes(mux)

	// Static files (catch-all)
	subFS, err := fs.Sub(staticFiles, "static")
	if err != nil {
		log.Fatalf("sub static FS: %v", err)
	}
	mux.Handle("/", http.FileServer(http.FS(subFS)))

	// Pick a random free port on localhost
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Fatalf("listen: %v", err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	url := fmt.Sprintf("http://127.0.0.1:%d", port)

	// Start HTTP server in background
	ready := make(chan struct{})
	go func() {
		close(ready)
		log.Printf("Desktop mode: listening on %s", url)
		if err := http.Serve(ln, corsMiddleware(mux)); err != nil {
			log.Printf("server: %v", err)
		}
	}()
	<-ready

	// Open native desktop window
	w := webview.NewWithOptions(webview.WebViewOptions{
		Debug:     false,
		AutoFocus: true,
		WindowOptions: webview.WindowOptions{
			Title:  "Study Session",
			Width:  393,
			Height: 852,
		},
	})
	defer w.Destroy()

	hwnd := uintptr(w.Window())

	user32 := syscall.NewLazyDLL("user32.dll")
	setWindowPos := user32.NewProc("SetWindowPos")

	const (
		hwndTopmost    = ^uintptr(0)
		hwndNotTopmost = ^uintptr(0) - 1
		swpNoMove      = 0x0002
		swpNoSize      = 0x0001
	)

	alwaysOnTop := false
	w.Bind("goToggleAlwaysOnTop", func() error {
		alwaysOnTop = !alwaysOnTop
		insertAfter := hwndNotTopmost
		if alwaysOnTop {
			insertAfter = hwndTopmost
		}
		setWindowPos.Call(hwnd, insertAfter, 0, 0, 0, 0, swpNoMove|swpNoSize)
		return nil
	})

	dwmapi := syscall.NewLazyDLL("dwmapi.dll")
	dwmSetAttr := dwmapi.NewProc("DwmSetWindowAttribute")

	type titleBarReq struct {
		IsDark bool   `json:"isDark"`
		Bg     string `json:"bg"`
	}
	w.Bind("goSetTitleBar", func(req titleBarReq) error {
		bg := hexToCOLORREF(req.Bg)
		dark := uint32(0)
		if req.IsDark {
			dark = 1
		}
		textColor := uint32(0x00DDDDDD)
		if !req.IsDark {
			textColor = 0x00222222
		}
		dwmSetAttr.Call(hwnd, 20, uintptr(unsafe.Pointer(&dark)), 4)
		dwmSetAttr.Call(hwnd, 35, uintptr(unsafe.Pointer(&bg)), 4)
		dwmSetAttr.Call(hwnd, 36, uintptr(unsafe.Pointer(&textColor)), 4)
		return nil
	})

	w.Bind("goNotify", func(req struct {
		Title   string `json:"title"`
		Message string `json:"message"`
	}) error {
		script := fmt.Sprintf(`
[Windows.UI.Notifications.ToastNotificationManager,Windows.UI.Notifications,ContentType=WindowsRuntime]|Out-Null
[Windows.Data.Xml.Dom.XmlDocument,Windows.Data.Xml.Dom.XmlDocument,ContentType=WindowsRuntime]|Out-Null
$t=[Windows.UI.Notifications.ToastTemplateType]::ToastImageAndText02
$x=[Windows.UI.Notifications.ToastNotificationManager]::GetTemplateContent($t)
$x.GetElementsByTagName('text')[0].AppendChild($x.CreateTextNode('%s'))|Out-Null
$x.GetElementsByTagName('text')[1].AppendChild($x.CreateTextNode('%s'))|Out-Null
[Windows.UI.Notifications.ToastNotificationManager]::CreateToastNotifier('Study Session').Show(
  [Windows.UI.Notifications.ToastNotification]::new($x))
`, req.Title, req.Message)
		cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", script)
		cmd.Start()
		return nil
	})

	w.SetSize(393, 852, webview.HintNone)
	w.Navigate(url)
	w.Run()
}
