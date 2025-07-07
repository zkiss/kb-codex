import React, { useState, useEffect } from "react";
import { BrowserRouter, Routes, Route, Navigate, Link, useNavigate, useParams } from "react-router-dom";
import { marked } from "marked";

function Login({ onLogin }) {
  const navigate = useNavigate();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError("");
    try {
      const resp = await fetch("/api/login", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ email, password }),
      });
      const data = await resp.json().catch(() => ({}));
      if (!resp.ok) {
        setError(data.error || resp.statusText);
        return;
      }
      onLogin(data.token);
      navigate("/kbs");
    } catch (err) {
      setError(err.message);
    }
  };

  return (
    <div className="row justify-content-center align-items-center vh-100">
      <div className="col-md-5 col-lg-4">
        <div className="card shadow rounded-3">
          <div className="card-body">
            <h2 className="card-title text-center mb-4">KB Codex</h2>
            <form onSubmit={handleSubmit}>
              <div className="mb-3">
                <label htmlFor="email" className="form-label">
                  Email address
                </label>
                <input
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  type="email"
                  className="form-control"
                  id="email"
                  placeholder="Enter email"
                  required
                />
              </div>
              <div className="mb-3">
                <label htmlFor="password" className="form-label">
                  Password
                </label>
                <input
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  type="password"
                  className="form-control"
                  id="password"
                  placeholder="Password"
                  required
                />
              </div>
              <button type="submit" className="btn btn-primary w-100">
                Login
              </button>
            </form>
            <p className="text-center mt-3">
              <Link to="/register">Don't have an account? Register</Link>
            </p>
            {error && (
              <div className="alert alert-danger mt-3" role="alert">
                {error}
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}

function Register() {
  const navigate = useNavigate();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError("");
    try {
      const resp = await fetch("/api/register", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ email, password }),
      });
      if (!resp.ok) {
        const data = await resp.json().catch(() => ({}));
        setError(data.error || resp.statusText);
        return;
      }
      alert("Registration successful, please login");
      navigate("/login");
    } catch (err) {
      setError(err.message);
    }
  };

  return (
    <div className="row justify-content-center align-items-center vh-100">
      <div className="col-md-5 col-lg-4">
        <div className="card shadow rounded-3">
          <div className="card-body">
            <h2 className="card-title text-center mb-4">Register</h2>
            <form onSubmit={handleSubmit}>
              <div className="mb-3">
                <label htmlFor="emailR" className="form-label">
                  Email address
                </label>
                <input
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  type="email"
                  className="form-control"
                  id="emailR"
                  placeholder="Enter email"
                  required
                />
              </div>
              <div className="mb-3">
                <label htmlFor="passwordR" className="form-label">
                  Password
                </label>
                <input
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  type="password"
                  className="form-control"
                  id="passwordR"
                  placeholder="Password"
                  required
                />
              </div>
              <button type="submit" className="btn btn-success w-100">
                Register
              </button>
            </form>
            <p className="text-center mt-3">
              <Link to="/login">Already have an account? Login</Link>
            </p>
            {error && (
              <div className="alert alert-danger mt-3" role="alert">
                {error}
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}

function KBLayout({ children, onLogout }) {
  return (
    <div className="container my-3">
      <nav className="navbar navbar-dark bg-primary mb-4 rounded">
        <div className="container-fluid">
          <span className="navbar-brand mb-0 h1">KB Codex</span>
          <button
            className="btn btn-sm btn-outline-light"
            onClick={onLogout}
          >
            Logout
          </button>
        </div>
      </nav>
      {children}
    </div>
  );
}

function KBList({ onLogout }) {
  const navigate = useNavigate();
  const [kbs, setKbs] = useState([]);
  const [newKBName, setNewKBName] = useState("");

  const fetchKBs = async () => {
    const res = await fetch("/api/kbs");
    setKbs(await res.json());
  };

  useEffect(() => {
    fetchKBs();
  }, []);

  const createKB = async () => {
    await fetch("/api/kbs", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ name: newKBName }),
    });
    setNewKBName("");
    fetchKBs();
  };

  return (
    <KBLayout onLogout={onLogout}>
      <h1 className="mb-4">Knowledge Bases</h1>
      <ul className="list-group mb-4">
        {kbs.map((kb) => (
          <li
            key={kb.id}
            className="list-group-item list-group-item-action d-flex justify-content-between align-items-center"
            onClick={() => navigate(`/kbs/${kb.id}`)}
            tabIndex="0"
          >
            {kb.name}
            <i className="bi bi-chevron-right"></i>
          </li>
        ))}
      </ul>
      <form
        onSubmit={(e) => {
          e.preventDefault();
          createKB();
        }}
        className="row g-2 mb-4"
      >
        <div className="col-auto flex-grow-1">
          <input
            value={newKBName}
            onChange={(e) => setNewKBName(e.target.value)}
            type="text"
            className="form-control"
            placeholder="New KB name"
            required
          />
        </div>
        <div className="col-auto">
          <button type="submit" className="btn btn-success">
            Create KB
          </button>
        </div>
      </form>
    </KBLayout>
  );
}

function KBDetail({ onLogout }) {
  const { kbID } = useParams();
  const navigate = useNavigate();
  const [kbName, setKbName] = useState("");
  const [files, setFiles] = useState([]);
  const [selectedFile, setSelectedFile] = useState(null);
  const [question, setQuestion] = useState("");
  const [messages, setMessages] = useState([]);
  const [loading, setLoading] = useState(false);
  const inputRef = React.useRef(null);

  const fetchMeta = async () => {
    const res = await fetch("/api/kbs");
    const list = await res.json();
    const kb = list.find((k) => String(k.id) === kbID);
    setKbName(kb ? kb.name : "");
  };

  const fetchFiles = async () => {
    const res = await fetch(`/api/kbs/${kbID}/files`);
    setFiles(await res.json());
  };

  useEffect(() => {
    fetchMeta();
    fetchFiles();
  }, [kbID]);

  useEffect(() => {
    inputRef.current && inputRef.current.focus();
  }, []);

  useEffect(() => {
    if (!loading) {
      inputRef.current && inputRef.current.focus();
    }
  }, [loading]);

  const uploadFile = async () => {
    if (!selectedFile) return;
    const form = new FormData();
    form.append("file", selectedFile);
    await fetch(`/api/kbs/${kbID}/files`, { method: "POST", body: form });
    setSelectedFile(null);
    fetchFiles();
  };

  const askQuestion = async () => {
    if (!question || loading) return;
    setLoading(true);
    const history = messages.map((m) => ({ role: m.role, content: m.content }));
    const resp = await fetch(`/api/kbs/${kbID}/ask`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ question, history }),
    });
    const data = await resp.json();
    const ans = data.answer || "No answer";
    setMessages([
      ...messages,
      { role: "user", content: question },
      { role: "assistant", content: ans, context: data.chunks || [] },
    ]);
    setQuestion("");
    setLoading(false);
    inputRef.current && inputRef.current.focus();
  };

  return (
    <KBLayout onLogout={onLogout}>
      <button className="btn btn-link mb-3" onClick={() => navigate("/kbs")}> 
        <i className="bi bi-arrow-left"></i> Back to list
      </button>
      <div className="row">
        <div className="col-md-8 col-lg-9 order-2 order-md-1">
          <h2 className="mb-3">Chat with {kbName}</h2>
          <div className="mb-3" style={{ height: "60vh", overflowY: "auto" }}>
            {messages.map((m, i) => (
              <div
                key={i}
                className={
              "d-flex mb-2 " +
              (m.role === "user"
                ? "justify-content-end"
                : "justify-content-start")
            }
          >
            <div
              className={
                "p-2 rounded-3 " +
                (m.role === "user" ? "bg-primary text-white" : "bg-light")
              }
              style={{ maxWidth: "80%" }}
            >
              {m.role === "assistant" ? (
                <div>
                  <span
                    dangerouslySetInnerHTML={{
                      __html: marked.parse(m.content),
                    }}
                  ></span>
                  {m.context && m.context.length > 0 && (
                    <>
                      <button
                        className="btn btn-sm btn-link mt-2"
                        type="button"
                        data-bs-toggle="collapse"
                        data-bs-target={`#ctx-${i}`}
                        aria-expanded="false"
                      >
                        <i className="bi bi-info-circle"></i> Show context
                      </button>
                      <div className="collapse" id={`ctx-${i}`}>
                        <ul className="list-group mt-2">
                          {m.context.map((c) => (
                            <li
                              key={c.file_name + "-" + c.index}
                              className="list-group-item"
                            >
                              <strong>
                                {c.file_name} [{c.index}]
                              </strong>
                              <pre className="mt-2 mb-0">{c.content}</pre>
                            </li>
                          ))}
                        </ul>
                      </div>
                    </>
                  )}
                </div>
              ) : (
                m.content
              )}
            </div>
          </div>
        ))}
          </div>
          <div className="input-group mb-3">
            <input
              ref={inputRef}
              value={question}
              onChange={(e) => setQuestion(e.target.value)}
              type="text"
              className="form-control"
              placeholder="Type your question"
              onKeyUp={(e) => (e.key === "Enter" ? askQuestion() : null)}
              disabled={loading}
            />
            <button
              className="btn btn-primary"
              onClick={askQuestion}
              disabled={loading}
            >
              {loading ? (
                <span className="spinner-border spinner-border-sm"></span>
              ) : (
                "Ask"
              )}
            </button>
          </div>
        </div>
        <div className="col-md-4 col-lg-3 order-1 order-md-2 mb-4">
          <h4>Files</h4>
          <ul className="list-group mb-3">
            {(files || []).map((file) => (
              <li key={file} className="list-group-item">
                {file}
              </li>
            ))}
          </ul>
          <div className="input-group">
            <input
              type="file"
              className="form-control"
              accept=".txt,.md"
              onChange={(e) => setSelectedFile(e.target.files[0])}
            />
            <button className="btn btn-primary" onClick={uploadFile}>
              Upload
            </button>
          </div>
        </div>
      </div>
    </KBLayout>
  );
}

function App() {
  const [token, setToken] = useState(localStorage.getItem("kb_jwt"));
  const handleLogin = (t) => {
    localStorage.setItem("kb_jwt", t);
    setToken(t);
  };
  const handleLogout = () => {
    localStorage.removeItem("kb_jwt");
    setToken(null);
  };

  return (
    <BrowserRouter>
      <Routes>
        <Route path="/login" element={<Login onLogin={handleLogin} />} />
        <Route path="/register" element={<Register />} />
        <Route
          path="/kbs"
          element={
            token ? (
              <KBList onLogout={handleLogout} />
            ) : (
              <Navigate to="/login" replace />
            )
          }
        />
        <Route
          path="/kbs/:kbID"
          element={
            token ? (
              <KBDetail onLogout={handleLogout} />
            ) : (
              <Navigate to="/login" replace />
            )
          }
        />
        <Route
          path="*"
          element={<Navigate to={token ? "/kbs" : "/login"} replace />}
        />
      </Routes>
    </BrowserRouter>
  );
}


export default App;
