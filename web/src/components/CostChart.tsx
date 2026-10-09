import React, { useState } from 'react';
import { DailyTrend, ModelDetail } from '../api';

interface CostChartProps {
  trends?: DailyTrend[];
  modelBreakdown?: Record<string, ModelDetail>;
}

const getTrendCost = (t?: DailyTrend | null): number => {
  if (!t) return 0;
  const val = t.cost ?? t.total_cost ?? 0;
  return typeof val === 'number' && !isNaN(val) ? val : 0;
};

const getTrendTokens = (t?: DailyTrend | null): number => {
  if (!t) return 0;
  const val = t.tokens ?? t.total_tokens ?? 0;
  return typeof val === 'number' && !isNaN(val) ? val : 0;
};

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

  // Consistent color generation for unknown models
  let hash = 0;
  for (let i = 0; i < model.length; i++) {
    hash = model.charCodeAt(i) + ((hash << 5) - hash);
  }
  const hue = Math.abs(hash) % 360;
  return `hsl(${hue}, 70%, 55%)`;
};

export const CostChart: React.FC<CostChartProps> = ({ trends = [], modelBreakdown = {} }) => {
  const [hoveredIdx, setHoveredIdx] = useState<number | null>(null);

  const safeTrends = Array.isArray(trends) ? trends : [];
  const safeBreakdown = modelBreakdown && typeof modelBreakdown === 'object' ? modelBreakdown : {};

  const maxCost = Math.max(...safeTrends.map(getTrendCost), 0.01);
  const totalCost = Object.values(safeBreakdown).reduce((acc, curr) => {
    const cost = curr?.cost ?? 0;
    return acc + (typeof cost === 'number' && !isNaN(cost) ? cost : 0);
  }, 0);

  const hoveredItem = hoveredIdx !== null && safeTrends[hoveredIdx] ? safeTrends[hoveredIdx] : null;

  return (
    <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
      {/* Daily Spend Trend SVG Chart */}
      <div className="lg:col-span-2 bg-[#14171f] border border-[#222836] rounded-xl p-5 shadow-sm flex flex-col justify-between">
        <div>
          <div className="flex items-center justify-between mb-4">
            <div>
              <h3 className="text-sm font-semibold text-white">Daily Reconciled Cloud Spend</h3>
              <p className="text-xs text-google-gray-400">Total daily cost attributed to Antigravity & Gemini usage</p>
            </div>
            {hoveredItem ? (
              <div className="text-right min-h-[42px] flex flex-col justify-center">
                <div className="flex items-baseline justify-end space-x-2">
                  <span className="text-xs text-google-gray-400 font-mono">{hoveredItem.date || 'Today'}</span>
                  <span className="text-sm font-bold text-google-blue font-mono">
                    ${getTrendCost(hoveredItem).toFixed(2)}
                  </span>
                </div>
                <div className="text-[10px] text-google-gray-500 font-mono">
                  {getTrendTokens(hoveredItem) > 0
                    ? `${getTrendTokens(hoveredItem).toLocaleString()} tokens`
                    : '0 tokens'}
                </div>
              </div>
            ) : (
              <div className="text-right min-h-[42px] flex flex-col justify-center">
                <div className="flex items-baseline justify-end space-x-2">
                  <span className="text-xs text-google-gray-400">
                    {safeTrends.length} {safeTrends.length === 1 ? 'day' : 'days'}
                  </span>
                  <span className="text-sm font-bold text-white font-mono">
                    ${safeTrends.reduce((a, b) => a + getTrendCost(b), 0).toFixed(2)}
                  </span>
                </div>
                <div className="text-[10px] text-google-gray-500 font-mono">
                  {safeTrends.reduce((a, b) => a + getTrendTokens(b), 0).toLocaleString()} tokens
                </div>
              </div>
            )}
          </div>

          {/* SVG Bar Chart */}
          {safeTrends.length === 0 ? (
            <div className="h-48 w-full flex items-center justify-center text-xs text-google-gray-500">
              No daily spend activity recorded in this period.
            </div>
          ) : (
            <div className="h-48 w-full flex items-end space-x-1.5 pt-4">
              {safeTrends.map((t, idx) => {
                const cost = getTrendCost(t);
                const heightPct = Math.min(Math.max((cost / maxCost) * 100, 6), 100);
                const isHovered = hoveredIdx === idx;
                return (
                  <div
                    key={t.date || idx}
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
        <h3 className="text-sm font-semibold text-white mb-1">Model Cost Distribution</h3>
        <p className="text-xs text-google-gray-400 mb-4">Proportional cost breakdown by underlying AI model</p>

        <div className="space-y-3">
          {Object.keys(safeBreakdown).length === 0 ? (
            <div className="text-xs text-google-gray-500 py-6 text-center">
              No AI model consumption recorded in this period.
            </div>
          ) : (
            Object.entries(safeBreakdown).map(([model, detail]) => {
              const cost = detail?.cost ?? 0;
              const pct = totalCost > 0 ? (cost / totalCost) * 100 : 0;
              const barColor = getModelColor(model);

              return (
                <div key={model} className="space-y-1">
                  <div className="flex items-center justify-between text-xs">
                    <div className="flex items-center space-x-2">
                      <span className="w-2.5 h-2.5 rounded-full flex-shrink-0" style={{ backgroundColor: barColor }} />
                      <span className="font-medium text-google-gray-200 truncate">{model}</span>
                    </div>
                    <div className="text-right whitespace-nowrap ml-2">
                      <span className="font-semibold text-white">${cost.toFixed(2)}</span>
                      <span className="text-google-gray-500 ml-1.5">({pct.toFixed(1)}%)</span>
                    </div>
                  </div>
                  {/* Progress bar */}
                  <div className="h-1.5 w-full bg-[#1e2330] rounded-full overflow-hidden">
                    <div
                      className="h-full rounded-full transition-all duration-500"
                      style={{ width: `${Math.min(Math.max(pct, 0), 100)}%`, backgroundColor: barColor }}
                    />
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
