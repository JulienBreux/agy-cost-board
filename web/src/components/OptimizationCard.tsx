import { FC } from 'react';
import type { OptimizationTip } from '../api';

interface OptimizationCardProps {
  tips?: OptimizationTip[];
  currency?: string;
  onApplyTip?: (tip: OptimizationTip) => void;
}

export const OptimizationCard: FC<OptimizationCardProps> = ({
  tips = [],
  currency = '$',
  onApplyTip,
}) => {
  const safeTips = Array.isArray(tips) ? tips : [];
  const currencySymbol = currency === 'EUR' ? '€' : currency === 'USD' ? '$' : currency;
  const totalSavings = safeTips.reduce((acc, tip) => acc + (tip.potential_savings_monthly || 0), 0);

  const getSeverityBadge = (severity: OptimizationTip['severity']) => {
    switch (severity) {
      case 'critical':
        return 'bg-red-500/10 text-red-400 border-red-500/20';
      case 'warning':
        return 'bg-amber-500/10 text-amber-400 border-amber-500/20';
      case 'info':
      default:
        return 'bg-blue-500/10 text-blue-400 border-blue-500/20';
    }
  };

  const getTypeLabel = (type: string) => {
    switch (type) {
      case 'model_switch':
        return 'Model Selection';
      case 'caching':
        return 'Context Caching';
      case 'budget':
        return 'Budget & Quota';
      default:
        return type.replace(/_/g, ' ').toUpperCase();
    }
  };

  return (
    <div className="bg-[#14171f] border border-[#222836] rounded-xl p-5 shadow-sm">
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 pb-4 border-b border-[#202530]">
        <div className="flex items-center gap-2.5">
          <div className="p-2 rounded-lg bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
            <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M13 10V3L4 14h7v7l9-11h-7z"
              />
            </svg>
          </div>
          <div>
            <h3 className="text-base font-semibold text-white">Optimization Insights</h3>
            <p className="text-xs text-google-gray-400">Actionable advice to drive cost efficiency</p>
          </div>
        </div>

        {totalSavings > 0 && (
          <div className="flex items-center gap-2 px-3 py-1.5 rounded-lg bg-emerald-500/10 border border-emerald-500/20 text-emerald-300 text-xs font-medium">
            <span>Potential Savings:</span>
            <span className="font-semibold text-emerald-400">
              +{currencySymbol}{totalSavings.toFixed(2)} / month
            </span>
          </div>
        )}
      </div>

      {safeTips.length === 0 ? (
        <div className="py-8 text-center">
          <div className="w-12 h-12 mx-auto mb-3 rounded-full bg-[#1e2330] flex items-center justify-center text-emerald-400">
            <svg className="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
            </svg>
          </div>
          <p className="text-sm font-medium text-google-gray-300">Your consumption is well-optimized!</p>
          <p className="text-xs text-google-gray-500 mt-1">No cost-saving recommendations right now.</p>
        </div>
      ) : (
        <div className="mt-4 space-y-3">
          {safeTips.map((tip, idx) => (
            <div
              key={idx}
              className="p-3.5 rounded-lg bg-[#181c26] border border-[#202530] hover:border-[#2e3748] transition flex flex-col md:flex-row items-start md:items-center justify-between gap-3"
            >
              <div className="space-y-1.5 flex-1">
                <div className="flex items-center gap-2 flex-wrap">
                  <span
                    className={`text-[10px] font-semibold px-2 py-0.5 rounded border uppercase tracking-wider ${getSeverityBadge(
                      tip.severity
                    )}`}
                  >
                    {tip.severity}
                  </span>
                  <span className="text-[10px] font-medium text-google-gray-400 uppercase tracking-wider">
                    {getTypeLabel(tip.type || '')}
                  </span>
                  <span className="text-sm font-medium text-white">{tip.title}</span>
                </div>
                <p className="text-xs text-google-gray-400 leading-relaxed">{tip.description}</p>
              </div>

              <div className="flex items-center gap-3 self-end md:self-center shrink-0">
                {(tip.potential_savings_monthly || 0) > 0 && (
                  <span className="px-2.5 py-1 rounded bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 text-xs font-semibold whitespace-nowrap">
                    Save {currencySymbol}{(tip.potential_savings_monthly || 0).toFixed(2)}/mo
                  </span>
                )}
                {onApplyTip && (
                  <button
                    type="button"
                    onClick={() => onApplyTip(tip)}
                    className="text-xs px-2.5 py-1 rounded bg-[#1e2330] hover:bg-[#202530] text-google-gray-300 hover:text-white border border-[#202530] transition font-medium"
                  >
                    Learn More
                  </button>
                )}
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
};
