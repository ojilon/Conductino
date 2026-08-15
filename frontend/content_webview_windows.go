//go:build windows

package main

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"syscall"
	"unsafe"

	"github.com/wailsapp/go-webview2/pkg/edge"
	"golang.org/x/sys/windows"
)

const (
	defaultChromeHeight = 82 // tabstrip 38 + toolbar 44
	contentClassName    = "ConductinoContentHost"
	COINIT_APARTMENTTHREADED = 0x2
	COINIT_MULTITHREADED     = 0x0
)

var (
	user32                       = windows.NewLazySystemDLL("user32.dll")
	procCreateWindowExW          = user32.NewProc("CreateWindowExW")
	procShowWindow               = user32.NewProc("ShowWindow")
	procSetWindowPos             = user32.NewProc("SetWindowPos")
	procGetClientRect            = user32.NewProc("GetClientRect")
	procRegisterClassExW         = user32.NewProc("RegisterClassExW")
	procDefWindowProcW           = user32.NewProc("DefWindowProcW")
	procGetWindowTextW           = user32.NewProc("GetWindowTextW")
	procIsWindowVisible          = user32.NewProc("IsWindowVisible")
	procEnumWindows              = user32.NewProc("EnumWindows")
	procGetWindowThreadProcessId = user32.NewProc("GetWindowThreadProcessId")

	kernel32             = windows.NewLazySystemDLL("kernel32.dll")
	procGetCurrentProcessId = kernel32.NewProc("GetCurrentProcessId")
	procGetModuleHandleW = kernel32.NewProc("GetModuleHandleW")

	ole32              = windows.NewLazySystemDLL("ole32.dll")
	procCoInitializeEx = ole32.NewProc("CoInitializeEx")
)

type rect struct {
	Left, Top, Right, Bottom int32
}

type wndClassEx struct {
	Size       uint32
	Style      uint32
	WndProc    uintptr
	ClsExtra   int32
	WndExtra   int32
	Instance   windows.Handle
	Icon       windows.Handle
	Cursor     windows.Handle
	Background windows.Handle
	MenuName   *uint16
	ClassName  *uint16
	IconSm     windows.Handle
}

var (
	contentOnce      sync.Once
	contentClassAtom uintptr
	contentMu        sync.Mutex
	contentBrowser   *ContentBrowser
	comOnce          sync.Once
)

// ensureCOM initializes COM on this thread (required for WebView2).
// S_OK (0) and S_FALSE (1 = already initialized) are both success.
func ensureCOM() {
	comOnce.Do(func() {})
	hr, _, _ := procCoInitializeEx.Call(0, uintptr(COINIT_APARTMENTTHREADED))
	// RPC_E_CHANGED_MODE = 0x80010106 — already init with different model; try MTA
	if hr == 0x80010106 {
		procCoInitializeEx.Call(0, uintptr(COINIT_MULTITHREADED))
	}
}

// ContentBrowser hosts a second WebView2 under the chrome strip.
type ContentBrowser struct {
	parent   uintptr
	host     uintptr
	chromium *edge.Chromium
	chromeH  int32
	visible  bool
	ready    bool
	lastURL  string
}

func contentWndProc(hwnd uintptr, msg uint32, wParam, lParam uintptr) uintptr {
	r, _, _ := procDefWindowProcW.Call(hwnd, uintptr(msg), wParam, lParam)
	return r
}

func registerContentClass() error {
	var err error
	contentOnce.Do(func() {
		name, e := windows.UTF16PtrFromString(contentClassName)
		if e != nil {
			err = e
			return
		}
		ret, _, callErr := procGetModuleHandleW.Call(0)
		if ret == 0 {
			err = fmt.Errorf("GetModuleHandleW failed: %v", callErr)
			return
		}
		hInst := windows.Handle(ret)
		wc := wndClassEx{
			Size:      uint32(unsafe.Sizeof(wndClassEx{})),
			WndProc:   syscall.NewCallback(contentWndProc),
			Instance:  hInst,
			ClassName: name,
		}
		atom, _, callErr := procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc)))
		if atom == 0 {
			err = fmt.Errorf("RegisterClassEx: %v", callErr)
			return
		}
		contentClassAtom = atom
	})
	return err
}

func findAppHWND(titleSubstr string) uintptr {
	r, _, _ := procGetCurrentProcessId.Call()
	pid := uint32(r)
	var found uintptr
	cb := syscall.NewCallback(func(hwnd uintptr, _ uintptr) uintptr {
		var winPid uint32
		procGetWindowThreadProcessId.Call(hwnd, uintptr(unsafe.Pointer(&winPid)))
		if winPid != pid {
			return 1
		}
		vis, _, _ := procIsWindowVisible.Call(hwnd)
		if vis == 0 {
			return 1
		}
		var buf [256]uint16
		procGetWindowTextW.Call(hwnd, uintptr(unsafe.Pointer(&buf[0])), 256)
		title := windows.UTF16ToString(buf[:])
		if strings.Contains(title, titleSubstr) {
			found = hwnd
			return 0
		}
		return 1
	})
	procEnumWindows.Call(cb, 0)
	return found
}

func (c *ContentBrowser) createHost() error {
	if err := registerContentClass(); err != nil {
		return err
	}
	className, err := windows.UTF16PtrFromString(contentClassName)
	if err != nil {
		return err
	}
	windowName, err := windows.UTF16PtrFromString("ConductinoContent")
	if err != nil {
		return err
	}
	ret, _, callErr := procGetModuleHandleW.Call(0)
	if ret == 0 {
		return fmt.Errorf("GetModuleHandleW failed: %v", callErr)
	}
	hInst := windows.Handle(ret)
	const (
		WS_CHILD        = 0x40000000
		WS_VISIBLE      = 0x10000000
		WS_CLIPSIBLINGS = 0x04000000
	)
	hwnd, _, callErr := procCreateWindowExW.Call(
		0,
		uintptr(unsafe.Pointer(className)),
		uintptr(unsafe.Pointer(windowName)),
		uintptr(WS_CHILD|WS_VISIBLE|WS_CLIPSIBLINGS),
		0, 0, 100, 100,
		c.parent,
		0,
		uintptr(hInst),
		0,
	)
	if hwnd == 0 {
		return fmt.Errorf("CreateWindowEx content host: %v", callErr)
	}
	c.host = hwnd
	return nil
}

func (c *ContentBrowser) layout() {
	if c.parent == 0 || c.host == 0 {
		return
	}
	var rc rect
	procGetClientRect.Call(c.parent, uintptr(unsafe.Pointer(&rc)))
	top := c.chromeH
	if top < 40 {
		top = defaultChromeHeight
	}
	w := rc.Right - rc.Left
	h := rc.Bottom - rc.Top - top
	if h < 50 {
		h = 50
	}
	const (
		SWP_NOZORDER   = 0x0004
		SWP_NOACTIVATE = 0x0010
		SWP_SHOWWINDOW = 0x0040
		SWP_HIDEWINDOW = 0x0080
	)
	flags := uintptr(SWP_NOZORDER | SWP_NOACTIVATE)
	if c.visible {
		flags |= SWP_SHOWWINDOW
	} else {
		flags |= SWP_HIDEWINDOW
	}
	procSetWindowPos.Call(c.host, 0, 0, uintptr(top), uintptr(w), uintptr(h), flags)
	if c.chromium != nil && c.ready {
		c.chromium.Resize()
	}
}

func (c *ContentBrowser) initChromium() error {
	// WebView2 requires COM on the creating thread (and callbacks).
	ensureCOM()

	c.chromium = edge.NewChromium()
	// Separate user-data folder so we do not clash with the Wails host WebView2 profile.
	c.chromium.DataPath = "" // library default next to exe is fine; edge uses a subfolder

	if !c.chromium.Embed(c.host) {
		return fmt.Errorf("content WebView2 Embed failed (COM/UI thread?)")
	}
	c.chromium.Resize()
	c.ready = true
	log.Printf("[content] WebView2 ready on child host")
	return nil
}

func (c *ContentBrowser) Navigate(url string) error {
	url = strings.TrimSpace(url)
	if url == "" {
		return fmt.Errorf("empty url")
	}
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") &&
		!strings.HasPrefix(url, "about:") {
		url = "https://" + url
	}
	if !c.ready {
		return fmt.Errorf("content browser not ready")
	}
	c.lastURL = url
	c.visible = true
	c.layout()
	c.chromium.Navigate(url)
	return nil
}

func (c *ContentBrowser) SetVisible(v bool) {
	c.visible = v
	c.layout()
	if c.host != 0 {
		const SW_SHOW = 5
		const SW_HIDE = 0
		if v {
			procShowWindow.Call(c.host, SW_SHOW)
		} else {
			procShowWindow.Call(c.host, SW_HIDE)
		}
	}
}

func (c *ContentBrowser) Eval(js string) {
	if c.chromium != nil && c.ready {
		c.chromium.Eval(js)
	}
}

func (c *ContentBrowser) LastURL() string {
	return c.lastURL
}

func ensureContentBrowser() (*ContentBrowser, error) {
	contentMu.Lock()
	defer contentMu.Unlock()
	if contentBrowser != nil && contentBrowser.ready {
		return contentBrowser, nil
	}
	ensureCOM()
	hwnd := findAppHWND("Conductino")
	if hwnd == 0 {
		return nil, fmt.Errorf("app window HWND not found")
	}
	cb := &ContentBrowser{
		parent:  hwnd,
		chromeH: defaultChromeHeight,
		visible: false,
	}
	if err := cb.createHost(); err != nil {
		return nil, err
	}
	if err := cb.initChromium(); err != nil {
		return nil, err
	}
	cb.SetVisible(false)
	contentBrowser = cb
	return cb, nil
}

func (a *App) ContentEnsure() error {
	_, err := ensureContentBrowser()
	return err
}

func (a *App) ContentNavigate(url string) error {
	cb, err := ensureContentBrowser()
	if err != nil {
		return err
	}
	return cb.Navigate(url)
}

func (a *App) ContentSetVisible(show bool) error {
	cb, err := ensureContentBrowser()
	if err != nil {
		return err
	}
	cb.SetVisible(show)
	return nil
}

func (a *App) ContentResize() error {
	contentMu.Lock()
	defer contentMu.Unlock()
	if contentBrowser == nil {
		return nil
	}
	contentBrowser.layout()
	return nil
}

func (a *App) ContentSetChromeHeight(px int) error {
	contentMu.Lock()
	defer contentMu.Unlock()
	if contentBrowser == nil {
		return nil
	}
	if px < 40 {
		px = defaultChromeHeight
	}
	contentBrowser.chromeH = int32(px)
	contentBrowser.layout()
	return nil
}

func (a *App) ContentEval(js string) error {
	cb, err := ensureContentBrowser()
	if err != nil {
		return err
	}
	cb.Eval(js)
	return nil
}

func (a *App) ContentLastURL() string {
	contentMu.Lock()
	defer contentMu.Unlock()
	if contentBrowser == nil {
		return ""
	}
	return contentBrowser.LastURL()
}

func (a *App) ContentCopySelection() error {
	js := `(function(){
  try {
    var t = String(window.getSelection() || "");
    if (!t) { return; }
    if (navigator.clipboard && navigator.clipboard.writeText) {
      navigator.clipboard.writeText(t);
    }
  } catch (e) {}
})();`
	return a.ContentEval(js)
}
