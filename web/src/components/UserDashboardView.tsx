import { FC, useState, useEffect, useCallback } from 'react';
import {
  fetchCurrentUser,
  fetchUserDashboard,
  fetchUserActivity,
  CurrentUserIdentity,
  UserConsumptionDriving,
  UserActivityLog,
} from '../api';
import { UserSelector } from './UserSelector';
import { LiveSyncBadge } from './LiveSyncBadge';
import { PersonalKPICards } from './PersonalKPICards';
import { BudgetProgressBar } from './BudgetProgressBar';
import { PersonalCostChart } from './PersonalCostChart';
import { OptimizationCard } from './OptimizationCard';
import { RecentActivityTable } from './RecentActivityTable';
import { RefreshCw, AlertTriangle } from 'lucide-react';

interface UserDashboardViewProps {
  days: number;
  availableUsers?: string[];
  initialUserId?: string;
  onUserChange?: (userId: string) => void;
}

export const UserDashboardView: FC<UserDashboardViewProps> = ({
  days,
  availableUsers = [],
  initialUserId,
  onUserChange,
}) => {
  const [currentUser, setCurrentUser] = useState<CurrentUserIdentity | null>(null);
  const [selectedUserId, setSelectedUserId] = useState<string>(() => {
    return initialUserId || localStorage.getItem('agy_selected_user_id') || '';
  });
  const [monthlyBudget, setMonthlyBudget] = useState<number>(() => {
    const saved = localStorage.getItem('agy_user_monthly_budget');
    return saved ? Number(saved) : 100;
  });

  const [dashboardData, setDashboardData] = useState<UserConsumptionDriving | null>(null);
  const [activityLogs, setActivityLogs] = useState<UserActivityLog[]>([]);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [activityLoading, setActivityLoading] = useState<boolean>(false);
  const [lastUpdated, setLastUpdated] = useState<Date | null>(null);
  const [error, setError] = useState<string | null>(null);

  // 1. Detect current user identity on mount
  useEffect(() => {
    let isMounted = true;
    fetchCurrentUser()
      .then((user) => {
        if (!isMounted) return;
        setCurrentUser(user);
        if (!selectedUserId) {
          const defaultUser = user.email || 'developer@example.com';
          setSelectedUserId(defaultUser);
          onUserChange?.(defaultUser);
        }
      })
      .catch((err) => {
        console.error('Failed to resolve current user identity:', err);
        if (!selectedUserId) {
          const fallback = 'developer@example.com';
          setSelectedUserId(fallback);
          onUserChange?.(fallback);
        }
      });

    return () => {
      isMounted = false;
    };
  }, []);

  // Update selection if prop changes
  useEffect(() => {
    if (initialUserId && initialUserId !== selectedUserId) {
      setSelectedUserId(initialUserId);
    }
  }, [initialUserId]);

  // 2. Fetch dashboard data & activity
  const loadDashboard = useCallback(
    async (isBackground = false) => {
      if (!selectedUserId) return;
      if (!isBackground) {
        setIsLoading(true);
        setError(null);
      } else {
        setActivityLoading(true);
      }

      try {
        const [dash, activity] = await Promise.all([
          fetchUserDashboard(selectedUserId, days, monthlyBudget),
          fetchUserActivity(selectedUserId, days, 20),
        ]);
        setDashboardData(dash);
        setActivityLogs(activity);
        setLastUpdated(new Date());
      } catch (err: unknown) {
        if (!isBackground) {
          setError(
            err instanceof Error ? err.message : 'Failed to fetch personal dashboard data'
          );
        }
      } finally {
        if (!isBackground) {
          setIsLoading(false);
        }
        setActivityLoading(false);
      }
    },
    [selectedUserId, days, monthlyBudget]
  );

  useEffect(() => {
    if (selectedUserId) {
      loadDashboard(false);
    }
  }, [selectedUserId, days, monthlyBudget, loadDashboard]);

  // 3. Live polling interval for activity stream (every 20s)
  useEffect(() => {
    if (!selectedUserId) return;
    const interval = setInterval(() => {
      loadDashboard(true);
    }, 20000);

    return () => clearInterval(interval);
  }, [selectedUserId, loadDashboard]);

  const handleUserSelect = (userId: string) => {
    setSelectedUserId(userId);
    localStorage.setItem('agy_selected_user_id', userId);
    onUserChange?.(userId);
  };

  const handleUpdateBudget = (newBudget: number) => {
    setMonthlyBudget(newBudget);
    localStorage.setItem('agy_user_monthly_budget', String(newBudget));
  };

  // Compile full user list for selector dropdown
  const allUsers = Array.from(
    new Set([
      ...(currentUser?.email ? [currentUser.email] : []),
      ...(selectedUserId ? [selectedUserId] : []),
      ...availableUsers,
    ])
  ).filter(Boolean);

  return (
    <div className="space-y-6">
      {/* Top Controls: Identity Selector & Live Status */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 bg-gray-900/60 border border-gray-800 rounded-xl p-4">
        <UserSelector
          availableUsers={allUsers}
          currentUser={currentUser?.email || ''}
          selectedUser={selectedUserId}
          onSelectUser={handleUserSelect}
        />
        <LiveSyncBadge
          isLive={true}
          lastUpdated={lastUpdated}
          onRefresh={() => loadDashboard(false)}
          isLoading={isLoading || activityLoading}
        />
      </div>

      {/* Error Alert */}
      {error && (
        <div className="p-4 bg-red-500/10 border border-red-500/30 rounded-xl flex items-center space-x-3 text-red-400">
          <AlertTriangle className="h-5 w-5 flex-shrink-0" />
          <div className="text-xs sm:text-sm font-medium">{error}</div>
          <button
            onClick={() => loadDashboard(false)}
            className="ml-auto underline text-xs font-semibold hover:text-white"
          >
            Retry
          </button>
        </div>
      )}

      {/* Loading Spinner */}
      {isLoading && !dashboardData && (
        <div className="py-24 flex flex-col items-center justify-center space-y-3">
          <RefreshCw className="h-8 w-8 text-blue-500 animate-spin" />
          <span className="text-sm text-gray-400 font-medium">Loading personal dashboard...</span>
        </div>
      )}

      {dashboardData && (
        <>
          {/* Top KPI Cards */}
          <PersonalKPICards
            kpis={dashboardData.kpis}
            currency={dashboardData.currency}
          />

          {/* Budget & Quota Gauge */}
          <BudgetProgressBar
            budget={dashboardData.budget}
            currency={dashboardData.currency}
            onUpdateBudget={handleUpdateBudget}
          />

          {/* Spend Trajectory & Model Distribution */}
          <PersonalCostChart
            trends={dashboardData.daily_trend}
            modelDistribution={dashboardData.model_distribution}
            currency={dashboardData.currency}
          />

          {/* Optimization Insights */}
          <OptimizationCard
            tips={dashboardData.optimization_tips}
            currency={dashboardData.currency}
          />

          {/* Recent Live Activity Stream */}
          <RecentActivityTable
            logs={activityLogs}
            currency={dashboardData.currency}
            isLoading={activityLoading}
          />
        </>
      )}
    </div>
  );
};
