import React from 'react';
import { DollarSign, Cpu, Users, TrendingDown } from 'lucide-react';
import { OverviewMetrics } from '../api';

interface KPICardsProps {
  metrics: OverviewMetrics;
}

export const KPICards: React.FC<KPICardsProps> = ({ metrics }) => {
  return (
    <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
      {/* Total Billed Spend */}
      <div className="bg-[#14171f] border border-[#222836] rounded-xl p-5 relative overflow-hidden shadow-sm">
        <div className="flex items-center justify-between">
          <span className="text-xs font-medium uppercase tracking-wider text-google-gray-500">Total Billed Cost</span>
          <div className="p-2 bg-google-blue/10 text-google-blue rounded-lg border border-google-blue/20">
            <DollarSign className="h-4 w-4" />
          </div>
        </div>
        <div className="mt-3">
          <div className="text-2xl sm:text-3xl font-bold text-white tracking-tight">
            ${metrics.total_billed_cost.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}
          </div>
          <div className="mt-1 flex items-center text-xs text-google-gray-400">
            <span>Reconciled with Cloud Billing</span>
          </div>
        </div>
      </div>

      {/* Total Tokens */}
      <div className="bg-[#14171f] border border-[#222836] rounded-xl p-5 relative overflow-hidden shadow-sm">
        <div className="flex items-center justify-between">
          <span className="text-xs font-medium uppercase tracking-wider text-google-gray-500">Inference Tokens</span>
          <div className="p-2 bg-indigo-500/10 text-indigo-400 rounded-lg border border-indigo-500/20">
            <Cpu className="h-4 w-4" />
          </div>
        </div>
        <div className="mt-3">
          <div className="text-2xl sm:text-3xl font-bold text-white tracking-tight">
            {(metrics.total_tokens / 1_000_000).toFixed(2)}M
          </div>
          <div className="mt-1 flex items-center text-xs text-google-gray-400">
            <span>{metrics.total_tokens.toLocaleString()} total tokens</span>
          </div>
        </div>
      </div>

      {/* Seat Utilization */}
      <div className="bg-[#14171f] border border-[#222836] rounded-xl p-5 relative overflow-hidden shadow-sm">
        <div className="flex items-center justify-between">
          <span className="text-xs font-medium uppercase tracking-wider text-google-gray-500">Seat Utilization</span>
          <div className="p-2 bg-google-green/10 text-google-green rounded-lg border border-google-green/20">
            <Users className="h-4 w-4" />
          </div>
        </div>
        <div className="mt-3">
          <div className="text-2xl sm:text-3xl font-bold text-white tracking-tight">
            {metrics.utilization_pct.toFixed(1)}%
          </div>
          <div className="mt-1 flex items-center text-xs text-google-gray-400">
            <span>{metrics.active_users_count} active / {metrics.seat_quota} quota</span>
          </div>
        </div>
      </div>

      {/* Reclamation Potential */}
      <div className="bg-[#14171f] border border-[#222836] rounded-xl p-5 relative overflow-hidden shadow-sm">
        <div className="flex items-center justify-between">
          <span className="text-xs font-medium uppercase tracking-wider text-google-gray-500">Monthly Savings Potential</span>
          <div className="p-2 bg-google-yellow/10 text-google-yellow rounded-lg border border-google-yellow/20">
            <TrendingDown className="h-4 w-4" />
          </div>
        </div>
        <div className="mt-3">
          <div className="text-2xl sm:text-3xl font-bold text-google-yellow tracking-tight">
            ${metrics.estimated_monthly_savings.toFixed(2)}
            <span className="text-sm font-normal text-google-gray-400">/mo</span>
          </div>
          <div className="mt-1 flex items-center text-xs text-google-gray-400">
            <span>Reclaiming dormant licenses</span>
          </div>
        </div>
      </div>
    </div>
  );
};
