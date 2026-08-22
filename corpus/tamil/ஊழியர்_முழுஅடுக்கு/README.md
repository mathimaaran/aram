# ஊழியர் விவரங்கள் — Niraluli + React + SQLite

A complete Tamil full-stack corpus:

```text
React grid → REST /api/employees → Niraluli HTTP server → SQLite
```

**Setup (start here):** [`SETUP.md`](SETUP.md) — prerequisites, backend, frontend, verify, and troubleshooting.

- UI labels, table name, column names, JSON fields, and text data are Tamil.
- IDs, salaries, and years remain numeric.
- API paths stay ASCII so browsers and command-line clients do not need
  percent-encoded Tamil URLs.

## Database schema

SQLite file: `build/ஊழியர்கள்.db` (created from the repository root).

```sql
CREATE TABLE ஊழியர்கள் (
    எண் INTEGER PRIMARY KEY,
    பெயர் TEXT NOT NULL,
    பதவி TEXT NOT NULL,
    துறை TEXT NOT NULL,
    ஊர் TEXT NOT NULL,
    சம்பளம் INTEGER NOT NULL,
    சேர்ந்தஆண்டு INTEGER NOT NULL
);
```

The backend uses `INSERT OR IGNORE`, so restarting it safely keeps existing
rows and only adds missing seed IDs.

## 1. Start the Niraluli backend

See **[`SETUP.md`](SETUP.md)** for the full walkthrough. Short form from the
repository root:

```bash
.tools/go/bin/go run ./cmd/uli run \
  corpus/tamil/ஊழியர்_முழுஅடுக்கு/backend
```

Expected startup:

```text
wrote build/backend.c and build/backend
நிரலுளி REST சேவை: http://127.0.0.1:8080
SQLite தரவுத்தளம்: build/ஊழியர்கள்.db
```

The process keeps running. Test it from another terminal:

```bash
curl http://127.0.0.1:8080/api/health
curl http://127.0.0.1:8080/api/employees
```

Endpoints:

| Method | Path | Purpose |
|--------|------|---------|
| `GET` | `/api/employees` | Tamil schema metadata and employee rows |
| `OPTIONS` | `/api/employees` | CORS preflight |
| `GET` | `/api/health` | Service health |

The employee response shape is:

```json
{
  "அட்டவணை": "ஊழியர்கள்",
  "நெடுவரிசைகள்": [
    "எண்",
    "பெயர்",
    "பதவி",
    "துறை",
    "ஊர்",
    "சம்பளம்",
    "சேர்ந்தஆண்டு"
  ],
  "பதிவுகள்": [
    {
      "எண்": 1,
      "பெயர்": "அருண்",
      "பதவி": "மென்பொருள் பொறியாளர்",
      "துறை": "தொழில்நுட்பம்",
      "ஊர்": "சென்னை",
      "சம்பளம்": 85000,
      "சேர்ந்தஆண்டு": 2021
    }
  ]
}
```

## 2. Start the React UI

See **[`SETUP.md`](SETUP.md)** for Node version notes and troubleshooting.
Short form:

```bash
cd corpus/tamil/ஊழியர்_முழுஅடுக்கு/frontend
npm install
npm run dev
```

Open <http://localhost:5173>. Vite proxies `/api` to the Niraluli server at
`127.0.0.1:8080`.

The UI includes:

- responsive employee grid;
- Tamil search across name, role, department, and city;
- department filtering;
- sortable columns;
- employee, department, and salary summaries;
- loading, empty, and backend-error states;
- manual refresh.

## Production UI configuration

When the frontend and API are hosted separately, create `.env`:

```text
VITE_API_URL=http://your-uli-api.example:8080
```

Then:

```bash
npm run build
```

The Niraluli backend sends permissive `Access-Control-Allow-Origin: *` because this
is a corpus example. Restrict that header to your real frontend origin before
production use.

## Reset the example

Stop the backend and delete `build/ஊழியர்கள்.db`. It will be recreated and
seeded on the next run.

## Implementation notes

- Each API request opens its own SQLite connection. This respects the v0.62
  rule that one database handle must not be shared concurrently across
  `இழை`.
- Prepared statements read rows through typed column functions.
- The backend includes small integer-to-decimal and JSON-string escaping
  helpers because Niraluli does not yet include general JSON/formatting packages.
- This corpus is read-only REST. POST/PUT/DELETE can be layered on the same
  prepared-statement API in a later example.
