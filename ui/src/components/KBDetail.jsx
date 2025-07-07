import React, { useState, useEffect } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { marked } from "marked";
import KBLayout from "./KBLayout";
import FileViewer from "./FileViewer";

function KBDetail({ onLogout }) {
  const { kbID, slug } = useParams();
  const navigate = useNavigate();
  const [kbName, setKbName] = useState("");
  const [files, setFiles] = useState([]);
  const [selectedFile, setSelectedFile] = useState(null);
  const [viewFile, setViewFile] = useState(null);
  const [question, setQuestion] = useState("");
  const [messages, setMessages] = useState([]);
  const [loading, setLoading] = useState(false);
  const inputRef = React.useRef(null);
  const [expanded, setExpanded] = useState({});

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
    if (!slug) {
      setViewFile(null);
      return;
    }
    if (files.length > 0) {
      const f = files.find((fi) => fi.slug === slug) || { slug, name: slug };
      fetchFile(f);
    }
  }, [slug, files]);

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

  const fetchFile = async (file) => {
    const res = await fetch(
      `/api/kbs/${kbID}/files/${encodeURIComponent(file.slug)}`,
    );
    if (res.ok) {
      const mimeType = res.headers.get("Content-Type") || "";
      const content = await res.text();
      setViewFile({ name: file.name, content, mimeType });
    }
  };

  const openFile = (file) => {
    if (!file) return;
    if (slug !== file.slug) {
      navigate(`/kbs/${kbID}/files/${encodeURIComponent(file.slug)}`);
    } else {
      fetchFile(file);
    }
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
      {viewFile && (
        <FileViewer
          file={viewFile}
          onClose={() => {
            setViewFile(null);
            navigate(`/kbs/${kbID}`);
          }}
        />
      )}
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
                            onClick={() =>
                              setExpanded((e) => ({ ...e, [i]: !e[i] }))
                            }
                          >
                            <i className="bi bi-info-circle"></i>{" "}
                            {expanded[i] ? "Hide context" : "Show context"}
                          </button>
                          {expanded[i] && (
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
                          )}
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
              <li
                key={file.slug}
                className="list-group-item list-group-item-action"
                onClick={() => openFile(file)}
                role="button"
              >
                {file.name}
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

export default KBDetail;
