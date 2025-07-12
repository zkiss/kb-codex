import React, { useEffect, useRef, useState } from "react";
import { marked } from "marked";

export default function ChatView({ kbID, kbName }) {
  const [question, setQuestion] = useState("");
  const [messages, setMessages] = useState([]);
  const [loading, setLoading] = useState(false);
  const [expanded, setExpanded] = useState({});
  const inputRef = useRef(null);

  useEffect(() => {
    inputRef.current && inputRef.current.focus();
  }, []);
  useEffect(() => {
    if (!loading) {
      inputRef.current && inputRef.current.focus();
    }
  }, [loading]);

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
  };

  return (
    <>
      <h2 className="mb-3">Chat with {kbName}</h2>
      <div className="mb-3" style={{ height: "60vh", overflowY: "auto" }}>
        {messages.map((m, i) => (
          <div
            key={i}
            className={
              "d-flex mb-2 " +
              (m.role === "user" ? "justify-content-end" : "justify-content-start")
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
                        onClick={() => setExpanded((e) => ({ ...e, [i]: !e[i] }))}
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
          aria-label="Your question"
          ref={inputRef}
          value={question}
          onChange={(e) => setQuestion(e.target.value)}
          type="text"
          className="form-control"
          placeholder="Type your question"
          onKeyUp={(e) => (e.key === "Enter" ? askQuestion() : null)}
          disabled={loading}
        />
        <button className="btn btn-primary" onClick={askQuestion} disabled={loading}>
          {loading ? (
            <span className="spinner-border spinner-border-sm"></span>
          ) : (
            "Ask"
          )}
        </button>
      </div>
    </>
  );
}
