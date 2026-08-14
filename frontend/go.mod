module frontend

go 1.22

require github.com/wailsapp/wails/v2 v2.14.0

// After pull, from frontend/:
//   go mod tidy
//   go install github.com/wailsapp/wails/v2/cmd/wails@latest   # optional CLI
//   wails dev   # or: go run .
