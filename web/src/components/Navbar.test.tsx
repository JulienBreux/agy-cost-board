import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { ThemeProvider } from '../context/ThemeContext';
import { Navbar } from './Navbar';

const renderNavbar = (props: Parameters<typeof Navbar>[0]) => {
  return render(
    <ThemeProvider>
      <Navbar {...props} />
    </ThemeProvider>
  );
};

describe('Navbar Component', () => {
  it('renders brand and navigation tabs correctly', () => {
    const setActiveTab = vi.fn();
    const setDays = vi.fn();

    renderNavbar({
      activeTab: 'overview',
      setActiveTab,
      days: 30,
      setDays,
    });

    expect(screen.getByText('Usage and Spend')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /overview/i })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /my consumption/i })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /attributed costs/i })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /seat governance/i })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /setup & health/i })).toBeInTheDocument();
  });

  it('triggers setActiveTab when a navigation button is clicked', () => {
    const setActiveTab = vi.fn();
    const setDays = vi.fn();

    renderNavbar({
      activeTab: 'overview',
      setActiveTab,
      days: 30,
      setDays,
    });

    fireEvent.click(screen.getByRole('button', { name: /seat governance/i }));
    expect(setActiveTab).toHaveBeenCalledWith('licenses');
  });

  it('triggers setDays when time window selector buttons are clicked', () => {
    const setActiveTab = vi.fn();
    const setDays = vi.fn();

    renderNavbar({
      activeTab: 'overview',
      setActiveTab,
      days: 30,
      setDays,
    });

    fireEvent.click(screen.getByRole('button', { name: '7d' }));
    expect(setDays).toHaveBeenCalledWith(7);

    fireEvent.click(screen.getByRole('button', { name: '14d' }));
    expect(setDays).toHaveBeenCalledWith(14);
  });

  it('renders "What\'s new" button and invokes onOpenChangelog when clicked', () => {
    const onOpenChangelog = vi.fn();

    renderNavbar({
      activeTab: 'overview',
      setActiveTab: vi.fn(),
      days: 30,
      setDays: vi.fn(),
      onOpenChangelog,
    });

    const whatsNewButton = screen.getByRole('button', { name: /what's new/i });
    expect(whatsNewButton).toBeInTheDocument();

    fireEvent.click(whatsNewButton);
    expect(onOpenChangelog).toHaveBeenCalledTimes(1);
  });

  it('prevents navigation items from wrapping onto multiple lines', () => {
    renderNavbar({
      activeTab: 'overview',
      setActiveTab: vi.fn(),
      days: 30,
      setDays: vi.fn(),
    });

    const brandTitle = screen.getByText('Usage and Spend');
    expect(brandTitle).toHaveClass('whitespace-nowrap');

    const navButtons = [
      screen.getByRole('button', { name: /overview/i }),
      screen.getByRole('button', { name: /my consumption/i }),
      screen.getByRole('button', { name: /attributed costs/i }),
      screen.getByRole('button', { name: /seat governance/i }),
      screen.getByRole('button', { name: /setup & health/i }),
    ];

    navButtons.forEach((btn) => {
      expect(btn).toHaveClass('whitespace-nowrap');
      expect(btn).toHaveClass('flex-shrink-0');
    });

    const nav = screen.getByRole('navigation', { name: /main navigation/i });
    expect(nav).toHaveClass('overflow-x-auto');
  });
});
