import { useCallback, useEffect, useMemo, useState } from "react";

const API_BASE = import.meta.env.VITE_API_URL ?? "";
const COLUMNS = [
  ["எண்", "எண்"],
  ["பெயர்", "பெயர்"],
  ["பதவி", "பதவி"],
  ["துறை", "துறை"],
  ["ஊர்", "ஊர்"],
  ["சம்பளம்", "மாதச் சம்பளம்"],
  ["சேர்ந்தஆண்டு", "சேர்ந்த ஆண்டு"],
];

const money = new Intl.NumberFormat("ta-IN-u-nu-latn", {
  style: "currency",
  currency: "INR",
  maximumFractionDigits: 0,
});

function SkeletonRows() {
  return Array.from({ length: 5 }, (_, row) => (
    <tr className="skeleton-row" key={row}>
      {COLUMNS.map(([key]) => (
        <td key={key}>
          <span className="skeleton" />
        </td>
      ))}
    </tr>
  ));
}

function App() {
  const [payload, setPayload] = useState({
    அட்டவணை: "ஊழியர்கள்",
    நெடுவரிசைகள்: COLUMNS.map(([key]) => key),
    பதிவுகள்: [],
  });
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [query, setQuery] = useState("");
  const [department, setDepartment] = useState("அனைத்தும்");
  const [sort, setSort] = useState({ key: "எண்", direction: "asc" });
  const [updatedAt, setUpdatedAt] = useState(null);

  const loadEmployees = useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      const response = await fetch(`${API_BASE}/api/employees`, {
        headers: { Accept: "application/json" },
      });
      const data = await response.json();
      if (!response.ok) {
        throw new Error(data.பிழை || `HTTP ${response.status}`);
      }
      setPayload(data);
      setUpdatedAt(new Date());
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : "தரவைப் பெற முடியவில்லை");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    loadEmployees();
  }, [loadEmployees]);

  const departments = useMemo(
    () => [
      "அனைத்தும்",
      ...new Set(payload.பதிவுகள்.map((employee) => employee.துறை)),
    ],
    [payload.பதிவுகள்],
  );

  const employees = useMemo(() => {
    const needle = query.trim().toLocaleLowerCase("ta");
    return payload.பதிவுகள்
      .filter((employee) => {
        const inDepartment =
          department === "அனைத்தும்" || employee.துறை === department;
        const searchable =
          `${employee.பெயர்} ${employee.பதவி} ${employee.துறை} ${employee.ஊர்}`.toLocaleLowerCase(
            "ta",
          );
        return inDepartment && (!needle || searchable.includes(needle));
      })
      .sort((left, right) => {
        const a = left[sort.key];
        const b = right[sort.key];
        const result =
          typeof a === "number"
            ? a - b
            : String(a).localeCompare(String(b), "ta");
        return sort.direction === "asc" ? result : -result;
      });
  }, [department, payload.பதிவுகள், query, sort]);

  const totalSalary = useMemo(
    () => employees.reduce((sum, employee) => sum + employee.சம்பளம், 0),
    [employees],
  );

  function changeSort(key) {
    setSort((current) => ({
      key,
      direction:
        current.key === key && current.direction === "asc" ? "desc" : "asc",
    }));
  }

  return (
    <main className="page-shell">
      <header className="hero">
        <div>
          <p className="eyebrow">அறம் · React · SQLite</p>
          <h1>ஊழியர் விவரங்கள்</h1>
          <p className="subtitle">
            அறம் REST சேவையிலிருந்து பெறப்பட்ட தமிழ் நிறுவனத் தரவுகள்
          </p>
        </div>
        <button
          className="refresh-button"
          disabled={loading}
          onClick={loadEmployees}
          type="button"
        >
          <span aria-hidden="true">↻</span>
          {loading ? "புதுப்பிக்கிறது…" : "புதுப்பி"}
        </button>
      </header>

      <section aria-label="சுருக்கம்" className="summary-grid">
        <article className="summary-card">
          <span>காணப்படும் ஊழியர்கள்</span>
          <strong>{employees.length}</strong>
        </article>
        <article className="summary-card">
          <span>துறைகள்</span>
          <strong>{Math.max(departments.length - 1, 0)}</strong>
        </article>
        <article className="summary-card salary-card">
          <span>மொத்த மாதச் சம்பளம்</span>
          <strong>{money.format(totalSalary)}</strong>
        </article>
      </section>

      <section className="data-panel">
        <div className="panel-heading">
          <div>
            <p className="table-label">SQLite அட்டவணை</p>
            <h2>{payload.அட்டவணை}</h2>
          </div>
          {updatedAt && (
            <time dateTime={updatedAt.toISOString()}>
              கடைசிப் புதுப்பிப்பு{" "}
              {updatedAt.toLocaleTimeString("ta-IN-u-nu-latn", {
                hour: "2-digit",
                minute: "2-digit",
              })}
            </time>
          )}
        </div>

        <div className="filters">
          <label className="search-field">
            <span>தேடல்</span>
            <input
              onChange={(event) => setQuery(event.target.value)}
              placeholder="பெயர், பதவி, துறை அல்லது ஊர்"
              type="search"
              value={query}
            />
          </label>
          <label>
            <span>துறை</span>
            <select
              onChange={(event) => setDepartment(event.target.value)}
              value={department}
            >
              {departments.map((item) => (
                <option key={item}>{item}</option>
              ))}
            </select>
          </label>
        </div>

        {error ? (
          <div className="error-state" role="alert">
            <strong>தரவைப் பெற முடியவில்லை</strong>
            <p>{error}</p>
            <p className="error-hint">
              அறம் backend சேவை 127.0.0.1:8080 முகவரியில் இயங்குகிறதா?
            </p>
            <button onClick={loadEmployees} type="button">
              மீண்டும் முயற்சி செய்
            </button>
          </div>
        ) : (
          <div className="table-wrap">
            <table>
              <thead>
                <tr>
                  {COLUMNS.map(([key, label]) => (
                    <th key={key} scope="col">
                      <button onClick={() => changeSort(key)} type="button">
                        {label}
                        <span aria-hidden="true">
                          {sort.key === key
                            ? sort.direction === "asc"
                              ? " ↑"
                              : " ↓"
                            : ""}
                        </span>
                      </button>
                    </th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {loading ? (
                  <SkeletonRows />
                ) : employees.length ? (
                  employees.map((employee) => (
                    <tr key={employee.எண்}>
                      <td className="number-cell">{employee.எண்}</td>
                      <td>
                        <div className="employee-name">
                          <span aria-hidden="true">
                            {employee.பெயர்.slice(0, 1)}
                          </span>
                          <strong>{employee.பெயர்}</strong>
                        </div>
                      </td>
                      <td>{employee.பதவி}</td>
                      <td>
                        <span className="department-pill">{employee.துறை}</span>
                      </td>
                      <td>{employee.ஊர்}</td>
                      <td className="number-cell">
                        {money.format(employee.சம்பளம்)}
                      </td>
                      <td className="number-cell">{employee.சேர்ந்தஆண்டு}</td>
                    </tr>
                  ))
                ) : (
                  <tr>
                    <td className="empty-state" colSpan={COLUMNS.length}>
                      தேடலுக்குப் பொருந்தும் ஊழியர்கள் இல்லை
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>
        )}

        <footer className="panel-footer">
          <span>{employees.length} பதிவுகள்</span>
          <span>API: /api/employees</span>
        </footer>
      </section>
    </main>
  );
}

export default App;
