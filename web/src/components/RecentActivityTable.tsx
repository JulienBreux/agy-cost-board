import { FC, useState, useMemo } from 'react';
import type { UserActivityLog } from '../api';

interface RecentActivityTableProps {
  logs: UserActivityLog[];
  currency?: string;
  isLoading?: boolean;
}

export const RecentActivityTable: FC<RecentActivityTableProps> = ({
  logs,
  currency = '$',
  isLoading = false,
}) => {
  const [selectedModel, setSelectedModel] = useState<string>('all');

  const models = useMemo(() => {
    const set = new Set<string>();
    logs.forEach((log) => {
      if (log.model) set.add(log.model);
    });
    return Array.from(set).sort();
  }, [logs]);

  const filteredLogs = useMemo(() => {
    if (selectedModel === 'all') return logs;
    return logs.filter((log) => log.model === selectedModel);
  }, [logs, selectedModel]);

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
    <div className="bg-gray-900 border border-gray-800 rounded-xl p-5 shadow-sm">
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 pb-4 border-b border-gray-800/80">
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
            <h3 className="text-base font-semibold text-gray-100">Recent Activity</h3>
            <p className="text-xs text-gray-400">Live stream of telemetry requests and token consumption</p>
          </div>
        </div>

        {models.length > 0 && (
          <div className="flex items-center gap-2">
            <label htmlFor="model-filter-select" className="text-xs text-gray-400">
              Filter:
            </label>
            <select
              id="model-filter-select"
              aria-label="Filter by model"
              value={selectedModel}
              onChange={(e) => setSelectedModel(e.target.value)}
              className="bg-gray-800 border border-gray-700 text-gray-200 text-xs rounded-lg px-2.5 py-1.5 focus:outline-none focus:ring-1 focus:ring-emerald-500"
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
        <div className="py-12 text-center text-gray-400 text-sm">Loading activity logs...</div>
      ) : filteredLogs.length === 0 ? (
        <div className="py-12 text-center">
          <div className="w-12 h-12 mx-auto mb-3 rounded-full bg-gray-800/60 flex items-center justify-center text-gray-500">
            <svg className="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
              />
            </svg>
          </div>
          <p className="text-sm font-medium text-gray-300">No recent activity recorded</p>
          <p className="text-xs text-gray-500 mt-1">Activity logs will stream in as you interact with Gemini models.</p>
        </div>
      ) : (
        <div className="mt-4 overflow-x-auto">
          <table className="w-full text-left text-xs text-gray-300">
            <thead className="text-[11px] uppercase tracking-wider text-gray-400 bg-gray-800/50 border-b border-gray-800">
              <tr>
                <th className="py-2.5 px-3">Time</th>
                <th className="py-2.5 px-3">Model</th>
                <th className="py-2.5 px-3 text-right">Input</th>
                <th className="py-2.5 px-3 text-right">Output</th>
                <th className="py-2.5 px-3 text-right">Total Tokens</th>
                <th className="py-2.5 px-3 text-right">Cost</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-800/60 font-mono">
              {filteredLogs.map((log, idx) => (
                <tr key={idx} className="hover:bg-gray-800/30 transition">
                  <td className="py-2.5 px-3 text-gray-400 font-sans whitespace-nowrap">
                    {formatTimestamp(log.timestamp)}
                  </td>
                  <td className="py-2.5 px-3">
                    <span className="inline-block px-2 py-0.5 rounded text-[11px] bg-gray-800 text-gray-200 border border-gray-700">
                      {log.model}
                    </span>
                  </td>
                  <td className="py-2.5 px-3 text-right text-gray-400">
                    {log.input_tokens.toLocaleString()}
                  </td>
                  <td className="py-2.5 px-3 text-right text-gray-400">
                    {log.output_tokens.toLocaleString()}
                  </td>
                  <td className="py-2.5 px-3 text-right font-semibold text-gray-200">
                    {log.total_tokens.toLocaleString()}
                  </td>
                  <td className="py-2.5 px-3 text-right font-semibold text-emerald-400 whitespace-nowrap">
                    {currency}{log.estimated_cost.toFixed(4)}
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
