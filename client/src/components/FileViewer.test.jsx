import { render, screen } from '@testing-library/react';
import { vi } from 'vitest';
import React from 'react';
import FileViewer from './FileViewer';

// Mock the viewer components
vi.mock('./PDFViewer', () => ({
  default: function MockPDFViewer({ file }) {
    return <div data-testid="pdf-viewer">PDF Viewer for {file.name}</div>;
  }
}));

vi.mock('./TextViewer', () => ({
  default: function MockTextViewer({ content }) {
    return <div data-testid="text-viewer">{content}</div>;
  }
}));

vi.mock('./MarkdownViewer', () => ({
  default: function MockMarkdownViewer({ content }) {
    return <div data-testid="markdown-viewer">{content}</div>;
  }
}));

describe('FileViewer', () => {
  it('renders text content viewer', () => {
    render(<FileViewer file={{ name: 't.txt', content: 'hello', mimeType: 'text/plain' }} onClose={() => {}} />);
    expect(screen.getByTestId('text-viewer')).toBeInTheDocument();
    expect(screen.getByText('hello')).toBeInTheDocument();
  });

  it('renders markdown content viewer', () => {
    render(<FileViewer file={{ name: 'm.md', content: '# Title', mimeType: 'text/markdown' }} onClose={() => {}} />);
    expect(screen.getByTestId('markdown-viewer')).toBeInTheDocument();
    expect(screen.getByText('# Title')).toBeInTheDocument();
  });

  it('renders pdf content viewer', () => {
    render(<FileViewer file={{ name: 'doc.pdf', content: 'pdfcontent', mimeType: 'application/pdf' }} onClose={() => {}} />);
    expect(screen.getByTestId('pdf-viewer')).toBeInTheDocument();
    expect(screen.getByText('PDF Viewer for doc.pdf')).toBeInTheDocument();
  });

  it('shows unsupported file type message', () => {
    render(<FileViewer file={{ name: 'file.xyz', content: 'content', mimeType: 'application/unknown' }} onClose={() => {}} />);
    expect(screen.getByText('Unsupported file type')).toBeInTheDocument();
  });
});
