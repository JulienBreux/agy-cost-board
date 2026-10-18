import React, { useState } from 'react';
import { TrendingUp, Calendar, AlertTriangle, CheckCircle2, RefreshCw } from 'lucide-react';
import { UsageForecast } from '../api';
import { formatCurrency, getCurrencySymbol } from '../utils/currency';

interface Props {
  forecast?: UsageForecast;
  currency?: string;
}

const formatWTU = (wtu?: number): string => {
  const n = Number(wtu ?? 0);
  if (n >= 1_000_000) return `${(n / 1_000_000).toFixed(2)}M WTU`;
  if (n >= 1_000) return `${(n / 1_000).toFixed(1)}K WTU`;
  return `${n.toLocaleString()} WTU`;
};

export const ForecastCard: React.FC<Props> = ({ forecast, currency = 'USD' }) => {
  const [unitMode, setUnitMode] = useState<'usd' | 'wtu'>('usd');

  if (!forecast || !forecast.forecast_days || forecast.forecast_days.length === 0) {
    return null;
  }

  const horizonDays = Number(forecast.horizon_days ?? 30);
  const dailyVelocityWTU = Number(forecast.daily_velocity_wtu ?? 0);
  const dailyVelocityUSD = Number(forecast.daily_velocity_usd ?? 0);
  const projectedTotalWTU = Number(forecast.projected_30d_total_wtu ?? 0);
  const projectedTotalUSD = Number(forecast.projected_30d_total_usd ?? 0);
  const projectedIncludedUSD = Number(forecast.projected_30d_included_usd ?? 0);
  const projectedPAYGUSD = Number(forecast.projected_30d_payg_usd ?? 0);
  const spilloverDate = forecast.spillover_date || '';
  const daysUntilSpillover = Number(forecast.days_until_spillover ?? -1);
  const hasSpillover = Boolean(spilloverDate) && daysUntilSpillover >= 0;

  const days = forecast.forecast_days;
  const maxDayTotalUSD = Math.max(
    ...days.map((d) => Number(d.included_cost_usd ?? 0) + Number(d.payg_overage_usd ?? 0)),
    0.01
  );
  const maxDayTotalWTU = Math.max(
    ...days.map((d) => Number(d.included_wtu ?? 0) + Number(d.payg_overage_wtu ?? 0)),
    1
  );

  const projectedIncludedWTU = days.reduce((acc, d) => acc + Number(d.included_wtu ?? 0), 0);
  const projectedPAYGWTU = days.reduce((acc, d) => acc + Number(d.payg_overage_wtu ?? 0), 0);

  return (
    <div className="bg-slate-800/40 backdrop-blur-md border border-slate-700/60 rounded-2xl p-6 shadow-xl">
      <div className="flex flex-col lg:flex-row lg:items-center lg:justify-between gap-4 mb-6">
        <div className="flex items-start gap-3">
          <div className="p-2 rounded-lg bg-cyan-500/15 border border-cyan-500/30 text-cyan-400">
            <TrendingUp className="w-5 h-5" />
          </div>
          <div>
            <div className="flex items-center gap-2.5 flex-wrap">
              <h2 className="text-lg font-semibold text-white">
                {horizonDays}-Day Window-Aware Usage &amp; PAYG Forecast
              </h2>
              {hasSpillover ? (
                <span className="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-xs font-medium bg-amber-500/15 text-amber-300 border border-amber-500/30">
                  <AlertTriangle className="w-3.5 h-3.5 text-amber-400" />
                  Spillover on {spilloverDate} ({daysUntilSpillover === 0 ? 'today' : `in ${daysUntilSpillover}d`})
                </span>
              ) : (
                <span className="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-xs font-medium bg-emerald-500/15 text-emerald-300 border border-emerald-500/30">
                  <CheckCircle2 className="w-3.5 h-3.5 text-emerald-400" />
                  Within Weekly Pool
                </span>
              )}
            </div>
            <p className="text-xs text-slate-400 mt-0.5">
              Trailing 7-day WTU velocity replayed across upcoming 7-day allowance pool resets to predict PAYG overage dates
            </p>
          </div>
        </div>

        <div className="flex items-center gap-2 self-start lg:self-auto">
          <div className="flex items-center bg-slate-900/70 rounded-xl p-1 border border-slate-700/60">
            <button
              type="button"
              onClick={() => setUnitMode('usd')}
              className={`px-3 py-1 rounded-lg text-xs font-medium transition-colors ${
                unitMode === 'usd'
                  ? 'bg-cyan-600 text-white shadow-sm'
                  : 'text-slate-400 hover:text-slate-200'
              }`}
            >
              {getCurrencySymbol(currency)} {currency}
            </button>
            <button
              type="button"
              onClick={() => setUnitMode('wtu')}
              className={`px-3 py-1 rounded-lg text-xs font-medium transition-colors ${
                unitMode === 'wtu'
                  ? 'bg-cyan-600 text-white shadow-sm'
                  : 'text-slate-400 hover:text-slate-200'
              }`}
            >
              WTU
            </button>
          </div>
        </div>
      </div>

      {/* 4 Forecast KPI Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
        <div className="bg-slate-900/50 border border-slate-700/50 rounded-xl p-4">
          <div className="text-xs text-slate-400 mb-1">Projected {horizonDays}d Total</div>
          <div className="text-xl font-bold text-white font-mono">
            {unitMode === 'usd'
              ? formatCurrency(projectedTotalUSD, currency)
              : formatWTU(projectedTotalWTU)}
          </div>
          <div className="text-xs text-slate-400 font-mono mt-0.5">
            {unitMode === 'usd'
              ? formatWTU(projectedTotalWTU)
              : formatCurrency(projectedTotalUSD, currency)}
          </div>
        </div>

        <div className="bg-slate-900/50 border border-emerald-500/30 rounded-xl p-4">
          <div className="text-xs text-emerald-300/80 mb-1">Projected Included Pool</div>
          <div className="text-xl font-bold text-emerald-400 font-mono">
            {unitMode === 'usd'
              ? formatCurrency(projectedIncludedUSD, currency)
              : formatWTU(projectedIncludedWTU)}
          </div>
          <div className="text-xs text-slate-400 font-mono mt-0.5">
            {projectedTotalUSD > 0
              ? `${((projectedIncludedUSD / projectedTotalUSD) * 100).toFixed(1)}% covered by weekly pools`
              : '0.0% covered'}
          </div>
        </div>

        <div
          className={`bg-slate-900/50 border rounded-xl p-4 ${
            projectedPAYGUSD > 0 ? 'border-rose-500/40' : 'border-slate-700/50'
          }`}
        >
          <div className="text-xs text-rose-300/80 mb-1">Projected PAYG Overage</div>
          <div
            className={`text-xl font-bold font-mono ${
              projectedPAYGUSD > 0 ? 'text-rose-400' : 'text-slate-200'
            }`}
          >
            {unitMode === 'usd'
              ? formatCurrency(projectedPAYGUSD, currency)
              : formatWTU(projectedPAYGWTU)}
          </div>
          <div className="text-xs text-slate-400 font-mono mt-0.5">
            {unitMode === 'usd'
              ? formatWTU(projectedPAYGWTU)
              : formatCurrency(projectedPAYGUSD, currency)}
          </div>
        </div>

        <div className="bg-slate-900/50 border border-cyan-500/30 rounded-xl p-4">
          <div className="flex items-center justify-between text-xs text-cyan-300/80 mb-1">
            <span>Trailing 7d Velocity</span>
            <Calendar className="w-3.5 h-3.5 text-cyan-400" />
          </div>
          <div className="text-xl font-bold text-cyan-300 font-mono">
            {formatCurrency(dailyVelocityUSD, currency)}/day
          </div>
          <div className="text-xs text-slate-400 font-mono mt-0.5">
            {formatWTU(dailyVelocityWTU)}/day
          </div>
        </div>
      </div>

      {/* 30-Day Stacked Bar Chart with Weekly Pool Reset Boundaries */}
      <div className="bg-slate-900/40 border border-slate-700/50 rounded-xl p-4">
        <div className="flex flex-wrap items-center justify-between gap-2 mb-3">
          <div className="text-xs font-semibold uppercase tracking-wider text-slate-400">
            Daily Projected Drawdown &amp; Weekly Pool Resets ({horizonDays} Days)
          </div>
          <div className="flex items-center gap-4 text-xs text-slate-400">
            <span className="inline-flex items-center gap-1.5">
              <span className="w-2.5 h-2.5 rounded-sm bg-emerald-500 inline-block" />
              Included Allowance
            </span>
            <span className="inline-flex items-center gap-1.5">
              <span className="w-2.5 h-2.5 rounded-sm bg-rose-500 inline-block" />
              PAYG Spillover
            </span>
            <span className="inline-flex items-center gap-1.5 text-indigo-300">
              <RefreshCw className="w-3 h-3" />
              7d Pool Reset
            </span>
          </div>
        </div>

        <div className="h-36 flex items-end gap-1 pt-4 pb-1 px-1 overflow-x-auto">
          {days.map((d, idx) => {
            const isNewWindow = idx > 0 && d.window_start !== days[idx - 1].window_start;
            const incVal =
              unitMode === 'usd' ? Number(d.included_cost_usd ?? 0) : Number(d.included_wtu ?? 0);
            const paygVal =
              unitMode === 'usd' ? Number(d.payg_overage_usd ?? 0) : Number(d.payg_overage_wtu ?? 0);
            const maxVal = unitMode === 'usd' ? maxDayTotalUSD : maxDayTotalWTU;
            const incHeightPct = Math.min(100, Math.max(0, (incVal / maxVal) * 100));
            const paygHeightPct = Math.min(100 - incHeightPct, Math.max(0, (paygVal / maxVal) * 100));

            return (
              <React.Fragment key={`${d.date}-${idx}`}>
                {isNewWindow && (
                  <div
                    className="h-full border-l border-dashed border-indigo-400/60 mx-0.5 relative flex items-start"
                    title={`Weekly Pool Reset (${d.window_start})`}
                  />
                )}
                <div
                  className="flex-1 min-w-[8px] h-full flex flex-col justify-end group relative"
                  title={`${d.date} (${d.window_start} → ${d.window_end})\nIncluded: ${formatCurrency(
                    Number(d.included_cost_usd ?? 0),
                    currency
                  )} (${formatWTU(d.included_wtu)})\nPAYG: ${formatCurrency(
                    Number(d.payg_overage_usd ?? 0),
                    currency
                  )} (${formatWTU(d.payg_overage_wtu)})`}
                >
                  {paygHeightPct > 0 && (
                    <div
                      className="w-full bg-rose-500/90 group-hover:bg-rose-400 rounded-t-sm transition-all"
                      style={{ height: `${Math.max(4, paygHeightPct)}%` }}
                    />
                  )}
                  {incHeightPct > 0 && (
                    <div
                      className={`w-full bg-emerald-500/85 group-hover:bg-emerald-400 transition-all ${
                        paygHeightPct === 0 ? 'rounded-t-sm' : ''
                      }`}
                      style={{ height: `${Math.max(4, incHeightPct)}%` }}
                    />
                  )}
                </div>
              </React.Fragment>
            );
          })}
        </div>

        <div className="flex items-center justify-between text-[11px] text-slate-400 font-mono mt-2 pt-2 border-t border-slate-800">
          <span>{days[0]?.date}</span>
          <span>Day +15 ({days[Math.min(14, days.length - 1)]?.date})</span>
          <span>{days[days.length - 1]?.date}</span>
        </div>
      </div>
    </div>
  );
};
