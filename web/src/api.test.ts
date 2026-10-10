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
        email: 'alex.turner@example.com',
        displayName: 'alex.turner@example.com',
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

  describe('fetchUserActivity', () => {
    it('fetches user activity logs with query params', async () => {
      const mockActivity = [
        {
          timestamp: '2026-10-09T14:30:00Z',
          user_id: 'alex.turner@example.com',
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

      const result = await fetchUserActivity('alex.turner@example.com', 7, 25);
      expect(result).toEqual(mockActivity);
      expect(fetch).toHaveBeenCalledWith('/api/v1/users/alex.turner%40example.com/activity?days=7&limit=25');
    });
  });

  describe('fetchUserDashboard', () => {
    it('fetches complete driving dashboard with budget', async () => {
      const mockDashboard = {
        user_id: 'alex.turner@example.com',
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

      const result = await fetchUserDashboard('alex.turner@example.com', 30, 100);
      expect(result).toEqual(mockDashboard);
      expect(fetch).toHaveBeenCalledWith('/api/v1/users/alex.turner%40example.com/dashboard?days=30&monthlyBudget=100');
    });

    it('normalizes flat backend response into rich UserConsumptionDriving format', async () => {
      const flatBackendResponse = {
        user_id: 'dev.intern@example.com',
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

      const result = await fetchUserDashboard('dev.intern@example.com', 30, 100);
      expect(result.user_id).toBe('dev.intern@example.com');
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
      setSelectedUser('sara.chen@example.com');
      expect(getSelectedUser()).toBe('sara.chen@example.com');
    });

    it('gets default budget if none is stored and updates correctly', () => {
      expect(getUserBudget('alex.turner@example.com')).toBe(100);
      setUserBudget('alex.turner@example.com', 150);
      expect(getUserBudget('alex.turner@example.com')).toBe(150);
    });
  });
});
