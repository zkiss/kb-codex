import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import React from 'react';
import KBDetail from '../components/KBDetail';
import { vi } from 'vitest';

let params = { kbID: '1', slug: undefined };
let rerender;
const mockNavigate = vi.fn((path) => {
  params.slug = decodeURIComponent(path.split('/').pop());
  rerender && rerender(<KBDetail onLogout={() => {}} />);
});

vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom');
  return { ...actual, useNavigate: () => mockNavigate, useParams: () => params };
});

describe('KBDetail', () => {
  beforeEach(() => {
    params = { kbID: '1', slug: undefined };
    rerender = undefined;
  });

  it('uploads file', async () => {
    fetch
      .mockResolvedValueOnce({ ok: true, json: () => Promise.resolve([{ id: 1, name: 'KB1' }]) })
      .mockResolvedValueOnce({ ok: true, json: () => Promise.resolve([]) });

    const utils = render(<KBDetail onLogout={() => {}} />);
    rerender = utils.rerender;

    await waitFor(() => expect(fetch).toHaveBeenCalledTimes(2));

    const file = new File(['hello'], 'test.txt', { type: 'text/plain' });
    fireEvent.change(screen.getByTestId('file-input'), { target: { files: [file] } });

    fetch.mockResolvedValueOnce({ ok: true });
    fetch.mockResolvedValueOnce({ ok: true, json: () => Promise.resolve([]) });

    fireEvent.click(screen.getByTestId('upload-btn'));

    await waitFor(() => expect(fetch).toHaveBeenCalledWith('/api/kbs/1/files', expect.objectContaining({ method: 'POST' })));
  });

  it('allows chatting', async () => {
    fetch
      .mockResolvedValueOnce({ ok: true, json: () => Promise.resolve([{ id: 1, name: 'KB1' }]) })
      .mockResolvedValueOnce({ ok: true, json: () => Promise.resolve([{ slug: 't.txt', name: 't.txt' }]) });

    const utils = render(<KBDetail onLogout={() => {}} />);
    rerender = utils.rerender;

    await waitFor(() => expect(fetch).toHaveBeenCalledTimes(2));

    fetch.mockResolvedValueOnce({ ok: true, json: () => Promise.resolve({ answer: 'hi' }) });
    fireEvent.change(screen.getByRole('textbox', { name: /your question/i }), { target: { value: 'Hi?' } });
    fireEvent.click(screen.getByRole('button', { name: /ask/i }));
    await waitFor(() => expect(screen.getByText('hi')).toBeInTheDocument());

  });

  it('lists files and opens them', async () => {
    fetch
      .mockResolvedValueOnce({ ok: true, json: () => Promise.resolve([{ id: 1, name: 'KB1' }]) })
      .mockResolvedValueOnce({ ok: true, json: () => Promise.resolve([{ slug: 't.txt', name: 't.txt' }]) });

    const utils = render(<KBDetail onLogout={() => {}} />);
    rerender = utils.rerender;

    await waitFor(() => expect(fetch).toHaveBeenCalledTimes(2));

    fetch.mockResolvedValueOnce({
      ok: true,
      headers: { get: () => 'text/plain' },
      text: () => Promise.resolve('content'),
    });

    fireEvent.click(screen.getByText('t.txt'));
    await waitFor(() => expect(screen.getByText('content')).toBeInTheDocument());
  });
});
