import React, { useEffect } from "react";
import PDFViewer from "./PDFViewer";
import usePDFController from "../hooks/usePDFController";
import TextViewer from "./TextViewer";
import MarkdownViewer from "./MarkdownViewer";

function renderContent(mimeType, content, file, pdfControls) {
  if (mimeType.startsWith("text/plain")) {
    return <TextViewer content={content} />;
  }
  if (mimeType.startsWith("text/markdown")) {
    return <MarkdownViewer content={content} />;
  }
  if (mimeType.startsWith("application/pdf")) {
    return <PDFViewer file={file} scale={pdfControls?.scale} />;
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
  const isPDF = file.mimeType.startsWith("application/pdf");
  const pdfControls = isPDF ? usePDFController(file) : null;

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
              <div className="d-flex align-items-center">
                {isPDF && pdfControls.renderUi()}
                <button type="button" className="btn-close" onClick={onClose}></button>
              </div>
            </div>
            <div className="modal-body">
              {renderContent(file.mimeType, file.content, file, pdfControls)}
            </div>
          </div>
        </div>
      </div>
      <div className="modal-backdrop fade show"></div>
    </>
  );
}
