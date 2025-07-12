import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import React from 'react';
import Register from '../components/Register';
import { vi } from 'vitest';

const mockNavigate = vi.fn();

vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom');
  return { ...actual, useNavigate: () => mockNavigate, Link: ({children, ...props}) => <a {...props}>{children}</a> };
});

describe('Register', () => {
  it('submits registration data', async () => {
    fetch.mockResolvedValueOnce({ ok: true, json: vi.fn() });
    render(<Register />);

    fireEvent.change(screen.getByLabelText(/Email address/i), { target: { value: 'a@b.com' } });
    fireEvent.change(screen.getByLabelText(/Password/i), { target: { value: 'secret' } });

    fireEvent.click(screen.getByRole('button', { name: /register/i }));

    await waitFor(() => expect(fetch).toHaveBeenCalled());
    expect(fetch).toHaveBeenCalledWith(
      '/api/register',
      expect.objectContaining({
        method: 'POST',
        body: JSON.stringify({ email: 'a@b.com', password: 'secret' }),
      })
    );
    await waitFor(() => expect(mockNavigate).toHaveBeenCalledWith('/login'));
  });
});
