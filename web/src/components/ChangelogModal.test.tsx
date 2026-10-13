import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { ChangelogModal } from './ChangelogModal';
import { ChangelogRelease } from '../data/changelog';

const mockReleases: ChangelogRelease[] = [
  {
    version: 'v0.6.0',
    date: 'October 13, 2026',
    title: 'Native Currency Support & Release Changelog Popup',
    items: [
      {
        title: 'Multi-Currency Engine',
        description: 'Support for EUR, GBP, JPY, CAD, AUD, USD.',
      },
      {
        title: 'What\'s New Changelog Popup',
        description: 'Instant notification popup highlighting new features.',
      },
    ],
  },
  {
    version: 'v0.5.1',
    date: 'October 10, 2026',
    title: 'Version derivation and release tag links',
    items: [
      {
        title: 'Version Synchronization',
        description: 'Dynamically load application build metadata.',
      },
    ],
  },
];

describe('ChangelogModal', () => {
  it('renders nothing when isOpen is false', () => {
    const { container } = render(
      <ChangelogModal isOpen={false} onClose={vi.fn()} releases={mockReleases} />
    );
    expect(container.firstChild).toBeNull();
  });

  it('renders modal dialog with header, badges, dates, and bullet list', () => {
    render(
      <ChangelogModal isOpen={true} onClose={vi.fn()} releases={mockReleases} />
    );

    expect(screen.getByRole('dialog')).toBeInTheDocument();
    expect(screen.getByText("What's new in AGY Cost Board")).toBeInTheDocument();
    expect(
      screen.getByText('Updates to AGY Cost Board — scroll for earlier versions.')
    ).toBeInTheDocument();

    // Check latest release headline and badge
    expect(screen.getByText('v0.6.0')).toBeInTheDocument();
    expect(screen.getByText('October 13, 2026')).toBeInTheDocument();
    expect(
      screen.getByText('Native Currency Support & Release Changelog Popup')
    ).toBeInTheDocument();
    expect(screen.getByText(/Multi-Currency Engine/)).toBeInTheDocument();

    // Check earlier release
    expect(screen.getByText('v0.5.1')).toBeInTheDocument();
    expect(screen.getByText('October 10, 2026')).toBeInTheDocument();
    expect(
      screen.getByText('Version derivation and release tag links')
    ).toBeInTheDocument();

    // Check "Got it" button
    expect(screen.getByRole('button', { name: /got it/i })).toBeInTheDocument();
  });

  it('calls onClose when "Got it" button is clicked', () => {
    const onClose = vi.fn();
    render(
      <ChangelogModal isOpen={true} onClose={onClose} releases={mockReleases} />
    );

    const gotItBtn = screen.getByRole('button', { name: /got it/i });
    fireEvent.click(gotItBtn);
    expect(onClose).toHaveBeenCalledTimes(1);
  });

  it('calls onClose when backdrop overlay is clicked', () => {
    const onClose = vi.fn();
    render(
      <ChangelogModal isOpen={true} onClose={onClose} releases={mockReleases} />
    );

    const backdrop = screen.getByTestId('changelog-backdrop');
    fireEvent.click(backdrop);
    expect(onClose).toHaveBeenCalledTimes(1);
  });

  it('calls onClose when Escape key is pressed', () => {
    const onClose = vi.fn();
    render(
      <ChangelogModal isOpen={true} onClose={onClose} releases={mockReleases} />
    );

    fireEvent.keyDown(window, { key: 'Escape' });
    expect(onClose).toHaveBeenCalledTimes(1);
  });
});
