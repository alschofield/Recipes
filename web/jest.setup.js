import '@testing-library/jest-dom'

jest.mock('next/link', () => {
  return function MockLink({ href, children, ...props }) {
    return require('react').createElement('a', { href, ...props }, children)
  }
})

Object.defineProperty(window, 'matchMedia', {
  writable: true,
  value: jest.fn().mockImplementation((query) => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: jest.fn(),
    removeListener: jest.fn(),
    addEventListener: jest.fn(),
    removeEventListener: jest.fn(),
    dispatchEvent: jest.fn(),
  })),
})
