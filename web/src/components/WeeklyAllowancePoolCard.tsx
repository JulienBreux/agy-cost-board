import React, { useState } from 'react';
import { Layers, Clock, Sliders, AlertTriangle, CheckCircle2 } from 'lucide-react';
import { WeeklyAllowancePool } from '../api';
import { formatCurrency } from '../utils/currency';

interface Props {
  pool?: WeeklyAllowancePool;
  currency?: string;
}

const formatWTU = (wtu?: number): string => {
  const n = Number(wtu ?? 0);
  if (n >= 1_000_000) return `${(n / 1_000_000).toFixed(2)}M WTU`;
  if (n >= 1_000) return `${(n / 1_000).toFixed(1)}K WTU`;
  return `${n.toLocaleString()} WTU`;
};

export const WeeklyAllowancePoolCard: React.FC<Props> = ({ pool, currency = 'USD' }) => {
  const [poolMode, setPoolMode] = useState<'pooled' | 'per_user'>(
    (pool?.pool_mode as 'pooled' | 'per_user') || 'pooled'
  );
  const [seatBasis, setSeatBasis] = useState<'assigned' | 'purchased'>(
    (pool?.seat_basis as 'assigned' | 'purchased') || 'assigned'
  );

  if (!pool) return null;

  const scenarioKey = `${poolMode}_${seatBasis}`;
  const activeScenario = pool.scenarios?.[scenarioKey];

  const poolLimitUSD = Number(activeScenario?.pool_limit_usd ?? pool.pool_limit_usd ?? 0);
  const poolLimitWTU = Number(activeScenario?.pool_limit_wtu ?? pool.pool_limit_wtu ?? 0);
  const includedUsedUSD = Number(activeScenario?.included_used_usd ?? pool.included_used_usd ?? 0);
  const includedUsedWTU = Number(activeScenario?.included_used_wtu ?? pool.included_used_wtu ?? 0);
  const paygOverageUSD = Number(activeScenario?.payg_overage_usd ?? pool.payg_overage_usd ?? 0);
  const paygOverageWTU = Number(activeScenario?.payg_overage_wtu ?? pool.payg_overage_wtu ?? 0);
  const remainingUSD = Number(activeScenario?.remaining_usd ?? pool.remaining_usd ?? 0);
  const remainingWTU = Number(activeScenario?.remaining_wtu ?? pool.remaining_wtu ?? 0);
  const utilizationPct = Number(activeScenario?.utilization_pct ?? pool.utilization_pct ?? 0);
  const windows = activeScenario?.windows ?? pool.windows ?? [];

  const clampedPct = Math.min(100, Math.max(0, utilizationPct));
  const hasOverage = paygOverageUSD > 0;

  return (
    <div className="bg-slate-800/40 backdrop-blur-md border border-slate-700/60 rounded-2xl p-6 shadow-xl">
      <div className="flex flex-col lg:flex-row lg:items-center lg:justify-between gap-4 mb-6">
        <div>
          <div className="flex items-center gap-2.5">
            <div className="p-2 rounded-lg bg-indigo-500/15 border border-indigo-500/30 text-indigo-400">
              <Layers className="w-5 h-5" />
            </div>
            <div>
              <div className="flex items-center gap-2.5 flex-wrap">
                <h2 className="text-lg font-semibold text-white">Weekly Allowance Pool</h2>
                <span className="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-xs font-medium bg-indigo-500/15 text-indigo-300 border border-indigo-500/30">
                  <Clock className="w-3.5 h-3.5" />
                  Resets in {pool.reset_countdown || `${pool.reset_in_hours ?? 0}h`}
                </span>
                <span className="text-xs text-slate-400 font-mono">
                  ({pool.window_start} → {pool.window_end})
                </span>
              </div>
              <p className="text-xs text-slate-400 mt-0.5">
                Gemini Enterprise 7-day floating allowance pool ({formatCurrency(2.5, currency)}/wk Standard · {formatCurrency(3.75, currency)}/wk Plus) &amp; What-If simulator
              </p>
            </div>
          </div>
        </div>

        {/* Interactive What-If Simulator Toggles */}
        <div className="flex flex-wrap items-center gap-2 bg-slate-900/70 p-1.5 rounded-xl border border-slate-700/60">
          <div className="flex items-center gap-1 px-2 text-xs text-slate-400 font-medium">
            <Sliders className="w-3.5 h-3.5 text-indigo-400" />
            <span>What-If:</span>
          </div>
          <div className="flex items-center bg-slate-800/80 rounded-lg p-0.5 border border-slate-700/50">
            <button
              type="button"
              onClick={() => setPoolMode('pooled')}
              className={`px-2.5 py-1 rounded-md text-xs font-medium transition-colors ${
                poolMode === 'pooled'
                  ? 'bg-indigo-600 text-white shadow-sm'
                  : 'text-slate-400 hover:text-slate-200'
              }`}
            >
              Pooled (Org)
            </button>
            <button
              type="button"
              onClick={() => setPoolMode('per_user')}
              className={`px-2.5 py-1 rounded-md text-xs font-medium transition-colors ${
                poolMode === 'per_user'
                  ? 'bg-indigo-600 text-white shadow-sm'
                  : 'text-slate-400 hover:text-slate-200'
              }`}
            >
              Per-User Cap
            </button>
          </div>

          <div className="flex items-center bg-slate-800/80 rounded-lg p-0.5 border border-slate-700/50">
            <button
              type="button"
              onClick={() => setSeatBasis('assigned')}
              className={`px-2.5 py-1 rounded-md text-xs font-medium transition-colors ${
                seatBasis === 'assigned'
                  ? 'bg-emerald-600 text-white shadow-sm'
                  : 'text-slate-400 hover:text-slate-200'
              }`}
            >
              Assigned Seats
            </button>
            <button
              type="button"
              onClick={() => setSeatBasis('purchased')}
              className={`px-2.5 py-1 rounded-md text-xs font-medium transition-colors ${
                seatBasis === 'purchased'
                  ? 'bg-emerald-600 text-white shadow-sm'
                  : 'text-slate-400 hover:text-slate-200'
              }`}
            >
              Purchased Quota
            </button>
          </div>
        </div>
      </div>

      {/* Drawdown Gauge Bar */}
      <div className="mb-6">
        <div className="flex items-center justify-between text-xs mb-2">
          <span className="text-slate-300 font-medium">
            7-Day Pool Drawdown ({poolMode === 'pooled' ? 'Shared Org Pool' : 'Individual Seat Caps'} ·{' '}
            {seatBasis === 'assigned' ? 'Assigned Seats' : 'Purchased Quota'})
          </span>
          <span className="font-mono font-semibold text-white">
            {utilizationPct.toFixed(1)}%
          </span>
        </div>
        <div className="w-full h-3.5 bg-slate-900/90 rounded-full overflow-hidden border border-slate-700/60 flex">
          <div
            className={`h-full transition-all duration-300 ${
              utilizationPct >= 95
                ? 'bg-gradient-to-r from-emerald-500 via-amber-500 to-rose-500'
                : utilizationPct >= 75
                ? 'bg-gradient-to-r from-emerald-500 to-amber-400'
                : 'bg-gradient-to-r from-emerald-500 to-indigo-500'
            }`}
            style={{ width: `${clampedPct}%` }}
          />
        </div>
      </div>

      {/* 4 Metric Mini-Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
        <div className="bg-slate-900/50 border border-slate-700/50 rounded-xl p-4">
          <div className="text-xs text-slate-400 mb-1">Weekly Pool Limit</div>
          <div className="text-xl font-bold text-white font-mono">
            {formatCurrency(poolLimitUSD, currency)}
          </div>
          <div className="text-xs text-slate-400 font-mono mt-0.5">{formatWTU(poolLimitWTU)}</div>
        </div>

        <div className="bg-slate-900/50 border border-emerald-500/30 rounded-xl p-4">
          <div className="text-xs text-emerald-300/80 mb-1">Included Allowance Used</div>
          <div className="text-xl font-bold text-emerald-400 font-mono">
            {formatCurrency(includedUsedUSD, currency)}
          </div>
          <div className="text-xs text-slate-400 font-mono mt-0.5">{formatWTU(includedUsedWTU)}</div>
        </div>

        <div className="bg-slate-900/50 border border-indigo-500/30 rounded-xl p-4">
          <div className="text-xs text-indigo-300/80 mb-1">Remaining Pool Allowance</div>
          <div className="text-xl font-bold text-indigo-300 font-mono">
            {formatCurrency(remainingUSD, currency)}
          </div>
          <div className="text-xs text-slate-400 font-mono mt-0.5">{formatWTU(remainingWTU)}</div>
        </div>

        <div
          className={`bg-slate-900/50 border rounded-xl p-4 ${
            hasOverage ? 'border-rose-500/40' : 'border-slate-700/50'
          }`}
        >
          <div className="flex items-center justify-between text-xs mb-1">
            <span className={hasOverage ? 'text-rose-300' : 'text-slate-400'}>
              7-Day Pool PAYG Overage
            </span>
            {hasOverage ? (
              <AlertTriangle className="w-3.5 h-3.5 text-rose-400" />
            ) : (
              <CheckCircle2 className="w-3.5 h-3.5 text-emerald-400" />
            )}
          </div>
          <div
            className={`text-xl font-bold font-mono ${
              hasOverage ? 'text-rose-400' : 'text-slate-200'
            }`}
          >
            {formatCurrency(paygOverageUSD, currency)}
          </div>
          <div className="text-xs text-slate-400 font-mono mt-0.5">{formatWTU(paygOverageWTU)}</div>
        </div>
      </div>

      {/* Cloud Monitoring Quota Meter vs. Log Estimate Reconciliation Banner */}
      {pool.monitoring_meter && (
        <div className="mb-6 bg-slate-900/60 border border-slate-700/60 rounded-xl p-3.5 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3">
          <div className="flex items-start sm:items-center gap-2.5">
            <span
              className={`inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-[11px] font-semibold uppercase tracking-wide border ${
                pool.monitoring_meter.status === 'synced'
                  ? 'bg-emerald-500/15 text-emerald-300 border-emerald-500/30'
                  : pool.monitoring_meter.status === 'minor_drift'
                  ? 'bg-amber-500/15 text-amber-300 border-amber-500/30'
                  : 'bg-rose-500/15 text-rose-300 border-rose-500/30'
              }`}
            >
              {pool.monitoring_meter.status === 'synced'
                ? 'Synced'
                : pool.monitoring_meter.status === 'minor_drift'
                ? 'Minor Drift'
                : 'Diverged'}
            </span>
            <div>
              <div className="text-xs font-medium text-slate-200">
                Cloud Monitoring Quota Meter vs. Log Estimate
              </div>
              <div className="text-[11px] text-slate-400 font-mono truncate max-w-xl">
                {pool.monitoring_meter.metric_type}
              </div>
            </div>
          </div>

          <div className="flex flex-wrap items-center gap-4 text-xs font-mono">
            <div>
              <span className="text-slate-400">Metered: </span>
              <span className="text-white font-semibold">
                {formatWTU(pool.monitoring_meter.metered_used_wtu)}
              </span>{' '}
              <span className="text-slate-400">
                ({formatCurrency(Number(pool.monitoring_meter.metered_used_usd ?? 0), currency)})
              </span>
            </div>
            <div>
              <span className="text-slate-400">Log Est.: </span>
              <span className="text-slate-200">
                {formatWTU(pool.monitoring_meter.log_estimated_wtu)}
              </span>
            </div>
            <div>
              <span className="text-slate-400">Drift: </span>
              <span
                className={`font-semibold ${
                  Math.abs(Number(pool.monitoring_meter.drift_pct ?? 0)) <= 5
                    ? 'text-emerald-400'
                    : Math.abs(Number(pool.monitoring_meter.drift_pct ?? 0)) <= 15
                    ? 'text-amber-400'
                    : 'text-rose-400'
                }`}
              >
                {Number(pool.monitoring_meter.drift_pct ?? 0) >= 0 ? '+' : ''}
                {Number(pool.monitoring_meter.drift_pct ?? 0).toFixed(2)}%
              </span>
            </div>
          </div>
        </div>
      )}

      {/* Rolling 7-Day Window History */}
      {windows.length > 0 && (
        <div className="border-t border-slate-700/50 pt-4">
          <div className="text-xs font-semibold uppercase tracking-wider text-slate-400 mb-3">
            Rolling 7-Day Pool Windows
          </div>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-3">
            {windows.slice(-4).map((w) => {
              const wPct = Math.min(100, Math.max(0, Number(w.utilization_pct ?? 0)));
              const wOverage = Number(w.payg_overage_usd ?? 0);
              return (
                <div
                  key={w.window_start}
                  className="bg-slate-900/40 border border-slate-700/40 rounded-lg p-3"
                >
                  <div className="flex items-center justify-between text-xs mb-1.5">
                    <span className="text-slate-300 font-medium">{w.window_label}</span>
                    <span className="font-mono text-slate-200">
                      {Number(w.utilization_pct ?? 0).toFixed(1)}%
                    </span>
                  </div>
                  <div className="w-full h-2 bg-slate-800 rounded-full overflow-hidden mb-2">
                    <div
                      className={`h-full ${
                        wOverage > 0 ? 'bg-rose-500' : 'bg-emerald-500'
                      }`}
                      style={{ width: `${wPct}%` }}
                    />
                  </div>
                  <div className="flex items-center justify-between text-[11px] text-slate-400 font-mono">
                    <span>
                      Used: {formatCurrency(Number(w.included_used_usd ?? 0), currency)} /{' '}
                      {formatCurrency(Number(w.pool_limit_usd ?? 0), currency)}
                    </span>
                    {wOverage > 0 && (
                      <span className="text-rose-400">
                        +{formatCurrency(wOverage, currency)} PAYG
                      </span>
                    )}
                  </div>
                </div>
              );
            })}
          </div>
        </div>
      )}
    </div>
  );
};
