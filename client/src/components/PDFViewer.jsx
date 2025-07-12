import React, { useEffect, useState } from "react";
import { Document, Page, pdfjs } from "react-pdf";
pdfjs.GlobalWorkerOptions.workerSrc = "/static/pdf.worker.min.js";

export default function PDFViewer({ file, scale: scaleProp }) {
  const scale = scaleProp;
  const [numPages, setNumPages] = useState(null);
  const [pdfData, setPdfData] = useState(null);

  useEffect(() => {
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
    }
  }, [file]);

  function onDocumentLoadSuccess({ numPages }) {
    setNumPages(numPages);
  }

  return (
    <div className="pdf-viewer">
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
      </div>
    </div>
  );
} 