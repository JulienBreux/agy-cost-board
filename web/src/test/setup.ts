import '@testing-library/jest-dom';

const storage: Record<string, string> = {};

const localStorageMock = {
  getItem: (key: string): string | null => {
    return key in storage ? storage[key] : null;
  },
  setItem: (key: string, value: string): void => {
    storage[key] = value;
  },
  removeItem: (key: string): void => {
    delete storage[key];
  },
  clear: (): void => {
    Object.keys(storage).forEach((k) => delete storage[k]);
  },
  get length(): number {
    return Object.keys(storage).length;
  },
  key: (index: number): string | null => {
    const keys = Object.keys(storage);
    return keys[index] ?? null;
  },
};

Object.defineProperty(globalThis, 'localStorage', {
  value: localStorageMock,
  writable: true,
});
