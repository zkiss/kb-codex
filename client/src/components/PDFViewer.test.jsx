import { render, screen } from '@testing-library/react';
import { vi } from 'vitest';
import React from 'react';
import PDFViewer from './PDFViewer';

// Mock browser APIs not present in jsdom
global.URL.createObjectURL = vi.fn(() => 'blob:mock-url');
global.URL.revokeObjectURL = vi.fn();

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

  it('shows zoom controls', () => {
    const file = {
      name: 'test.pdf',
      content: btoa('fake-pdf-content-for-testing'),
      mimeType: 'application/pdf'
    };
    render(<PDFViewer file={file} />);
    expect(screen.getByText('Zoom')).toBeInTheDocument();
    expect(screen.getByText('-')).toBeInTheDocument();
    expect(screen.getByText('+')).toBeInTheDocument();
  });

  it('has download button', () => {
    const file = {
      name: 'test.pdf',
      content: btoa('fake-pdf-content-for-testing'),
      mimeType: 'application/pdf'
    };
    render(<PDFViewer file={file} />);
    const downloadLink = screen.getByRole('link', { name: /download pdf/i });
    expect(downloadLink).toBeInTheDocument();
    expect(downloadLink).toHaveAttribute('href', 'blob:mock-url');
    expect(downloadLink).toHaveAttribute('download', 'test.pdf');
  });
}); 