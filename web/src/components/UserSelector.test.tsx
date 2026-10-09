import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { UserSelector } from './UserSelector';

describe('UserSelector Component', () => {
  const users = [
    'alex.turner@example.com',
    'sara.chen@example.com',
    'marcus.vance@example.com',
  ];

  it('renders currently selected user and current user indicator', () => {
    render(
      <UserSelector
        currentUser="alex.turner@example.com"
        selectedUser="alex.turner@example.com"
        availableUsers={users}
        onSelectUser={vi.fn()}
      />
    );

    expect(screen.getByText('alex.turner@example.com')).toBeInTheDocument();
    expect(screen.getByText('You')).toBeInTheDocument();
  });

  it('opens dropdown, searches users and selects a user', () => {
    const handleSelect = vi.fn();
    render(
      <UserSelector
        currentUser="alex.turner@example.com"
        selectedUser="alex.turner@example.com"
        availableUsers={users}
        onSelectUser={handleSelect}
      />
    );

    // Open dropdown
    const trigger = screen.getByRole('button', { name: /switch user/i });
    fireEvent.click(trigger);

    // Filter by search
    const searchInput = screen.getByPlaceholderText(/search developer/i);
    fireEvent.change(searchInput, { target: { value: 'sara' } });

    expect(screen.getByText('sara.chen@example.com')).toBeInTheDocument();
    expect(screen.queryByText('marcus.vance@example.com')).not.toBeInTheDocument();

    // Click to select
    fireEvent.click(screen.getByText('sara.chen@example.com'));
    expect(handleSelect).toHaveBeenCalledWith('sara.chen@example.com');
  });

  it('renders empty search state when query does not match', () => {
    render(
      <UserSelector
        currentUser="alex.turner@example.com"
        selectedUser="alex.turner@example.com"
        availableUsers={users}
        onSelectUser={vi.fn()}
      />
    );

    fireEvent.click(screen.getByRole('button', { name: /switch user/i }));
    const searchInput = screen.getByPlaceholderText(/search developer/i);
    fireEvent.change(searchInput, { target: { value: 'nonexistent' } });

    expect(screen.getByText(/no developers found/i)).toBeInTheDocument();
  });

  it('shows reset to me button when viewing another user', () => {
    const handleSelect = vi.fn();
    render(
      <UserSelector
        currentUser="alex.turner@example.com"
        selectedUser="sara.chen@example.com"
        availableUsers={users}
        onSelectUser={handleSelect}
      />
    );

    const resetBtn = screen.getByRole('button', { name: /view my dashboard/i });
    expect(resetBtn).toBeInTheDocument();
    fireEvent.click(resetBtn);
    expect(handleSelect).toHaveBeenCalledWith('alex.turner@example.com');
  });
});
