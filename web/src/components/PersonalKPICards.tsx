import React from 'react';
import { DollarSign, Flame, Cpu, PieChart } from 'lucide-react';
import { UserConsumptionKPIs } from '../api';

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
  const currencySymbol = currency === 'EUR' ? '€' : currency === 'USD' ? '$' : currency;

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
            {currencySymbol}
            {(safeKpis.mtd_spend || 0).toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}
          </div>
          <div className="mt-1 flex items-center text-xs text-google-gray-400">
            <span>
              {currencySymbol}{(safeKpis.daily_burn_rate || 0).toFixed(2)}/day average velocity
            </span>
          </div>
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
            {currencySymbol}
            {(safeKpis.weekly_burn_rate || 0).toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}
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
