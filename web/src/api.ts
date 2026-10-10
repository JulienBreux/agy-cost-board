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
  const rawList = await res.json();
  if (!Array.isArray(rawList)) return [];
  return rawList.map((item: any) => ({
    timestamp: item.timestamp || '',
    user_id: item.user_id || userId,
    model: item.model || 'unknown',
    input_tokens: item.input_tokens ?? item.prompt_tokens ?? 0,
    output_tokens: item.output_tokens ?? item.completion_tokens ?? 0,
    total_tokens: item.total_tokens ?? ((item.prompt_tokens || 0) + (item.completion_tokens || 0)),
    estimated_cost: item.estimated_cost ?? 0,
    currency: item.currency || 'USD',
  }));
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
  const raw = await res.json();

  if (raw && raw.kpis && raw.budget) {
    return raw as UserConsumptionDriving;
  }

  const rawTrends: Array<{
    date: string;
    total_cost?: number;
    cost?: number;
    total_tokens?: number;
    tokens?: number;
    by_model?: Record<string, number>;
  }> = raw.daily_trends || [];

  const modelDist: Record<string, ModelDetail> = {};
  let totalModelCost = 0;
  rawTrends.forEach((t) => {
    if (t.by_model) {
      Object.entries(t.by_model).forEach(([m, cost]) => {
        const numCost = Number(cost) || 0;
        totalModelCost += numCost;
        if (!modelDist[m]) {
          modelDist[m] = { tokens: 0, cost: 0, share: 0 };
        }
        modelDist[m].cost += numCost;
      });
    }
  });

  Object.keys(modelDist).forEach((m) => {
    modelDist[m].share = totalModelCost > 0 ? (modelDist[m].cost / totalModelCost) * 100 : 0;
    if (totalModelCost > 0 && raw.total_tokens_in_window) {
      modelDist[m].tokens = Math.round((modelDist[m].cost / totalModelCost) * raw.total_tokens_in_window);
    }
  });

  const dailyTrend: PersonalDailyTrend[] = rawTrends.map((t) => ({
    date: t.date,
    cost: t.total_cost ?? t.cost ?? 0,
    total_tokens: t.total_tokens ?? t.tokens ?? 0,
    models: t.by_model || {},
  }));

  const recommendations = raw.recommendations || [];
  const optimizationTips: OptimizationTip[] = recommendations.map((r: any) => ({
    type: r.id || 'optimization',
    title: r.title || 'Optimization Insight',
    description: r.description || '',
    potential_savings_monthly: r.estimated_savings_usd || 0,
    severity: (r.severity === 'success' || r.severity === 'info' || r.severity === 'warning' || r.severity === 'critical')
      ? (r.severity === 'success' ? 'info' : r.severity)
      : 'info',
  }));

  const mtdSpend = raw.total_spend_in_window ?? 0;
  const budgetThreshold = raw.monthly_budget ?? monthlyBudget ?? 100;
  const projectedSpend = raw.projected_month_end_spend ?? mtdSpend;
  const overage = Math.max(0, projectedSpend - budgetThreshold);
  const onTrack = raw.budget_status ? raw.budget_status !== 'exceeded' : projectedSpend <= budgetThreshold;

  return {
    user_id: raw.user_id || userId,
    window_days: days,
    currency: raw.currency || 'USD',
    kpis: {
      mtd_spend: mtdSpend,
      daily_burn_rate: raw.daily_burn_rate ?? 0,
      weekly_burn_rate: raw.weekly_burn_rate ?? 0,
      total_user_tokens: raw.total_tokens_in_window ?? 0,
      org_spend_share_percent: raw.org_spend_share_pct ?? 0,
    },
    budget: {
      threshold: budgetThreshold,
      current_spend: mtdSpend,
      utilization_percent: raw.budget_consumed_pct ?? (budgetThreshold > 0 ? (mtdSpend / budgetThreshold) * 100 : 0),
      projected_month_end_spend: projectedSpend,
      projected_overage: overage,
      on_track: onTrack,
    },
    daily_trend: dailyTrend,
    model_distribution: modelDist,
    optimization_tips: optimizationTips,
  };
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


