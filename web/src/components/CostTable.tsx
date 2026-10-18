import React, { useState, useMemo } from 'react';
import { Search, Filter, ArrowUpDown, Download } from 'lucide-react';
import { AllocatedUserCost } from '../api';
import { formatCurrency } from '../utils/currency';

interface CostTableProps {
  costs: AllocatedUserCost[];
  onSelectUser: (userId: string) => void;
}

type PaygMode = 'fair_share' | 'when_used';

export const CostTable: React.FC<CostTableProps> = ({ costs, onSelectUser }) => {
  const [search, setSearch] = useState('');
  const [selectedModel, setSelectedModel] = useState<string>('all');
  const [sortBy, setSortBy] = useState<'cost' | 'tokens' | 'date'>('cost');
  const [sortOrder, setSortOrder] = useState<'asc' | 'desc'>('desc');
  const [paygMode, setPaygMode] = useState<PaygMode>('fair_share');
  const [page, setPage] = useState(1);
  const pageSize = 15;

  const getIncludedCost = (r: AllocatedUserCost): number => {
    if (paygMode === 'when_used') {
      return r.when_used_included_cost ?? r.included_cost ?? 0;
    }
    return r.fair_share_included_cost ?? r.included_cost ?? 0;
  };

  const getPaygCost = (r: AllocatedUserCost): number => {
    if (paygMode === 'when_used') {
      return r.when_used_payg_cost ?? r.payg_cost ?? 0;
    }
    return r.fair_share_payg_cost ?? r.payg_cost ?? 0;
  };

  const safeCosts = Array.isArray(costs) ? costs : [];

  const models = useMemo(() => {
    const set = new Set<string>();
    safeCosts.forEach((c) => {
      if (c?.model) set.add(c.model);
    });
    return Array.from(set);
  }, [safeCosts]);

  const filteredAndSorted = useMemo(() => {
    return safeCosts
      .filter((c) => {
        const matchesUser = (c?.user_id || '').toLowerCase().includes(search.toLowerCase());
        const matchesModel = selectedModel === 'all' || c?.model === selectedModel;
        return matchesUser && matchesModel;
      })
      .sort((a, b) => {
        let diff = 0;
        if (sortBy === 'cost') diff = a.allocated_cost - b.allocated_cost;
        if (sortBy === 'tokens') diff = a.user_tokens - b.user_tokens;
        if (sortBy === 'date') diff = a.usage_date.localeCompare(b.usage_date);
        return sortOrder === 'desc' ? -diff : diff;
      });
  }, [safeCosts, search, selectedModel, sortBy, sortOrder]);

  const totalPages = Math.ceil(filteredAndSorted.length / pageSize) || 1;
  const paginatedRows = filteredAndSorted.slice((page - 1) * pageSize, page * pageSize);

  const toggleSort = (col: 'cost' | 'tokens' | 'date') => {
    if (sortBy === col) {
      setSortOrder(sortOrder === 'desc' ? 'asc' : 'desc');
    } else {
      setSortBy(col);
      setSortOrder('desc');
    }
  };

  const handleExportCsv = () => {
    const headers = [
      'usage_date',
      'user_id',
      'model',
      'user_tokens',
      'wtu',
      'cache_hit_rate',
      'token_share',
      'allocated_cost',
      'payg_mode',
      'included_cost',
      'payg_cost',
      'fair_share_included_cost',
      'fair_share_payg_cost',
      'when_used_included_cost',
      'when_used_payg_cost',
      'currency',
    ];
    const rows = filteredAndSorted.map((r) => [
      r.usage_date,
      `"${(r.user_id || '').replace(/"/g, '""')}"`,
      `"${(r.model || '').replace(/"/g, '""')}"`,
      r.user_tokens ?? 0,
      Math.round(r.wtu ?? 0),
      (r.cache_hit_rate ?? 0).toFixed(2),
      (r.token_share ?? 0).toFixed(4),
      (r.allocated_cost ?? 0).toFixed(4),
      paygMode,
      getIncludedCost(r).toFixed(4),
      getPaygCost(r).toFixed(4),
      (r.fair_share_included_cost ?? r.included_cost ?? 0).toFixed(4),
      (r.fair_share_payg_cost ?? r.payg_cost ?? 0).toFixed(4),
      (r.when_used_included_cost ?? r.included_cost ?? 0).toFixed(4),
      (r.when_used_payg_cost ?? r.payg_cost ?? 0).toFixed(4),
      r.currency || 'USD',
    ]);
    const csv = [headers.join(','), ...rows.map((row) => row.join(','))].join('\n');
    try {
      const blob = new Blob([csv], { type: 'text/csv;charset=utf-8;' });
      if (typeof URL !== 'undefined' && typeof URL.createObjectURL === 'function') {
        const url = URL.createObjectURL(blob);
        const link = document.createElement('a');
        link.href = url;
        link.setAttribute('download', `attributed-costs-${paygMode}.csv`);
        document.body.appendChild(link);
        link.click();
        document.body.removeChild(link);
        if (typeof URL.revokeObjectURL === 'function') {
          URL.revokeObjectURL(url);
        }
      }
    } catch {
      // No-op in headless test environments
    }
  };

  return (
    <div className="bg-[#14171f] border border-[#222836] rounded-xl overflow-hidden shadow-sm">
      {/* Controls Bar */}
      <div className="p-4 border-b border-[#222836] flex flex-col sm:flex-row items-stretch sm:items-center justify-between gap-3">
        {/* Search */}
        <div className="relative flex-1 max-w-md">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-google-gray-500" />
          <input
            type="text"
            placeholder="Filter by developer email..."
            value={search}
            onChange={(e) => {
              setSearch(e.target.value);
              setPage(1);
            }}
            className="w-full bg-[#1a1e28] border border-[#282f40] rounded-lg pl-9 pr-4 py-2 text-xs sm:text-sm text-white placeholder-google-gray-500 focus:outline-none focus:border-google-blue"
          />
        </div>

        {/* Filters + Chargeback Mode + CSV Export */}
        <div className="flex flex-wrap items-center gap-2">
          <div className="inline-flex items-center bg-[#181c26] border border-[#282f40] rounded-lg p-0.5 text-xs">
            <button
              type="button"
              onClick={() => setPaygMode('fair_share')}
              className={`px-2.5 py-1 rounded-md font-medium transition-colors ${
                paygMode === 'fair_share'
                  ? 'bg-google-blue text-white shadow-sm'
                  : 'text-google-gray-400 hover:text-white'
              }`}
              title="Distribute 7-day pool overage proportionally across licensed pool users"
            >
              Fair Share
            </button>
            <button
              type="button"
              onClick={() => setPaygMode('when_used')}
              className={`px-2.5 py-1 rounded-md font-medium transition-colors ${
                paygMode === 'when_used'
                  ? 'bg-google-blue text-white shadow-sm'
                  : 'text-google-gray-400 hover:text-white'
              }`}
              title="Attribute PAYG chronologically as the weekly pool exhausts"
            >
              When Used
            </button>
          </div>

          <div className="flex items-center space-x-1.5 bg-[#1a1e28] border border-[#282f40] rounded-lg px-2.5 py-1.5 text-xs text-google-gray-400">
            <Filter className="h-3.5 w-3.5" />
            <select
              value={selectedModel}
              onChange={(e) => {
                setSelectedModel(e.target.value);
                setPage(1);
              }}
              className="bg-transparent text-white focus:outline-none cursor-pointer"
            >
              <option value="all" className="bg-[#1a1e28]">All AI Models</option>
              {models.map((m) => (
                <option key={m} value={m} className="bg-[#1a1e28]">{m}</option>
              ))}
            </select>
          </div>

          <button
            type="button"
            onClick={handleExportCsv}
            className="inline-flex items-center space-x-1.5 bg-[#181c26] hover:bg-[#202532] border border-[#282f40] text-google-gray-200 px-3 py-1.5 rounded-lg text-xs font-medium transition-colors"
          >
            <Download className="h-3.5 w-3.5" />
            <span>Export CSV</span>
          </button>
        </div>
      </div>

      {/* Table */}
      <div className="overflow-x-auto">
        <table className="w-full text-left text-xs sm:text-sm">
          <thead className="bg-[#181c26] text-google-gray-400 text-[11px] uppercase tracking-wider border-b border-[#222836]">
            <tr>
              <th className="py-3 px-4 font-semibold">Developer</th>
              <th className="py-3 px-4 font-semibold">Model</th>
              <th
                onClick={() => toggleSort('date')}
                className="py-3 px-4 font-semibold cursor-pointer hover:text-white"
              >
                <div className="flex items-center space-x-1">
                  <span>Usage Date</span>
                  <ArrowUpDown className="h-3 w-3" />
                </div>
              </th>
              <th
                onClick={() => toggleSort('tokens')}
                className="py-3 px-4 font-semibold cursor-pointer hover:text-white text-right"
              >
                <div className="flex items-center justify-end space-x-1">
                  <span>Inference Tokens</span>
                  <ArrowUpDown className="h-3 w-3" />
                </div>
              </th>
              <th className="py-3 px-4 font-semibold text-right">Cache Hit %</th>
              <th className="py-3 px-4 font-semibold text-center">Model Share</th>
              <th
                onClick={() => toggleSort('cost')}
                className="py-3 px-4 font-semibold cursor-pointer hover:text-white text-right"
              >
                <div className="flex items-center justify-end space-x-1">
                  <span>Attributed Cost</span>
                  <ArrowUpDown className="h-3 w-3" />
                </div>
              </th>
            </tr>
          </thead>
          <tbody className="divide-y divide-[#1e2330]">
            {paginatedRows.length === 0 ? (
              <tr>
                <td colSpan={7} className="py-8 text-center text-google-gray-500">
                  No attributed cost records found matching your filters.
                </td>
              </tr>
            ) : (
              paginatedRows.map((r, idx) => (
                <tr
                  key={`${r.user_id}-${r.model}-${r.usage_date}-${idx}`}
                  className="hover:bg-[#1a1e28]/70 transition-colors cursor-pointer group"
                  onClick={() => onSelectUser(r.user_id)}
                >
                  <td className="py-3 px-4 font-medium text-white flex items-center space-x-2">
                    <span className="w-6 h-6 rounded-full bg-google-blue/20 text-google-blue flex items-center justify-center text-xs font-bold uppercase">
                      {r.user_id[0]}
                    </span>
                    <span>{r.user_id}</span>
                  </td>
                  <td className="py-3 px-4 text-google-gray-300 font-mono text-xs">
                    {r.model}
                  </td>
                  <td className="py-3 px-4 text-google-gray-400">
                    {r.usage_date}
                  </td>
                  <td className="py-3 px-4 text-google-gray-200 text-right font-mono">
                    <div>{r.user_tokens.toLocaleString()}</div>
                    {(r.wtu ?? 0) > 0 && (
                      <div className="text-[10px] text-indigo-400">
                        {Math.round(r.wtu ?? 0).toLocaleString()} WTU
                      </div>
                    )}
                  </td>
                  <td className="py-3 px-4 text-right font-mono text-emerald-400">
                    {(r.cache_hit_rate ?? 0) > 0 ? `${Number(r.cache_hit_rate).toFixed(1)}%` : '—'}
                  </td>
                  <td className="py-3 px-4 text-center">
                    <div className="inline-flex items-center space-x-2">
                      <div className="w-16 h-1.5 bg-[#202532] rounded-full overflow-hidden">
                        <div
                          className="h-full bg-google-blue rounded-full"
                          style={{ width: `${Math.min(r.token_share * 100, 100)}%` }}
                        />
                      </div>
                      <span className="text-xs text-google-gray-400 font-mono">
                        {(r.token_share * 100).toFixed(1)}%
                      </span>
                    </div>
                  </td>
                  <td className="py-3 px-4 text-right font-semibold text-white font-mono">
                    <div>{formatCurrency(r.allocated_cost, r.currency)}</div>
                    {(getIncludedCost(r) > 0 || getPaygCost(r) > 0) && (
                      <div className="text-[10px] font-normal text-google-gray-400">
                        Inc: {formatCurrency(getIncludedCost(r), r.currency)} · PAYG:{' '}
                        {formatCurrency(getPaygCost(r), r.currency)}
                      </div>
                    )}
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      {/* Pagination Footer */}
      <div className="p-3 border-t border-[#222836] bg-[#181c26] flex items-center justify-between text-xs text-google-gray-400">
        <div>
          Showing {paginatedRows.length} of {filteredAndSorted.length} entries
        </div>
        <div className="flex items-center space-x-1">
          <button
            onClick={() => setPage((p) => Math.max(p - 1, 1))}
            disabled={page === 1}
            className="px-2.5 py-1 rounded border border-[#282f40] disabled:opacity-40 hover:bg-[#202532] text-white"
          >
            Prev
          </button>
          <span className="px-2">Page {page} of {totalPages}</span>
          <button
            onClick={() => setPage((p) => Math.min(p + 1, totalPages))}
            disabled={page === totalPages}
            className="px-2.5 py-1 rounded border border-[#282f40] disabled:opacity-40 hover:bg-[#202532] text-white"
          >
            Next
          </button>
        </div>
      </div>
    </div>
  );
};

