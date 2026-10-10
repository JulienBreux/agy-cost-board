import React, { useState } from 'react';
import { Target, AlertTriangle, CheckCircle, Edit3, Check } from 'lucide-react';
import { UserBudgetMetrics } from '../api';

interface BudgetProgressBarProps {
  budget?: UserBudgetMetrics;
  currency?: string;
  onUpdateBudget: (newBudget: number) => void;
}

export const BudgetProgressBar: React.FC<BudgetProgressBarProps> = ({
  budget,
  currency = 'USD',
  onUpdateBudget,
}) => {
  const safeBudget: UserBudgetMetrics = budget || {
    threshold: 100,
    current_spend: 0,
    utilization_percent: 0,
    projected_month_end_spend: 0,
    projected_overage: 0,
    on_track: true,
  };

  const [isEditing, setIsEditing] = useState(false);
  const [inputValue, setInputValue] = useState(safeBudget.threshold.toString());

  const currencySymbol = currency === 'EUR' ? '€' : currency === 'USD' ? '$' : currency;
  const pct = Math.min(Math.max(safeBudget.utilization_percent || 0, 0), 100);

  const handleSave = () => {
    const val = parseFloat(inputValue);
    if (!isNaN(val) && val > 0) {
      onUpdateBudget(val);
      setIsEditing(false);
    }
  };

  const getProgressColor = () => {
    if (!safeBudget.on_track || (safeBudget.utilization_percent || 0) >= 100) return 'bg-rose-500';
    if ((safeBudget.utilization_percent || 0) >= 80) return 'bg-amber-400';
    return 'bg-google-blue';
  };

  return (
    <div className="bg-[#14171f] border border-[#222836] rounded-xl p-5 shadow-sm">
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
        <div className="flex items-center space-x-2.5">
          <div className="p-2 bg-google-blue/10 text-google-blue rounded-lg border border-google-blue/20">
            <Target className="h-4 w-4" />
          </div>
          <div>
            <h3 className="text-sm font-semibold text-white tracking-tight">
              Monthly Budget & Quota Driving
            </h3>
            <p className="text-xs text-google-gray-400">
              Personal threshold and projection velocity
            </p>
          </div>
        </div>

        <div className="flex items-center space-x-2">
          {safeBudget.on_track ? (
            <span className="inline-flex items-center space-x-1 px-2.5 py-1 text-xs font-medium rounded-full bg-google-green/15 text-google-green border border-google-green/30">
              <CheckCircle className="w-3.5 h-3.5" />
              <span>On Track</span>
            </span>
          ) : (
            <span className="inline-flex items-center space-x-1 px-2.5 py-1 text-xs font-medium rounded-full bg-rose-500/15 text-rose-400 border border-rose-500/30">
              <AlertTriangle className="w-3.5 h-3.5" />
              <span>Risk of Overage</span>
            </span>
          )}

          {!isEditing ? (
            <button
              type="button"
              onClick={() => {
                setInputValue(safeBudget.threshold.toString());
                setIsEditing(true);
              }}
              aria-label="Set Budget"
              className="flex items-center space-x-1 px-2.5 py-1 text-xs font-medium text-google-gray-300 hover:text-white bg-[#1a1e27] hover:bg-[#202530] border border-[#222836] rounded-lg transition-colors"
            >
              <Edit3 className="w-3.5 h-3.5 text-google-gray-400" />
              <span>Set Budget</span>
            </button>
          ) : (
            <div className="flex items-center space-x-1.5 animate-in fade-in duration-100">
              <span className="text-xs text-google-gray-400">{currencySymbol}</span>
              <input
                type="number"
                value={inputValue}
                onChange={(e) => setInputValue(e.target.value)}
                placeholder="100"
                className="w-20 bg-[#101218] border border-[#2d3548] rounded px-2 py-0.5 text-xs text-white focus:outline-none focus:border-google-blue"
              />
              <button
                type="button"
                onClick={handleSave}
                aria-label="Save"
                className="p-1 bg-google-blue/20 hover:bg-google-blue/30 text-google-blue rounded border border-google-blue/40"
              >
                <Check className="w-3.5 h-3.5" />
              </button>
            </div>
          )}
        </div>
      </div>

      {/* Spend vs Threshold numbers */}
      <div className="mt-4 flex flex-col sm:flex-row sm:items-baseline justify-between gap-1">
        <div className="flex items-baseline space-x-2">
          <span className="text-2xl font-bold text-white tracking-tight">
            {currencySymbol}
            {(safeBudget.current_spend || 0).toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}
          </span>
          <span className="text-xs text-google-gray-400">
            of {currencySymbol}
            {(safeBudget.threshold || 0).toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}
          </span>
        </div>
        <span className="text-xs font-semibold text-google-gray-300">
          {(safeBudget.utilization_percent || 0).toFixed(1)}% utilized
        </span>
      </div>

      {/* Progress bar */}
      <div className="mt-2.5 w-full bg-[#1b202c] rounded-full h-2.5 overflow-hidden">
        <div
          className={`h-2.5 rounded-full transition-all duration-500 ${getProgressColor()}`}
          style={{ width: `${pct}%` }}
        />
      </div>

      {/* Forecast and contextual notes */}
      <div className="mt-3 flex flex-col sm:flex-row sm:items-center justify-between text-xs text-google-gray-400 gap-1 border-t border-[#1d222e] pt-2.5">
        <div>
          <span>
            Projected {currencySymbol}{(safeBudget.projected_month_end_spend || 0).toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })} at month end
          </span>
        </div>

        {(safeBudget.projected_overage || 0) > 0 ? (
          <span className="text-rose-400 font-medium">
            Over budget by {currencySymbol}{(safeBudget.projected_overage || 0).toFixed(2)}
          </span>
        ) : (
          <span className="text-google-green font-medium">
            Projected within budget
          </span>
        )}
      </div>
    </div>
  );
};
