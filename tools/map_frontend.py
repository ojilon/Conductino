from pathlib import Path

# ==========================================================
# Configuration
# ==========================================================

TARGET_FOLDER = "frontend-ui"      # <-- Change only this
OUTPUT_FILE = "frontend_combined.txt"

INCLUDE = {
    ".html",
    ".css",
    ".js",
    ".ts",
    ".go",
    ".json",
    ".txt",
    ".md",
}

# ==========================================================
# Locate project root
# ==========================================================

current = Path.cwd()

while current != current.parent:
    if (current / ".git").exists():
        PROJECT_ROOT = current
        break
    current = current.parent
else:
    raise RuntimeError("Could not locate project root (.git folder not found).")

# ==========================================================
# Locate requested folder
# ==========================================================

matches = [
    p for p in PROJECT_ROOT.rglob(TARGET_FOLDER)
    if p.is_dir()
]

if not matches:
    raise RuntimeError(f"Folder '{TARGET_FOLDER}' was not found.")

target = matches[0]

print(f"Project Root : {PROJECT_ROOT}")
print(f"Target Folder: {target}")

# ==========================================================
# Collect files
# ==========================================================

files = sorted(
    [
        p for p in target.rglob("*")
        if p.is_file() and p.suffix.lower() in INCLUDE
    ]
)

# ==========================================================
# Write output
# ==========================================================

output = PROJECT_ROOT / OUTPUT_FILE

total_files = 0
total_lines = 0

with output.open("w", encoding="utf-8") as out:

    out.write("=" * 90 + "\n")
    out.write("PROJECT SOURCE EXPORT\n")
    out.write("=" * 90 + "\n")
    out.write(f"Project : {PROJECT_ROOT}\n")
    out.write(f"Folder  : {target.relative_to(PROJECT_ROOT)}\n")
    out.write("=" * 90 + "\n\n")

    for file in files:

        rel = file.relative_to(PROJECT_ROOT)

        try:
            text = file.read_text(encoding="utf-8")
        except UnicodeDecodeError:
            print(f"Skipping binary file: {rel}")
            continue

        lines = text.splitlines()

        total_files += 1
        total_lines += len(lines)

        out.write("\n")
        out.write("=" * 90 + "\n")
        out.write(f"FILE : {rel}\n")
        out.write(f"LINES: {len(lines)}\n")
        out.write("=" * 90 + "\n\n")

        width = len(str(len(lines)))

        for number, line in enumerate(lines, start=1):
            out.write(f"{number:>{width}} | {line}\n")

        out.write("\n\n")

    out.write("\n")
    out.write("=" * 90 + "\n")
    out.write("SUMMARY\n")
    out.write("=" * 90 + "\n")
    out.write(f"Files Exported : {total_files}\n")
    out.write(f"Total Lines    : {total_lines}\n")

print("\nDone!")
print(f"Output: {output}")
print(f"Files : {total_files}")
print(f"Lines : {total_lines}")