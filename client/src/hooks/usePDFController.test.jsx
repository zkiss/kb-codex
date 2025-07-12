import { render, screen, fireEvent } from '@testing-library/react';
import React from 'react';
import usePDFController from './usePDFController';

describe('usePDFController', () => {
  const file = {
    name: 'test.pdf',
    content: btoa('fake-pdf-content-for-testing'),
    mimeType: 'application/pdf'
  };

  function TestComponent({ file }) {
    const pdf = usePDFController(file);
    return (
      <>
        <div data-testid="scale">{pdf.scale}</div>
        {pdf.renderUi()}
      </>
    );
  }

  it('initializes with default scale', () => {
    render(<TestComponent file={file} />);
    expect(screen.getByTestId('scale').textContent).toBe('1.2');
  });

  it('zoom in increases scale', () => {
    render(<TestComponent file={file} />);
    fireEvent.click(screen.getByTestId('zoom-in'));
    expect(screen.getByTestId('scale').textContent).toBe('1.4');
  });

  it('zoom out decreases scale', () => {
    render(<TestComponent file={file} />);
    fireEvent.click(screen.getByTestId('zoom-out'));
    expect(screen.getByTestId('scale').textContent).toBe('1');
  });

  it('download button has correct attributes', () => {
    render(<TestComponent file={file} />);
    const downloadBtn = screen.getByTestId('download-btn');
    expect(downloadBtn).toHaveAttribute('href', 'blob:mock-url');
    expect(downloadBtn).toHaveAttribute('download', 'test.pdf');
  });

  it('renders all controls', () => {
    render(<TestComponent file={file} />);
    expect(screen.getByTestId('zoom-in')).toBeInTheDocument();
    expect(screen.getByTestId('zoom-out')).toBeInTheDocument();
    expect(screen.getByTestId('download-btn')).toBeInTheDocument();
  });
}); 