# Setup instructions — ஊழியர் விவரங்கள் (Niraluli + React + SQLite)

Step-by-step guide for this full-stack corpus. You need **two terminals**:
one for the Niraluli backend, one for the React frontend.

```text
Terminal A: Niraluli REST API  →  http://127.0.0.1:8080
Terminal B: React UI       →  http://localhost:5173
```

Overview docs and schema notes: [`README.md`](README.md).

---

## Prerequisites

| Tool | Why | Check |
|------|-----|--------|
| Linux + GCC/Clang | Niraluli C backend | `gcc --version` or `cc --version` |
| Go toolchain | Run `cmd/uli` | `.tools/go/bin/go version` or `go version` |
| `libsqlite3.so.0` | SQLite at runtime | `ldconfig -p \| grep libsqlite3` |
| Node.js **18+** | Vite 5 frontend | `node --version` |
| npm | Install frontend deps | `npm --version` |

Optional:

```bash
# From the repository root, if you use the portable Go tree:
source tools/env.sh
```

Install SQLite runtime if missing (Debian/Ubuntu):

```bash
sudo apt install libsqlite3-0
```

---

## Layout

```text
corpus/tamil/ஊழியர்_முழுஅடுக்கு/
├── SETUP.md          ← this file
├── README.md         ← design / schema / API notes
├── backend/          ← Niraluli REST server
│   └── தொடக்கம்.uli
└── frontend/         ← React + Vite UI
    ├── package.json
    ├── vite.config.js
    └── src/
```

SQLite file created on first backend run (from **repo root**):

```text
build/ஊழியர்கள்.db
```

---

## Step 1 — Start the Niraluli backend (Terminal A)

Always start from the **repository root** (`go_spec/`), not from `frontend/`:

```bash
cd /path/to/go_spec

.tools/go/bin/go run ./cmd/uli run \
  corpus/tamil/ஊழியர்_முழுஅடுக்கு/backend
```

If `.tools/go` is not present and system Go works:

```bash
go run ./cmd/uli run \
  corpus/tamil/ஊழியர்_முழுஅடுக்கு/backend
```

**Leave this terminal open.** The server must keep running.

Expected output:

```text
wrote build/backend.c and build/backend
நிரலுளி REST சேவை: http://127.0.0.1:8080
SQLite தரவுத்தளம்: build/ஊழியர்கள்.db
```

### Verify the API (Terminal B or a third terminal)

```bash
curl http://127.0.0.1:8080/api/health
curl http://127.0.0.1:8080/api/employees
```

Healthy responses:

- `/api/health` → JSON with `"நிலை":"நலம்"`
- `/api/employees` → JSON with `"அட்டவணை":"ஊழியர்கள்"` and 6 seeded rows

Endpoints:

| Method | Path | Purpose |
|--------|------|---------|
| `GET` | `/api/employees` | Employee grid data |
| `OPTIONS` | `/api/employees` | CORS preflight |
| `GET` | `/api/health` | Health check |

---

## Step 2 — Start the React frontend (Terminal B)

Node **18.x or newer** is required. Vite 5 is pinned in `package.json` for Node 18.

```bash
cd /path/to/go_spec/corpus/tamil/ஊழியர்_முழுஅடுக்கு/frontend

# First time, or after package.json changes:
rm -rf node_modules package-lock.json
npm install

npm run dev
```

Expected output includes something like:

```text
  VITE v5.x.x  ready in …
  ➜  Local:   http://localhost:5173/
```

Open:

```text
http://localhost:5173
```

Vite proxies `/api` → `http://127.0.0.1:8080`, so the UI talks to the Niraluli backend without CORS issues in development.

---

## Step 3 — Use the UI

You should see:

- Tamil title **ஊழியர் விவரங்கள்**
- Summary cards (count, departments, total salary)
- Searchable / filterable / sortable employee grid
- Data loaded from SQLite through Niraluli REST

If the grid shows an error, the backend is usually not running or not on port 8080.

---

## Stop / restart

| Action | How |
|--------|-----|
| Stop frontend | `Ctrl+C` in Terminal B |
| Stop backend | `Ctrl+C` in Terminal A |
| Reset SQLite seed DB | stop backend, then `rm -f build/ஊழியர்கள்.db` from repo root; restart backend |

---

## Troubleshooting

### Backend: `libsqlite3` / SQLite runtime unavailable

Install the runtime library and restart:

```bash
sudo apt install libsqlite3-0
```

### Backend: address already in use (`8080`)

Something else is bound to 8080. Find and stop it:

```bash
ss -ltnp | grep 8080
# or
lsof -i :8080
```

### Frontend: `styleText` / Rolldown error on Node 18

You had an old `latest` Vite install. Reinstall with the pinned Vite 5 deps:

```bash
cd corpus/tamil/ஊழியர்_முழுஅடுக்கு/frontend
rm -rf node_modules package-lock.json
npm install
npm run dev
```

### Frontend: empty grid / “தரவைப் பெற முடியவில்லை”

1. Confirm backend is running (`curl http://127.0.0.1:8080/api/health`).
2. Confirm you started backend from the **repo root** so `build/ஊழியர்கள்.db` is created in the right place.
3. Hard-refresh the browser.

### Frontend: Node too old

```bash
node --version   # need >= 18
```

Upgrade Node (nvm, NodeSource, or your distro packages), then reinstall frontend deps.

### Wrong working directory for backend

If you run `uli` from inside `frontend/`, the relative DB path `build/ஊழியர்கள்.db` will not land in the repo `build/` folder. Always run the Niraluli command from the repository root.

---

## Quick copy-paste (two terminals)

**Terminal A — backend**

```bash
cd /path/to/go_spec
.tools/go/bin/go run ./cmd/uli run \
  corpus/tamil/ஊழியர்_முழுஅடுக்கு/backend
```

**Terminal B — frontend**

```bash
cd /path/to/go_spec/corpus/tamil/ஊழியர்_முழுஅடுக்கு/frontend
npm install
npm run dev
```

Then open <http://localhost:5173>.
