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
