export interface DailyTrend {
  date: string;
  cost: number;
  tokens: number;
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

