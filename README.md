# Sentra

Developer-first CLI to scan for `.env*` files across local git repositories.

## Usage

```bash
go run ./cmd/sentra scan
```

`sentra scan` scans `~/dev`, detects git projects (directories containing a `.git/` folder), and lists detected `.env*` files grouped by project.
