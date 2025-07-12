import React, { useEffect, useState } from "react";
import { Document, Page, pdfjs } from "react-pdf";
pdfjs.GlobalWorkerOptions.workerSrc = "/static/pdf.worker.min.js";

export default function PDFViewer({ file }) {
  const [numPages, setNumPages] = useState(null);
  const [scale, setScale] = useState(1.2);
  const [pdfData, setPdfData] = useState(null);
  const [downloadUrl, setDownloadUrl] = useState(null);

  useEffect(() => {
    let blobUrl = null;
    if (file && file.content) {
      let arrayBuffer;
      if (file.content instanceof ArrayBuffer) {
        arrayBuffer = file.content;
      } else if (typeof file.content === "string") {
        // If base64 string
        const byteCharacters = atob(file.content);
        const byteNumbers = new Array(byteCharacters.length);
        for (let i = 0; i < byteCharacters.length; i++) {
          byteNumbers[i] = byteCharacters.charCodeAt(i);
        }
        arrayBuffer = new Uint8Array(byteNumbers);
      }
      setPdfData(arrayBuffer);
      // Create blob URL for download
      const blob = new Blob([
        arrayBuffer instanceof ArrayBuffer ? arrayBuffer : new Uint8Array(arrayBuffer)
      ], { type: file.mimeType });
      blobUrl = URL.createObjectURL(blob);
      setDownloadUrl(blobUrl);
    }
    return () => {
      if (blobUrl) {
        URL.revokeObjectURL(blobUrl);
      }
    };
  }, [file]);

  function onDocumentLoadSuccess({ numPages }) {
    setNumPages(numPages);
  }

  return (
    <div className="pdf-viewer">
      <div style={{ marginBottom: 8 }}>
        <button style={{ marginLeft: 16 }} onClick={() => setScale((s) => Math.max(0.5, s - 0.2))}>-</button>
        <span style={{ margin: '0 8px' }}>Zoom</span>
        <button onClick={() => setScale((s) => Math.min(3, s + 0.2))}>+</button>
      </div>
      <div style={{ border: '1px solid #eee', display: 'flex', flexDirection: 'column', alignItems: 'center' }}>
        <Document
          file={pdfData}
          onLoadSuccess={onDocumentLoadSuccess}
          loading={<div>Loading PDF...</div>}
          error={<div>Failed to load PDF.</div>}
        >
          {numPages && Array.from(new Array(numPages), (el, index) => (
            <Page key={`page_${index + 1}`} pageNumber={index + 1} scale={scale} />
          ))}
        </Document>
        <div style={{ margin: '16px 0' }}>
          <a href={downloadUrl} download={file.name} rel="noopener noreferrer" className="btn btn-primary">Download PDF</a>
        </div>
      </div>
    </div>
  );
} 