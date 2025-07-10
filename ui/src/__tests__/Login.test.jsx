import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import React from 'react';
import Login from '../components/Login';
import { vi } from 'vitest';

const mockNavigate = vi.fn();
const mockOnLogin = vi.fn();

vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom');
  return { ...actual, useNavigate: () => mockNavigate, Link: ({children, ...props}) => <a {...props}>{children}</a> };
});

describe('Login', () => {
  it('logs in and navigates to KB list', async () => {
    fetch.mockResolvedValueOnce({ ok: true, json: () => Promise.resolve({ token: 'access-token' }) });
    render(<Login onLogin={mockOnLogin} />);

    fireEvent.change(screen.getByLabelText(/Email address/i), { target: { value: 'a@b.com' } });
    fireEvent.change(screen.getByLabelText(/Password/i), { target: { value: 'secret' } });

    fireEvent.click(screen.getByRole('button', { name: /login/i }));

    await waitFor(() => expect(fetch).toHaveBeenCalled());
    expect(fetch).toHaveBeenCalledWith(
      '/api/login',
      expect.objectContaining({
        method: 'POST',
        body: JSON.stringify({ email: 'a@b.com', password: 'secret' }),
      })
    );
    await waitFor(() => expect(mockOnLogin).toHaveBeenCalledWith('access-token'));
    await waitFor(() => expect(mockNavigate).toHaveBeenCalledWith('/kbs'));
  });
});
