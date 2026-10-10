import React, { useState } from 'react';
import { PersonalDailyTrend, ModelDetail } from '../api';

interface PersonalCostChartProps {
  trends?: PersonalDailyTrend[];
  modelDistribution?: Record<string, ModelDetail>;
  currency?: string;
}

const getModelColor = (model: string): string => {
  const palette: Record<string, string> = {
    'gemini-3.8-flash': '#34a853',
    'gemini-3.5-flash-lite': '#4285f4',
    'gemini-2.5-flash': '#34a853',
    'gemini-2.5-pro': '#1a73e8',
    'gemini-1.5-pro': '#1a73e8',
    'gemini-1.5-flash': '#34a853',
    'claude-3-5-sonnet-v2': '#ea4335',
    'claude-3-7-sonnet': '#ea4335',
  };
  if (palette[model]) return palette[model];

  let hash = 0;
  for (let i = 0; i < model.length; i++) {
    hash = model.charCodeAt(i) + ((hash << 5) - hash);
  }
  const hue = Math.abs(hash) % 360;
  return `hsl(${hue}, 70%, 55%)`;
};

export const PersonalCostChart: React.FC<PersonalCostChartProps> = ({
  trends = [],
  modelDistribution = {},
  currency = 'USD',
}) => {
  const [hoveredIdx, setHoveredIdx] = useState<number | null>(null);

  const currencySymbol = currency === 'EUR' ? '€' : currency === 'USD' ? '$' : currency;
  const safeTrends = Array.isArray(trends) ? trends : [];
  const safeDistribution = modelDistribution && typeof modelDistribution === 'object' ? modelDistribution : {};

  const maxCost = Math.max(...safeTrends.map((t) => t.cost || 0), 0.01);
  const totalPeriodCost = safeTrends.reduce((acc, t) => acc + (t.cost || 0), 0);
  const totalPeriodTokens = safeTrends.reduce((acc, t) => acc + (t.total_tokens || 0), 0);

  const hoveredItem = hoveredIdx !== null && safeTrends[hoveredIdx] ? safeTrends[hoveredIdx] : null;

  return (
    <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
      {/* Daily Spend Trajectory */}
      <div className="lg:col-span-2 bg-[#14171f] border border-[#222836] rounded-xl p-5 shadow-sm flex flex-col justify-between">
        <div>
          <div className="flex items-center justify-between mb-4">
            <div>
              <h3 className="text-sm font-semibold text-white">Daily Spend Trajectory</h3>
              <p className="text-xs text-google-gray-400">
                Personal daily spend attributed to model invocations
              </p>
            </div>
            {hoveredItem ? (
              <div className="text-right min-h-[42px] flex flex-col justify-center">
                <div className="flex items-baseline justify-end space-x-2">
                  <span className="text-xs text-google-gray-400 font-mono">{hoveredItem.date || 'Today'}</span>
                  <span className="text-sm font-bold text-google-blue font-mono">
                    {currencySymbol}{(hoveredItem.cost || 0).toFixed(2)}
                  </span>
                </div>
                <div className="text-[10px] text-google-gray-500 font-mono">
                  {(hoveredItem.total_tokens || 0).toLocaleString()} tokens
                </div>
              </div>
            ) : (
              <div className="text-right min-h-[42px] flex flex-col justify-center">
                <div className="flex items-baseline justify-end space-x-2">
                  <span className="text-xs text-google-gray-400">
                    {safeTrends.length} {safeTrends.length === 1 ? 'day' : 'days'}
                  </span>
                  <span className="text-sm font-bold text-white font-mono">
                    {currencySymbol}{(totalPeriodCost || 0).toFixed(2)}
                  </span>
                </div>
                <div className="text-[10px] text-google-gray-500 font-mono">
                  {(totalPeriodTokens || 0).toLocaleString()} tokens
                </div>
              </div>
            )}
          </div>

          {/* SVG Bar Chart */}
          {safeTrends.length === 0 ? (
            <div className="h-48 w-full flex items-center justify-center text-xs text-google-gray-500">
              No daily spend recorded in this period.
            </div>
          ) : (
            <div className="h-48 w-full flex items-end space-x-1.5 pt-4">
              {safeTrends.map((t, idx) => {
                const cost = t.cost || 0;
                const heightPct = Math.min(Math.max((cost / maxCost) * 100, 6), 100);
                const isHovered = hoveredIdx === idx;
                return (
                  <div
                    key={t.date || idx}
                    data-testid="trend-bar"
                    className="flex-1 flex flex-col items-center h-full justify-end group relative cursor-pointer"
                    onMouseEnter={() => setHoveredIdx(idx)}
                    onMouseLeave={() => setHoveredIdx(null)}
                  >
                    <div
                      style={{ height: `${heightPct}%` }}
                      className={`w-full rounded-t transition-all ${
                        isHovered
                          ? 'bg-google-blue brightness-125 shadow-lg shadow-google-blue/30'
                          : 'bg-google-blue/70 hover:bg-google-blue/90'
                      }`}
                    />
                  </div>
                );
              })}
            </div>
          )}
        </div>

        {/* Timeline Axis Labels */}
        {safeTrends.length > 0 && (
          <div className="flex justify-between mt-2 pt-2 border-t border-[#222836] text-[10px] text-google-gray-500">
            {safeTrends.length === 1 ? (
              <span className="mx-auto">{safeTrends[0]?.date || ''}</span>
            ) : (
              <>
                <span>{safeTrends[0]?.date || ''}</span>
                <span>{safeTrends[Math.floor(safeTrends.length / 2)]?.date || ''}</span>
                <span>{safeTrends[safeTrends.length - 1]?.date || ''}</span>
              </>
            )}
          </div>
        )}
      </div>

      {/* Model Distribution Card */}
      <div className="bg-[#14171f] border border-[#222836] rounded-xl p-5 shadow-sm">
        <h3 className="text-sm font-semibold text-white mb-1">Personal Model Breakdown</h3>
        <p className="text-xs text-google-gray-400 mb-4">Cost and token volume by model</p>

        <div className="space-y-3">
          {Object.keys(safeDistribution).length === 0 ? (
            <div className="text-xs text-google-gray-500 py-6 text-center">
              No model activity recorded in this period.
            </div>
          ) : (
            Object.entries(safeDistribution).map(([model, detail]) => {
              const cost = detail?.cost ?? 0;
              const share = detail?.share ?? 0;
              const tokens = detail?.tokens ?? 0;
              const barColor = getModelColor(model);

              return (
                <div key={model} className="space-y-1">
                  <div className="flex items-center justify-between text-xs">
                    <div className="flex items-center space-x-2 truncate">
                      <span className="w-2.5 h-2.5 rounded-full flex-shrink-0" style={{ backgroundColor: barColor }} />
                      <span className="font-medium text-google-gray-200 truncate">{model}</span>
                    </div>
                    <div className="text-right whitespace-nowrap ml-2">
                      <span className="font-semibold text-white">
                        {currencySymbol}{cost.toFixed(2)}
                      </span>
                      <span className="text-google-gray-500 ml-1.5">({share.toFixed(1)}%)</span>
                    </div>
                  </div>
                  {/* Progress bar */}
                  <div className="h-1.5 w-full bg-[#1e2330] rounded-full overflow-hidden">
                    <div
                      className="h-full rounded-full transition-all duration-500"
                      style={{ width: `${Math.min(Math.max(share, 0), 100)}%`, backgroundColor: barColor }}
                    />
                  </div>
                  <div className="text-[10px] text-google-gray-500 text-right">
                    {tokens.toLocaleString()} tokens
                  </div>
                </div>
              );
            })
          )}
        </div>
      </div>
    </div>
  );
};
