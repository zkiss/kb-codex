import React, { useEffect, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import KBLayout from "./KBLayout";
import ChatView from "./ChatView";
import FileManagementView from "./FileManagementView";

export default function KBDetail({ onLogout }) {
  const { kbID, slug } = useParams();
  const navigate = useNavigate();
  const [kbName, setKbName] = useState("");
  const [files, setFiles] = useState([]);

  useEffect(() => {
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
    fetchMeta();
    fetchFiles();
  }, [kbID]);

  return (
    <KBLayout onLogout={onLogout}>
      <button className="btn btn-link mb-3" onClick={() => navigate("/kbs")}>
        <i className="bi bi-arrow-left"></i> Back to list
      </button>
      <div className="row">
        <div className="col-md-8 col-lg-9 order-2 order-md-1">
          <ChatView kbID={kbID} kbName={kbName} />
        </div>
        <div className="col-md-4 col-lg-3 order-1 order-md-2 mb-4">
          <FileManagementView kbID={kbID} slug={slug} files={files} refreshFiles={setFiles} />
        </div>
      </div>
    </KBLayout>
  );
}
