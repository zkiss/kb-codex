import React from "react";

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

export default KBLayout; 