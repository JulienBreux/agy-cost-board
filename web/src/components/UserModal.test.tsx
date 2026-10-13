import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { UserModal } from './UserModal';
import { UserSummary } from '../api';

const mockUserUSD: UserSummary = {
  user_id: 'alice@google.com',
  seat_status: 'active',
  last_active: '2026-10-10T12:00:00Z',
  total_cost: 123.45,
  total_tokens: 500000,
  currency: 'USD',
  model_breakdown: {
    'gemini-2.5-pro': {
      tokens: 400000,
      cost: 100.0,
      share: 0.8,
    },
    'gemini-2.5-flash': {
      tokens: 100000,
      cost: 23.45,
      share: 0.2,
    },
  },
};

const mockUserEUR: UserSummary = {
  user_id: 'bob@google.com',
  seat_status: 'dormant',
  last_active: '2026-10-09T08:30:00Z',
  total_cost: 88.5,
  total_tokens: 250000,
  currency: 'EUR',
  model_breakdown: {
    'gemini-2.5-flash': {
      tokens: 250000,
      cost: 88.5,
      share: 1.0,
    },
  },
};

describe('UserModal', () => {
  it('renders nothing when user is null', () => {
    const { container } = render(<UserModal user={null} onClose={vi.fn()} />);
    expect(container.firstChild).toBeNull();
  });

  it('renders user details and formats monetary values in USD', () => {
    render(<UserModal user={mockUserUSD} onClose={vi.fn()} />);

    expect(screen.getByText('alice@google.com')).toBeInTheDocument();
    expect(screen.getByText('ACTIVE SEAT')).toBeInTheDocument();
    expect(screen.getByText('500,000')).toBeInTheDocument();

    // Check formatted total cost and model breakdown costs
    expect(screen.getByText('$123.45')).toBeInTheDocument();
    expect(screen.getByText('$100.00')).toBeInTheDocument();
    expect(screen.getByText('$23.45')).toBeInTheDocument();
  });

  it('renders user details and formats monetary values in EUR', () => {
    render(<UserModal user={mockUserEUR} onClose={vi.fn()} />);

    expect(screen.getByText('bob@google.com')).toBeInTheDocument();
    expect(screen.getByText('DORMANT SEAT')).toBeInTheDocument();

    // Check EUR formatted currency
    expect(screen.getAllByText(/88\.50/).length).toBe(2);
    expect(screen.getAllByText(/€/).length).toBeGreaterThanOrEqual(1);
  });

  it('handles onClose and onOpenDashboard triggers', () => {
    const onClose = vi.fn();
    const onOpenDashboard = vi.fn();

    render(
      <UserModal
        user={mockUserUSD}
        onClose={onClose}
        onOpenDashboard={onOpenDashboard}
      />
    );

    const openDashBtn = screen.getByRole('button', {
      name: /open personal dashboard/i,
    });
    fireEvent.click(openDashBtn);
    expect(onOpenDashboard).toHaveBeenCalledWith('alice@google.com');

    const closeBtn = screen.getByRole('button', { name: /^close$/i });
    fireEvent.click(closeBtn);
    expect(onClose).toHaveBeenCalled();
  });
});
