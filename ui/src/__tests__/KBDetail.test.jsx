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

    fetch
      .mockResolvedValueOnce({ ok: true })
      .mockResolvedValueOnce({ ok: true, json: () => Promise.resolve([{ slug: 'test.txt', name: 'test.txt' }]) });

    fireEvent.click(screen.getByTestId('upload-btn'));

    await waitFor(() => {
      const call = fetch.mock.calls.find(c => c[0] === '/api/kbs/1/files' && c[1]?.method === 'POST');
      expect(call).toBeDefined();
      expect(call[1].body).toBeInstanceOf(FormData);
      expect(call[1].body.get('file')).toBe(file);
    });
  });

  it('uploads PDF file', async () => {
    fetch
      .mockResolvedValueOnce({ ok: true, json: () => Promise.resolve([{ id: 1, name: 'KB1' }]) })
      .mockResolvedValueOnce({ ok: true, json: () => Promise.resolve([]) });

    const utils = render(<KBDetail onLogout={() => {}} />);
    rerender = utils.rerender;

    await waitFor(() => expect(fetch).toHaveBeenCalledTimes(2));

    const pdfFile = new File(['%PDF-1.1'], 'test.pdf', { type: 'application/pdf' });
    fireEvent.change(screen.getByTestId('file-input'), { target: { files: [pdfFile] } });

    fetch
      .mockResolvedValueOnce({ ok: true })
      .mockResolvedValueOnce({ ok: true, json: () => Promise.resolve([{ slug: 'test.pdf', name: 'test.pdf' }]) });

    fireEvent.click(screen.getByTestId('upload-btn'));

    await waitFor(() => {
      const call = fetch.mock.calls.find(c => c[0] === '/api/kbs/1/files' && c[1]?.method === 'POST');
      expect(call).toBeDefined();
      expect(call[1].body).toBeInstanceOf(FormData);
      expect(call[1].body.get('file')).toBe(pdfFile);
    });
  });

  it('starts new chat and renders answer', async () => {
    fetch
      .mockResolvedValueOnce({ ok: true, json: () => Promise.resolve([{ id: 1, name: 'KB1' }]) })
      .mockResolvedValueOnce({ ok: true, json: () => Promise.resolve([]) });

    const utils = render(<KBDetail onLogout={() => {}} />);
    rerender = utils.rerender;

    await waitFor(() => expect(fetch).toHaveBeenCalledTimes(2));

    fetch.mockResolvedValueOnce({ ok: true, json: () => Promise.resolve({ answer: 'hi' }) });

    fireEvent.change(screen.getByRole('textbox', { name: /your question/i }), { target: { value: 'Hi?' } });
    fireEvent.click(screen.getByRole('button', { name: /ask/i }));

    await waitFor(() =>
      expect(fetch).toHaveBeenCalledWith(
        '/api/kbs/1/ask',
        expect.objectContaining({ method: 'POST', body: JSON.stringify({ question: 'Hi?', history: [] }) })
      )
    );
    await waitFor(() => expect(screen.getByText('hi')).toBeInTheDocument());
  });

  it('continues existing chat using history', async () => {
    fetch
      .mockResolvedValueOnce({ ok: true, json: () => Promise.resolve([{ id: 1, name: 'KB1' }]) })
      .mockResolvedValueOnce({ ok: true, json: () => Promise.resolve([]) });

    const utils = render(<KBDetail onLogout={() => {}} />);
    rerender = utils.rerender;

    await waitFor(() => expect(fetch).toHaveBeenCalledTimes(2));

    fetch.mockResolvedValueOnce({ ok: true, json: () => Promise.resolve({ answer: 'hi' }) });
    fireEvent.change(screen.getByRole('textbox', { name: /your question/i }), { target: { value: 'Hi?' } });
    fireEvent.click(screen.getByRole('button', { name: /ask/i }));
    await waitFor(() => expect(screen.getByText('hi')).toBeInTheDocument());

    fetch.mockResolvedValueOnce({ ok: true, json: () => Promise.resolve({ answer: 'there' }) });
    fireEvent.change(screen.getByRole('textbox', { name: /your question/i }), { target: { value: 'How?' } });
    fireEvent.click(screen.getByRole('button', { name: /ask/i }));

    await waitFor(() => {
      const body = JSON.parse(fetch.mock.calls.at(-1)[1].body);
      expect(body.history).toEqual([
        { role: 'user', content: 'Hi?' },
        { role: 'assistant', content: 'hi' },
      ]);
    });
    await waitFor(() => expect(screen.getByText('there')).toBeInTheDocument());
  });

  it('opens file viewer when clicking file', async () => {
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
