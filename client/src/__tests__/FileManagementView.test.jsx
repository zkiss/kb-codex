import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import React from 'react';
import { vi } from 'vitest';
import FileManagementView from '../components/FileManagementView';

let slug;
let files;
let rerender;
const setFiles = vi.fn((f) => {
  files = f;
});
const mockNavigate = vi.fn((path) => {
  slug = decodeURIComponent(path.split('/').pop());
  rerender(<FileManagementView kbID="1" slug={slug} files={files} refreshFiles={setFiles} />);
});

vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom');
  return { ...actual, useNavigate: () => mockNavigate };
});

describe('FileManagementView', () => {
  beforeEach(() => {
    slug = undefined;
    files = [];
    rerender = undefined;
    mockNavigate.mockClear();
    setFiles.mockClear();
  });

  it('uploads file', async () => {
    const utils = render(<FileManagementView kbID="1" slug={slug} files={files} refreshFiles={setFiles} />);
    rerender = utils.rerender;

    const file = new File(['hello'], 't.txt', { type: 'text/plain' });
    fireEvent.change(screen.getByTestId('file-input'), { target: { files: [file] } });

    fetch.mockResolvedValueOnce({ ok: true });
    fetch.mockResolvedValueOnce({ ok: true, json: () => Promise.resolve([]) });

    fireEvent.click(screen.getByTestId('upload-btn'));

    await waitFor(() => {
      const call = fetch.mock.calls.find(c => c[0] === '/api/kbs/1/files' && c[1]?.method === 'POST');
      expect(call).toBeDefined();
      expect(call[1].body).toBeInstanceOf(FormData);
      expect(call[1].body.get('file')).toBe(file);
    });
  });

  it('opens file viewer when clicking file', async () => {
    files = [{ slug: 't.txt', name: 't.txt' }];
    const utils = render(<FileManagementView kbID="1" slug={slug} files={files} refreshFiles={setFiles} />);
    rerender = utils.rerender;

    fetch.mockResolvedValueOnce({
      ok: true,
      headers: { get: () => 'text/plain' },
      text: () => Promise.resolve('content'),
    });

    fireEvent.click(screen.getByText('t.txt'));

    await waitFor(() => expect(screen.getByText('content')).toBeInTheDocument());
  });
});
