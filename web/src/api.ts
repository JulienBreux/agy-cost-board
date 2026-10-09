export interface DailyTrend {
  date: string;
  cost?: number;
  total_cost?: number;
  tokens?: number;
  total_tokens?: number;
  active_users?: number;
}

export interface ModelDetail {
  tokens: number;
  cost: number;
  share: number;
}

export interface OverviewMetrics {
  total_billed_cost: number;
  total_tokens: number;
  active_users_count: number;
  seat_quota: number;
  assigned_seats: number;
  utilization_pct: number;
  estimated_monthly_savings: number;
  daily_trends: DailyTrend[];
  model_breakdown: Record<string, ModelDetail>;
  currency: string;
}

export interface AllocatedUserCost {
  user_id: string;
  model: string;
  usage_date: string;
  user_tokens: number;
  total_model_tokens: number;
  token_share: number;
  allocated_cost: number;
  currency: string;
}

export interface LicenseSeat {
  user_id: string;
  status: 'active' | 'at_risk' | 'dormant';
  last_activity: string;
  total_tokens_in_window: number;
}

export interface LicenseGovernance {
  seat_quota: number;
  assigned_seats: number;
  active_seats: number;
  dormant_seats: number;
  utilization_pct: number;
  estimated_monthly_savings: number;
  dormant_users: LicenseSeat[];
}

export interface UserSummary {
  user_id: string;
  total_tokens: number;
  total_cost: number;
  currency: string;
  last_active: string;
  seat_status: 'active' | 'at_risk' | 'dormant';
  model_breakdown: Record<string, ModelDetail>;
}

export const fetchOverview = async (days = 30): Promise<OverviewMetrics> => {
  const res = await fetch(`/api/v1/metrics/overview?days=${days}`);
  if (!res.ok) throw new Error(`HTTP ${res.status}: Failed to fetch overview metrics`);
  return res.json();
};

export const fetchAttributedCosts = async (days = 30, model = ''): Promise<AllocatedUserCost[]> => {
  const url = `/api/v1/costs/users?days=${days}${model ? `&model=${encodeURIComponent(model)}` : ''}`;
  const res = await fetch(url);
  if (!res.ok) throw new Error(`HTTP ${res.status}: Failed to fetch attributed costs`);
  return res.json();
};

export const fetchLicenseGovernance = async (days = 30): Promise<LicenseGovernance> => {
  const res = await fetch(`/api/v1/licenses/status?days=${days}`);
  if (!res.ok) throw new Error(`HTTP ${res.status}: Failed to fetch license governance`);
  return res.json();
};

export const fetchUserSummary = async (userId: string, days = 30): Promise<UserSummary> => {
  const res = await fetch(`/api/v1/users/${encodeURIComponent(userId)}?days=${days}`);
  if (!res.ok) throw new Error(`HTTP ${res.status}: Failed to fetch user summary`);
  return res.json();
};

export interface DiagnosticCheck {
  id: string;
  name: string;
  status: 'OK' | 'WARNING' | 'ERROR';
  message: string;
  duration: number;
  details?: Record<string, unknown>;
  remediation_command?: string;
}

export interface DiagnosticReport {
  project_id: string;
  timestamp: string;
  overall_status: 'OK' | 'WARNING' | 'ERROR';
  checks: DiagnosticCheck[];
  passed_count: number;
  warning_count: number;
  error_count: number;
}

export const fetchSetupStatus = async (): Promise<DiagnosticReport> => {
  const res = await fetch('/api/v1/setup/status');
  if (!res.ok) throw new Error(`HTTP ${res.status}: Failed to fetch setup status`);
  return res.json();
};

export interface CurrentUserIdentity {
  email: string;
  displayName: string;
  authenticated: boolean;
  source: string;
}

export interface UserActivityLog {
  timestamp: string;
  user_id: string;
  model: string;
  input_tokens: number;
  output_tokens: number;
  total_tokens: number;
  estimated_cost: number;
  currency: string;
}

export interface PersonalDailyTrend {
  date: string;
  cost: number;
  total_tokens: number;
  models?: Record<string, number>;
}

export interface OptimizationTip {
  type: string;
  title: string;
  description: string;
  potential_savings_monthly: number;
  severity: 'info' | 'warning' | 'critical';
}

export interface UserBudgetMetrics {
  threshold: number;
  current_spend: number;
  utilization_percent: number;
  projected_month_end_spend: number;
  projected_overage: number;
  on_track: boolean;
}

export interface UserConsumptionKPIs {
  mtd_spend: number;
  daily_burn_rate: number;
  weekly_burn_rate: number;
  total_user_tokens: number;
  org_spend_share_percent: number;
}

export interface UserConsumptionDriving {
  user_id: string;
  window_days: number;
  currency: string;
  kpis: UserConsumptionKPIs;
  budget: UserBudgetMetrics;
  daily_trend: PersonalDailyTrend[];
  model_distribution: Record<string, ModelDetail>;
  optimization_tips: OptimizationTip[];
}

export const fetchCurrentUser = async (): Promise<CurrentUserIdentity> => {
  const res = await fetch('/api/v1/me');
  if (!res.ok) throw new Error(`HTTP ${res.status}: Failed to fetch current user identity`);
  return res.json();
};

export const fetchUserActivity = async (
  userId: string,
  days = 30,
  limit = 20
): Promise<UserActivityLog[]> => {
  const res = await fetch(
    `/api/v1/users/${encodeURIComponent(userId)}/activity?days=${days}&limit=${limit}`
  );
  if (!res.ok) throw new Error(`HTTP ${res.status}: Failed to fetch user activity logs`);
  return res.json();
};

export const fetchUserDashboard = async (
  userId: string,
  days = 30,
  monthlyBudget = 100
): Promise<UserConsumptionDriving> => {
  const res = await fetch(
    `/api/v1/users/${encodeURIComponent(userId)}/dashboard?days=${days}&monthlyBudget=${monthlyBudget}`
  );
  if (!res.ok) throw new Error(`HTTP ${res.status}: Failed to fetch user consumption dashboard`);
  return res.json();
};

const STORAGE_KEY_SELECTED_USER = 'agy_dashboard_selected_user';
const STORAGE_KEY_USER_BUDGET_PREFIX = 'agy_dashboard_user_budget_';

export const getSelectedUser = (): string | null => {
  try {
    return localStorage.getItem(STORAGE_KEY_SELECTED_USER);
  } catch {
    return null;
  }
};

export const setSelectedUser = (userId: string): void => {
  try {
    localStorage.setItem(STORAGE_KEY_SELECTED_USER, userId);
  } catch {
    // ignore
  }
};

export const getUserBudget = (userId: string, defaultBudget = 100): number => {
  try {
    const val = localStorage.getItem(`${STORAGE_KEY_USER_BUDGET_PREFIX}${userId}`);
    if (val !== null) {
      const parsed = parseFloat(val);
      if (!isNaN(parsed) && parsed > 0) return parsed;
    }
  } catch {
    // ignore
  }
  return defaultBudget;
};

export const setUserBudget = (userId: string, budget: number): void => {
  try {
    localStorage.setItem(`${STORAGE_KEY_USER_BUDGET_PREFIX}${userId}`, budget.toString());
  } catch {
    // ignore
  }
};


