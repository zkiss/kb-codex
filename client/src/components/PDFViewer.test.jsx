import { render, screen } from '@testing-library/react';
import { vi } from 'vitest';
import React from 'react';
import PDFViewer from './PDFViewer';

// Mock react-pdf components
vi.mock('react-pdf', () => ({
  Document: ({ children, loading, error }) => (
    <div data-testid="pdf-document">
      {loading}
      {error}
      {children}
    </div>
  ),
  Page: ({ pageNumber, scale }) => (
    <div data-testid={`pdf-page-${pageNumber}`} data-scale={scale}>
      Page {pageNumber}
    </div>
  ),
  pdfjs: {
    GlobalWorkerOptions: {
      workerSrc: ''
    }
  }
}));

describe('PDFViewer', () => {
  it('renders without crashing', () => {
    const file = {
      name: 'test.pdf',
      content: btoa('fake-pdf-content-for-testing'),
      mimeType: 'application/pdf'
    };
    render(<PDFViewer file={file} />);
    expect(screen.getByTestId('pdf-document')).toBeInTheDocument();
  });
}); 