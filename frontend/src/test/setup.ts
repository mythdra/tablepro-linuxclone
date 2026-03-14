import '@testing-library/jest-dom'
import { vi } from 'vitest'

Object.defineProperty(globalThis, 'EventsOn', { value: vi.fn(), writable: true })
Object.defineProperty(globalThis, 'EventsOff', { value: vi.fn(), writable: true })

Object.defineProperty(window, 'Wails', {
  value: {
    EventsOn: vi.fn(),
    EventsOff: vi.fn(),
  },
  writable: true,
})