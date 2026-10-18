import React from 'react';
import { TokenCostShare, ModelEffortStats, ClientStats } from '../api';
import { formatCurrency } from '../utils/currency';
import { Cpu, Brain, Terminal, Layers } from 'lucide-react';

interface TokensEffortClientSectionProps {
  tokenCostShares?: TokenCostShare[];
  modelEffortStats?: ModelEffortStats[];
  clientBreakdown?: ClientStats[];
  currency?: string;
}

const getCategoryColor = (categoryOrLabel: string): string => {
  const lower = (categoryOrLabel || '').toLowerCase();
  if (lower.includes('uncached')) return '#4285f4';
  if (lower.includes('cached')) return '#34a853';
  if (lower.includes('output')) return '#fbbc04';
  if (lower.includes('thinking') || lower.includes('thought')) return '#ea4335';
  return '#8ab4f8';
};

const getEffortBadgeClass = (effort: string): string => {
  switch ((effort || '').toLowerCase()) {
    case 'max':
    case 'xhigh':
      return 'bg-google-red/15 text-google-red border-google-red/30';
    case 'high':
      return 'bg-amber-500/15 text-amber-400 border-amber-500/30';
    case 'medium':
      return 'bg-google-blue/15 text-google-blue border-google-blue/30';
    case 'low':
    case 'minimal':
      return 'bg-google-green/15 text-google-green border-google-green/30';
    default:
      return 'bg-[#1e2330] text-google-gray-300 border-[#2e3748]';
  }
};

const formatCompactNumber = (num?: number): string => {
  const safeNum = Number(num ?? 0);
  if (safeNum >= 1_000_000) return `${(safeNum / 1_000_000).toFixed(2)}M`;
  if (safeNum >= 1_000) return `${(safeNum / 1_000).toFixed(1)}k`;
  return safeNum.toLocaleString();
};

export const TokensEffortClientSection: React.FC<TokensEffortClientSectionProps> = ({
  tokenCostShares = [],
  modelEffortStats = [],
  clientBreakdown = [],
  currency = 'USD',
}) => {
  const hasShares = tokenCostShares && tokenCostShares.length > 0;
  const hasEffort = modelEffortStats && modelEffortStats.length > 0;
  const hasClients = clientBreakdown && clientBreakdown.length > 0;

  if (!hasShares && !hasEffort && !hasClients) {
    return null;
  }

  const totalClientCost = (clientBreakdown || []).reduce(
    (sum, c) => sum + Number(c.total_cost ?? c.allocated_cost ?? 0),
    0
  );

  return (
    <div className="space-y-6">
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Card 1: Where Tokens and Money Go */}
        {hasShares && (
          <div className="bg-[#14171f] border border-[#222836] rounded-xl p-5 shadow-sm flex flex-col justify-between">
            <div>
              <div className="flex items-center justify-between mb-1">
                <h3 className="text-sm font-semibold text-white flex items-center space-x-1.5">
                  <Layers className="h-4 w-4 text-google-blue" />
                  <span>Where Tokens and Money Go</span>
                </h3>
                <span className="text-[10px] font-mono uppercase px-1.5 py-0.5 rounded bg-indigo-500/15 text-indigo-300 border border-indigo-500/30">
                  Volume vs Spend
                </span>
              </div>
              <p className="text-xs text-google-gray-400 mb-4">
                Compare Token Share (%) against Reconciled Cost Share (%) across prompt cache &amp; reasoning tiers
              </p>

              <div className="space-y-3.5">
                {tokenCostShares.map((item) => {
                  const displayLabel = item.label || item.category;
                  const color = getCategoryColor(displayLabel);
                  const itemCost = Number(item.cost ?? item.cost_usd ?? 0);
                  const tokenShare = Number(item.token_share ?? 0);
                  const costShare = Number(item.cost_share ?? 0);
                  return (
                    <div
                      key={item.category}
                      className="p-2.5 rounded-lg bg-[#181c26] border border-[#202530]"
                    >
                      <div className="flex items-center justify-between text-xs mb-1.5">
                        <div className="flex items-center space-x-2">
                          <span
                            className="w-2.5 h-2.5 rounded-full flex-shrink-0"
                            style={{ backgroundColor: color }}
                          />
                          <span className="font-medium text-white">{displayLabel}</span>
                        </div>
                        <span className="font-mono text-xs text-google-gray-300">
                          {formatCompactNumber(item.tokens)} tok ·{' '}
                          <strong className="text-white">{formatCurrency(itemCost, currency)}</strong>
                        </span>
                      </div>

                      {/* Token Share Bar */}
                      <div className="flex items-center space-x-2 text-[11px] font-mono mb-1">
                        <span className="w-14 text-google-gray-400">Tokens</span>
                        <div className="flex-1 h-1.5 bg-[#202530] rounded-full overflow-hidden">
                          <div
                            className="h-full rounded-full opacity-75"
                            style={{
                              width: `${Math.min(Math.max(tokenShare, 0), 100)}%`,
                              backgroundColor: color,
                            }}
                          />
                        </div>
                        <span className="w-12 text-right text-google-gray-300">
                          {tokenShare.toFixed(1)}%
                        </span>
                      </div>

                      {/* Cost Share Bar */}
                      <div className="flex items-center space-x-2 text-[11px] font-mono">
                        <span className="w-14 text-google-gray-400">Spend</span>
                        <div className="flex-1 h-1.5 bg-[#202530] rounded-full overflow-hidden">
                          <div
                            className="h-full rounded-full"
                            style={{
                              width: `${Math.min(Math.max(costShare, 0), 100)}%`,
                              backgroundColor: color,
                            }}
                          />
                        </div>
                        <span className="w-12 text-right font-semibold text-white">
                          {costShare.toFixed(1)}%
                        </span>
                      </div>
                    </div>
                  );
                })}
              </div>
            </div>
          </div>
        )}

        {/* Card 2: Models & Thinking Effort */}
        {hasEffort && (
          <div className="lg:col-span-2 bg-[#14171f] border border-[#222836] rounded-xl p-5 shadow-sm flex flex-col justify-between">
            <div>
              <div className="flex items-center justify-between mb-1">
                <h3 className="text-sm font-semibold text-white flex items-center space-x-1.5">
                  <Brain className="h-4 w-4 text-amber-400" />
                  <span>Models &amp; Thinking Effort</span>
                </h3>
                <span className="text-[10px] font-mono text-google-gray-400">
                  WTU · Cache Hit % · Included vs PAYG
                </span>
              </div>
              <p className="text-xs text-google-gray-400 mb-4">
                Breakdown by foundation model and reasoning effort tier (minimal to max)
              </p>

              <div className="overflow-x-auto">
                <table className="w-full text-left border-collapse text-xs">
                  <thead>
                    <tr className="border-b border-[#222836] text-[10px] font-semibold uppercase tracking-wider text-google-gray-500">
                      <th className="py-2 pr-3">Model</th>
                      <th className="py-2 px-2">Effort</th>
                      <th className="py-2 px-2 text-right">Tokens</th>
                      <th className="py-2 px-2 text-right">Cache Hit</th>
                      <th className="py-2 px-2 text-right">WTU</th>
                      <th className="py-2 px-2 text-right">Included / PAYG</th>
                      <th className="py-2 pl-2 text-right">Total Cost</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-[#1e2330] font-mono">
                    {modelEffortStats.slice(0, 8).map((row, idx) => {
                      const rowTokens = Number(row.tokens ?? row.total_tokens ?? 0);
                      const rowCost = Number(row.total_cost ?? row.allocated_cost ?? 0);
                      const rowCacheHit = Number(row.cache_hit_rate ?? 0);
                      return (
                        <tr
                          key={`${row.model}-${row.effort_level}-${idx}`}
                          className="hover:bg-[#1a1f2c]/60 transition-colors"
                        >
                          <td className="py-2.5 pr-3 font-sans font-medium text-white">
                            {row.model}
                          </td>
                          <td className="py-2.5 px-2">
                            <span
                              className={`inline-flex items-center px-2 py-0.5 rounded text-[10px] font-semibold border ${getEffortBadgeClass(
                                row.effort_level
                              )}`}
                            >
                              {row.effort_level}
                            </span>
                          </td>
                          <td className="py-2.5 px-2 text-right text-google-gray-300">
                            {formatCompactNumber(rowTokens)}
                          </td>
                          <td className="py-2.5 px-2 text-right text-cyan-300">
                            {rowCacheHit.toFixed(1)}%
                          </td>
                          <td className="py-2.5 px-2 text-right text-indigo-300">
                            {formatCompactNumber(row.wtu)}
                          </td>
                          <td className="py-2.5 px-2 text-right text-[11px]">
                            <span className="text-google-green">
                              {formatCurrency(row.included_cost ?? 0, currency)}
                            </span>
                            <span className="text-google-gray-500 mx-1">/</span>
                            <span className="text-amber-400">
                              {formatCurrency(row.payg_cost ?? 0, currency)}
                            </span>
                          </td>
                          <td className="py-2.5 pl-2 text-right font-semibold text-white">
                            {formatCurrency(rowCost, currency)}
                          </td>
                        </tr>
                      );
                    })}
                  </tbody>
                </table>
              </div>
            </div>

            {/* Bottom Sub-Section: Antigravity Clients */}
            {hasClients && (
              <div className="mt-5 pt-4 border-t border-[#222836]">
                <div className="flex items-center justify-between mb-2.5">
                  <h4 className="text-xs font-semibold text-white flex items-center space-x-1.5">
                    <Terminal className="h-3.5 w-3.5 text-google-green" />
                    <span>Antigravity Clients</span>
                  </h4>
                  <span className="text-[10px] font-mono text-google-gray-400">
                    IDE &amp; CLI Surfaces
                  </span>
                </div>
                <div className="grid grid-cols-1 sm:grid-cols-3 gap-2.5">
                  {clientBreakdown.map((client, idx) => {
                    const clientTokens = Number(client.tokens ?? client.total_tokens ?? 0);
                    const clientCost = Number(client.total_cost ?? client.allocated_cost ?? 0);
                    const clientShare =
                      client.share_of_cost !== undefined
                        ? Number(client.share_of_cost)
                        : client.token_share !== undefined
                        ? Number(client.token_share)
                        : totalClientCost > 0
                        ? (clientCost / totalClientCost) * 100
                        : 0;
                    const displayName = client.client_version
                      ? `${client.client_name} (${client.client_version})`
                      : client.client_name;
                    return (
                      <div
                        key={`${client.client_name}-${client.client_version || idx}`}
                        className="p-2.5 rounded-lg bg-[#181c26] border border-[#202530] flex items-center justify-between"
                      >
                        <div className="flex items-center space-x-2">
                          <Cpu className="h-3.5 w-3.5 text-google-blue flex-shrink-0" />
                          <div>
                            <div className="text-xs font-semibold text-white">
                              {displayName}
                            </div>
                            <div className="text-[10px] font-mono text-google-gray-400">
                              {formatCompactNumber(clientTokens)} tok · {formatCompactNumber(client.wtu)} WTU
                            </div>
                          </div>
                        </div>
                        <div className="text-right font-mono">
                          <div className="text-xs font-bold text-white">
                            {formatCurrency(clientCost, currency)}
                          </div>
                          <div className="text-[10px] text-google-blue">
                            {clientShare.toFixed(1)}%
                          </div>
                        </div>
                      </div>
                    );
                  })}
                </div>
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  );
};
