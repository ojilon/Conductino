# Changes applied to frontend/content_webview_windows.go

This file was modified on branch `add_foundation_for_ai_related_tools` to fix two issues that prevented the dual WebView2 path from reliably creating and using a content WebView2 controller on Windows.

Summary of changes

1) Fix GWLP_WNDPROC conversion
- Problem: the previous code attempted to build a `uintptr` value for GWLP_WNDPROC by converting a negative int32 to uint32 and then to uintptr. That produced a large unsigned value instead of the intended -4 index, so `SetWindowLongPtrW` never installed the subclass correctly.
- Change: replaced the `signedMinusFour/uint32` conversion with a clear constant and conversion that preserves the negative index:
  const GWLP_WNDPROC = -4
  var gwlpWndProc = uintptr(int32(GWLP_WNDPROC))
- Effect: subclassParent now passes the correct index to `SetWindowLongPtrW` and the custom `parentSubclassProc` will be called for the target window.

2) Use PostMessage instead of SendMessage for UI requests
- Problem: the code used `SendMessageW` to send creation / navigation / eval / visibility messages. `SendMessage` blocks and can cause nested window-proc re-entrancy and COM/message-pump issues when creating WebView2 controllers on the UI thread.
- Change: switched message dispatch to `PostMessageW` (non-blocking). The code already uses `createDone` / `navDone` channels to wait for completion, so queuing the request with `PostMessage` is the safer pattern.
- Concrete edits:
  - Replaced the direct `procSendMessageW.Call(parentHWND, wmCreateContent, 0, 0)` call with `procPostMessageW.Call(...)` in `ensureContentBrowser`.
  - Replaced the implementation of `sendToUI` to call `procPostMessageW.Call(...)` instead of `procSendMessageW`.
  - Removed the now-unused `procSendMessageW` variable declaration.

Files changed

- frontend/content_webview_windows.go (modified)
  - GWLP_WNDPROC conversion corrected
  - Message sending changed from SendMessage -> PostMessage
  - Minor comment and log improvements

Why both fixes are required

- If subclassing isn't installed correctly (GWLP bug), posted messages won't be handled by your subclass proc and the create/navigation messages never run.
- If subclassing is correct but you keep using `SendMessage`, you can still hit COM initialization / nested-pump re-entrancy issues when creating the WebView2 controller. Using `PostMessage` avoids those nested-pump problems.

Testing checklist (local)

1) Build and run on Windows with the WebView2 runtime installed.
2) Start the app from the `add_foundation_for_ai_related_tools` branch.
3) Check logs for messages added by the file:
   - "[content] subclassed parent HWND=..." should appear after subclassing.
   - "[content] UI-thread Embed host=..." and "[content] WebView2 ready on UI thread" should appear during creation.
   - "[content] Navigate ..." should appear when navigation is requested.
4) If the content view is not visible, use Spy++ or WinSpy to confirm a child HWND named "ConductinoContent" exists and has non-zero bounds and visible style.
5) Try navigation to a remote url and verify the page loads.

Notes and follow-ups

- Z-order / clipping: once subclassing + PostMessage are fixed, verify the child window has correct Z-order and the Wails host doesn't cover it. If needed, tune SetWindowPos flags or WS_ styles.
- Ready gating: if you observe races where Embed returns but the controller isn't fully usable, consider using Chromium callbacks (if available) or increasing the readiness gating beyond the current sleep.

If you'd like, I can also:
- Run a repo-wide search to find any other occurrences of SendMessageW / SetWindowLongPtrW that may need the same treatment.
- Open a pull request on branch `add_foundation_for_ai_related_tools` with these changes.

