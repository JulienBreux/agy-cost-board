import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { LiveSyncBadge } from './LiveSyncBadge';

describe('LiveSyncBadge', () => {
  it('renders live status with pulsating indicator', () => {
    render(<LiveSyncBadge lastUpdated={new Date('2026-10-09T15:30:00Z')} isLive={true} />);

    expect(screen.getByText(/Live Telemetry/i)).toBeInTheDocument();
  });

  it('renders pause/offline state when isLive is false', () => {
    render(<LiveSyncBadge isLive={false} />);

    expect(screen.getByText(/Paused|Offline/i)).toBeInTheDocument();
  });

  it('calls onRefresh when refresh button is clicked', () => {
    const onRefresh = vi.fn();
    render(<LiveSyncBadge onRefresh={onRefresh} isLive={true} />);

    const refreshBtn = screen.getByRole('button', { name: /refresh/i });
    fireEvent.click(refreshBtn);

    expect(onRefresh).toHaveBeenCalledTimes(1);
  });
});
