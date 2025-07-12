import { render, screen } from '@testing-library/react';
import React from 'react';
import MarkdownViewer from './MarkdownViewer';

describe('MarkdownViewer', () => {
  it('renders markdown content as HTML', () => {
    const content = '# Test Title\n\nThis is a paragraph.';
    render(<MarkdownViewer content={content} />);
    expect(screen.getByRole('heading', { name: 'Test Title' })).toBeInTheDocument();
    expect(screen.getByText('This is a paragraph.')).toBeInTheDocument();
  });

  it('renders empty content', () => {
    render(<MarkdownViewer content="" />);
    // Should render without crashing
    expect(document.querySelector('.mt-2')).toBeInTheDocument();
  });
}); 