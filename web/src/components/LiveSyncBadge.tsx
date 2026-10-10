import { FC } from 'react';

interface LiveSyncBadgeProps {
  isLive?: boolean;
  lastUpdated?: Date | null;
  onRefresh?: () => void;
  isLoading?: boolean;
}

export const LiveSyncBadge: FC<LiveSyncBadgeProps> = ({
  isLive = true,
  lastUpdated,
  onRefresh,
  isLoading = false,
}) => {
  return (
    <div className="flex items-center gap-2.5 bg-[#1a1e28] border border-[#202530] rounded-lg px-3 py-1.5 text-xs shadow-sm">
      <div className="flex items-center gap-1.5">
        {isLive ? (
          <>
            <span className="relative flex h-2 w-2">
              <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75"></span>
              <span className="relative inline-flex rounded-full h-2 w-2 bg-emerald-500"></span>
            </span>
            <span className="font-medium text-emerald-400">Live Telemetry</span>
          </>
        ) : (
          <>
            <span className="h-2 w-2 rounded-full bg-google-gray-500"></span>
            <span className="font-medium text-google-gray-400">Sync Paused</span>
          </>
        )}
      </div>

      {lastUpdated && (
        <span className="text-[11px] text-google-gray-400 border-l border-[#202530] pl-2 hidden sm:inline">
          Updated {lastUpdated.toLocaleTimeString()}
        </span>
      )}

      {onRefresh && (
        <button
          type="button"
          onClick={onRefresh}
          aria-label="Refresh telemetry data"
          disabled={isLoading}
          className="text-google-gray-400 hover:text-white transition p-0.5 rounded hover:bg-[#202530] disabled:opacity-50"
          title="Refresh telemetry data"
        >
          <svg
            className={`w-3.5 h-3.5 ${isLoading ? 'animate-spin text-emerald-400' : ''}`}
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
            />
          </svg>
        </button>
      )}
    </div>
  );
};
