import { describe, it, expect } from 'vitest';
import { render, screen, fireEvent, within } from '@testing-library/react';
import { RecentActivityTable } from './RecentActivityTable';
import type { UserActivityLog } from '../api';

const mockLogs: UserActivityLog[] = [
    {
        timestamp: '2026-10-09T14:20:00Z',
        user_id: 'alice@google.com',
        model: 'gemini-2.5-flash',
        input_tokens: 1200,
        output_tokens: 300,
        total_tokens: 1500,
        estimated_cost: 0.0018,
        currency: '$'
    },
    {
        timestamp: '2026-10-09T14:15:00Z',
        user_id: 'alice@google.com',
        model: 'gemini-2.5-pro',
        input_tokens: 4500,
        output_tokens: 1200,
        total_tokens: 5700,
        estimated_cost: 0.0385,
        currency: '$'
    }
];

describe('RecentActivityTable', () => {
    it('renders table headers and activity rows correctly', () => {
        render(<RecentActivityTable logs={mockLogs} />);

        expect(screen.getByText('Recent Activity')).toBeInTheDocument();

        const table = screen.getByRole('table');
        expect(within(table).getByText('gemini-2.5-flash')).toBeInTheDocument();
        expect(within(table).getByText('gemini-2.5-pro')).toBeInTheDocument();

        // Check token numbers formatting
        expect(screen.getByText('1,500')).toBeInTheDocument();
        expect(screen.getByText('5,700')).toBeInTheDocument();

        // Check cost formatting
        expect(screen.getByText('$0.0018')).toBeInTheDocument();
        expect(screen.getByText('$0.0385')).toBeInTheDocument();
    });

    it('filters rows by selected model', () => {
        render(<RecentActivityTable logs={mockLogs} />);

        const select = screen.getByRole('combobox', { name: /filter by model/i });
        fireEvent.change(select, { target: { value: 'gemini-2.5-pro' } });

        const table = screen.getByRole('table');
        expect(within(table).queryByText('gemini-2.5-flash')).not.toBeInTheDocument();
        expect(within(table).getByText('gemini-2.5-pro')).toBeInTheDocument();
    });

    it('renders empty state when there are no activity logs', () => {
        render(<RecentActivityTable logs={[]} />);

        expect(screen.getByText(/no recent activity recorded/i)).toBeInTheDocument();
    });
});
