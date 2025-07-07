import React, { useState } from "react";
import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom";
import Login from "./components/Login";
import Register from "./components/Register";
import KBList from "./components/KBList";
import KBDetail from "./components/KBDetail";

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
