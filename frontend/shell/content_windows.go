//go:build windows

package shell

import (
	"log"
	"os"
	"path/filepath"
	"sync"
	"syscall"
	"unsafe"

	"github.com/wailsapp/go-webview2/pkg/edge"
	"golang.org/x/sys/windows"
)

// DefaultChromeTopPx is the reserved height for tab strip + toolbar
// (no custom titlebar — OS caption is used).
const DefaultChromeTopPx int32 = 96

var (
	user32                       = windows.NewLazySystemDLL("user32.dll")
	procCreateWindowExW          = user32.NewProc("CreateWindowExW")
	procDestroyWindow            = user32.NewProc("DestroyWindow")
	procSetWindowPos             = user32.NewProc("SetWindowPos")
	procGetClientRect            = user32.NewProc("GetClientRect")
	procShowWindow               = user32.NewProc("ShowWindow")
	procRegisterClassExW         = user32.NewProc("RegisterClassExW")
	procDefWindowProcW           = user32.NewProc("DefWindowProcW")
	procGetModuleHandleW         = windows.NewLazySystemDLL("kernel32.dll").NewProc("GetModuleHandleW")
	contentClassRegistered       bool
	contentClassName             = windows.StringToUTF16Ptr("ConductinoContentHost")
)

const (
	wsChild      = 0x40000000
	wsVisible    = 0x10000000
	wsClipSibs   = 0x04000000
	swShow       = 5
	swHide       = 0
	swpNoZOrder  = 0x0004
	swpNoActivate = 0x0010
	swpShowWindow = 0x0040
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

func contentWndProc(hwnd windows.Handle, msg uint32, wParam, lParam uintptr) uintptr {
	r, _, _ := procDefWindowProcW.Call(uintptr(hwnd), uintptr(msg), wParam, lParam)
	return r
}

func ensureContentClass() error {
	if contentClassRegistered {
		return nil
	}
	hInstance, _, _ := procGetModuleHandleW.Call(0)
	wc := wndClassEx{
		Size:      uint32(unsafe.Sizeof(wndClassEx{})),
		WndProc:   syscall.NewCallback(contentWndProc),
		Instance:  windows.Handle(hInstance),
		ClassName: contentClassName,
	}
	r, _, err := procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc)))
	if r == 0 {
		// class may already exist
		if errno, ok := err.(syscall.Errno); ok && errno != 0 && errno != 1410 {
			return err
		}
	}
	contentClassRegistered = true
	return nil
}

func createContentChild(parent uintptr, topPad int32) (uintptr, error) {
	if err := ensureContentClass(); err != nil {
		return 0, err
	}
	var rc rect
	procGetClientRect.Call(parent, uintptr(unsafe.Pointer(&rc)))
	w := rc.Right - rc.Left
	h := rc.Bottom - rc.Top - topPad
	if h < 50 {
		h = 50
	}
	hInstance, _, _ := procGetModuleHandleW.Call(0)
	r, _, err := procCreateWindowExW.Call(
		0,
		uintptr(unsafe.Pointer(contentClassName)),
		0,
		uintptr(wsChild|wsVisible|wsClipSibs),
		0,
		uintptr(topPad),
		uintptr(w),
		uintptr(h),
		parent,
		0,
		hInstance,
		0,
	)
	if r == 0 {
		return 0, err
	}
	return r, nil
}

// ContentWebView is a native WebView2 hosted in a child HWND under the chrome band.
type ContentWebView struct {
	mu         sync.Mutex
	chromium   *edge.Chromium
	childHWND  uintptr
	parentHWND uintptr
	topPad     int32
	bounds     Bounds
	ready      bool
	hidden     bool
	lastURL    string
}

func NewContentWebView() *ContentWebView {
	return &ContentWebView{hidden: true, topPad: DefaultChromeTopPx}
}

// Embed creates a child HWND below the chrome band and attaches WebView2 to it.
// The child window paints above the chrome webview in that region.
func (c *ContentWebView) Embed(parentHWND uintptr, dataDir string) bool {
	if parentHWND == 0 {
		log.Printf("[content] Embed: null HWND")
		return false
	}
	if dataDir == "" {
		dataDir = filepath.Join(os.TempDir(), "conductino-content-wv2")
	}
	_ = os.MkdirAll(dataDir, 0o755)

	child, err := createContentChild(parentHWND, c.topPad)
	if err != nil || child == 0 {
		log.Printf("[content] create child HWND failed: %v", err)
		return false
	}
	log.Printf("[content] child HWND=%v under parent=%v topPad=%d", child, parentHWND, c.topPad)

	cr := edge.NewChromium()
	cr.DataPath = dataDir
	cr.Debug = false

	log.Printf("[content] embedding WebView2 on child HWND data=%s", dataDir)
	ok := false
	func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("[content] Embed panic: %v", r)
				ok = false
			}
		}()
		ok = cr.Embed(child)
	}()
	if !ok {
		log.Printf("[content] Embed failed")
		procDestroyWindow.Call(child)
		return false
	}

	// Fill the child entirely (no extra padding — child is already below chrome).
	cr.SetPadding(edge.Rect{})
	cr.Resize()
	_ = cr.Hide()
	procShowWindow.Call(child, uintptr(swHide))

	c.mu.Lock()
	c.chromium = cr
	c.childHWND = child
	c.parentHWND = parentHWND
	c.ready = true
	c.hidden = true
	c.mu.Unlock()

	log.Printf("[content] WebView2 ready on child host")
	return true
}

func (c *ContentWebView) layoutChild() {
	c.mu.Lock()
	parent := c.parentHWND
	child := c.childHWND
	pad := c.topPad
	c.mu.Unlock()
	if parent == 0 || child == 0 {
		return
	}
	var rc rect
	procGetClientRect.Call(parent, uintptr(unsafe.Pointer(&rc)))
	w := rc.Right - rc.Left
	h := rc.Bottom - rc.Top - pad
	if h < 50 {
		h = 50
	}
	procSetWindowPos.Call(
		child, 0,
		0, uintptr(pad), uintptr(w), uintptr(h),
		uintptr(swpNoZOrder|swpNoActivate|swpShowWindow),
	)
	c.mu.Lock()
	cr := c.chromium
	c.mu.Unlock()
	if cr != nil {
		cr.Resize()
	}
}

func (c *ContentWebView) Navigate(url string) {
	c.mu.Lock()
	cr := c.chromium
	ready := c.ready
	child := c.childHWND
	c.lastURL = url
	c.mu.Unlock()
	if !ready || cr == nil {
		log.Printf("[content] Navigate skipped (not ready): %s", url)
		return
	}
	c.layoutChild()
	log.Printf("[content] Navigate → %s", url)
	cr.Navigate(url)
	procShowWindow.Call(child, uintptr(swShow))
	if err := cr.Show(); err != nil {
		log.Printf("[content] Show: %v", err)
	}
	c.mu.Lock()
	c.hidden = false
	c.mu.Unlock()
}

func (c *ContentWebView) Reload() {
	c.mu.Lock()
	url := c.lastURL
	c.mu.Unlock()
	if url == "" {
		return
	}
	c.Navigate(url)
}

func (c *ContentWebView) Hide() {
	c.mu.Lock()
	cr := c.chromium
	ready := c.ready
	child := c.childHWND
	c.mu.Unlock()
	if !ready {
		return
	}
	if cr != nil {
		_ = cr.Hide()
	}
	if child != 0 {
		procShowWindow.Call(child, uintptr(swHide))
	}
	c.mu.Lock()
	c.hidden = true
	c.mu.Unlock()
}

func (c *ContentWebView) Show() {
	c.mu.Lock()
	cr := c.chromium
	ready := c.ready
	child := c.childHWND
	c.mu.Unlock()
	if !ready {
		return
	}
	c.layoutChild()
	if child != 0 {
		procShowWindow.Call(child, uintptr(swShow))
	}
	if cr != nil {
		_ = cr.Show()
	}
	c.mu.Lock()
	c.hidden = false
	c.mu.Unlock()
}

func (c *ContentWebView) Bounds() Bounds {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.bounds
}

func (c *ContentWebView) SetBounds(b Bounds) {
	c.mu.Lock()
	c.bounds = b
	if b.Y > 0 {
		c.topPad = int32(b.Y)
	}
	c.mu.Unlock()
	c.layoutChild()
}

func (c *ContentWebView) Ready() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.ready
}
