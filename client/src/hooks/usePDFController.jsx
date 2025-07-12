import { useEffect, useState } from "react";

export default function usePDFController(file) {
  const [scale, setScale] = useState(1.2);
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

  const handleZoomIn = () => setScale((s) => Math.min(3, s + 0.2));
  const handleZoomOut = () => setScale((s) => Math.max(0.5, s - 0.2));

  function renderUi() {
    return (
      <div className="d-flex align-items-center">
        <button
          type="button"
          className="btn btn-sm btn-outline-secondary me-2"
          onClick={handleZoomOut}
          title="Zoom Out"
          data-testid="zoom-out"
        >
          <i className="bi bi-dash fs-4"></i>
        </button>
        <button
          type="button"
          className="btn btn-sm btn-outline-secondary me-2"
          onClick={handleZoomIn}
          title="Zoom In"
          data-testid="zoom-in"
        >
          <i className="bi bi-plus fs-4"></i>
        </button>
        <a
          href={downloadUrl}
          rel="noopener noreferrer"
          className="btn btn-sm btn-outline-primary me-2"
          title="Download"
          data-testid="download-btn"
          download={file.name}
        >
          <i className="bi bi-download fs-4"></i>
        </a>
      </div>
    );
  }

  return { scale, renderUi };
} 