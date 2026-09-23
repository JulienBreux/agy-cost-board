import React, { useState, useMemo } from 'react';
import { Search, Filter, ArrowUpDown, ChevronRight } from 'lucide-react';
import { AllocatedUserCost } from '../api';

interface CostTableProps {
  costs: AllocatedUserCost[];
  onSelectUser: (userId: string) => void;
}

export const CostTable: React.FC<CostTableProps> = ({ costs, onSelectUser }) => {
  const [search, setSearch] = useState('');
  const [selectedModel, setSelectedModel] = useState<string>('all');
  const [sortBy, setSortBy] = useState<'cost' | 'tokens' | 'date'>('cost');
  const [sortOrder, setSortOrder] = useState<'asc' | 'desc'>('desc');
  const [page, setPage] = useState(1);
  const pageSize = 15;

  const models = useMemo(() => {
    const set = new Set<string>();
    costs.forEach((c) => set.add(c.model));
    return Array.from(set);
  }, [costs]);

  const filteredAndSorted = useMemo(() => {
    return costs
      .filter((c) => {
        const matchesUser = c.user_id.toLowerCase().includes(search.toLowerCase());
        const matchesModel = selectedModel === 'all' || c.model === selectedModel;
        return matchesUser && matchesModel;
      })
      .sort((a, b) => {
        let diff = 0;
        if (sortBy === 'cost') diff = a.allocated_cost - b.allocated_cost;
        if (sortBy === 'tokens') diff = a.user_tokens - b.user_tokens;
        if (sortBy === 'date') diff = a.usage_date.localeCompare(b.usage_date);
        return sortOrder === 'desc' ? -diff : diff;
      });
  }, [costs, search, selectedModel, sortBy, sortOrder]);

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

        {/* Filters */}
        <div className="flex items-center space-x-2">
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
              <th className="py-3 px-4 text-center">Action</th>
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
                    {r.user_tokens.toLocaleString()}
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
                    ${r.allocated_cost.toFixed(2)}
                  </td>
                  <td className="py-3 px-4 text-center">
                    <button
                      onClick={(e) => {
                        e.stopPropagation();
                        onSelectUser(r.user_id);
                      }}
                      className="p-1 rounded hover:bg-google-blue/20 text-google-gray-400 hover:text-google-blue transition-colors"
                    >
                      <ChevronRight className="h-4 w-4" />
                    </button>
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
