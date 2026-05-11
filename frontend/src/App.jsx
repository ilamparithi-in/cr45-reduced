import { useEffect, useState } from "react";

const CLASS_ID = import.meta.env.VITE_CLASS_ID || "107125";
const TOKEN_STORAGE_KEY = "cr45_reduced_token";

async function requestJSON(path, options = {}) {
  const response = await fetch(path, options);
  const contentType = response.headers.get("content-type") || "";
  const isJSON = contentType.includes("application/json");
  const payload = isJSON ? await response.json() : null;
  if (!response.ok) {
    const message = payload?.error || `Request failed (${response.status})`;
    throw new Error(message);
  }
  if (!isJSON && response.status !== 204) {
    throw new Error(`Expected application/json but received ${contentType || "(missing Content-Type)"}`);
  }
  return {
    data: payload,
    headers: {
      contentType,
      cacheControl: response.headers.get("cache-control") || "",
      contentLength: response.headers.get("content-length") || "",
      xContentTypeOptions: response.headers.get("x-content-type-options") || ""
    }
  };
}

export function App() {
  const [slots, setSlots] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [headerNotes, setHeaderNotes] = useState(null);
  const [adminMessage, setAdminMessage] = useState("");
  const [loginMessage, setLoginMessage] = useState("");
  const [token, setToken] = useState(() => window.localStorage.getItem(TOKEN_STORAGE_KEY) || "");
  const [loginForm, setLoginForm] = useState({ username: "admin", password: "" });
  const [form, setForm] = useState({
    slot_index: "2",
    course_code: "CSIR11",
    start_time: "10:30",
    end_time: "11:20",
    venue: "LH210",
    status: "updated"
  });

  async function refreshTimetable() {
    const result = await requestJSON(`/api/timetable?class_id=${encodeURIComponent(CLASS_ID)}`);
    setSlots(result.data?.slots || []);
    setHeaderNotes(result.headers);
    return result.data?.slots || [];
  }

  const nextSlotIndex = String(slots.length + 1);

  useEffect(() => {
    let mounted = true;

    async function boot() {
      setLoading(true);
      setError("");
      try {
        const result = await requestJSON(`/api/timetable?class_id=${encodeURIComponent(CLASS_ID)}`);
        if (mounted) {
          setSlots(result.data?.slots || []);
          setHeaderNotes(result.headers);
          setForm((prev) => ({ ...prev, slot_index: String((result.data?.slots || []).length + 1) }));
        }
      } catch (err) {
        if (mounted) {
          setError(err.message || "Failed to load timetable");
        }
      } finally {
        if (mounted) {
          setLoading(false);
        }
      }
    }

    void boot();
    return () => {
      mounted = false;
    };
  }, []);

  function logout() {
    window.localStorage.removeItem(TOKEN_STORAGE_KEY);
    setToken("");
    setLoginMessage("");
    setAdminMessage("");
  }

  async function submitLogin(event) {
    event.preventDefault();
    setLoginMessage("");

    try {
      const result = await requestJSON("/api/auth/login", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Accept: "application/json"
        },
        body: JSON.stringify(loginForm)
      });
      const nextToken = result.data?.token || "";
      window.localStorage.setItem(TOKEN_STORAGE_KEY, nextToken);
      setToken(nextToken);
      setLoginMessage("Logged in.");
    } catch (err) {
      setLoginMessage(err.message || "Login failed");
    }
  }

  async function submitOverride(event) {
    event.preventDefault();
    setAdminMessage("");

    try {
      await requestJSON("/api/admin/override", {
        method: "POST",
        headers: {
          Authorization: `Bearer ${token}`,
          "Content-Type": "application/json",
          Accept: "application/json"
        },
        body: JSON.stringify({
          class_id: CLASS_ID,
          slot_index: Number(form.slot_index),
          course_code: form.course_code,
          start_time: form.start_time,
          end_time: form.end_time,
          venue: form.venue,
          status: form.status
        })
      });

      const nextSlots = await refreshTimetable();
      setForm((prev) => ({ ...prev, slot_index: String(nextSlots.length + 1) }));
      setAdminMessage("Override saved.");
    } catch (err) {
      setAdminMessage(err.message || "Override failed");
    }
  }

  async function deleteSlot(slotIndex) {
    setAdminMessage("");

    try {
      await requestJSON(`/api/admin/slot?class_id=${encodeURIComponent(CLASS_ID)}&slot_index=${slotIndex}`, {
        method: "DELETE",
        headers: {
          Authorization: `Bearer ${token}`,
          Accept: "application/json"
        }
      });

      const nextSlots = await refreshTimetable();
      setForm((prev) => ({ ...prev, slot_index: String(nextSlots.length + 1) }));
      setAdminMessage(`Slot ${slotIndex} deleted and remaining slots reindexed.`);
    } catch (err) {
      setAdminMessage(err.message || "Delete failed");
    }
  }

  return (
    <main className="shell">
      <h1>cr45-reduced</h1>
      <p className="subtitle">Docker training app: one frontend, one Go backend.</p>

      <section className="card">
        <h2>HTTP Header Observations</h2>
        <p>The frontend expects JSON API responses and will fail loudly if response headers are wrong or browser security headers block requests.</p>
        {headerNotes ? (
          <dl className="facts">
            <div>
              <dt>Content-Type</dt>
              <dd>{headerNotes.contentType || "(missing)"}</dd>
            </div>
            <div>
              <dt>Cache-Control</dt>
              <dd>{headerNotes.cacheControl || "(missing)"}</dd>
            </div>
            <div>
              <dt>X-Content-Type-Options</dt>
              <dd>{headerNotes.xContentTypeOptions || "(missing)"}</dd>
            </div>
            <div>
              <dt>Request Content-Type</dt>
              <dd>application/json</dd>
            </div>
          </dl>
        ) : (
          <p className="error">No successful API response yet. Inspect browser network headers when requests fail.</p>
        )}
      </section>

      <section className="card">
        <h2>Today Timetable ({CLASS_ID})</h2>
        {loading && <p>Loading...</p>}
        {error && <p className="error">{error}</p>}
        {!loading && !error && (
          <table>
            <thead>
              <tr>
                <th>Slot</th>
                <th>Course</th>
                <th>Time</th>
                <th>Venue</th>
                <th>Status</th>
                {token && <th>Actions</th>}
              </tr>
            </thead>
            <tbody>
              {slots.map((slot) => (
                <tr key={slot.slot_index}>
                  <td>{slot.slot_index}</td>
                  <td>{slot.course_code}</td>
                  <td>{slot.start_time} - {slot.end_time}</td>
                  <td>{slot.venue}</td>
                  <td>{slot.status}</td>
                  {token && (
                    <td>
                      <button type="button" className="danger" onClick={() => deleteSlot(slot.slot_index)}>
                        Delete
                      </button>
                    </td>
                  )}
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </section>

      <section className="card">
        <h2>Admin Override</h2>
        {!token ? (
          <>
            <p>Login is required to add or update overrides. Without login, only the timetable is visible.</p>
            <form onSubmit={submitLogin} className="grid">
              <label>
                Username
                <input value={loginForm.username} onChange={(e) => setLoginForm({ ...loginForm, username: e.target.value })} />
              </label>
              <label>
                Password
                <input type="password" value={loginForm.password} onChange={(e) => setLoginForm({ ...loginForm, password: e.target.value })} />
              </label>
              <button type="submit">Login</button>
            </form>
            <p className={loginMessage.includes("Logged in") ? "ok" : loginMessage ? "error" : "muted"}>{loginMessage || "Use the seeded admin account created by the database migration."}</p>
          </>
        ) : (
          <>
            <div className="toolbar">
              <p className="ok">Authenticated. Override editing is enabled.</p>
              <button type="button" className="secondary" onClick={logout}>Logout</button>
            </div>
            <p className="muted">To append a new slot beyond the current timetable, use slot index {nextSlotIndex}.</p>
            <form onSubmit={submitOverride} className="grid">
              <label>
                Slot
                <input value={form.slot_index} onChange={(e) => setForm({ ...form, slot_index: e.target.value })} />
              </label>
              <label>
                Course
                <input value={form.course_code} onChange={(e) => setForm({ ...form, course_code: e.target.value })} />
              </label>
              <label>
                Start
                <input value={form.start_time} onChange={(e) => setForm({ ...form, start_time: e.target.value })} />
              </label>
              <label>
                End
                <input value={form.end_time} onChange={(e) => setForm({ ...form, end_time: e.target.value })} />
              </label>
              <label>
                Venue
                <input value={form.venue} onChange={(e) => setForm({ ...form, venue: e.target.value })} />
              </label>
              <label>
                Status
                <input value={form.status} onChange={(e) => setForm({ ...form, status: e.target.value })} />
              </label>
              <button type="button" className="secondary" onClick={() => setForm((prev) => ({ ...prev, slot_index: nextSlotIndex }))}>Use Next Slot</button>
              <button type="submit">Save Override</button>
            </form>
            {adminMessage && <p className={adminMessage.includes("saved") ? "ok" : "error"}>{adminMessage}</p>}
          </>
        )}
      </section>
    </main>
  );
}
