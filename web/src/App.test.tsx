import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import App from './App';
import * as api from './api';

const mockOverview: api.OverviewMetrics = {
    total_billed_cost: 1250.50,
    total_tokens: 45000000,
    active_users_count: 12,
    seat_quota: 20,
    assigned_seats: 15,
    utilization_pct: 75.0,
    estimated_monthly_savings: 150.00,
    daily_trends: [],
    model_breakdown: {},
    currency: 'USD'
};

const mockCosts: api.AllocatedUserCost[] = [
    {
        user_id: 'alice@google.com',
        model: 'gemini-2.5-pro',
        usage_date: '2026-10-09',
        user_tokens: 500000,
        total_model_tokens: 1000000,
        token_share: 0.5,
        allocated_cost: 25.00,
        currency: 'USD'
    }
];

const mockGovernance: api.LicenseGovernance = {
    seat_quota: 20,
    assigned_seats: 15,
    active_seats: 12,
    dormant_seats: 3,
    utilization_pct: 75.0,
    estimated_monthly_savings: 150.00,
    dormant_users: []
};

const mockCurrentUser: api.CurrentUserIdentity = {
    email: 'alice@google.com',
    displayName: 'Alice Example',
    authenticated: true,
    source: 'iap'
};

const mockDashboard: api.UserConsumptionDriving = {
    user_id: 'alice@google.com',
    window_days: 30,
    currency: '$',
    kpis: {
        mtd_spend: 25.00,
        daily_burn_rate: 0.83,
        weekly_burn_rate: 5.80,
        total_user_tokens: 500000,
        org_spend_share_percent: 2.0
    },
    budget: {
        threshold: 100,
        current_spend: 25.00,
        utilization_percent: 25.0,
        projected_month_end_spend: 35.00,
        projected_overage: 0,
        on_track: true
    },
    daily_trend: [],
    model_distribution: {},
    optimization_tips: []
};

describe('App Component Navigation', () => {
    beforeEach(() => {
        window.history.replaceState(null, '', '/');
        vi.restoreAllMocks();
        vi.spyOn(api, 'fetchOverview').mockResolvedValue(mockOverview);
        vi.spyOn(api, 'fetchAttributedCosts').mockResolvedValue(mockCosts);
        vi.spyOn(api, 'fetchLicenseGovernance').mockResolvedValue(mockGovernance);
        vi.spyOn(api, 'fetchCurrentUser').mockResolvedValue(mockCurrentUser);
        vi.spyOn(api, 'fetchUserDashboard').mockResolvedValue(mockDashboard);
        vi.spyOn(api, 'fetchUserActivity').mockResolvedValue([]);
        vi.spyOn(api, 'fetchVersion').mockResolvedValue({
            version: 'v0.4.0-dirty',
            commit: 'a3fefe9',
            build_date: '2026-10-10T19:40:54Z'
        });
        vi.spyOn(api, 'fetchSetupStatus').mockResolvedValue({
            project_id: 'test-project',
            timestamp: '2026-10-09T00:00:00Z',
            overall_status: 'OK',
            checks: [],
            passed_count: 0,
            warning_count: 0,
            error_count: 0
        });
    });

    it('renders navbar with My Consumption tab and navigates on click', async () => {
        render(<App />);

        // Check navbar tab exists
        const myConsumptionTab = screen.getByRole('button', { name: /my consumption/i });
        expect(myConsumptionTab).toBeInTheDocument();

        // Click My Consumption tab
        fireEvent.click(myConsumptionTab);

        // Verify personal dashboard view is displayed
        await waitFor(() => {
            expect(screen.getByText('Month-to-Date Spend')).toBeInTheDocument();
            expect(screen.getByText('Monthly Budget & Quota Driving')).toBeInTheDocument();
        });
    });

    it('navigates from user drilldown modal directly to personal consumption dashboard', async () => {
        vi.spyOn(api, 'fetchUserSummary').mockResolvedValue({
            user_id: 'alice@google.com',
            total_tokens: 500000,
            total_cost: 25.00,
            currency: 'USD',
            last_active: '2026-10-09T12:00:00Z',
            seat_status: 'active',
            model_breakdown: {}
        });

        render(<App />);

        await waitFor(() => {
            expect(screen.getByText('alice@google.com')).toBeInTheDocument();
        });

        // Click on user in the table to open modal
        fireEvent.click(screen.getByText('alice@google.com'));

        // Wait for modal button to appear
        await waitFor(() => {
            expect(screen.getByRole('button', { name: /open personal dashboard/i })).toBeInTheDocument();
        });

        // Click Open Personal Dashboard
        fireEvent.click(screen.getByRole('button', { name: /open personal dashboard/i }));

        // Verify view changed to My Consumption dashboard
        await waitFor(() => {
            expect(screen.getByText('Month-to-Date Spend')).toBeInTheDocument();
        });
    });

    it('fetches and renders the dynamic version in the footer', async () => {
        render(<App />);

        await waitFor(() => {
            expect(screen.getByText('v0.4.0-dirty')).toBeInTheDocument();
        });
    });
});

