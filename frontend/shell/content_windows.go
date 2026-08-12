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

const DefaultChromeTopPx int32 = 96

// DefaultSidebarWidthPx matches CSS --sidebar-w (260px).
const DefaultSidebarWidthPx int32 = 260

var (
	user32               = windows.NewLazySystemDLL("user32.dll")
	procCreateWindowExW  = user32.NewProc("CreateWindowExW")
	procDestroyWindow    = user32.NewProc("DestroyWindow")
	procSetWindowPos     = user32.NewProc("SetWindowPos")
	procGetClientRect    = user32.NewProc("GetClientRect")
	procClientToScreen   = user32.NewProc("ClientToScreen")
	procShowWindow       = user32.NewProc("ShowWindow")
	procBringWindowToTop = user32.NewProc("BringWindowToTop")
	procRegisterClassExW = user32.NewProc("RegisterClassExW")
	procDefWindowProcW   = user32.NewProc("DefWindowProcW")
	procGetModuleHandleW = windows.NewLazySystemDLL("kernel32.dll").NewProc("GetModuleHandleW")
	contentClassRegistered bool
	contentClassName     = windows.StringToUTF16Ptr("ConductinoContentHost")
)

const (
	wsPopup        = 0x80000000
	wsClipChildren = 0x02000000
	wsClipSiblings = 0x04000000
	swShow         = 5
	swHide         = 0
	swpNoActivate  = 0x0010
	swpShowWindow  = 0x0040
	swpNoCopyBits  = 0x0100
	hwndTop        = 0
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
		if errno, ok := err.(syscall.Errno); ok && errno != 0 && errno != 1410 {
			return err
		}
	}
	contentClassRegistered = true
	return nil
}

func createContentHost(parent uintptr, topPad, leftPad int32) (uintptr, error) {
	if err := ensureContentClass(); err != nil {
		return 0, err
	}
	var rc rect
	procGetClientRect.Call(parent, uintptr(unsafe.Pointer(&rc)))
	w := rc.Right - rc.Left - leftPad
	h := rc.Bottom - rc.Top - topPad
	if w < 50 {
		w = 50
	}
	if h < 50 {
		h = 50
	}
	pt := point{X: leftPad, Y: topPad}
	procClientToScreen.Call(parent, uintptr(unsafe.Pointer(&pt)))

	hInstance, _, _ := procGetModuleHandleW.Call(0)
	r, _, err := procCreateWindowExW.Call(
		0,
		uintptr(unsafe.Pointer(contentClassName)),
		0,
		uintptr(wsPopup|wsClipChildren|wsClipSiblings),
		uintptr(pt.X),
		uintptr(pt.Y),
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

type ContentWebView struct {
	mu         sync.Mutex
	chromium   *edge.Chromium
	hostHWND   uintptr
	parentHWND uintptr
	topPad     int32
	leftPad    int32 // sidebar inset when open
	bounds     Bounds
	ready      bool
	hidden     bool
	lastURL    string
}

func NewContentWebView() *ContentWebView {
	return &ContentWebView{hidden: true, topPad: DefaultChromeTopPx}
}

func (c *ContentWebView) Embed(parentHWND uintptr, dataDir string) bool {
	if parentHWND == 0 {
		log.Printf("[content] Embed: null HWND")
		return false
	}
	if dataDir == "" {
		dataDir = filepath.Join(os.TempDir(), "conductino-content-wv2")
	}
	_ = os.MkdirAll(dataDir, 0o755)

	host, err := createContentHost(parentHWND, c.topPad, c.leftPad)
	if err != nil || host == 0 {
		log.Printf("[content] create host HWND failed: %v", err)
		return false
	}
	log.Printf("[content] popup host HWND=%v owner=%v topPad=%d", host, parentHWND, c.topPad)

	cr := edge.NewChromium()
	cr.DataPath = dataDir
	cr.Debug = false

	ok := false
	func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("[content] Embed panic: %v", r)
				ok = false
			}
		}()
		ok = cr.Embed(host)
	}()
	if !ok {
		procDestroyWindow.Call(host)
		return false
	}

	cr.SetPadding(edge.Rect{})
	cr.Resize()
	_ = cr.Hide()
	procShowWindow.Call(host, uintptr(swHide))

	c.mu.Lock()
	c.chromium = cr
	c.hostHWND = host
	c.parentHWND = parentHWND
	c.ready = true
	c.hidden = true
	c.mu.Unlock()

	log.Printf("[content] WebView2 ready on popup host")
	return true
}

// SetLeftInset reserves horizontal space for a floating sidebar (px).
func (c *ContentWebView) SetLeftInset(px int32) {
	c.mu.Lock()
	if px < 0 {
		px = 0
	}
	c.leftPad = px
	hidden := c.hidden
	c.mu.Unlock()
	log.Printf("[content] left inset = %d", px)
	c.layoutHost()
	if !hidden {
		c.raise()
	}
}

func (c *ContentWebView) layoutHost() {
	c.mu.Lock()
	parent := c.parentHWND
	host := c.hostHWND
	pad := c.topPad
	left := c.leftPad
	c.mu.Unlock()
	if parent == 0 || host == 0 {
		return
	}
	var rc rect
	procGetClientRect.Call(parent, uintptr(unsafe.Pointer(&rc)))
	w := rc.Right - rc.Left - left
	h := rc.Bottom - rc.Top - pad
	if w < 50 {
		w = 50
	}
	if h < 50 {
		h = 50
	}
	pt := point{X: left, Y: pad}
	procClientToScreen.Call(parent, uintptr(unsafe.Pointer(&pt)))

	procSetWindowPos.Call(
		host,
		hwndTop,
		uintptr(pt.X),
		uintptr(pt.Y),
		uintptr(w),
		uintptr(h),
		uintptr(swpNoActivate|swpShowWindow|swpNoCopyBits),
	)
	procBringWindowToTop.Call(host)

	c.mu.Lock()
	cr := c.chromium
	c.mu.Unlock()
	if cr != nil {
		cr.Resize()
	}
}

func (c *ContentWebView) raise() {
	c.mu.Lock()
	host := c.hostHWND
	cr := c.chromium
	c.mu.Unlock()
	if host == 0 {
		return
	}
	c.layoutHost()
	procShowWindow.Call(host, uintptr(swShow))
	procBringWindowToTop.Call(host)
	if cr != nil {
		_ = cr.Show()
		cr.Resize()
	}
}

func (c *ContentWebView) Navigate(url string) {
	c.mu.Lock()
	cr := c.chromium
	ready := c.ready
	c.lastURL = url
	c.mu.Unlock()
	if !ready || cr == nil {
		log.Printf("[content] Navigate skipped (not ready): %s", url)
		return
	}
	log.Printf("[content] Navigate → %s", url)
	cr.Navigate(url)
	c.raise()
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
	host := c.hostHWND
	c.mu.Unlock()
	if !ready {
		return
	}
	if cr != nil {
		_ = cr.Hide()
	}
	if host != 0 {
		procShowWindow.Call(host, uintptr(swHide))
	}
	c.mu.Lock()
	c.hidden = true
	c.mu.Unlock()
}

func (c *ContentWebView) Show() {
	c.mu.Lock()
	ready := c.ready
	c.mu.Unlock()
	if !ready {
		return
	}
	c.raise()
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
	if b.X > 0 {
		c.leftPad = int32(b.X)
	}
	c.mu.Unlock()
	c.layoutHost()
}

func (c *ContentWebView) Ready() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.ready
}
