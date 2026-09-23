import React, { useState } from 'react';
import { DailyTrend, ModelDetail } from '../api';

interface CostChartProps {
  trends: DailyTrend[];
  modelBreakdown: Record<string, ModelDetail>;
}

export const CostChart: React.FC<CostChartProps> = ({ trends, modelBreakdown }) => {
  const [hoveredIdx, setHoveredIdx] = useState<number | null>(null);

  const maxCost = Math.max(...trends.map((t) => t.cost), 1);
  const totalCost = Object.values(modelBreakdown).reduce((acc, curr) => acc + curr.cost, 0);

  // Model colors
  const modelColors: Record<string, string> = {
    'gemini-1.5-pro': '#1a73e8',
    'gemini-1.5-flash': '#34a853',
    'claude-3-5-sonnet-v2': '#ea4335',
  };

  return (
    <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
      {/* Daily Spend Trend SVG Chart */}
      <div className="lg:col-span-2 bg-[#14171f] border border-[#222836] rounded-xl p-5 shadow-sm">
        <div className="flex items-center justify-between mb-4">
          <div>
            <h3 className="text-sm font-semibold text-white">Daily Reconciled Cloud Spend</h3>
            <p className="text-xs text-google-gray-400">Total daily cost attributed to Antigravity & Gemini usage</p>
          </div>
          {hoveredIdx !== null && trends[hoveredIdx] && (
            <div className="text-right">
              <span className="text-xs text-google-gray-400">{trends[hoveredIdx].date}</span>
              <div className="text-sm font-bold text-google-blue">
                ${trends[hoveredIdx].cost.toFixed(2)}
              </div>
            </div>
          )}
        </div>

        {/* SVG Bar Chart */}
        <div className="h-48 w-full flex items-end space-x-1.5 pt-4">
          {trends.map((t, idx) => {
            const heightPct = Math.max((t.cost / maxCost) * 100, 4);
            const isHovered = hoveredIdx === idx;
            return (
              <div
                key={t.date}
                className="flex-1 flex flex-col items-center h-full justify-end group relative cursor-pointer"
                onMouseEnter={() => setHoveredIdx(idx)}
                onMouseLeave={() => setHoveredIdx(null)}
              >
                <div
                  style={{ height: `${heightPct}%` }}
                  className={`w-full rounded-t transition-all ${
                    isHovered
                      ? 'bg-google-blue brightness-125'
                      : 'bg-google-blue/70 hover:bg-google-blue/90'
                  }`}
                />
              </div>
            );
          })}
        </div>

        {/* Timeline Axis Labels */}
        <div className="flex justify-between mt-2 pt-2 border-t border-[#222836] text-[10px] text-google-gray-500">
          <span>{trends[0]?.date || ''}</span>
          <span>{trends[Math.floor(trends.length / 2)]?.date || ''}</span>
          <span>{trends[trends.length - 1]?.date || ''}</span>
        </div>
      </div>

      {/* Model Distribution Card */}
      <div className="bg-[#14171f] border border-[#222836] rounded-xl p-5 shadow-sm">
        <h3 className="text-sm font-semibold text-white mb-1">Model Cost Distribution</h3>
        <p className="text-xs text-google-gray-400 mb-4">Proportional cost breakdown by underlying AI model</p>

        <div className="space-y-3">
          {Object.entries(modelBreakdown).map(([model, detail]) => {
            const pct = totalCost > 0 ? (detail.cost / totalCost) * 100 : 0;
            const barColor = modelColors[model] || '#fbbc04';

            return (
              <div key={model} className="space-y-1">
                <div className="flex items-center justify-between text-xs">
                  <div className="flex items-center space-x-2">
                    <span className="w-2.5 h-2.5 rounded-full" style={{ backgroundColor: barColor }} />
                    <span className="font-medium text-google-gray-200">{model}</span>
                  </div>
                  <div className="text-right">
                    <span className="font-semibold text-white">${detail.cost.toFixed(2)}</span>
                    <span className="text-google-gray-500 ml-1.5">({pct.toFixed(1)}%)</span>
                  </div>
                </div>
                {/* Progress bar */}
                <div className="h-1.5 w-full bg-[#1e2330] rounded-full overflow-hidden">
                  <div
                    className="h-full rounded-full transition-all duration-500"
                    style={{ width: `${pct}%`, backgroundColor: barColor }}
                  />
                </div>
              </div>
            );
          })}
        </div>
      </div>
    </div>
  );
};
