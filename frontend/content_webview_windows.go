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
	wmAppBase           = 0x8000
	wmCreateContent     = wmAppBase + 0x51
	wmNavigateContent   = wmAppBase + 0x52
	wmResizeContent     = wmAppBase + 0x53
	wmSetVisibleContent = wmAppBase + 0x54
	wmEvalContent       = wmAppBase + 0x55

	defaultChromeHeight = 82
	contentClassName    = "ConductinoContentHost"

	COINIT_APARTMENTTHREADED = 0x2
	COINIT_MULTITHREADED     = 0x0

	WS_CHILD        = 0x40000000
	WS_CLIPSIBLINGS = 0x04000000
	WS_CLIPCHILDREN = 0x02000000

	SWP_NOZORDER   = 0x0004
	SWP_NOACTIVATE = 0x0010
	SWP_SHOWWINDOW = 0x0040
	SWP_HIDEWINDOW = 0x0080
	SW_SHOW        = 5
	SW_HIDE        = 0
)

// GWLP_WNDPROC index (-4). Used as window long pointer value.
const GWLP_WNDPROC = -4

var gwlpWndProc int

func init() {
	// Initialize gwlpWndProc with proper -4 to uintptr conversion
	// -4 as int32 converted to uintptr gives the correct bit pattern for GWLP_WNDPROC
	gwlpWndProc = int(int32(GWLP_WNDPROC))
}


var (
	user32                       = windows.NewLazySystemDLL("user32.dll")
	procCreateWindowExW          = user32.NewProc("CreateWindowExW")
	procShowWindow               = user32.NewProc("ShowWindow")
	procSetWindowPos             = user32.NewProc("SetWindowPos")
	procGetClientRect            = user32.NewProc("GetClientRect")
	procRegisterClassExW         = user32.NewProc("RegisterClassExW")
	procDefWindowProcW           = user32.NewProc("DefWindowProcW")
	procCallWindowProcW          = user32.NewProc("CallWindowProcW")
	procGetWindowTextW           = user32.NewProc("GetWindowTextW")
	procIsWindowVisible          = user32.NewProc("IsWindowVisible")
	procEnumWindows              = user32.NewProc("EnumWindows")
	procGetWindowThreadProcessId = user32.NewProc("GetWindowThreadProcessId")
	procSetWindowLongPtrW        = user32.NewProc("SetWindowLongPtrW")
	procPostMessageW             = user32.NewProc("PostMessageW")

	kernel32                 = windows.NewLazySystemDLL("kernel32.dll")
	procGetCurrentProcessId = kernel32.NewProc("GetCurrentProcessId")
	procGetModuleHandleW    = kernel32.NewProc("GetModuleHandleW")

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

	subclassMu  sync.Mutex
	origWndProc uintptr
	subclassed  bool
	parentHWND  uintptr

	pendingURL    string
	pendingShow   bool
	pendingEval   string
	pendingChrome int32 = defaultChromeHeight

	createDone chan error
	navDone    chan error
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
}

func contentWndProc(hwnd uintptr, msg uint32, wParam, lParam uintptr) uintptr {
	r, _, _ := procDefWindowProcW.Call(hwnd, uintptr(msg), wParam, lParam)
	return r
}

func parentSubclassProc(hwnd uintptr, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case wmCreateContent:
		err := uiCreateContent(hwnd)
		if createDone != nil {
			select {
			case createDone <- err:
			default:
			}
		}
		return 0
	case wmNavigateContent:
		err := uiNavigate()
		if navDone != nil {
			select {
			case navDone <- err:
			default:
			}
		}
		return 0
	case wmResizeContent:
		uiResize()
		return 0
	case wmSetVisibleContent:
		uiSetVisible(pendingShow)
		return 0
	case wmEvalContent:
		uiEval(pendingEval)
		return 0
	case 0x0005: // WM_SIZE
		uiResize()
	}
	if origWndProc != 0 {
		r, _, _ := procCallWindowProcW.Call(origWndProc, hwnd, uintptr(msg), wParam, lParam)
		return r
	}
	r, _, _ := procDefWindowProcW.Call(hwnd, uintptr(msg), wParam, lParam)
	return r
}

var parentSubclassCallback = syscall.NewCallback(parentSubclassProc)

func subclassParent(hwnd uintptr) error {
	subclassMu.Lock()
	defer subclassMu.Unlock()
	if subclassed && parentHWND == hwnd {
		return nil
	}
	prev, _, _ := procSetWindowLongPtrW.Call(hwnd, uintptr(gwlpWndProc), parentSubclassCallback)
	origWndProc = prev
	parentHWND = hwnd
	subclassed = true
	log.Printf("[content] subclassed parent HWND=%v", hwnd)
	return nil
}

func postToUI(msg uint32) {
	if parentHWND == 0 {
		return
	}
	procPostMessageW.Call(parentHWND, uintptr(msg), 0, 0)
}

func sendToUI(msg uint32, _ uintptr) {
	// Use PostMessage to avoid nested message-pump re-entrancy and COM initialization
	// issues that can happen with SendMessage. The handlers use createDone/navDone
	// to signal completion.
	if parentHWND == 0 {
		return
	}
	procPostMessageW.Call(parentHWND, uintptr(msg), 0, 0)
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

func uiCreateContent(parent uintptr) error {
	ensureCOM()
	if err := registerContentClass(); err != nil {
		return err
	}

	contentMu.Lock()
	if contentBrowser != nil && contentBrowser.ready {
		contentMu.Unlock()
		return nil
	}
	contentMu.Unlock()

	className, _ := windows.UTF16PtrFromString(contentClassName)
	windowName, _ := windows.UTF16PtrFromString("ConductinoContent")
	ret, _, callErr := procGetModuleHandleW.Call(0)
	if ret == 0 {
		return fmt.Errorf("GetModuleHandleW: %v", callErr)
	}
	hInst := windows.Handle(ret)

	chromeH := int32(defaultChromeHeight)
	if pendingChrome >= 40 {
		chromeH = pendingChrome
	}

	var rc rect
	procGetClientRect.Call(parent, uintptr(unsafe.Pointer(&rc)))
	w := rc.Right - rc.Left
	h := rc.Bottom - rc.Top - chromeH
	if w < 200 {
		w = 200
	}
	if h < 200 {
		h = 200
	}

	host, _, callErr := procCreateWindowExW.Call(
		0,
		uintptr(unsafe.Pointer(className)),
		uintptr(unsafe.Pointer(windowName)),
		uintptr(WS_CHILD|WS_CLIPSIBLINGS|WS_CLIPCHILDREN),
		0, uintptr(chromeH), uintptr(w), uintptr(h),
		parent,
		0,
		uintptr(hInst),
		0,
	)
    
    if host == 0 {
        return fmt.Errorf("CreateWindowEx content host: %v", callErr)
    }
    // Fix Z-order: ensure content host is above the Wails webview controller.
    // The Wails host fills the client area; we must insert our child host above it.
    // SWP_NOMOVE = 0x0002, SWP_NOSIZE = 0x0001 - keep current position/size
    procSetWindowPos.Call(host, 0, 0, 0, 0, 0, uintptr(0x0002)|uintptr(0x0001)|SWP_NOACTIVATE|SWP_SHOWWINDOW)

    procShowWindow.Call(host, uintptr(SW_SHOW))

    chromium := edge.NewChromium()
	chromium.DataPath = contentDataPath()
	log.Printf("[content] UI-thread Embed host=%v", host)

	if !chromium.Embed(host) {
		return fmt.Errorf("Embed returned false on UI thread")
	}

	// Check that the WebView2 controller is available after Embed.
	// Embed() may return before the controller is fully initialized.
	// We verify the controller exists; if not, Embed didn't create one.
	if chromium.GetController() == nil {
		return fmt.Errorf("WebView2 controller is nil after Embed")
	}

	// Brief delay to allow WebView2 to finish initial internal setup.
	// After this, we re-verify the controller is still valid.
	time.Sleep(500 * time.Millisecond)

	if chromium.GetController() == nil {
		return fmt.Errorf("WebView2 controller became nil after initialization delay")
	}

	chromium.Resize()

	cb := &ContentBrowser{
		parent:   parent,
		host:     host,
		chromium: chromium,
		chromeH:  chromeH,
		visible:  false,
		ready:    true,
	}
	contentMu.Lock()
	contentBrowser = cb
	contentMu.Unlock()

	procShowWindow.Call(host, uintptr(SW_HIDE))
	log.Printf("[content] WebView2 ready on UI thread")
	return nil
}

func uiNavigate() error {
	contentMu.Lock()
	cb := contentBrowser
	url := pendingURL
	contentMu.Unlock()
	if cb == nil || !cb.ready || cb.chromium == nil {
		return fmt.Errorf("content browser not ready")
	}
	cb.lastURL = url
	cb.visible = true
	uiLayout(cb)
	procShowWindow.Call(cb.host, uintptr(SW_SHOW))
	cb.chromium.Navigate(url)
	log.Printf("[content] Navigate %s", url)
	return nil
}

func uiLayout(cb *ContentBrowser) {
	if cb == nil || cb.parent == 0 || cb.host == 0 {
		return
	}
	var rc rect
	procGetClientRect.Call(cb.parent, uintptr(unsafe.Pointer(&rc)))
	top := cb.chromeH
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
	flags := uintptr(SWP_NOZORDER | SWP_NOACTIVATE)
	if cb.visible {
		flags |= SWP_SHOWWINDOW
	} else {
		flags |= SWP_HIDEWINDOW
	}
	procSetWindowPos.Call(cb.host, 0, 0, uintptr(top), uintptr(w), uintptr(h), flags)
	if cb.chromium != nil && cb.ready {
		cb.chromium.Resize()
	}
}

func uiResize() {
	contentMu.Lock()
	cb := contentBrowser
	contentMu.Unlock()
	if cb != nil {
		uiLayout(cb)
	}
}

func uiSetVisible(show bool) {
	contentMu.Lock()
	cb := contentBrowser
	contentMu.Unlock()
	if cb == nil {
		return
	}
	cb.visible = show
	uiLayout(cb)
	if show {
		procShowWindow.Call(cb.host, uintptr(SW_SHOW))
	} else {
		procShowWindow.Call(cb.host, uintptr(SW_HIDE))
	}
}

func uiEval(js string) {
	contentMu.Lock()
	cb := contentBrowser
	contentMu.Unlock()
	if cb != nil && cb.chromium != nil && cb.ready {
		cb.chromium.Eval(js)
	}
}

func ensureContentBrowser() (*ContentBrowser, error) {
	contentMu.Lock()
	if contentBrowser != nil && contentBrowser.ready {
		cb := contentBrowser
		contentMu.Unlock()
		return cb, nil
	}
	contentMu.Unlock()

	hwnd := findAppHWND("Conductino")
	if hwnd == 0 {
		return nil, fmt.Errorf("app window HWND not found")
	}
	if err := subclassParent(hwnd); err != nil {
		return nil, err
	}

	createDone = make(chan error, 1)
	procPostMessageW.Call(parentHWND, uintptr(wmCreateContent), 0, 0)

	select {
	case err := <-createDone:
		if err != nil {
			return nil, err
		}
	case <-time.After(12 * time.Second):
		contentMu.Lock()
		cb := contentBrowser
		contentMu.Unlock()
		if cb != nil && cb.ready {
			return cb, nil
		}
		return nil, fmt.Errorf("timeout waiting for UI-thread WebView2 create")
	}

	contentMu.Lock()
	defer contentMu.Unlock()
	if contentBrowser == nil || !contentBrowser.ready {
		return nil, fmt.Errorf("content browser still not ready")
	}
	return contentBrowser, nil
}

func (a *App) ContentEnsure() error {
	_, err := ensureContentBrowser()
	return err
}

func (a *App) ContentNavigate(url string) error {
	url = strings.TrimSpace(url)
	if url == "" {
		return fmt.Errorf("empty url")
	}
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") &&
		!strings.HasPrefix(url, "about:") {
		url = "https://" + url
	}
	if _, err := ensureContentBrowser(); err != nil {
		return err
	}
	pendingURL = url
	navDone = make(chan error, 1)
	sendToUI(wmNavigateContent, 5000)
	select {
	case err := <-navDone:
		return err
	case <-time.After(5 * time.Second):
		return nil
	}
}

func (a *App) ContentSetVisible(show bool) error {
	if _, err := ensureContentBrowser(); err != nil {
		return err
	}
	pendingShow = show
	sendToUI(wmSetVisibleContent, 2000)
	return nil
}

func (a *App) ContentResize() error {
	if parentHWND == 0 {
		return nil
	}
	postToUI(wmResizeContent)
	return nil
}

func (a *App) ContentSetChromeHeight(px int) error {
	if px < 40 {
		px = defaultChromeHeight
	}
	pendingChrome = int32(px)
	contentMu.Lock()
	if contentBrowser != nil {
		contentBrowser.chromeH = int32(px)
	}
	contentMu.Unlock()
	postToUI(wmResizeContent)
	return nil
}

func (a *App) ContentEval(js string) error {
	if _, err := ensureContentBrowser(); err != nil {
		return err
	}
	pendingEval = js
	sendToUI(wmEvalContent, 2000)
	return nil
}

func (a *App) ContentLastURL() string {
	contentMu.Lock()
	defer contentMu.Unlock()
	if contentBrowser == nil {
		return ""
	}
	return contentBrowser.lastURL
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

func (a *App) ContentRecreateAsPopup() error {
	return a.ContentEnsure()
}
