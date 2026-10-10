import { describe, it, expect, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { ThemeProvider } from '../context/ThemeContext';
import { ThemeToggle } from './ThemeToggle';

describe('ThemeToggle Component', () => {
  beforeEach(() => {
    localStorage.clear();
    document.documentElement.className = '';
    document.documentElement.removeAttribute('data-theme');
  });

  it('renders the theme toggle button with initial label', () => {
    render(
      <ThemeProvider>
        <ThemeToggle />
      </ThemeProvider>
    );

    const button = screen.getByTestId('theme-toggle-btn');
    expect(button).toBeInTheDocument();
    expect(button).toHaveAttribute('aria-haspopup', 'true');
    expect(button).toHaveAttribute('aria-expanded', 'false');
  });

  it('opens dropdown menu on click and lists Light, Dark, and Auto options', () => {
    render(
      <ThemeProvider>
        <ThemeToggle />
      </ThemeProvider>
    );

    const button = screen.getByTestId('theme-toggle-btn');
    fireEvent.click(button);

    expect(button).toHaveAttribute('aria-expanded', 'true');
    expect(screen.getByTestId('theme-option-light')).toBeInTheDocument();
    expect(screen.getByTestId('theme-option-dark')).toBeInTheDocument();
    expect(screen.getByTestId('theme-option-auto')).toBeInTheDocument();
  });

  it('switches to light theme when Light option is clicked', () => {
    render(
      <ThemeProvider>
        <ThemeToggle />
      </ThemeProvider>
    );

    const button = screen.getByTestId('theme-toggle-btn');
    fireEvent.click(button);

    const lightOption = screen.getByTestId('theme-option-light');
    fireEvent.click(lightOption);

    // Verify localStorage was updated
    expect(localStorage.getItem('agy-theme')).toBe('light');

    // Verify documentElement has class 'light' and not 'dark'
    expect(document.documentElement.classList.contains('light')).toBe(true);
    expect(document.documentElement.classList.contains('dark')).toBe(false);
    expect(document.documentElement.getAttribute('data-theme')).toBe('light');

    // Menu should be closed
    expect(screen.queryByTestId('theme-option-light')).not.toBeInTheDocument();
  });

  it('switches to dark theme when Dark option is clicked', () => {
    render(
      <ThemeProvider>
        <ThemeToggle />
      </ThemeProvider>
    );

    const button = screen.getByTestId('theme-toggle-btn');
    fireEvent.click(button);

    const darkOption = screen.getByTestId('theme-option-dark');
    fireEvent.click(darkOption);

    // Verify localStorage was updated
    expect(localStorage.getItem('agy-theme')).toBe('dark');

    // Verify documentElement has class 'dark' and not 'light'
    expect(document.documentElement.classList.contains('dark')).toBe(true);
    expect(document.documentElement.classList.contains('light')).toBe(false);
    expect(document.documentElement.getAttribute('data-theme')).toBe('dark');
  });

  it('switches to auto theme when Auto option is clicked', () => {
    // First set to light
    localStorage.setItem('agy-theme', 'light');

    render(
      <ThemeProvider>
        <ThemeToggle />
      </ThemeProvider>
    );

    const button = screen.getByTestId('theme-toggle-btn');
    fireEvent.click(button);

    const autoOption = screen.getByTestId('theme-option-auto');
    fireEvent.click(autoOption);

    expect(localStorage.getItem('agy-theme')).toBe('auto');
    expect(document.documentElement.getAttribute('data-theme')).toBe('auto');
  });

  it('closes dropdown when Escape key is pressed', () => {
    render(
      <ThemeProvider>
        <ThemeToggle />
      </ThemeProvider>
    );

    const button = screen.getByTestId('theme-toggle-btn');
    fireEvent.click(button);
    expect(screen.getByTestId('theme-option-light')).toBeInTheDocument();

    fireEvent.keyDown(document, { key: 'Escape' });
    expect(screen.queryByTestId('theme-option-light')).not.toBeInTheDocument();
  });

  it('closes dropdown when clicking outside', () => {
    render(
      <div>
        <div data-testid="outside-element">Outside</div>
        <ThemeProvider>
          <ThemeToggle />
        </ThemeProvider>
      </div>
    );

    const button = screen.getByTestId('theme-toggle-btn');
    fireEvent.click(button);
    expect(screen.getByTestId('theme-option-light')).toBeInTheDocument();

    fireEvent.mouseDown(screen.getByTestId('outside-element'));
    expect(screen.queryByTestId('theme-option-light')).not.toBeInTheDocument();
  });
});
