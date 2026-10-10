import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { StatuslineHelpBox } from './StatuslineHelpBox';

describe('StatuslineHelpBox', () => {
  beforeEach(() => {
    vi.restoreAllMocks();
    localStorage.clear();
  });

  it('renders integration title, direct script link, and example of rendering', () => {
    render(
      <StatuslineHelpBox
        userId="alex@google.com"
        days={30}
        currentSpend={42.5}
        currency="$"
      />
    );

    // Header check
    expect(
      screen.getByText('Antigravity CLI Status Bar Integration')
    ).toBeInTheDocument();
    expect(screen.getByText('Live Spend in Statusline')).toBeInTheDocument();

    // Script link check
    const scriptLink = screen.getByRole('link', { name: /direct script link/i });
    expect(scriptLink).toBeInTheDocument();
    expect(scriptLink).toHaveAttribute(
      'href',
      expect.stringContaining('/statusline.sh?user=alex%40google.com&days=30&ttl=300')
    );

    // Quick install command check
    expect(
      screen.getByText(/curl -sS.*\/statusline\.sh.*-o ~\/\.gemini\/antigravity-cli\/statusline\.sh/i)
    ).toBeInTheDocument();

    // Example rendering check
    expect(screen.getByText('Example of Rendering in Antigravity Status Bar')).toBeInTheDocument();
    expect(screen.getByText('READY')).toBeInTheDocument();
    expect(screen.getByText('Gemini 3.8 Flash (High)')).toBeInTheDocument();
    expect(screen.getByText('cost $43')).toBeInTheDocument();
  });

  it('renders fallback cost of $18 when spend is 0 or undefined', () => {
    render(<StatuslineHelpBox userId="alex@google.com" />);

    expect(screen.getByText('cost $18')).toBeInTheDocument();
  });

  it('allows copying install command to clipboard', async () => {
    const writeTextMock = vi.fn().mockResolvedValue(undefined);
    Object.assign(navigator, {
      clipboard: {
        writeText: writeTextMock,
      },
    });

    render(<StatuslineHelpBox userId="alex@google.com" />);

    const copyBtn = screen.getByRole('button', { name: /copy install command/i });
    fireEvent.click(copyBtn);

    expect(writeTextMock).toHaveBeenCalledWith(
      expect.stringContaining('curl -sS')
    );

    await waitFor(() => {
      expect(screen.getByText('Copied!')).toBeInTheDocument();
    });
  });

  it('toggles collapse and expand when clicking header toggle button', () => {
    render(<StatuslineHelpBox userId="alex@google.com" />);

    // Initially open
    expect(screen.getByText('One-Line Quick Install')).toBeInTheDocument();

    // Click collapse
    const toggleBtn = screen.getByRole('button', { name: /collapse statusline help/i });
    fireEvent.click(toggleBtn);

    // Content should now be hidden
    expect(screen.queryByText('One-Line Quick Install')).not.toBeInTheDocument();
    expect(screen.getByText('Show Instructions')).toBeInTheDocument();

    // Click expand
    fireEvent.click(screen.getByRole('button', { name: /expand statusline help/i }));
    expect(screen.getByText('One-Line Quick Install')).toBeInTheDocument();
  });
});
