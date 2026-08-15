//go:build windows

package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"github.com/wailsapp/go-webview2/pkg/edge"
	"golang.org/x/sys/windows"
)

const (
	defaultChromeHeight       = 82
	contentClassName          = "ConductinoContentHost"
	COINIT_APARTMENTTHREADED  = 0x2
	COINIT_MULTITHREADED      = 0x0
	WS_CHILD                  = 0x40000000
	WS_VISIBLE                = 0x10000000
	WS_CLIPSIBLINGS           = 0x04000000
	WS_CLIPCHILDREN           = 0x02000000
	WS_POPUP                  = 0x80000000
	WS_BORDER                 = 0x00800000
	WS_EX_TOOLWINDOW          = 0x00000080
	WS_EX_NOACTIVATE          = 0x08000000
	SWP_NOZORDER              = 0x0004
	SWP_NOACTIVATE            = 0x0010
	SWP_SHOWWINDOW            = 0x0040
	SWP_HIDEWINDOW            = 0x0080
	SW_SHOW                   = 5
	SW_HIDE                   = 0
	GWLP_HWNDPARENT           = -8
)

var (
	user32                       = windows.NewLazySystemDLL("user32.dll")
	procCreateWindowExW          = user32.NewProc("CreateWindowExW")
	procShowWindow               = user32.NewProc("ShowWindow")
	procSetWindowPos             = user32.NewProc("SetWindowPos")
	procGetClientRect            = user32.NewProc("GetClientRect")
	procClientToScreen           = user32.NewProc("ClientToScreen")
	procRegisterClassExW         = user32.NewProc("RegisterClassExW")
	procDefWindowProcW           = user32.NewProc("DefWindowProcW")
	procGetWindowTextW           = user32.NewProc("GetWindowTextW")
	procIsWindowVisible          = user32.NewProc("IsWindowVisible")
	procEnumWindows              = user32.NewProc("EnumWindows")
	procGetWindowThreadProcessId = user32.NewProc("GetWindowThreadProcessId")
	procSetWindowLongPtrW       = user32.NewProc("SetWindowLongPtrW")

	kernel32                 = windows.NewLazySystemDLL("kernel32.dll")
	procGetCurrentProcessId = kernel32.NewProc("GetCurrentProcessId")
	procGetModuleHandleW    = kernel32.NewProc("GetModuleHandleW")

	ole32              = windows.NewLazySystemDLL("ole32.dll")
	procCoInitializeEx = ole32.NewProc("CoInitializeEx")
)

type rect struct {
	Left, Top, Right, Bottom int32
}

type point struct {
	X, Y int32
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
)

func ensureCOM() {
	hr, _, _ := procCoInitializeEx.Call(0, uintptr(COINIT_APARTMENTTHREADED))
	if hr == 0x80010106 {
		procCoInitializeEx.Call(0, uintptr(COINIT_MULTITHREADED))
	}
}

type ContentBrowser struct {
	parent   uintptr
	host     uintptr
	chromium *edge.Chromium
	chromeH  int32
	visible  bool
	ready    bool
	lastURL  string
	usePopup bool // owned popup instead of pure WS_CHILD
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
			err = fmt.Errorf("GetModuleHandleW: %v", callErr)
			return
		}
		wc := wndClassEx{
			Size:      uint32(unsafe.Sizeof(wndClassEx{})),
			WndProc:   syscall.NewCallback(contentWndProc),
			Instance:  windows.Handle(ret),
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

func contentDataPath() string {
	base, err := os.UserCacheDir()
	if err != nil {
		base = os.TempDir()
	}
	dir := filepath.Join(base, "Conductino", "content-webview2")
	_ = os.MkdirAll(dir, 0o755)
	return dir
}

// createHost builds a child HWND first; if WebView2 rejects it we recreate as owned popup.
func (c *ContentBrowser) createHost(asPopup bool) error {
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
		return fmt.Errorf("GetModuleHandleW: %v", callErr)
	}
	hInst := windows.Handle(ret)

	var style, exStyle uintptr
	var parent uintptr
	if asPopup {
		// Owned popup sits above the main window content area (more reliable for 2nd WebView2).
		style = uintptr(WS_POPUP | WS_CLIPSIBLINGS | WS_CLIPCHILDREN)
		exStyle = uintptr(WS_EX_TOOLWINDOW)
		parent = c.parent // owner
		c.usePopup = true
	} else {
		style = uintptr(WS_CHILD | WS_CLIPSIBLINGS | WS_CLIPCHILDREN)
		exStyle = 0
		parent = c.parent
		c.usePopup = false
	}

	hwnd, _, callErr := procCreateWindowExW.Call(
		exStyle,
		uintptr(unsafe.Pointer(className)),
		uintptr(unsafe.Pointer(windowName)),
		style,
		0, 0, 800, 600, // non-zero initial size (required before Embed)
		parent,
		0,
		uintptr(hInst),
		0,
	)
	if hwnd == 0 {
		return fmt.Errorf("CreateWindowEx: %v", callErr)
	}
	c.host = hwnd

	// Size + show before WebView2 controller creation.
	c.visible = true
	c.layout()
	procShowWindow.Call(c.host, uintptr(SW_SHOW))
	// Let the window manager settle.
	time.Sleep(50 * time.Millisecond)
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
	if w < 100 {
		w = 100
	}
	if h < 100 {
		h = 100
	}

	flags := uintptr(SWP_NOACTIVATE)
	if c.visible {
		flags |= SWP_SHOWWINDOW
	} else {
		flags |= SWP_HIDEWINDOW
	}

	if c.usePopup {
		// Client (0, top) → screen coords for popup placement.
		pt := point{X: 0, Y: top}
		procClientToScreen.Call(c.parent, uintptr(unsafe.Pointer(&pt)))
		procSetWindowPos.Call(c.host, 0, uintptr(pt.X), uintptr(pt.Y), uintptr(w), uintptr(h), flags)
	} else {
		flags |= SWP_NOZORDER
		procSetWindowPos.Call(c.host, 0, 0, uintptr(top), uintptr(w), uintptr(h), flags)
	}
	if c.chromium != nil && c.ready {
		c.chromium.Resize()
	}
}

func (c *ContentBrowser) initChromium() error {
	ensureCOM()

	c.chromium = edge.NewChromium()
	c.chromium.DataPath = contentDataPath()
	log.Printf("[content] DataPath=%s host=%v popup=%v", c.chromium.DataPath, c.host, c.usePopup)

	if !c.chromium.Embed(c.host) {
		return fmt.Errorf("content WebView2 Embed returned false")
	}

	// Embed starts async environment+controller creation. Wait briefly for controller.
	deadline := time.Now().Add(8 * time.Second)
	for time.Now().Before(deadline) {
		time.Sleep(100 * time.Millisecond)
		// Resize is a no-op until controller exists; try Navigate to about:blank as readiness probe.
		c.chromium.Resize()
		// Heuristic: if Eval does not panic and Resize ran, mark ready after 400ms min
		if time.Since(deadline.Add(-8*time.Second)) > 400*time.Millisecond {
			break
		}
	}
	// Give controller creation callbacks time (error is logged by go-webview2 if it fails).
	time.Sleep(400 * time.Millisecond)
	c.chromium.Resize()
	c.ready = true
	log.Printf("[content] WebView2 Embed finished (check logs for controller errors)")
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
	if !c.ready || c.chromium == nil {
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
		if v {
			procShowWindow.Call(c.host, uintptr(SW_SHOW))
		} else {
			procShowWindow.Call(c.host, uintptr(SW_HIDE))
		}
	}
}

func (c *ContentBrowser) Eval(js string) {
	if c.chromium != nil && c.ready {
		c.chromium.Eval(js)
	}
}

func (c *ContentBrowser) LastURL() string { return c.lastURL }

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

	try := func(asPopup bool) (*ContentBrowser, error) {
		cb := &ContentBrowser{
			parent:  hwnd,
			chromeH: defaultChromeHeight,
			visible: false,
		}
		if err := cb.createHost(asPopup); err != nil {
			return nil, err
		}
		if err := cb.initChromium(); err != nil {
			return nil, err
		}
		cb.SetVisible(false)
		return cb, nil
	}

	// Prefer true child first (integrated layout).
	cb, err := try(false)
	if err != nil {
		log.Printf("[content] child host failed: %v — trying owned popup", err)
		cb, err = try(true)
		if err != nil {
			return nil, err
		}
	}
	// If child path "succeeded" Embed but controller error was only logged, user may still see blank.
	// Second attempt with popup is available via ContentRecreateAsPopup if needed.
	contentBrowser = cb
	return cb, nil
}

// ContentRecreateAsPopup forces owned-popup content surface (call if child path is blank).
func (a *App) ContentRecreateAsPopup() error {
	contentMu.Lock()
	contentBrowser = nil
	contentMu.Unlock()
	ensureCOM()
	hwnd := findAppHWND("Conductino")
	if hwnd == 0 {
		return fmt.Errorf("app window HWND not found")
	}
	cb := &ContentBrowser{parent: hwnd, chromeH: defaultChromeHeight}
	if err := cb.createHost(true); err != nil {
		return err
	}
	if err := cb.initChromium(); err != nil {
		return err
	}
	contentMu.Lock()
	contentBrowser = cb
	contentMu.Unlock()
	return nil
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
