# Release checklist

1. `npm run build` and `wails build` succeed on Windows.
2. Version bump in `frontend/wails.json` `info.productVersion`.
3. Smoke test [TESTING.md](TESTING.md).
4. Tag release; attach binary if distributing.
5. Note PDF extract still optional without DLL.
