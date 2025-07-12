import { render, screen } from '@testing-library/react';
import React from 'react';
import FileViewer from '../components/FileViewer';

describe('FileViewer', () => {
  it('renders text content', () => {
    render(<FileViewer file={{ name: 't.txt', content: 'hello', mimeType: 'text/plain' }} onClose={() => {}} />);
    expect(screen.getByText('hello')).toBeInTheDocument();
  });

  it('renders markdown content', () => {
    render(<FileViewer file={{ name: 'm.md', content: '# Title', mimeType: 'text/markdown' }} onClose={() => {}} />);
    expect(screen.getByRole('heading', { name: 'Title' })).toBeInTheDocument();
  });
});
