import React, { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import KBLayout from "./KBLayout";

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

export default KBList; 