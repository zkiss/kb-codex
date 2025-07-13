import React, { useEffect, useRef, useState, useCallback } from "react";
import { Document, Page, pdfjs } from "react-pdf";
pdfjs.GlobalWorkerOptions.workerSrc = "/static/pdf.worker.min.js";

const PAGE_WINDOW = 2; // current ±2

export default function PDFViewer({ file, scale: scaleProp }) {
  const scale = scaleProp;
  const [numPages, setNumPages] = useState(null);
  const [pdfData, setPdfData] = useState(null);
  const [visiblePage, setVisiblePage] = useState(1);
  const [pageHeights, setPageHeights] = useState({});
  const containerRef = useRef();
  const pageRefs = useRef([]);

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

  // Intersection Observer to track visible page
  useEffect(() => {
    if (!numPages) return;
    const observer = new window.IntersectionObserver(
      (entries) => {
        // Find the entry most in view
        let maxRatio = 0;
        let mostVisible = visiblePage;
        entries.forEach((entry) => {
          if (entry.intersectionRatio > maxRatio) {
            maxRatio = entry.intersectionRatio;
            mostVisible = Number(entry.target.dataset.pagenum);
          }
        });
        setVisiblePage(mostVisible);
      },
      {
        root: containerRef.current,
        threshold: [0.1, 0.5, 0.9],
      }
    );
    pageRefs.current.forEach((ref) => {
      if (ref) observer.observe(ref);
    });
    return () => observer.disconnect();
  }, [numPages]);

  // Callback to cache page heights
  const onPageRenderSuccess = useCallback(
    (page, pageNumber) => {
      const height = page.height * scale;
      setPageHeights((prev) => {
        if (prev[pageNumber] === height) return prev;
        return { ...prev, [pageNumber]: height };
      });
    },
    [scale]
  );

  // Helper to get placeholder height
  const getPageHeight = (pageNumber) => {
    // Use cached height, or fallback to first page's height, or 800px
    return (
      pageHeights[pageNumber] || pageHeights[1] || 800
    );
  };

  return (
    <div className="pdf-viewer" style={{ width: "100%" }}>
      <div
        ref={containerRef}
        style={{
          border: "1px solid #eee",
          display: "flex",
          flexDirection: "column",
          alignItems: "center",
          height: "80vh",
          overflowY: "auto",
          width: "100%",
        }}
      >
        <Document
          file={pdfData}
          onLoadSuccess={onDocumentLoadSuccess}
          loading={<div>Loading PDF...</div>}
          error={<div>Failed to load PDF.</div>}
        >
          {numPages &&
            Array.from(new Array(numPages), (el, index) => {
              const pageNumber = index + 1;
              const inWindow =
                Math.abs(pageNumber - visiblePage) <= PAGE_WINDOW;
              return (
                <div
                  key={`pagewrap_${pageNumber}`}
                  ref={(el) => (pageRefs.current[index] = el)}
                  data-pagenum={pageNumber}
                  style={{
                    width: "100%",
                    display: "flex",
                    justifyContent: "center",
                    margin: "0 auto 16px auto",
                  }}
                >
                  {inWindow ? (
                    <Page
                      pageNumber={pageNumber}
                      scale={scale}
                      onRenderSuccess={(page) =>
                        onPageRenderSuccess(page, pageNumber)
                      }
                    />
                  ) : (
                    <div
                      style={{
                        background: "#fff",
                        width: "90%",
                        height: getPageHeight(pageNumber),
                        boxShadow: "0 0 2px #ccc",
                      }}
                    />
                  )}
                </div>
              );
            })}
        </Document>
      </div>
    </div>
  );
} 