import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import React from 'react';
import ChatView from '../components/ChatView';

describe('ChatView', () => {
  it('starts new chat and renders answer', async () => {
    fetch.mockResolvedValueOnce({ ok: true, json: () => Promise.resolve({ answer: 'hi' }) });
    render(<ChatView kbID="1" kbName="KB1" />);

    fireEvent.change(screen.getByRole('textbox', { name: /your question/i }), { target: { value: 'Hi?' } });
    fireEvent.click(screen.getByRole('button', { name: /ask/i }));

    await waitFor(() =>
      expect(fetch).toHaveBeenCalledWith(
        '/api/kbs/1/ask',
        expect.objectContaining({ method: 'POST' })
      )
    );
    await waitFor(() => expect(screen.getByText('hi')).toBeInTheDocument());
  });

  it('continues chat using history', async () => {
    fetch.mockResolvedValueOnce({ ok: true, json: () => Promise.resolve({ answer: 'hi' }) });
    render(<ChatView kbID="1" kbName="KB1" />);

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

  it('shows answer context', async () => {
    fetch.mockResolvedValueOnce({
      ok: true,
      json: () => Promise.resolve({
        answer: 'ans',
        chunks: [{ file_name: 'f.txt', index: 0, content: 'ctx' }],
      }),
    });
    render(<ChatView kbID="1" kbName="KB1" />);

    fireEvent.change(screen.getByRole('textbox', { name: /your question/i }), { target: { value: 'Q?' } });
    fireEvent.click(screen.getByRole('button', { name: /ask/i }));

    await waitFor(() => expect(screen.getByText('ans')).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: /show context/i }));
    await waitFor(() => expect(screen.getByText('ctx')).toBeInTheDocument());
  });
});
