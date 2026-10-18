export interface DailyTrend {
  date: string;
  cost?: number;
  total_cost?: number;
  included_cost?: number;
  payg_cost?: number;
  wtu?: number;
  tokens?: number;
  total_tokens?: number;
  active_users?: number;
  by_model?: Record<string, number>;
}

export interface ModelDetail {
  tokens: number;
  cost: number;
  share: number;
}

export interface TokenCostShare {
  category: string;
  label?: string;
  tokens: number;
  token_share: number;
  cost?: number;
  cost_usd?: number;
  cost_share: number;
}

export interface ModelEffortStats {
  model: string;
  effort_level: string;
  calls: number;
  tokens?: number;
  total_tokens?: number;
  prompt_tokens?: number;
  cached_tokens: number;
  thinking_tokens?: number;
  cache_hit_rate: number;
  wtu: number;
  included_cost: number;
  payg_cost: number;
  total_cost?: number;
  allocated_cost?: number;
}

export interface ClientStats {
  client_name: string;
  client_version?: string;
  calls: number;
  tokens?: number;
  total_tokens?: number;
  token_share?: number;
  wtu: number;
  included_cost?: number;
  payg_cost?: number;
  total_cost?: number;
  allocated_cost?: number;
  share_of_cost?: number;
}

export interface WeeklyPoolWindow {
  window_label: string;
  window_start: string;
  window_end: string;
  pool_limit_usd: number;
  pool_limit_wtu: number;
  included_used_usd: number;
  included_used_wtu: number;
  payg_overage_usd: number;
  payg_overage_wtu: number;
  utilization_pct: number;
}

export interface PoolWhatIfScenario {
  pool_mode: 'pooled' | 'per_user' | string;
  seat_basis: 'assigned' | 'purchased' | string;
  pool_limit_usd: number;
  pool_limit_wtu: number;
  included_used_usd: number;
  included_used_wtu: number;
  payg_overage_usd: number;
  payg_overage_wtu: number;
  remaining_usd: number;
  remaining_wtu: number;
  utilization_pct: number;
  windows?: WeeklyPoolWindow[];
}

export interface MonitoringQuotaMeter {
  metric_type: string;
  metered_used_wtu: number;
  metered_used_usd: number;
  log_estimated_wtu: number;
  log_estimated_usd: number;
  drift_wtu: number;
  drift_pct: number;
  status: 'synced' | 'minor_drift' | 'diverged' | 'unavailable' | string;
  last_synced_at?: string;
}

export interface ForecastDay {
  date: string;
  window_start: string;
  window_end: string;
  projected_wtu: number;
  included_wtu: number;
  payg_overage_wtu: number;
  included_cost_usd: number;
  payg_overage_usd: number;
  cumulative_pool_wtu: number;
  pool_limit_wtu: number;
  spillover_triggered: boolean;
}

export interface UsageForecast {
  horizon_days: number;
  daily_velocity_wtu: number;
  daily_pool_eligible_wtu?: number;
  daily_always_payg_wtu?: number;
  daily_velocity_usd: number;
  projected_30d_total_wtu: number;
  projected_30d_total_usd: number;
  projected_30d_included_usd: number;
  projected_30d_payg_usd: number;
  will_spillover_in_current_window: boolean;
  spillover_date?: string;
  days_until_spillover: number;
  forecast_days: ForecastDay[];
}

export interface WeeklyAllowancePool {
  window_start: string;
  window_end: string;
  reset_in_hours: number;
  reset_countdown: string;
  pool_mode: 'pooled' | 'per_user' | string;
  seat_basis: 'assigned' | 'purchased' | string;
  pool_limit_usd: number;
  pool_limit_wtu: number;
  included_used_usd: number;
  included_used_wtu: number;
  payg_overage_usd: number;
  payg_overage_wtu: number;
  remaining_usd: number;
  remaining_wtu: number;
  utilization_pct: number;
  windows?: WeeklyPoolWindow[];
  scenarios?: Record<string, PoolWhatIfScenario>;
  monitoring_meter?: MonitoringQuotaMeter;
}

export interface OverviewMetrics {
  monitored_projects?: string[];
  selected_project?: string;
  total_billed_cost: number;
  total_wtu?: number;
  included_cost?: number;
  payg_cost?: number;
  total_tokens: number;
  total_cached_tokens?: number;
  cache_hit_rate?: number;
  active_users_count: number;
  seat_quota: number;
  assigned_seats: number;
  utilization_pct: number;
  estimated_monthly_savings: number;
  daily_trends: DailyTrend[];
  model_breakdown: Record<string, ModelDetail>;
  token_cost_shares?: TokenCostShare[];
  model_effort_stats?: ModelEffortStats[];
  client_breakdown?: ClientStats[];
  weekly_pool?: WeeklyAllowancePool;
  forecast?: UsageForecast;
  currency: string;
}

export interface AllocatedUserCost {
  project_id?: string;
  user_id: string;
  model: string;
  usage_date: string;
  user_tokens: number;
  cached_tokens?: number;
  cache_hit_rate?: number;
  total_model_tokens: number;
  token_share: number;
  allocated_cost: number;
  wtu?: number;
  included_cost?: number;
  payg_cost?: number;
  fair_share_included_cost?: number;
  fair_share_payg_cost?: number;
  when_used_included_cost?: number;
  when_used_payg_cost?: number;
  currency: string;
}

export interface LicenseSeat {
  project_id?: string;
  user_id: string;
  assigned?: boolean;
  status: 'active' | 'at_risk' | 'dormant';
  tier?: 'standard' | 'plus' | string;
  weekly_allowance_usd?: number;
  weekly_allowance_wtu?: number;
  location?: 'global' | 'eu' | 'us' | string;
  is_trial?: boolean;
  trial_ends_at?: string;
  last_activity: string;
  total_tokens_in_window: number;
}

export interface LicenseGovernance {
  monitored_projects?: string[];
  selected_project?: string;
  seat_quota: number;
  assigned_seats: number;
  active_seats: number;
  dormant_seats: number;
  standard_seats?: number;
  plus_seats?: number;
  trial_seats?: number;
  locations?: string[];
  utilization_pct: number;
  estimated_monthly_savings: number;
  dormant_users: LicenseSeat[];
  all_seats?: LicenseSeat[];
  weekly_pool?: WeeklyAllowancePool;
  currency?: string;
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

export interface LeaderboardEntry {
  rank: number;
  project_id?: string;
  user_id: string;
  total_cost: number;
  included_cost?: number;
  payg_cost?: number;
  fair_share_included_cost?: number;
  fair_share_payg_cost?: number;
  when_used_included_cost?: number;
  when_used_payg_cost?: number;
  wtu?: number;
  spend_share_pct: number;
  share_of_spend?: number;
  total_tokens: number;
  cached_tokens?: number;
  cache_hit_rate?: number;
  token_share_pct: number;
  share_of_tokens?: number;
  active_days: number;
  cost_per_1k_tokens: number;
  top_model: string;
  top_effort?: string;
  model_breakdown: Record<string, ModelDetail | number>;
  currency: string;
}

export const normalizeLeaderboardEntry = (raw: any, idx = 0): LeaderboardEntry => {
  const totalTokens = Number(raw?.total_tokens ?? 0);
  const totalCost = Number(raw?.total_cost ?? 0);
  const spendShare = Number(raw?.spend_share_pct ?? raw?.share_of_spend ?? 0);
  const tokenShare = Number(raw?.token_share_pct ?? raw?.share_of_tokens ?? 0);
  const cachedTokens = Number(raw?.cached_tokens ?? 0);
  const cacheHitRate =
    raw?.cache_hit_rate !== undefined
      ? Number(raw.cache_hit_rate)
      : totalTokens > 0
      ? Number(((cachedTokens / totalTokens) * 100).toFixed(1))
      : 0;

  const normalizedBreakdown: Record<string, ModelDetail> = {};
  if (raw?.model_breakdown && typeof raw.model_breakdown === 'object') {
    for (const [model, val] of Object.entries(raw.model_breakdown)) {
      if (typeof val === 'number') {
        // Go backend sends percentage 0..100 (e.g., 23.2 for 23.2%)
        const ratio = val > 1 ? val / 100 : val;
        normalizedBreakdown[model] = {
          tokens: Math.round(totalTokens * ratio),
          cost: Number((totalCost * ratio).toFixed(2)),
          share: ratio,
        };
      } else if (val && typeof val === 'object') {
        const md = val as Partial<ModelDetail>;
        const rawShare = Number(md.share ?? 0);
        const ratio = rawShare > 1 ? rawShare / 100 : rawShare;
        normalizedBreakdown[model] = {
          tokens: Number(md.tokens ?? Math.round(totalTokens * ratio)),
          cost: Number(md.cost ?? Number((totalCost * ratio).toFixed(2))),
          share: ratio,
        };
      }
    }
  }

  const includedCost = Number(raw?.included_cost ?? 0);
  const paygCost = Number(raw?.payg_cost ?? 0);

  return {
    rank: Number(raw?.rank ?? idx + 1),
    user_id: String(raw?.user_id ?? 'unknown'),
    total_cost: totalCost,
    included_cost: includedCost,
    payg_cost: paygCost,
    fair_share_included_cost: Number(raw?.fair_share_included_cost ?? includedCost),
    fair_share_payg_cost: Number(raw?.fair_share_payg_cost ?? paygCost),
    when_used_included_cost: Number(raw?.when_used_included_cost ?? includedCost),
    when_used_payg_cost: Number(raw?.when_used_payg_cost ?? paygCost),
    wtu: Number(raw?.wtu ?? 0),
    spend_share_pct: spendShare,
    share_of_spend: spendShare,
    total_tokens: totalTokens,
    cached_tokens: cachedTokens,
    cache_hit_rate: cacheHitRate,
    token_share_pct: tokenShare,
    share_of_tokens: tokenShare,
    active_days: Number(raw?.active_days ?? 0),
    cost_per_1k_tokens: Number(raw?.cost_per_1k_tokens ?? 0),
    top_model: String(raw?.top_model ?? ''),
    top_effort: String(raw?.top_effort ?? ''),
    model_breakdown: normalizedBreakdown,
    currency: String(raw?.currency || 'USD'),
  };
};

export const aggregateLeaderboardFromCosts = (costs: AllocatedUserCost[]): LeaderboardEntry[] => {
  if (!costs || costs.length === 0) return [];
  const byUser = new Map<
    string,
    {
      totalCost: number;
      includedCost: number;
      paygCost: number;
      fairShareIncludedCost: number;
      fairSharePaygCost: number;
      whenUsedIncludedCost: number;
      whenUsedPaygCost: number;
      wtu: number;
      totalTokens: number;
      cachedTokens: number;
      days: Set<string>;
      models: Record<string, ModelDetail>;
      currency: string;
    }
  >();
  let orgCost = 0;
  let orgTokens = 0;

  for (const c of costs) {
    orgCost += c.allocated_cost;
    orgTokens += c.user_tokens;
    let u = byUser.get(c.user_id);
    if (!u) {
      u = {
        totalCost: 0,
        includedCost: 0,
        paygCost: 0,
        fairShareIncludedCost: 0,
        fairSharePaygCost: 0,
        whenUsedIncludedCost: 0,
        whenUsedPaygCost: 0,
        wtu: 0,
        totalTokens: 0,
        cachedTokens: 0,
        days: new Set<string>(),
        models: {},
        currency: c.currency || 'USD',
      };
      byUser.set(c.user_id, u);
    }
    const inc = c.included_cost ?? 0;
    const payg = c.payg_cost ?? 0;
    u.totalCost += c.allocated_cost;
    u.includedCost += inc;
    u.paygCost += payg;
    u.fairShareIncludedCost += c.fair_share_included_cost ?? inc;
    u.fairSharePaygCost += c.fair_share_payg_cost ?? payg;
    u.whenUsedIncludedCost += c.when_used_included_cost ?? inc;
    u.whenUsedPaygCost += c.when_used_payg_cost ?? payg;
    u.wtu += c.wtu ?? 0;
    u.totalTokens += c.user_tokens;
    u.cachedTokens += c.cached_tokens ?? 0;
    if (c.usage_date) u.days.add(c.usage_date);
    const md = u.models[c.model] || { tokens: 0, cost: 0, share: 0 };
    md.tokens += c.user_tokens;
    md.cost += c.allocated_cost;
    u.models[c.model] = md;
  }

  const entries: LeaderboardEntry[] = [];
  for (const [userId, u] of byUser.entries()) {
    let topModel = '';
    let topCost = -1;
    let topTokens = -1;
    for (const [m, md] of Object.entries(u.models)) {
      md.share = u.totalTokens > 0 ? md.tokens / u.totalTokens : 0;
      if (md.cost > topCost || (md.cost === topCost && md.tokens > topTokens)) {
        topCost = md.cost;
        topTokens = md.tokens;
        topModel = m;
      }
    }
    const spendShare = orgCost > 0 ? Number(((u.totalCost / orgCost) * 100).toFixed(2)) : 0;
    const tokenShare = orgTokens > 0 ? Number(((u.totalTokens / orgTokens) * 100).toFixed(2)) : 0;
    const cacheHitRate =
      u.totalTokens > 0 ? Number(((u.cachedTokens / u.totalTokens) * 100).toFixed(1)) : 0;
    entries.push({
      rank: 0,
      user_id: userId,
      total_cost: Number(u.totalCost.toFixed(2)),
      included_cost: Number(u.includedCost.toFixed(2)),
      payg_cost: Number(u.paygCost.toFixed(2)),
      fair_share_included_cost: Number(u.fairShareIncludedCost.toFixed(2)),
      fair_share_payg_cost: Number(u.fairSharePaygCost.toFixed(2)),
      when_used_included_cost: Number(u.whenUsedIncludedCost.toFixed(2)),
      when_used_payg_cost: Number(u.whenUsedPaygCost.toFixed(2)),
      wtu: u.wtu,
      spend_share_pct: spendShare,
      share_of_spend: spendShare,
      total_tokens: u.totalTokens,
      cached_tokens: u.cachedTokens,
      cache_hit_rate: cacheHitRate,
      token_share_pct: tokenShare,
      share_of_tokens: tokenShare,
      active_days: u.days.size,
      cost_per_1k_tokens:
        u.totalTokens > 0 ? Number(((u.totalCost * 1000) / u.totalTokens).toFixed(4)) : 0,
      top_model: topModel,
      model_breakdown: u.models,
      currency: u.currency,
    });
  }

  entries.sort((a, b) => {
    if (b.total_cost !== a.total_cost) return b.total_cost - a.total_cost;
    if (b.total_tokens !== a.total_tokens) return b.total_tokens - a.total_tokens;
    return a.user_id.localeCompare(b.user_id);
  });

  entries.forEach((e, idx) => {
    e.rank = idx + 1;
  });

  return entries;
};

export const fetchLeaderboard = async (
  days = 30,
  paygMode?: 'fair_share' | 'when_used',
  project?: string
): Promise<LeaderboardEntry[]> => {
  const projectParam = project && project !== 'all' ? `&project=${encodeURIComponent(project)}` : '';
  const url = `/api/v1/leaderboard?days=${days}${paygMode ? `&payg_mode=${encodeURIComponent(paygMode)}` : ''}${projectParam}`;
  const res = await fetch(url);
  if (!res.ok) throw new Error(`HTTP ${res.status}: Failed to fetch developer leaderboard`);
  const data = await res.json();
  if (!Array.isArray(data)) return [];
  return data.map((item, idx) => normalizeLeaderboardEntry(item, idx));
};

export const fetchOverview = async (days = 30, project?: string): Promise<OverviewMetrics> => {
  const projectParam = project && project !== 'all' ? `&project=${encodeURIComponent(project)}` : '';
  const res = await fetch(`/api/v1/metrics/overview?days=${days}${projectParam}`);
  if (!res.ok) throw new Error(`HTTP ${res.status}: Failed to fetch overview metrics`);
  return res.json();
};

export const fetchAttributedCosts = async (
  days = 30,
  model = '',
  paygMode?: 'fair_share' | 'when_used',
  project?: string
): Promise<AllocatedUserCost[]> => {
  const projectParam = project && project !== 'all' ? `&project=${encodeURIComponent(project)}` : '';
  const url = `/api/v1/costs/users?days=${days}${model ? `&model=${encodeURIComponent(model)}` : ''}${
    paygMode ? `&payg_mode=${encodeURIComponent(paygMode)}` : ''
  }${projectParam}`;
  const res = await fetch(url);
  if (!res.ok) throw new Error(`HTTP ${res.status}: Failed to fetch attributed costs`);
  return res.json();
};

export const fetchLicenseGovernance = async (days = 30, project?: string): Promise<LicenseGovernance> => {
  const projectParam = project && project !== 'all' ? `&project=${encodeURIComponent(project)}` : '';
  const res = await fetch(`/api/v1/licenses/status?days=${days}${projectParam}`);
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

export interface AppVersionInfo {
  version: string;
  commit?: string;
  build_date?: string;
}

export const fetchVersion = async (): Promise<AppVersionInfo> => {
  const res = await fetch('/api/v1/version');
  if (!res.ok) throw new Error(`HTTP ${res.status}: Failed to fetch version`);
  return res.json();
};

export interface McpStatusInfo {
  name: string;
  version: string;
  protocol_version: string;
  mcp_enabled: boolean;
  endpoint: string;
  tools: string[];
}

export const fetchMcpStatus = async (): Promise<McpStatusInfo> => {
  const res = await fetch('/api/v1/mcp/status');
  if (!res.ok) throw new Error(`HTTP ${res.status}: Failed to fetch MCP status`);
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
  by_model?: Record<string, number>;
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
  included_cost?: number;
  payg_cost?: number;
  total_wtu?: number;
  daily_burn_rate: number;
  weekly_burn_rate: number;
  total_user_tokens: number;
  cached_tokens?: number;
  cache_hit_rate?: number;
  top_effort?: string;
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
    by_model: t.by_model || {},
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
      included_cost: raw.included_cost ?? 0,
      payg_cost: raw.payg_cost ?? 0,
      total_wtu: raw.total_wtu ?? 0,
      daily_burn_rate: raw.daily_burn_rate ?? 0,
      weekly_burn_rate: raw.weekly_burn_rate ?? 0,
      total_user_tokens: raw.total_tokens_in_window ?? 0,
      cached_tokens: raw.cached_tokens_in_window ?? 0,
      cache_hit_rate: raw.cache_hit_rate ?? 0,
      top_effort: raw.top_effort ?? '',
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


