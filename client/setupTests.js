import '@testing-library/jest-dom/vitest';
import { vi, beforeEach } from 'vitest';

beforeEach(() => {
  vi.restoreAllMocks();
  global.fetch = vi.fn();
  global.alert = vi.fn();
});

// Polyfill DOMMatrix for react-pdf in jsdom
global.DOMMatrix = global.DOMMatrix || class DOMMatrix {
  constructor() {}
};
