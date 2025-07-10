import '@testing-library/jest-dom/vitest';
import { vi, beforeEach } from 'vitest';

beforeEach(() => {
  vi.restoreAllMocks();
  global.fetch = vi.fn();
  global.alert = vi.fn();
});
