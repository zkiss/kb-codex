import { render, screen } from '@testing-library/react';
import React from 'react';
import TextViewer from './TextViewer';

describe('TextViewer', () => {
  it('renders text content', () => {
    const content = 'Hello, this is a test file content';
    render(<TextViewer content={content} />);
    expect(screen.getByText(content)).toBeInTheDocument();
  });

  it('renders empty content', () => {
    render(<TextViewer content="" />);
    // Should render a <pre> element even if content is empty
    const preElements = screen.getAllByRole('generic');
    expect(preElements.some(el => el.tagName === 'PRE')).toBe(true);
  });
}); 