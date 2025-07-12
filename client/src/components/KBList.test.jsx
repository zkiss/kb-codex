import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import React from 'react';
import KBList from '../components/KBList';
import { vi } from 'vitest';

const mockNavigate = vi.fn();

vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom');
  return { ...actual, useNavigate: () => mockNavigate };
});

describe('KBList', () => {
  it('fetches and creates knowledge bases', async () => {
    fetch.mockResolvedValueOnce({ ok: true, json: () => Promise.resolve([]) });
    render(<KBList onLogout={() => {}} />);

    await waitFor(() => expect(fetch).toHaveBeenCalledWith('/api/kbs'));

    fetch.mockResolvedValueOnce({ ok: true });
    fetch.mockResolvedValueOnce({ ok: true, json: () => Promise.resolve([{ id: 1, name: 'Test KB' }]) });

    fireEvent.change(screen.getByRole('textbox', { name: /New knowledge base name/i }), { target: { value: 'Test KB' } });
    fireEvent.click(screen.getByRole('button', { name: /Create KB/i }));

    await waitFor(() =>
      expect(fetch).toHaveBeenCalledWith(
        '/api/kbs',
        expect.objectContaining({
          method: 'POST',
          body: JSON.stringify({ name: 'Test KB' }),
        })
      )
    );
    await waitFor(() => expect(screen.getByText('Test KB')).toBeInTheDocument());

    fireEvent.click(screen.getByText('Test KB'));
    expect(mockNavigate).toHaveBeenCalledWith('/kbs/1');
  });
});
