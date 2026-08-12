package bridge

import (
	"sync"
)

// Tab is one browser tab tracked on the Go side.
// History is a simple linear stack used until multi-webview lands.
type Tab struct {
	ID       int      `json:"id"`
	Title    string   `json:"title"`
	URL      string   `json:"url"`
	History  []string `json:"history"`
	HistIdx  int      `json:"histIdx"`
	CanBack  bool     `json:"canBack"`
	CanFwd   bool     `json:"canFwd"`
}

// TabSnapshot is a JSON-friendly view for the chrome.
type TabSnapshot struct {
	ID      int    `json:"id"`
	Title   string `json:"title"`
	URL     string `json:"url"`
	CanBack bool   `json:"canBack"`
	CanFwd  bool   `json:"canFwd"`
	Active  bool   `json:"active"`
}

// TabManager is the source of truth for open tabs.
type TabManager struct {
	mu        sync.Mutex
	tabs      []*Tab
	activeID  int
	nextID    int
}

func NewTabManager() *TabManager {
	m := &TabManager{nextID: 1}
	m.NewTab("New Tab", "")
	return m
}

func (m *TabManager) recomputeFlags(t *Tab) {
	t.CanBack = t.HistIdx > 0
	t.CanFwd = t.HistIdx >= 0 && t.HistIdx < len(t.History)-1
}

// NewTab creates and activates a tab. Returns the new tab id.
func (m *TabManager) NewTab(title, url string) int {
	m.mu.Lock()
	defer m.mu.Unlock()
	id := m.nextID
	m.nextID++
	t := &Tab{ID: id, Title: title, URL: url, HistIdx: -1}
	if url != "" {
		t.History = []string{url}
		t.HistIdx = 0
		t.Title = titleFromURL(url)
	}
	m.recomputeFlags(t)
	m.tabs = append(m.tabs, t)
	m.activeID = id
	return id
}

// CloseTab removes a tab. Ensures at least one remains.
func (m *TabManager) CloseTab(id int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	idx := -1
	for i, t := range m.tabs {
		if t.ID == id {
			idx = i
			break
		}
	}
	if idx < 0 {
		return
	}
	m.tabs = append(m.tabs[:idx], m.tabs[idx+1:]...)
	if len(m.tabs) == 0 {
		// recreate under lock
		nid := m.nextID
		m.nextID++
		t := &Tab{ID: nid, Title: "New Tab", HistIdx: -1}
		m.tabs = []*Tab{t}
		m.activeID = nid
		return
	}
	if m.activeID == id {
		if idx >= len(m.tabs) {
			idx = len(m.tabs) - 1
		}
		m.activeID = m.tabs[idx].ID
	}
}

// Activate sets the active tab.
func (m *TabManager) Activate(id int) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, t := range m.tabs {
		if t.ID == id {
			m.activeID = id
			return true
		}
	}
	return false
}

// Active returns a copy of the active tab, or nil.
func (m *TabManager) Active() *Tab {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, t := range m.tabs {
		if t.ID == m.activeID {
			cp := *t
			cp.History = append([]string(nil), t.History...)
			return &cp
		}
	}
	return nil
}

// Navigate records url on the active tab and returns it.
func (m *TabManager) Navigate(url string) *Tab {
	m.mu.Lock()
	defer m.mu.Unlock()
	t := m.activeLocked()
	if t == nil {
		return nil
	}
	// Drop any forward entries when navigating from mid-stack.
	if t.HistIdx >= 0 && t.HistIdx < len(t.History)-1 {
		t.History = t.History[:t.HistIdx+1]
	}
	t.History = append(t.History, url)
	t.HistIdx = len(t.History) - 1
	t.URL = url
	t.Title = titleFromURL(url)
	m.recomputeFlags(t)
	cp := *t
	cp.History = append([]string(nil), t.History...)
	return &cp
}

// Back moves history back on the active tab. ok=false if none.
func (m *TabManager) Back() (url string, ok bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	t := m.activeLocked()
	if t == nil || t.HistIdx <= 0 {
		return "", false
	}
	t.HistIdx--
	t.URL = t.History[t.HistIdx]
	t.Title = titleFromURL(t.URL)
	m.recomputeFlags(t)
	return t.URL, true
}

// Forward moves history forward on the active tab.
func (m *TabManager) Forward() (url string, ok bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	t := m.activeLocked()
	if t == nil || t.HistIdx < 0 || t.HistIdx >= len(t.History)-1 {
		return "", false
	}
	t.HistIdx++
	t.URL = t.History[t.HistIdx]
	t.Title = titleFromURL(t.URL)
	m.recomputeFlags(t)
	return t.URL, true
}

// CurrentURL of the active tab.
func (m *TabManager) CurrentURL() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	t := m.activeLocked()
	if t == nil {
		return ""
	}
	return t.URL
}

// Snapshot returns all tabs for the chrome UI.
func (m *TabManager) Snapshot() []TabSnapshot {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]TabSnapshot, 0, len(m.tabs))
	for _, t := range m.tabs {
		out = append(out, TabSnapshot{
			ID:      t.ID,
			Title:   t.Title,
			URL:     t.URL,
			CanBack: t.CanBack,
			CanFwd:  t.CanFwd,
			Active:  t.ID == m.activeID,
		})
	}
	return out
}

func (m *TabManager) activeLocked() *Tab {
	for _, t := range m.tabs {
		if t.ID == m.activeID {
			return t
		}
	}
	return nil
}

func titleFromURL(url string) string {
	if url == "" {
		return "New Tab"
	}
	// strip scheme and path for a short tab title
	s := url
	for _, p := range []string{"https://", "http://"} {
		if len(s) > len(p) && s[:len(p)] == p {
			s = s[len(p):]
			break
		}
	}
	for i := 0; i < len(s); i++ {
		if s[i] == '/' || s[i] == '?' {
			s = s[:i]
			break
		}
	}
	if s == "" {
		return "Tab"
	}
	return s
}
