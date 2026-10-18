import React, { useState, useMemo } from 'react';
import { Trophy, Medal, Search, ArrowUpDown, Calendar, Cpu, Coins, Flame } from 'lucide-react';
import { LeaderboardEntry } from '../api';
import { formatCurrency } from '../utils/currency';

interface LeaderboardViewProps {
  entries: LeaderboardEntry[];
  onSelectUser: (userId: string) => void;
  days: number;
}

type SortField = 'cost' | 'tokens' | 'active_days' | 'efficiency';

const MODEL_COLORS: Record<string, string> = {
  'gemini-2.5-pro': 'bg-google-blue',
  'gemini-4.0-pro': 'bg-google-blue',
  'gemini-2.5-flash': 'bg-emerald-400',
  'gemini-4.0-flash': 'bg-emerald-400',
  'gemini-3.8-flash': 'bg-teal-400',
  'gemini-2.5-flash-lite': 'bg-teal-400',
  'claude-3-7-sonnet': 'bg-amber-400',
  'claude-3-5-sonnet': 'bg-orange-400',
  'claude-5.5-sonnet-medium': 'bg-amber-400',
  'claude-5.5-opus-high': 'bg-orange-400',
  'claude-5.5-opus-max': 'bg-rose-400',
};

const getModelColorClass = (model: string, idx: number): string => {
  if (MODEL_COLORS[model]) return MODEL_COLORS[model];
  const lower = model.toLowerCase();
  if (lower.includes('flash')) return 'bg-emerald-400';
  if (lower.includes('pro')) return 'bg-google-blue';
  if (lower.includes('claude')) return 'bg-amber-400';
  const fallback = ['bg-purple-400', 'bg-pink-400', 'bg-indigo-400', 'bg-cyan-400'];
  return fallback[idx % fallback.length];
};

const getSpendSharePct = (entry: LeaderboardEntry): number => {
  return Number(entry.spend_share_pct ?? entry.share_of_spend ?? 0);
};

const getModelSharePct = (val: unknown): number => {
  if (typeof val === 'number') {
    return val > 1 ? val : val * 100;
  }
  if (val && typeof val === 'object' && 'share' in val) {
    const s = Number((val as { share?: number }).share ?? 0);
    return s > 1 ? s : s * 100;
  }
  return 0;
};

const getModelSortWeight = (val: unknown): number => {
  if (val && typeof val === 'object' && 'tokens' in val) {
    const t = Number((val as { tokens?: number }).tokens ?? 0);
    if (t > 0) return t;
  }
  return getModelSharePct(val);
};

const formatCompactTokens = (tokens: number): string => {
  const safeTokens = Number(tokens ?? 0);
  if (safeTokens >= 1_000_000) return `${(safeTokens / 1_000_000).toFixed(2)}M`;
  if (safeTokens >= 1_000) return `${(safeTokens / 1_000).toFixed(1)}K`;
  return safeTokens.toLocaleString();
};

export const LeaderboardView: React.FC<LeaderboardViewProps> = ({
  entries,
  onSelectUser,
  days,
}) => {
  const [search, setSearch] = useState('');
  const [sortBy, setSortBy] = useState<SortField>('cost');
  const [sortOrder, setSortOrder] = useState<'asc' | 'desc'>('desc');

  const topThree = useMemo(() => {
    const byCost = [...entries].sort((a, b) => a.rank - b.rank);
    return byCost.slice(0, 3);
  }, [entries]);

  const filteredAndSorted = useMemo(() => {
    return entries
      .filter((e) => (e.user_id || '').toLowerCase().includes(search.toLowerCase()))
      .sort((a, b) => {
        let diff = 0;
        if (sortBy === 'cost') diff = (a.total_cost ?? 0) - (b.total_cost ?? 0);
        else if (sortBy === 'tokens') diff = (a.total_tokens ?? 0) - (b.total_tokens ?? 0);
        else if (sortBy === 'active_days') diff = (a.active_days ?? 0) - (b.active_days ?? 0);
        else if (sortBy === 'efficiency')
          diff = (a.cost_per_1k_tokens ?? 0) - (b.cost_per_1k_tokens ?? 0);
        if (diff === 0) {
          return (a.rank ?? 0) - (b.rank ?? 0);
        }
        return sortOrder === 'desc' ? -diff : diff;
      });
  }, [entries, search, sortBy, sortOrder]);

  const toggleSort = (col: SortField) => {
    if (sortBy === col) {
      setSortOrder(sortOrder === 'desc' ? 'asc' : 'desc');
    } else {
      setSortBy(col);
      setSortOrder('desc');
    }
  };

  const podiumStyles = [
    {
      rank: 1,
      label: '1st Place',
      border: 'border-amber-500/40 hover:border-amber-400/70',
      bg: 'bg-gradient-to-br from-amber-500/10 via-[#14171f] to-[#14171f]',
      badgeBg: 'bg-amber-500/20 text-amber-400 border-amber-500/40',
      iconColor: 'text-amber-400',
    },
    {
      rank: 2,
      label: '2nd Place',
      border: 'border-slate-400/40 hover:border-slate-300/70',
      bg: 'bg-gradient-to-br from-slate-400/10 via-[#14171f] to-[#14171f]',
      badgeBg: 'bg-slate-400/20 text-slate-300 border-slate-400/40',
      iconColor: 'text-slate-300',
    },
    {
      rank: 3,
      label: '3rd Place',
      border: 'border-orange-500/40 hover:border-orange-400/70',
      bg: 'bg-gradient-to-br from-orange-500/10 via-[#14171f] to-[#14171f]',
      badgeBg: 'bg-orange-500/20 text-orange-400 border-orange-500/40',
      iconColor: 'text-orange-400',
    },
  ];

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-2">
        <div>
          <div className="flex items-center space-x-2">
            <Trophy className="h-5 w-5 text-amber-400" />
            <h2 className="text-lg font-bold text-white tracking-tight">
              Developer Consumption Leaderboard
            </h2>
          </div>
          <p className="text-xs text-google-gray-400 mt-0.5">
            Ranked developer AI spend, token throughput, active days, and model mix efficiency over the last {days} days.
          </p>
        </div>
      </div>

      {/* Top 3 Podium Cards */}
      {topThree.length > 0 && (
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          {topThree.map((entry, idx) => {
            const style = podiumStyles[idx] || podiumStyles[2];
            const spendSharePct = getSpendSharePct(entry);
            return (
              <div
                key={entry.user_id}
                data-testid={`podium-card-${idx + 1}`}
                onClick={() => onSelectUser(entry.user_id)}
                className={`relative rounded-xl border ${style.border} ${style.bg} p-5 shadow-sm transition-all cursor-pointer group`}
              >
                <div className="flex items-start justify-between gap-2 mb-3">
                  <div className="flex items-center space-x-2 min-w-0">
                    <span
                      className={`inline-flex items-center space-x-1 px-2.5 py-0.5 rounded-full text-xs font-bold border ${style.badgeBg}`}
                    >
                      {idx === 0 ? (
                        <Trophy className="h-3.5 w-3.5 mr-1" />
                      ) : (
                        <Medal className="h-3.5 w-3.5 mr-1" />
                      )}
                      <span>#{entry.rank}</span>
                    </span>
                    <span className="text-xs text-google-gray-400 font-medium">
                      {style.label}
                    </span>
                  </div>
                  <span className="text-xs font-mono text-google-gray-400 bg-[#1a1e28] px-2 py-0.5 rounded border border-[#282f40]">
                    {spendSharePct.toFixed(1)}% of org
                  </span>
                </div>

                <div className="mb-4">
                  <div className="text-sm font-bold text-white truncate group-hover:text-google-blue transition-colors">
                    {entry.user_id}
                  </div>
                  <div className="text-2xl font-extrabold text-white font-mono mt-1">
                    {formatCurrency(entry.total_cost ?? 0, entry.currency)}
                  </div>
                  {((entry.included_cost ?? 0) > 0 || (entry.payg_cost ?? 0) > 0) && (
                    <div className="flex flex-wrap items-center gap-1.5 mt-1.5 text-[11px] font-mono">
                      <span className="px-1.5 py-0.5 rounded bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
                        Inc: {formatCurrency(entry.included_cost ?? 0, entry.currency)}
                      </span>
                      <span className="px-1.5 py-0.5 rounded bg-amber-500/10 text-amber-400 border border-amber-500/20">
                        PAYG: {formatCurrency(entry.payg_cost ?? 0, entry.currency)}
                      </span>
                    </div>
                  )}
                </div>

                <div className="grid grid-cols-3 gap-2 pt-3 border-t border-[#222836]/80 text-xs">
                  <div>
                    <div className="text-google-gray-500 flex items-center space-x-1">
                      <Coins className="h-3 w-3" />
                      <span>Tokens</span>
                    </div>
                    <div className="font-mono font-semibold text-google-gray-200 mt-0.5">
                      {formatCompactTokens(entry.total_tokens)}
                    </div>
                    {(entry.cache_hit_rate ?? 0) > 0 && (
                      <div className="text-[10px] font-mono text-emerald-400">
                        {Number(entry.cache_hit_rate).toFixed(1)}% cache
                      </div>
                    )}
                  </div>
                  <div>
                    <div className="text-google-gray-500 flex items-center space-x-1">
                      <Calendar className="h-3 w-3" />
                      <span>Active</span>
                    </div>
                    <div className="font-mono font-semibold text-google-gray-200 mt-0.5">
                      {entry.active_days ?? 0}d / {days}d
                    </div>
                  </div>
                  <div>
                    <div className="text-google-gray-500 flex items-center space-x-1">
                      <Cpu className="h-3 w-3" />
                      <span>Top Model</span>
                    </div>
                    <div
                      className="font-mono font-semibold text-google-gray-200 mt-0.5 truncate"
                      title={entry.top_model}
                    >
                      {entry.top_model || 'N/A'}
                    </div>
                    {entry.top_effort && (
                      <div className="mt-0.5">
                        <span className="text-[10px] px-1.5 py-0.2 rounded bg-purple-500/10 text-purple-300 border border-purple-500/20 font-mono">
                          {entry.top_effort}
                        </span>
                      </div>
                    )}
                  </div>
                </div>
              </div>
            );
          })}
        </div>
      )}

      {/* Ranked Leaderboard Table */}
      <div className="bg-[#14171f] border border-[#222836] rounded-xl overflow-hidden shadow-sm">
        {/* Controls Bar */}
        <div className="p-4 border-b border-[#222836] flex flex-col sm:flex-row items-stretch sm:items-center justify-between gap-3">
          <div className="relative flex-1 max-w-md">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-google-gray-500" />
            <input
              type="text"
              placeholder="Search developer in leaderboard..."
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              className="w-full bg-[#1a1e28] border border-[#282f40] rounded-lg pl-9 pr-4 py-2 text-xs sm:text-sm text-white placeholder-google-gray-500 focus:outline-none focus:border-google-blue"
            />
          </div>
          <div className="flex items-center space-x-3 text-xs text-google-gray-400">
            <span className="flex items-center space-x-1.5">
              <span className="w-2.5 h-2.5 rounded-full bg-google-blue inline-block" />
              <span>Gemini Pro</span>
            </span>
            <span className="flex items-center space-x-1.5">
              <span className="w-2.5 h-2.5 rounded-full bg-emerald-400 inline-block" />
              <span>Gemini Flash</span>
            </span>
            <span className="flex items-center space-x-1.5">
              <span className="w-2.5 h-2.5 rounded-full bg-amber-400 inline-block" />
              <span>Claude / Other</span>
            </span>
          </div>
        </div>

        {/* Table */}
        <div className="overflow-x-auto">
          <table className="w-full text-left text-xs sm:text-sm">
            <thead className="bg-[#181c26] text-google-gray-400 text-[11px] uppercase tracking-wider border-b border-[#222836]">
              <tr>
                <th className="py-3 px-4 font-semibold w-16">Rank</th>
                <th className="py-3 px-4 font-semibold">Developer</th>
                <th
                  onClick={() => toggleSort('cost')}
                  className="py-3 px-4 font-semibold cursor-pointer hover:text-white text-right"
                >
                  <div className="flex items-center justify-end space-x-1">
                    <span>Spend</span>
                    <ArrowUpDown className="h-3 w-3" />
                  </div>
                </th>
                <th className="py-3 px-4 font-semibold text-center">Spend Share</th>
                <th
                  onClick={() => toggleSort('tokens')}
                  className="py-3 px-4 font-semibold cursor-pointer hover:text-white text-right"
                >
                  <div className="flex items-center justify-end space-x-1">
                    <span>Tokens</span>
                    <ArrowUpDown className="h-3 w-3" />
                  </div>
                </th>
                <th className="py-3 px-4 font-semibold text-right">Cache Hit %</th>
                <th
                  onClick={() => toggleSort('active_days')}
                  className="py-3 px-4 font-semibold cursor-pointer hover:text-white text-right"
                >
                  <div className="flex items-center justify-end space-x-1">
                    <span>Active Days</span>
                    <ArrowUpDown className="h-3 w-3" />
                  </div>
                </th>
                <th
                  onClick={() => toggleSort('efficiency')}
                  className="py-3 px-4 font-semibold cursor-pointer hover:text-white text-right"
                >
                  <div className="flex items-center justify-end space-x-1">
                    <span>Cost / 1K</span>
                    <ArrowUpDown className="h-3 w-3" />
                  </div>
                </th>
                <th className="py-3 px-4 font-semibold">Model Mix</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-[#1e2330]">
              {filteredAndSorted.length === 0 ? (
                <tr>
                  <td colSpan={9} className="py-8 text-center text-google-gray-500">
                    No developers found matching your filter.
                  </td>
                </tr>
              ) : (
                filteredAndSorted.map((entry) => {
                  const spendSharePct = getSpendSharePct(entry);
                  const modelEntries = Object.entries(entry.model_breakdown || {}).sort(
                    (a, b) => getModelSortWeight(b[1]) - getModelSortWeight(a[1])
                  );
                  return (
                    <tr
                      key={entry.user_id}
                      data-user-id={entry.user_id}
                      onClick={() => onSelectUser(entry.user_id)}
                      className="hover:bg-[#1a1e28]/70 transition-colors cursor-pointer group"
                    >
                      <td className="py-3 px-4 font-mono font-bold">
                        {entry.rank === 1 ? (
                          <span className="inline-flex items-center justify-center w-6 h-6 rounded-full bg-amber-500/20 text-amber-400 border border-amber-500/40 text-xs">
                            #1
                          </span>
                        ) : entry.rank === 2 ? (
                          <span className="inline-flex items-center justify-center w-6 h-6 rounded-full bg-slate-400/20 text-slate-300 border border-slate-400/40 text-xs">
                            #2
                          </span>
                        ) : entry.rank === 3 ? (
                          <span className="inline-flex items-center justify-center w-6 h-6 rounded-full bg-orange-500/20 text-orange-400 border border-orange-500/40 text-xs">
                            #3
                          </span>
                        ) : (
                          <span className="text-google-gray-400 pl-1.5">#{entry.rank}</span>
                        )}
                      </td>
                      <td className="py-3 px-4 font-medium text-white">
                        <div className="flex items-center space-x-2">
                          <span className="w-6 h-6 rounded-full bg-google-blue/20 text-google-blue flex items-center justify-center text-xs font-bold uppercase flex-shrink-0">
                            {(entry.user_id || '?')[0]}
                          </span>
                          <span className="group-hover:text-google-blue transition-colors">
                            {entry.user_id}
                          </span>
                        </div>
                      </td>
                      <td className="py-3 px-4 text-right font-semibold text-white font-mono">
                        <div>{formatCurrency(entry.total_cost ?? 0, entry.currency)}</div>
                        {((entry.included_cost ?? 0) > 0 || (entry.payg_cost ?? 0) > 0) && (
                          <div className="text-[10px] font-normal text-google-gray-400">
                            Inc: {formatCurrency(entry.included_cost ?? 0, entry.currency)} · PAYG:{' '}
                            {formatCurrency(entry.payg_cost ?? 0, entry.currency)}
                          </div>
                        )}
                      </td>
                      <td className="py-3 px-4 text-center">
                        <div className="inline-flex items-center space-x-2">
                          <div className="w-16 h-1.5 bg-[#202532] rounded-full overflow-hidden">
                            <div
                              className="h-full bg-google-blue rounded-full"
                              style={{ width: `${Math.min(spendSharePct, 100)}%` }}
                            />
                          </div>
                          <span className="text-xs text-google-gray-400 font-mono w-12 text-right">
                            {spendSharePct.toFixed(1)}%
                          </span>
                        </div>
                      </td>
                      <td className="py-3 px-4 text-right text-google-gray-200 font-mono">
                        <div>{Number(entry.total_tokens ?? 0).toLocaleString()}</div>
                        {(entry.wtu ?? 0) > 0 && (
                          <div className="text-[10px] text-indigo-400">
                            {formatCompactTokens(entry.wtu ?? 0)} WTU
                          </div>
                        )}
                      </td>
                      <td className="py-3 px-4 text-right font-mono text-emerald-400">
                        {(entry.cache_hit_rate ?? 0) > 0
                          ? `${Number(entry.cache_hit_rate).toFixed(1)}%`
                          : '—'}
                      </td>
                      <td className="py-3 px-4 text-right font-mono text-google-gray-300">
                        <span className="inline-flex items-center space-x-1">
                          <Flame className="h-3.5 w-3.5 text-amber-400/80" />
                          <span>{entry.active_days ?? 0}d</span>
                        </span>
                      </td>
                      <td className="py-3 px-4 text-right font-mono text-google-gray-300">
                        {formatCurrency(entry.cost_per_1k_tokens ?? 0, entry.currency, {
                          minimumFractionDigits: 4,
                          maximumFractionDigits: 4,
                        })}
                      </td>
                      <td className="py-3 px-4 min-w-[160px]">
                        <div className="space-y-1">
                          <div
                            data-testid="model-mix-bar"
                            className="w-full h-2 bg-[#202532] rounded-full overflow-hidden flex"
                          >
                            {modelEntries.map(([model, detail], idx) => {
                              const pct = getModelSharePct(detail);
                              return (
                                <div
                                  key={model}
                                  className={`${getModelColorClass(model, idx)} h-full`}
                                  style={{ width: `${Math.max(pct, 4)}%` }}
                                  title={`${model}: ${pct.toFixed(1)}%`}
                                />
                              );
                            })}
                          </div>
                          <div className="flex items-center justify-between gap-1 text-[11px] text-google-gray-400 font-mono">
                            <span className="truncate">{entry.top_model}</span>
                            {entry.top_effort && (
                              <span className="px-1.5 py-0.2 rounded bg-purple-500/10 text-purple-300 border border-purple-500/20 text-[10px]">
                                {entry.top_effort}
                              </span>
                            )}
                          </div>
                        </div>
                      </td>
                    </tr>
                  );
                })
              )}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
};
