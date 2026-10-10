import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import { UserDashboardView } from './UserDashboardView';
import * as api from '../api';

const mockCurrentUser: api.CurrentUserIdentity = {
    email: 'testuser@google.com',
    displayName: 'Test User',
    authenticated: true,
    source: 'iap'
};

const mockDashboard: api.UserConsumptionDriving = {
    user_id: 'testuser@google.com',
    window_days: 30,
    currency: '$',
    kpis: {
        mtd_spend: 34.50,
        daily_burn_rate: 1.15,
        weekly_burn_rate: 8.05,
        total_user_tokens: 1250000,
        org_spend_share_percent: 18.2
    },
    budget: {
        threshold: 100,
        current_spend: 34.50,
        utilization_percent: 34.5,
        projected_month_end_spend: 48.00,
        projected_overage: 0,
        on_track: true
    },
    daily_trend: [
        { date: '2026-10-08', cost: 1.20, total_tokens: 45000 }
    ],
    model_distribution: {
        'gemini-2.5-pro': { cost: 25.00, tokens: 600000, share: 0.72 },
        'gemini-2.5-flash': { cost: 9.50, tokens: 650000, share: 0.28 }
    },
    optimization_tips: [
        {
            type: 'model_switch',
            title: 'Optimize Model Mix',
            description: 'Switch routine tasks to Flash to save money.',
            potential_savings_monthly: 12.00,
            severity: 'info'
        }
    ]
};

const mockActivity: api.UserActivityLog[] = [
    {
        timestamp: '2026-10-09T14:00:00Z',
        user_id: 'testuser@google.com',
        model: 'gemini-2.5-flash',
        input_tokens: 1000,
        output_tokens: 250,
        total_tokens: 1250,
        estimated_cost: 0.0015,
        currency: '$'
    }
];

describe('UserDashboardView', () => {
    beforeEach(() => {
        vi.restoreAllMocks();
        vi.spyOn(api, 'fetchCurrentUser').mockResolvedValue(mockCurrentUser);
        vi.spyOn(api, 'fetchUserDashboard').mockResolvedValue(mockDashboard);
        vi.spyOn(api, 'fetchUserActivity').mockResolvedValue(mockActivity);
    });

    it('loads and renders all dashboard components with fetched user data', async () => {
        render(<UserDashboardView days={30} availableUsers={['testuser@google.com', 'other@google.com']} />);

        // Verify loading spinner shows initially
        expect(screen.getByText(/loading personal dashboard/i)).toBeInTheDocument();

        // Verify main components appear after fetch
        await waitFor(() => {
            expect(screen.getByText('testuser@google.com')).toBeInTheDocument();
            expect(screen.getByText('Month-to-Date Spend')).toBeInTheDocument();
            expect(screen.getByText('Monthly Budget & Quota Driving')).toBeInTheDocument();
            expect(screen.getByText('Antigravity CLI Status Bar Integration')).toBeInTheDocument();
            expect(screen.getByText('Daily Spend Trajectory')).toBeInTheDocument();
            expect(screen.getByText('Optimization Insights')).toBeInTheDocument();
            expect(screen.getByText('Recent Activity')).toBeInTheDocument();
            expect(screen.getByText('Live Telemetry')).toBeInTheDocument();
        });
    });

    it('allows switching selected user and re-fetches dashboard data', async () => {
        render(<UserDashboardView days={30} availableUsers={['testuser@google.com', 'other@google.com']} initialUserId="testuser@google.com" />);

        await waitFor(() => {
            expect(screen.getByText('testuser@google.com')).toBeInTheDocument();
        });

        const trigger = screen.getByRole('button', { name: /switch user/i });
        fireEvent.click(trigger);

        const userOption = screen.getByText('other@google.com');
        fireEvent.click(userOption);

        await waitFor(() => {
            expect(api.fetchUserDashboard).toHaveBeenCalledWith('other@google.com', 30, expect.any(Number));
        });
    });
});
