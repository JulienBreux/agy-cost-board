import { describe, it, expect, vi, beforeEach } from 'vitest';
import {
    fetchCurrentUser,
    fetchUserActivity,
    fetchUserDashboard,
    getSelectedUser,
    setSelectedUser,
    getUserBudget,
    setUserBudget,
} from './api';

describe('User Personal Consumption API & Persistence', () => {
    beforeEach(() => {
        localStorage.clear();
        vi.restoreAllMocks();
    });

    describe('fetchCurrentUser', () => {
        it('fetches current user identity from /api/v1/me', async () => {
            const mockUser = {
                email: 'alex.turner@google.com',
                displayName: 'alex.turner@google.com',
                authenticated: true,
                source: 'iap',
            };

            vi.spyOn(globalThis, 'fetch').mockResolvedValueOnce({
                ok: true,
                json: async () => mockUser,
            } as Response);

            const result = await fetchCurrentUser();
            expect(result).toEqual(mockUser);
            expect(fetch).toHaveBeenCalledWith('/api/v1/me');
        });

        it('throws error when response is not ok', async () => {
            vi.spyOn(globalThis, 'fetch').mockResolvedValueOnce({
                ok: false,
                status: 500,
            } as Response);

            await expect(fetchCurrentUser()).rejects.toThrow('HTTP 500: Failed to fetch current user identity');
        });
    });

    describe('fetchVersion', () => {
        it('fetches version info from /api/v1/version', async () => {
            const mockVersion = {
                version: 'v0.4.0-dirty',
                commit: 'a3fefe9',
                build_date: '2026-10-10T19:40:54Z',
            };

            vi.spyOn(globalThis, 'fetch').mockResolvedValueOnce({
                ok: true,
                json: async () => mockVersion,
            } as Response);

            const result = await (await import('./api')).fetchVersion();
            expect(result).toEqual(mockVersion);
            expect(fetch).toHaveBeenCalledWith('/api/v1/version');
        });
    });

    describe('fetchMcpStatus', () => {
        it('fetches MCP status info from /api/v1/mcp/status', async () => {
            const mockMcpStatus = {
                name: 'agy-cost-board',
                version: 'v0.5.0',
                protocol_version: '2025-03-26',
                mcp_enabled: true,
                endpoint: '/mcp',
                tools: [
                    'get_overview_metrics',
                    'get_license_status',
                    'get_users_costs',
                    'get_user_dashboard',
                ],
            };

            vi.spyOn(globalThis, 'fetch').mockResolvedValueOnce({
                ok: true,
                json: async () => mockMcpStatus,
            } as Response);

            const result = await (await import('./api')).fetchMcpStatus();
            expect(result).toEqual(mockMcpStatus);
            expect(fetch).toHaveBeenCalledWith('/api/v1/mcp/status');
        });
    });

    describe('fetchUserActivity', () => {
        it('fetches user activity logs with query params', async () => {
            const mockActivity = [
                {
                    timestamp: '2026-10-09T14:30:00Z',
                    user_id: 'alex.turner@google.com',
                    model: 'gemini-1.5-pro',
                    input_tokens: 1200,
                    output_tokens: 450,
                    total_tokens: 1650,
                    estimated_cost: 0.0055,
                    currency: 'USD',
                },
            ];

            vi.spyOn(globalThis, 'fetch').mockResolvedValueOnce({
                ok: true,
                json: async () => mockActivity,
            } as Response);

            const result = await fetchUserActivity('alex.turner@google.com', 7, 25);
            expect(result).toEqual(mockActivity);
            expect(fetch).toHaveBeenCalledWith('/api/v1/users/alex.turner%40google.com/activity?days=7&limit=25');
        });
    });

    describe('fetchUserDashboard', () => {
        it('fetches complete driving dashboard with budget', async () => {
            const mockDashboard = {
                user_id: 'alex.turner@google.com',
                window_days: 30,
                currency: 'USD',
                kpis: {
                    mtd_spend: 34.5,
                    daily_burn_rate: 1.15,
                    weekly_burn_rate: 8.05,
                    total_user_tokens: 8500000,
                    org_spend_share_percent: 4.8,
                },
                budget: {
                    threshold: 100,
                    current_spend: 34.5,
                    utilization_percent: 34.5,
                    projected_month_end_spend: 48.2,
                    projected_overage: 0,
                    on_track: true,
                },
                daily_trend: [],
                model_distribution: {},
                optimization_tips: [],
            };

            vi.spyOn(globalThis, 'fetch').mockResolvedValueOnce({
                ok: true,
                json: async () => mockDashboard,
            } as Response);

            const result = await fetchUserDashboard('alex.turner@google.com', 30, 100);
            expect(result).toEqual(mockDashboard);
            expect(fetch).toHaveBeenCalledWith('/api/v1/users/alex.turner%40google.com/dashboard?days=30&monthlyBudget=100');
        });

        it('normalizes flat backend response into rich UserConsumptionDriving format', async () => {
            const flatBackendResponse = {
                user_id: 'dev.intern@google.com',
                currency: 'USD',
                total_spend_in_window: 126.94,
                total_tokens_in_window: 320656,
                daily_burn_rate: 4.23,
                weekly_burn_rate: 29.61,
                org_spend_share_pct: 5.3,
                monthly_budget: 100,
                projected_month_end_spend: 126.94,
                budget_consumed_pct: 126.9,
                budget_status: 'exceeded',
                recommendations: [
                    {
                        id: 'optimal-usage',
                        title: 'Healthy Consumption Profile',
                        description: 'Your usage is balanced.',
                        severity: 'success',
                        estimated_savings_usd: 0,
                    },
                ],
                daily_trends: [
                    {
                        date: '2026-09-10',
                        total_cost: 20.89,
                        total_tokens: 50056,
                        by_model: {
                            'gemini-1.5-flash': 6.88,
                            'gemini-1.5-pro': 14.01,
                        },
                    },
                ],
            };

            vi.spyOn(globalThis, 'fetch').mockResolvedValueOnce({
                ok: true,
                json: async () => flatBackendResponse,
            } as Response);

            const result = await fetchUserDashboard('dev.intern@google.com', 30, 100);
            expect(result.user_id).toBe('dev.intern@google.com');
            expect(result.kpis.mtd_spend).toBe(126.94);
            expect(result.kpis.daily_burn_rate).toBe(4.23);
            expect(result.kpis.weekly_burn_rate).toBe(29.61);
            expect(result.kpis.total_user_tokens).toBe(320656);
            expect(result.kpis.org_spend_share_percent).toBe(5.3);

            expect(result.budget.threshold).toBe(100);
            expect(result.budget.current_spend).toBe(126.94);
            expect(result.budget.utilization_percent).toBe(126.9);
            expect(result.budget.on_track).toBe(false);
            expect(result.budget.projected_overage).toBeCloseTo(26.94);

            expect(result.daily_trend).toHaveLength(1);
            expect(result.daily_trend[0].cost).toBe(20.89);
            expect(result.optimization_tips).toHaveLength(1);
            expect(result.model_distribution['gemini-1.5-flash']).toBeDefined();
        });
    });

    describe('Identity & Budget LocalStorage Persistence', () => {
        it('gets and sets selected user in localStorage', () => {
            expect(getSelectedUser()).toBeNull();
            setSelectedUser('sara.chen@google.com');
            expect(getSelectedUser()).toBe('sara.chen@google.com');
        });

        it('gets default budget if none is stored and updates correctly', () => {
            expect(getUserBudget('alex.turner@google.com')).toBe(100);
            setUserBudget('alex.turner@google.com', 150);
            expect(getUserBudget('alex.turner@google.com')).toBe(150);
        });
    });

    describe('fetchLeaderboard', () => {
        it('normalizes Go backend LeaderboardEntry payload (share_of_spend and numeric model_breakdown)', async () => {
            const rawGoPayload = [
                {
                    rank: 1,
                    user_id: 'elena.rostova@google.com',
                    total_cost: 460.24,
                    share_of_spend: 19.32,
                    total_tokens: 1088425,
                    share_of_tokens: 18.97,
                    active_days: 30,
                    cost_per_1k_tokens: 0.4228,
                    top_model: 'claude-5.5-sonnet-medium',
                    model_breakdown: {
                        'claude-5.5-sonnet-medium': 60.0,
                        'gemini-4.0-pro': 40.0,
                    },
                    currency: 'EUR',
                },
            ];

            vi.spyOn(globalThis, 'fetch').mockResolvedValueOnce({
                ok: true,
                json: async () => rawGoPayload,
            } as Response);

            const { fetchLeaderboard } = await import('./api');
            const result = await fetchLeaderboard(30);
            expect(result).toHaveLength(1);
            expect(result[0].spend_share_pct).toBe(19.32);
            expect(result[0].token_share_pct).toBe(18.97);
            const sonnet = result[0].model_breakdown['claude-5.5-sonnet-medium'];
            expect(typeof sonnet).toBe('object');
            if (typeof sonnet === 'object') {
                expect(sonnet.share).toBeCloseTo(0.6);
            }
        });

        it('normalizes WTU, Included/PAYG cost, cache hit rate, and top effort in Leaderboard and UserDashboard', async () => {
            const rawGoPayload = [
                {
                    rank: 1,
                    user_id: 'elena.rostova@google.com',
                    total_cost: 460.24,
                    included_cost: 300.00,
                    payg_cost: 160.24,
                    wtu: 85000,
                    share_of_spend: 19.32,
                    total_tokens: 1000000,
                    cached_tokens: 450000,
                    cache_hit_rate: 45.0,
                    share_of_tokens: 18.97,
                    active_days: 30,
                    cost_per_1k_tokens: 0.4602,
                    top_model: 'gemini-2.5-pro',
                    top_effort: 'high',
                    model_breakdown: {
                        'gemini-2.5-pro': 100.0,
                    },
                    currency: 'USD',
                },
            ];

            vi.spyOn(globalThis, 'fetch').mockResolvedValueOnce({
                ok: true,
                json: async () => rawGoPayload,
            } as Response);

            const { fetchLeaderboard } = await import('./api');
            const lb = await fetchLeaderboard(30);
            expect(lb[0].included_cost).toBe(300.00);
            expect(lb[0].payg_cost).toBe(160.24);
            expect(lb[0].wtu).toBe(85000);
            expect(lb[0].cached_tokens).toBe(450000);
            expect(lb[0].cache_hit_rate).toBe(45.0);
            expect(lb[0].top_effort).toBe('high');

            const flatUserDashboard = {
                user_id: 'elena.rostova@google.com',
                currency: 'USD',
                total_spend_in_window: 460.24,
                included_cost: 300.00,
                payg_cost: 160.24,
                total_wtu: 85000,
                total_tokens_in_window: 1000000,
                cached_tokens_in_window: 450000,
                cache_hit_rate: 45.0,
                top_effort: 'high',
                daily_burn_rate: 15.34,
                weekly_burn_rate: 107.38,
                org_spend_share_pct: 19.32,
                monthly_budget: 500,
                projected_month_end_spend: 460.24,
                budget_consumed_pct: 92.0,
                budget_status: 'warning',
                recommendations: [],
                daily_trends: [],
            };

            vi.spyOn(globalThis, 'fetch').mockResolvedValueOnce({
                ok: true,
                json: async () => flatUserDashboard,
            } as Response);

            const dash = await fetchUserDashboard('elena.rostova@google.com', 30, 500);
            expect(dash.kpis.included_cost).toBe(300.00);
            expect(dash.kpis.payg_cost).toBe(160.24);
            expect(dash.kpis.total_wtu).toBe(85000);
            expect(dash.kpis.cached_tokens).toBe(450000);
            expect(dash.kpis.cache_hit_rate).toBe(45.0);
            expect(dash.kpis.top_effort).toBe('high');
        });
    });
});


