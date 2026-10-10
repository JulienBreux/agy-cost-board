import { FC, useState, useMemo } from 'react';
import type { UserActivityLog } from '../api';

interface RecentActivityTableProps {
  logs?: UserActivityLog[];
  currency?: string;
  isLoading?: boolean;
}

export const RecentActivityTable: FC<RecentActivityTableProps> = ({
  logs = [],
  currency = '$',
  isLoading = false,
}) => {
  const [selectedModel, setSelectedModel] = useState<string>('all');
  const safeLogs = Array.isArray(logs) ? logs : [];
  const currencySymbol = currency === 'EUR' ? '€' : currency === 'USD' ? '$' : currency;

  const models = useMemo(() => {
    const set = new Set<string>();
    safeLogs.forEach((log) => {
      if (log.model) set.add(log.model);
    });
    return Array.from(set).sort();
  }, [safeLogs]);

  const filteredLogs = useMemo(() => {
    if (selectedModel === 'all') return safeLogs;
    return safeLogs.filter((log) => log.model === selectedModel);
  }, [safeLogs, selectedModel]);

  const formatTimestamp = (iso: string) => {
    try {
      const d = new Date(iso);
      return d.toLocaleDateString(undefined, {
        month: 'short',
        day: 'numeric',
        hour: '2-digit',
        minute: '2-digit',
      });
    } catch {
      return iso;
    }
  };

  return (
    <div className="bg-[#14171f] border border-[#222836] rounded-xl p-5 shadow-sm">
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 pb-4 border-b border-[#202530]">
        <div className="flex items-center gap-2.5">
          <div className="p-2 rounded-lg bg-blue-500/10 text-blue-400 border border-blue-500/20">
            <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"
              />
            </svg>
          </div>
          <div>
            <h3 className="text-base font-semibold text-white">Recent Activity</h3>
            <p className="text-xs text-google-gray-400">Live stream of telemetry requests and token consumption</p>
          </div>
        </div>

        {models.length > 0 && (
          <div className="flex items-center gap-2">
            <label htmlFor="model-filter-select" className="text-xs text-google-gray-400">
              Filter:
            </label>
            <select
              id="model-filter-select"
              aria-label="Filter by model"
              value={selectedModel}
              onChange={(e) => setSelectedModel(e.target.value)}
              className="bg-[#1a1e28] border border-[#202530] text-google-gray-200 text-xs rounded-lg px-2.5 py-1.5 focus:outline-none focus:ring-1 focus:ring-emerald-500"
            >
              <option value="all">All Models</option>
              {models.map((m) => (
                <option key={m} value={m}>
                  {m}
                </option>
              ))}
            </select>
          </div>
        )}
      </div>

      {isLoading ? (
        <div className="py-12 text-center text-google-gray-400 text-sm">Loading activity logs...</div>
      ) : filteredLogs.length === 0 ? (
        <div className="py-12 text-center">
          <div className="w-12 h-12 mx-auto mb-3 rounded-full bg-[#1e2330] flex items-center justify-center text-google-gray-500">
            <svg className="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
              />
            </svg>
          </div>
          <p className="text-sm font-medium text-google-gray-300">No recent activity recorded</p>
          <p className="text-xs text-google-gray-500 mt-1">Activity logs will stream in as you interact with Gemini models.</p>
        </div>
      ) : (
        <div className="mt-4 overflow-x-auto">
          <table className="w-full text-left text-xs text-google-gray-300">
            <thead className="text-[11px] uppercase tracking-wider text-google-gray-400 bg-[#1a1e28] border-b border-[#202530]">
              <tr>
                <th className="py-2.5 px-3">Time</th>
                <th className="py-2.5 px-3">Model</th>
                <th className="py-2.5 px-3 text-right">Input</th>
                <th className="py-2.5 px-3 text-right">Output</th>
                <th className="py-2.5 px-3 text-right">Total Tokens</th>
                <th className="py-2.5 px-3 text-right">Cost</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-[#202530] font-mono">
              {filteredLogs.map((log, idx) => (
                <tr key={idx} className="hover:bg-[#1a1e28]/70 transition">
                  <td className="py-2.5 px-3 text-google-gray-400 font-sans whitespace-nowrap">
                    {formatTimestamp(log.timestamp)}
                  </td>
                  <td className="py-2.5 px-3">
                    <span className="inline-block px-2 py-0.5 rounded text-[11px] bg-[#1e2330] text-google-gray-200 border border-[#202530]">
                      {log.model}
                    </span>
                  </td>
                  <td className="py-2.5 px-3 text-right text-google-gray-400">
                    {(log.input_tokens ?? 0).toLocaleString()}
                  </td>
                  <td className="py-2.5 px-3 text-right text-google-gray-400">
                    {(log.output_tokens ?? 0).toLocaleString()}
                  </td>
                  <td className="py-2.5 px-3 text-right font-semibold text-white">
                    {(log.total_tokens ?? 0).toLocaleString()}
                  </td>
                  <td className="py-2.5 px-3 text-right font-semibold text-emerald-400 whitespace-nowrap">
                    {currencySymbol}{(log.estimated_cost ?? 0).toFixed(4)}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
};
