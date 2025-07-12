import React, { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import FileViewer from "./FileViewer";

export default function FileManagementView({ kbID, slug, files, refreshFiles }) {
  const navigate = useNavigate();
  const [selectedFile, setSelectedFile] = useState(null);
  const [viewFile, setViewFile] = useState(null);

  const fetchFiles = async () => {
    const res = await fetch(`/api/kbs/${kbID}/files`);
    const list = await res.json();
    refreshFiles(list);
  };


  const fetchFile = async (file) => {
    const res = await fetch(
      `/api/kbs/${kbID}/files/${encodeURIComponent(file.slug)}`,
    );
    if (res.ok) {
      const mimeType = res.headers.get("Content-Type") || "";
      let content;
      if (mimeType.startsWith("application/pdf")) {
        content = await res.arrayBuffer();
      } else {
        content = await res.text();
      }
      setViewFile({ name: file.name, content, mimeType });
    }
  };

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

  const uploadFile = async () => {
    if (!selectedFile) return;
    const form = new FormData();
    form.append("file", selectedFile);
    await fetch(`/api/kbs/${kbID}/files`, { method: "POST", body: form });
    setSelectedFile(null);
    fetchFiles();
  };

  const openFile = (file) => {
    if (!file) return;
    if (slug !== file.slug) {
      navigate(`/kbs/${kbID}/files/${encodeURIComponent(file.slug)}`);
    } else {
      fetchFile(file);
    }
  };

  return (
    <>
      {viewFile && (
        <FileViewer
          file={viewFile}
          onClose={() => {
            setViewFile(null);
            navigate(`/kbs/${kbID}`);
          }}
        />
      )}
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
          data-testid="file-input"
          type="file"
          className="form-control"
          accept=".txt,.md,.pdf"
          onChange={(e) => setSelectedFile(e.target.files[0])}
        />
        <button data-testid="upload-btn" className="btn btn-primary" onClick={uploadFile}>
          Upload
        </button>
      </div>
    </>
  );
}
