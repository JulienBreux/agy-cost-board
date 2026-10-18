import React from 'react';
import { DollarSign, Flame, Cpu, PieChart } from 'lucide-react';
import { UserConsumptionKPIs } from '../api';
import { formatCurrency } from '../utils/currency';

interface PersonalKPICardsProps {
  kpis?: UserConsumptionKPIs;
  currency?: string;
}

export const PersonalKPICards: React.FC<PersonalKPICardsProps> = ({ kpis, currency = 'USD' }) => {
  const safeKpis = kpis || {
    mtd_spend: 0,
    daily_burn_rate: 0,
    weekly_burn_rate: 0,
    total_user_tokens: 0,
    org_spend_share_percent: 0,
  };

  return (
    <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
      {/* Month-to-Date Spend */}
      <div className="bg-[#14171f] border border-[#222836] rounded-xl p-5 relative overflow-hidden shadow-sm">
        <div className="flex items-center justify-between">
          <span className="text-xs font-medium uppercase tracking-wider text-google-gray-500">
            Month-to-Date Spend
          </span>
          <div className="p-2 bg-google-blue/10 text-google-blue rounded-lg border border-google-blue/20">
            <DollarSign className="h-4 w-4" />
          </div>
        </div>
        <div className="mt-3">
          <div className="text-2xl sm:text-3xl font-bold text-white tracking-tight">
            {formatCurrency(safeKpis.mtd_spend, currency)}
          </div>
          <div className="mt-1 flex items-center text-xs text-google-gray-400">
            <span>
              {formatCurrency(safeKpis.daily_burn_rate, currency)}/day average velocity
            </span>
          </div>
          {((safeKpis.included_cost ?? 0) > 0 || (safeKpis.payg_cost ?? 0) > 0) && (
            <div className="mt-2 flex flex-wrap items-center gap-1.5 text-[11px] font-mono">
              <span className="px-1.5 py-0.5 rounded bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
                Included: {formatCurrency(safeKpis.included_cost ?? 0, currency)}
              </span>
              <span className="px-1.5 py-0.5 rounded bg-amber-500/10 text-amber-400 border border-amber-500/20">
                PAYG: {formatCurrency(safeKpis.payg_cost ?? 0, currency)}
              </span>
            </div>
          )}
        </div>
      </div>

      {/* Weekly Burn Rate */}
      <div className="bg-[#14171f] border border-[#222836] rounded-xl p-5 relative overflow-hidden shadow-sm">
        <div className="flex items-center justify-between">
          <span className="text-xs font-medium uppercase tracking-wider text-google-gray-500">
            Weekly Burn Rate
          </span>
          <div className="p-2 bg-amber-500/10 text-amber-400 rounded-lg border border-amber-500/20">
            <Flame className="h-4 w-4" />
          </div>
        </div>
        <div className="mt-3">
          <div className="text-2xl sm:text-3xl font-bold text-amber-400 tracking-tight">
            {formatCurrency(safeKpis.weekly_burn_rate, currency)}
          </div>
          <div className="mt-1 flex items-center text-xs text-google-gray-400">
            <span>Last 7 days run rate</span>
          </div>
        </div>
      </div>

      {/* Personal Inference Tokens */}
      <div className="bg-[#14171f] border border-[#222836] rounded-xl p-5 relative overflow-hidden shadow-sm">
        <div className="flex items-center justify-between">
          <span className="text-xs font-medium uppercase tracking-wider text-google-gray-500">
            Personal Inference Tokens
          </span>
          <div className="p-2 bg-indigo-500/10 text-indigo-400 rounded-lg border border-indigo-500/20">
            <Cpu className="h-4 w-4" />
          </div>
        </div>
        <div className="mt-3">
          <div className="text-2xl sm:text-3xl font-bold text-white tracking-tight">
            {((safeKpis.total_user_tokens || 0) / 1_000_000).toFixed(2)}M
          </div>
          <div className="mt-1 flex items-center text-xs text-google-gray-400">
            <span>{(safeKpis.total_user_tokens || 0).toLocaleString()} tokens total</span>
          </div>
          {((safeKpis.total_wtu ?? 0) > 0 || (safeKpis.cache_hit_rate ?? 0) > 0) && (
            <div className="mt-2 flex flex-wrap items-center gap-1.5 text-[11px] font-mono">
              {(safeKpis.total_wtu ?? 0) > 0 && (
                <span className="px-1.5 py-0.5 rounded bg-indigo-500/10 text-indigo-400 border border-indigo-500/20">
                  {((safeKpis.total_wtu ?? 0) / 1_000_000).toFixed(2)}M WTU
                </span>
              )}
              {(safeKpis.cache_hit_rate ?? 0) > 0 && (
                <span className="px-1.5 py-0.5 rounded bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
                  {Number(safeKpis.cache_hit_rate).toFixed(1)}% cache
                </span>
              )}
            </div>
          )}
        </div>
      </div>

      {/* Org Spend Share */}
      <div className="bg-[#14171f] border border-[#222836] rounded-xl p-5 relative overflow-hidden shadow-sm">
        <div className="flex items-center justify-between">
          <span className="text-xs font-medium uppercase tracking-wider text-google-gray-500">
            Org Spend Share
          </span>
          <div className="p-2 bg-google-green/10 text-google-green rounded-lg border border-google-green/20">
            <PieChart className="h-4 w-4" />
          </div>
        </div>
        <div className="mt-3">
          <div className="text-2xl sm:text-3xl font-bold text-white tracking-tight">
            {(safeKpis.org_spend_share_percent || 0).toFixed(1)}%
          </div>
          <div className="mt-1 flex items-center text-xs text-google-gray-400">
            <span>Share of company AI spend</span>
          </div>
        </div>
      </div>
    </div>
  );
};
