import React, { useEffect } from "react";
import { marked } from "marked";

function renderContent(mimeType, content) {
  if (mimeType.startsWith("text/plain")) {
    return <pre className="mt-2">{content}</pre>;
  }
  if (mimeType.startsWith("text/markdown")) {
    return (
      <div
        className="mt-2"
        dangerouslySetInnerHTML={{ __html: marked.parse(content) }}
      ></div>
    );
  }
  return <p className="mt-2">Unsupported file type</p>;
}

export default function FileViewer({ file, onClose }) {
  useEffect(() => {
    const handler = (e) => {
      if (e.key === "Escape") onClose();
    };
    document.addEventListener("keydown", handler);
    return () => document.removeEventListener("keydown", handler);
  }, [onClose]);

  if (!file) return null;
  return (
    <>
      <div
        className="modal fade show"
        style={{ display: "block" }}
        tabIndex="-1"
        role="dialog"
      >
        <div className="modal-dialog modal-xl modal-dialog-scrollable" role="document">
          <div className="modal-content">
            <div className="modal-header">
              <h5 className="modal-title">{file.name}</h5>
              <button type="button" className="btn-close" onClick={onClose}></button>
            </div>
            <div className="modal-body">
              {renderContent(file.mimeType, file.content)}
            </div>
          </div>
        </div>
      </div>
      <div className="modal-backdrop fade show"></div>
    </>
  );
}
