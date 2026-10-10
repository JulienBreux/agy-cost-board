import React, { useState, useMemo } from 'react';
import { PersonalDailyTrend, ModelDetail } from '../api';
import { Layers, BarChart3, Coins, Sparkles, PieChart } from 'lucide-react';

interface PersonalCostChartProps {
  trends?: PersonalDailyTrend[];
  modelDistribution?: Record<string, ModelDetail>;
  currency?: string;
}

export const getModelColor = (model: string): string => {
  const palette: Record<string, string> = {
    'gemini-3.8-flash': '#34a853',
    'gemini-3.5-flash-lite': '#4285f4',
    'gemini-2.5-flash': '#34a853',
    'gemini-2.5-pro': '#1a73e8',
    'gemini-1.5-pro': '#1a73e8',
    'gemini-1.5-flash': '#34a853',
    'claude-3-5-sonnet-v2': '#ea4335',
    'claude-3-7-sonnet': '#ea4335',
    'claude-3-5-haiku': '#fbbc04',
  };
  if (palette[model]) return palette[model];

  let hash = 0;
  for (let i = 0; i < model.length; i++) {
    hash = model.charCodeAt(i) + ((hash << 5) - hash);
  }
  const hue = Math.abs(hash) % 360;
  return `hsl(${hue}, 70%, 55%)`;
};

const formatTokens = (tokens: number): string => {
  if (tokens >= 1_000_000) {
    return `${(tokens / 1_000_000).toFixed(2)}M`;
  }
  if (tokens >= 1_000) {
    return `${(tokens / 1_000).toFixed(1)}k`;
  }
  return tokens.toLocaleString();
};

export const PersonalCostChart: React.FC<PersonalCostChartProps> = ({
  trends = [],
  modelDistribution = {},
  currency = 'USD',
}) => {
  const [hoveredIdx, setHoveredIdx] = useState<number | null>(null);
  const [focusedModel, setFocusedModel] = useState<string | null>(null);
  const [chartMode, setChartMode] = useState<'stacked' | 'aggregate'>('stacked');
  const [metricMode, setMetricMode] = useState<'cost' | 'tokens'>('cost');

  const currencySymbol = currency === 'EUR' ? '€' : currency === 'USD' ? '$' : currency;
  const safeTrends = Array.isArray(trends) ? trends : [];
  const safeDistribution = modelDistribution && typeof modelDistribution === 'object' ? modelDistribution : {};

  // Compute breakdown totals
  const totalPeriodCost = useMemo(() => {
    return safeTrends.reduce((acc, t) => acc + (t.cost || 0), 0);
  }, [safeTrends]);

  const totalPeriodTokens = useMemo(() => {
    return safeTrends.reduce((acc, t) => acc + (t.total_tokens || 0), 0);
  }, [safeTrends]);

  const totalDistCost = useMemo(() => {
    const sum = Object.values(safeDistribution).reduce((acc, curr) => acc + (curr?.cost ?? 0), 0);
    return sum > 0 ? sum : totalPeriodCost;
  }, [safeDistribution, totalPeriodCost]);

  const totalDistTokens = useMemo(() => {
    const sum = Object.values(safeDistribution).reduce((acc, curr) => acc + (curr?.tokens ?? 0), 0);
    return sum > 0 ? sum : totalPeriodTokens;
  }, [safeDistribution, totalPeriodTokens]);

  // Max value for scaling daily bars
  const maxVal = useMemo(() => {
    if (metricMode === 'cost') {
      return Math.max(...safeTrends.map((t) => t.cost || 0), 0.01);
    }
    return Math.max(...safeTrends.map((t) => t.total_tokens || 0), 1);
  }, [safeTrends, metricMode]);

  // Sorted model entries by cost/tokens descending
  const sortedModels = useMemo(() => {
    return Object.entries(safeDistribution).sort(([, a], [, b]) => {
      if (metricMode === 'cost') {
        return (b?.cost ?? 0) - (a?.cost ?? 0);
      }
      return (b?.tokens ?? 0) - (a?.tokens ?? 0);
    });
  }, [safeDistribution, metricMode]);

  const hoveredItem = hoveredIdx !== null && safeTrends[hoveredIdx] ? safeTrends[hoveredIdx] : null;

  // Donut geometry calculations
  const donutRadius = 52;
  const donutCircumference = 2 * Math.PI * donutRadius; // ~326.726

  const donutSlices = useMemo(() => {
    const total = metricMode === 'cost' ? totalDistCost : totalDistTokens;
    if (total <= 0) return [];

    let accumulatedOffset = 0;
    return sortedModels.map(([model, detail]) => {
      const val = metricMode === 'cost' ? (detail?.cost ?? 0) : (detail?.tokens ?? 0);
      const ratio = total > 0 ? Math.max(0, val / total) : 0;
      const strokeDash = ratio * donutCircumference;
      const sliceOffset = accumulatedOffset;
      accumulatedOffset += strokeDash;

      return {
        model,
        detail,
        value: val,
        ratio,
        strokeDash,
        strokeOffset: sliceOffset,
        color: getModelColor(model),
      };
    });
  }, [sortedModels, metricMode, totalDistCost, totalDistTokens, donutCircumference]);

  const activeDonutModel = focusedModel ? safeDistribution[focusedModel] : null;

  return (
    <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
      {/* Daily Spend Trajectory SVG Chart */}
      <div className="lg:col-span-2 bg-[#14171f] border border-[#222836] rounded-xl p-5 shadow-sm flex flex-col justify-between">
        <div>
          {/* Top Bar: Title & Controls */}
          <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3 mb-4">
            <div>
              <div className="flex items-center space-x-2">
                <h3 className="text-sm font-semibold text-white">Daily Spend Trajectory</h3>
                <span className="text-[10px] font-semibold tracking-wide uppercase px-1.5 py-0.5 rounded bg-google-blue/15 text-google-blue border border-google-blue/30">
                  {chartMode === 'stacked' ? 'Model Breakdown' : 'Aggregate'}
                </span>
              </div>
              <p className="text-xs text-google-gray-400 mt-0.5">
                Personal daily spend attributed to model invocations
              </p>
            </div>

            {/* View Mode & Metric Toggles */}
            <div className="flex items-center space-x-2 self-start sm:self-auto">
              <div className="bg-[#1a1f2c] p-0.5 rounded-lg border border-[#2a3245] flex items-center text-xs">
                <button
                  onClick={() => setChartMode('stacked')}
                  className={`px-2.5 py-1 rounded-md flex items-center space-x-1.5 transition-colors ${
                    chartMode === 'stacked'
                      ? 'bg-google-blue text-white font-medium shadow-sm'
                      : 'text-google-gray-400 hover:text-white'
                  }`}
                  title="Show daily cost stacked by AI model"
                >
                  <Layers className="h-3 w-3" />
                  <span className="hidden sm:inline">Stacked</span>
                </button>
                <button
                  onClick={() => setChartMode('aggregate')}
                  className={`px-2.5 py-1 rounded-md flex items-center space-x-1.5 transition-colors ${
                    chartMode === 'aggregate'
                      ? 'bg-google-blue text-white font-medium shadow-sm'
                      : 'text-google-gray-400 hover:text-white'
                  }`}
                  title="Show aggregate daily total"
                >
                  <BarChart3 className="h-3 w-3" />
                  <span className="hidden sm:inline">Aggregate</span>
                </button>
              </div>

              <div className="bg-[#1a1f2c] p-0.5 rounded-lg border border-[#2a3245] flex items-center text-xs">
                <button
                  onClick={() => setMetricMode('cost')}
                  className={`px-2 py-1 rounded-md flex items-center space-x-1 transition-colors ${
                    metricMode === 'cost'
                      ? 'bg-[#2a3245] text-white font-medium'
                      : 'text-google-gray-400 hover:text-white'
                  }`}
                  title="View spend in currency"
                >
                  <Coins className="h-3 w-3" />
                  <span>$</span>
                </button>
                <button
                  onClick={() => setMetricMode('tokens')}
                  className={`px-2 py-1 rounded-md flex items-center space-x-1 transition-colors ${
                    metricMode === 'tokens'
                      ? 'bg-[#2a3245] text-white font-medium'
                      : 'text-google-gray-400 hover:text-white'
                  }`}
                  title="View token throughput"
                >
                  <Sparkles className="h-3 w-3" />
                  <span className="text-[11px]">Tokens</span>
                </button>
              </div>
            </div>
          </div>

          {/* Dynamic Summary / Hover Metrics Banner */}
          <div className="bg-[#0f1115] border border-[#222836] rounded-lg px-4 py-2.5 mb-4 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-2 min-h-[52px]">
            {hoveredItem ? (
              <>
                <div className="flex items-center space-x-3">
                  <div className="w-2 h-2 rounded-full bg-google-blue animate-pulse" />
                  <span className="text-xs font-mono font-medium text-google-gray-300">
                    {hoveredItem.date || 'Today'}
                  </span>
                </div>

                <div className="flex items-center space-x-4">
                  <div className="text-right">
                    <span className="text-xs text-google-gray-400 mr-2">Spend:</span>
                    <span className="text-sm font-bold text-google-blue font-mono">
                      {currencySymbol}{(hoveredItem.cost || 0).toFixed(2)}
                    </span>
                  </div>
                  <div className="text-right border-l border-[#222836] pl-4">
                    <span className="text-xs text-google-gray-400 mr-2">Tokens:</span>
                    <span className="text-xs font-semibold text-white font-mono">
                      {(hoveredItem.total_tokens || 0).toLocaleString()} tokens
                    </span>
                  </div>
                </div>
              </>
            ) : (
              <>
                <div className="flex items-center space-x-3">
                  <div className="w-2 h-2 rounded-full bg-google-green" />
                  <span className="text-xs text-google-gray-400">
                    Window Total ({safeTrends.length} {safeTrends.length === 1 ? 'day' : 'days'}):
                  </span>
                  <span className="text-sm font-bold text-white font-mono">
                    {currencySymbol}{(totalPeriodCost || 0).toFixed(2)}
                  </span>
                </div>

                <div className="flex items-center space-x-4 text-xs font-mono text-google-gray-400">
                  <div>
                    Avg:{' '}
                    <span className="text-white font-semibold">
                      {currencySymbol}
                      {(safeTrends.length > 0 ? totalPeriodCost / safeTrends.length : 0).toFixed(2)}
                      /day
                    </span>
                  </div>
                  <div className="border-l border-[#222836] pl-4">
                    Tokens:{' '}
                    <span className="text-white font-semibold">
                      {(totalPeriodTokens || 0).toLocaleString()} tokens
                    </span>
                  </div>
                </div>
              </>
            )}
          </div>

          {/* Model Repartition Pill Bar for Selected Day (if hovered) */}
          {hoveredItem && (hoveredItem.by_model || hoveredItem.models) &&
            Object.keys(hoveredItem.by_model || hoveredItem.models || {}).length > 0 && (
              <div className="flex flex-wrap items-center gap-2 mb-3 px-1">
                <span className="text-[10px] text-google-gray-500 uppercase tracking-wider font-semibold mr-1">
                  Day Repartition:
                </span>
                {Object.entries(hoveredItem.by_model || hoveredItem.models || {}).map(([model, cost]) => {
                  const dayCost = hoveredItem.cost || 0;
                  const pct = dayCost > 0 ? (cost / dayCost) * 100 : 0;
                  const mColor = getModelColor(model);
                  const isFocused = focusedModel === model;

                  return (
                    <button
                      key={model}
                      onClick={() => setFocusedModel(isFocused ? null : model)}
                      onMouseEnter={() => setFocusedModel(model)}
                      onMouseLeave={() => setFocusedModel(null)}
                      className={`inline-flex items-center space-x-1.5 px-2 py-0.5 rounded-full text-[11px] border transition-all ${
                        isFocused
                          ? 'bg-[#1e2330] border-white/40 text-white shadow-sm'
                          : 'bg-[#14171f] border-[#222836] text-google-gray-300 hover:border-google-gray-500'
                      }`}
                    >
                      <span className="w-2 h-2 rounded-full" style={{ backgroundColor: mColor }} />
                      <span className="font-medium truncate max-w-[120px]">{model}</span>
                      <span className="font-mono text-white font-semibold">{currencySymbol}{cost.toFixed(2)}</span>
                      <span className="text-[10px] text-google-gray-500">({pct.toFixed(0)}%)</span>
                    </button>
                  );
                })}
              </div>
            )}

          {/* Bar Chart Rendering */}
          {safeTrends.length === 0 ? (
            <div className="h-52 w-full flex items-center justify-center text-xs text-google-gray-500">
              No daily spend recorded in this period.
            </div>
          ) : (
            <div className="h-52 w-full flex items-end space-x-1 sm:space-x-1.5 pt-4">
              {safeTrends.map((t, idx) => {
                const dayVal = metricMode === 'cost' ? (t.cost || 0) : (t.total_tokens || 0);
                const heightPct = Math.min(Math.max((dayVal / maxVal) * 100, 4), 100);
                const isHovered = hoveredIdx === idx;
                const dayModels = t.by_model || t.models;
                const hasModelBreakdown = dayModels && Object.keys(dayModels).length > 0;

                return (
                  <div
                    key={t.date || idx}
                    data-testid="trend-bar"
                    className="flex-1 flex flex-col items-center h-full justify-end group relative cursor-pointer"
                    onMouseEnter={() => setHoveredIdx(idx)}
                    onMouseLeave={() => setHoveredIdx(null)}
                  >
                    {/* Bar Container */}
                    <div
                      style={{ height: `${heightPct}%` }}
                      className={`w-full rounded-t overflow-hidden flex flex-col-reverse transition-all duration-200 ${
                        isHovered
                          ? 'ring-2 ring-google-blue shadow-lg shadow-google-blue/30 z-10 brightness-110'
                          : 'hover:brightness-110'
                      }`}
                    >
                      {/* Stacked Model Segments */}
                      {chartMode === 'stacked' && hasModelBreakdown ? (
                        Object.entries(dayModels!).map(([model, cost]) => {
                          const totalDayCost = t.cost || 0;
                          const segPct = totalDayCost > 0 ? (cost / totalDayCost) * 100 : 0;
                          const mColor = getModelColor(model);
                          const isMFocused = focusedModel === model;
                          const isOtherFocused = focusedModel !== null && !isMFocused;

                          return (
                            <div
                              key={model}
                              style={{
                                height: `${Math.max(segPct, 0)}%`,
                                backgroundColor: mColor,
                                opacity: isOtherFocused ? 0.3 : 1,
                              }}
                              className="w-full transition-all duration-200"
                              title={`${model}: ${currencySymbol}${cost.toFixed(2)} (${segPct.toFixed(1)}%)`}
                            />
                          );
                        })
                      ) : (
                        /* Aggregate Single Bar */
                        <div
                          className={`w-full h-full transition-all ${
                            isHovered
                              ? 'bg-google-blue brightness-125 shadow-lg shadow-google-blue/30'
                              : 'bg-google-blue/70 hover:bg-google-blue/90'
                          }`}
                        />
                      )}
                    </div>
                  </div>
                );
              })}
            </div>
          )}
        </div>

        {/* Timeline Axis Labels */}
        {safeTrends.length > 0 && (
          <div className="flex justify-between mt-3 pt-2.5 border-t border-[#222836] text-[10px] text-google-gray-500 font-mono">
            {safeTrends.length === 1 ? (
              <span className="mx-auto">{safeTrends[0]?.date || ''}</span>
            ) : (
              <>
                <span>{safeTrends[0]?.date || ''}</span>
                {safeTrends.length > 6 && (
                  <span>{safeTrends[Math.floor(safeTrends.length / 2)]?.date || ''}</span>
                )}
                <span>{safeTrends[safeTrends.length - 1]?.date || ''}</span>
              </>
            )}
          </div>
        )}
      </div>

      {/* Model Distribution & Donut Card */}
      <div className="bg-[#14171f] border border-[#222836] rounded-xl p-5 shadow-sm flex flex-col justify-between">
        <div>
          <div className="flex items-center justify-between mb-1">
            <h3 className="text-sm font-semibold text-white flex items-center space-x-1.5">
              <PieChart className="h-4 w-4 text-google-blue" />
              <span>Personal Model Breakdown</span>
            </h3>
            <span className="text-[10px] font-mono text-google-gray-400">
              {sortedModels.length} {sortedModels.length === 1 ? 'Model' : 'Models'}
            </span>
          </div>
          <p className="text-xs text-google-gray-400 mb-4">
            Cost and token volume by model
          </p>

          {sortedModels.length === 0 ? (
            <div className="text-xs text-google-gray-500 py-12 text-center">
              No model activity recorded in this period.
            </div>
          ) : (
            <>
              {/* Interactive SVG Donut Chart */}
              <div className="relative flex items-center justify-center py-2 mb-4">
                <svg
                  viewBox="0 0 140 140"
                  className="w-36 h-36 transform -rotate-90 drop-shadow-sm"
                >
                  {/* Background Track */}
                  <circle
                    cx="70"
                    cy="70"
                    r={donutRadius}
                    fill="transparent"
                    stroke="#1e2330"
                    strokeWidth="14"
                  />
                  {/* Slices */}
                  {donutSlices.map((slice) => {
                    const isFocused = focusedModel === slice.model;
                    const isOtherFocused = focusedModel !== null && !isFocused;

                    return (
                      <circle
                        key={slice.model}
                        cx="70"
                        cy="70"
                        r={donutRadius}
                        fill="transparent"
                        stroke={slice.color}
                        strokeWidth={isFocused ? 18 : 14}
                        strokeDasharray={`${slice.strokeDash} ${donutCircumference - slice.strokeDash}`}
                        strokeDashoffset={-slice.strokeOffset}
                        strokeLinecap="butt"
                        className="transition-all duration-300 cursor-pointer"
                        style={{
                          opacity: isOtherFocused ? 0.35 : 1,
                        }}
                        onMouseEnter={() => setFocusedModel(slice.model)}
                        onMouseLeave={() => setFocusedModel(null)}
                      />
                    );
                  })}
                </svg>

                {/* Donut Center Label */}
                <div className="absolute inset-0 flex flex-col items-center justify-center text-center pointer-events-none px-4">
                  {activeDonutModel ? (
                    <>
                      <span className="text-[10px] text-google-gray-400 truncate max-w-[85px]">
                        {focusedModel}
                      </span>
                      <span className="text-sm font-bold text-white font-mono">
                        {metricMode === 'cost'
                          ? `${currencySymbol}${(activeDonutModel.cost ?? 0).toFixed(2)}`
                          : formatTokens(activeDonutModel.tokens ?? 0)}
                      </span>
                      <span className="text-[10px] font-semibold text-google-blue">
                        {totalDistCost > 0
                          ? `${(((activeDonutModel.cost ?? 0) / totalDistCost) * 100).toFixed(1)}%`
                          : '0%'}
                      </span>
                    </>
                  ) : (
                    <>
                      <span className="text-[10px] text-google-gray-400 font-medium">Total Spend</span>
                      <span className="text-sm font-bold text-white font-mono">
                        {currencySymbol}{totalDistCost.toFixed(2)}
                      </span>
                      <span className="text-[10px] text-google-gray-500 font-mono">
                        {formatTokens(totalDistTokens)} tokens
                      </span>
                    </>
                  )}
                </div>
              </div>

              {/* Detailed Breakdown List */}
              <div className="space-y-2.5">
                {sortedModels.map(([model, detail]) => {
                  const cost = detail?.cost ?? 0;
                  const share = detail?.share ?? 0;
                  const tokens = detail?.tokens ?? 0;
                  const mColor = getModelColor(model);
                  const isFocused = focusedModel === model;

                  return (
                    <div
                      key={model}
                      data-testid={`model-item-${model}`}
                      onMouseEnter={() => setFocusedModel(model)}
                      onMouseLeave={() => setFocusedModel(null)}
                      className={`p-2 rounded-lg border transition-all cursor-pointer ${
                        isFocused
                          ? 'bg-[#1a1f2c] border-google-blue/40 shadow-sm'
                          : 'bg-[#12151c]/60 border-[#1f2430] hover:border-[#2e3748]'
                      }`}
                    >
                      <div className="flex items-center justify-between text-xs mb-1">
                        <div className="flex items-center space-x-2 truncate">
                          <span
                            className="w-2.5 h-2.5 rounded-full flex-shrink-0 transition-transform"
                            style={{
                              backgroundColor: mColor,
                              transform: isFocused ? 'scale(1.2)' : 'scale(1)',
                            }}
                          />
                          <span
                            className={`font-medium truncate transition-colors ${
                              isFocused ? 'text-white' : 'text-google-gray-200'
                            }`}
                          >
                            {model}
                          </span>
                        </div>
                        <div className="text-right whitespace-nowrap ml-2">
                          <span className="font-semibold text-white font-mono">
                            {currencySymbol}{cost.toFixed(2)}
                          </span>
                          <span className="text-google-gray-500 ml-1.5 font-mono">
                            ({share.toFixed(1)}%)
                          </span>
                        </div>
                      </div>

                      {/* Proportional Progress Bar */}
                      <div className="h-1.5 w-full bg-[#1e2330] rounded-full overflow-hidden">
                        <div
                          className="h-full rounded-full transition-all duration-500"
                          style={{
                            width: `${Math.min(Math.max(share, 0), 100)}%`,
                            backgroundColor: mColor,
                          }}
                        />
                      </div>

                      <div className="flex items-center justify-between text-[10px] text-google-gray-500 mt-1 font-mono">
                        <span>{tokens.toLocaleString()} tokens</span>
                        <span>
                          {tokens > 0 ? `${currencySymbol}${((cost / tokens) * 1000).toFixed(4)}/1k tok` : '—'}
                        </span>
                      </div>
                    </div>
                  );
                })}
              </div>
            </>
          )}
        </div>
      </div>
    </div>
  );
};
