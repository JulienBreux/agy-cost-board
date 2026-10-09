import React from 'react';
import { X, User, DollarSign, Cpu, Clock, AlertCircle, CheckCircle2 } from 'lucide-react';
import { UserSummary } from '../api';

interface UserModalProps {
  user: UserSummary | null;
  onClose: () => void;
  onOpenDashboard?: (userId: string) => void;
}

export const UserModal: React.FC<UserModalProps> = ({ user, onClose, onOpenDashboard }) => {
  if (!user) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/70 backdrop-blur-sm animate-fade-in">
      <div className="bg-[#14171f] border border-[#282f40] rounded-2xl w-full max-w-xl overflow-hidden shadow-2xl">
        {/* Header */}
        <div className="p-5 border-b border-[#222836] flex items-center justify-between">
          <div className="flex items-center space-x-3">
            <div className="w-10 h-10 rounded-xl bg-google-blue/15 text-google-blue border border-google-blue/30 flex items-center justify-center font-bold text-lg">
              <User className="h-5 w-5" />
            </div>
            <div>
              <h3 className="text-base font-bold text-white tracking-tight">{user.user_id}</h3>
              <div className="flex items-center space-x-2 mt-0.5">
                <span
                  className={`inline-flex items-center space-x-1 text-[11px] font-semibold px-2 py-0.5 rounded ${
                    user.seat_status === 'active'
                      ? 'bg-google-green/15 text-google-green border border-google-green/30'
                      : 'bg-google-red/15 text-google-red border border-google-red/30'
                  }`}
                >
                  {user.seat_status === 'active' ? (
                    <CheckCircle2 className="h-3 w-3" />
                  ) : (
                    <AlertCircle className="h-3 w-3" />
                  )}
                  <span>{user.seat_status.toUpperCase()} SEAT</span>
                </span>
                <span className="text-xs text-google-gray-500">•</span>
                <span className="text-xs text-google-gray-400 flex items-center space-x-1">
                  <Clock className="h-3 w-3" />
                  <span>
                    Last active: {user.last_active ? new Date(user.last_active).toLocaleString() : 'Never'}
                  </span>
                </span>
              </div>
            </div>
          </div>
          <button
            onClick={onClose}
            className="p-1.5 rounded-lg text-google-gray-400 hover:text-white hover:bg-[#1e2330] transition-colors"
          >
            <X className="h-5 w-5" />
          </button>
        </div>

        {/* Content */}
        <div className="p-6 space-y-6">
          {/* Top Metric Cards */}
          <div className="grid grid-cols-2 gap-4">
            <div className="bg-[#181c26] border border-[#222836] rounded-xl p-4">
              <span className="text-xs text-google-gray-400">Total Attributed Cost</span>
              <div className="text-2xl font-bold text-white mt-1 flex items-baseline space-x-1">
                <DollarSign className="h-5 w-5 text-google-blue" />
                <span>{user.total_cost.toFixed(2)}</span>
                <span className="text-xs text-google-gray-500 font-normal">{user.currency}</span>
              </div>
            </div>

            <div className="bg-[#181c26] border border-[#222836] rounded-xl p-4">
              <span className="text-xs text-google-gray-400">Total Tokens</span>
              <div className="text-2xl font-bold text-white mt-1 flex items-baseline space-x-1">
                <Cpu className="h-5 w-5 text-indigo-400" />
                <span>{user.total_tokens.toLocaleString()}</span>
              </div>
            </div>
          </div>

          {/* Model Breakdown */}
          <div>
            <h4 className="text-xs font-semibold text-google-gray-400 uppercase tracking-wider mb-3">
              AI Model Consumption Breakdown
            </h4>
            <div className="space-y-3">
              {Object.entries(user.model_breakdown).length === 0 ? (
                <div className="text-sm text-google-gray-500 italic py-2">
                  No model consumption history found in this window.
                </div>
              ) : (
                Object.entries(user.model_breakdown).map(([model, detail]) => (
                  <div
                    key={model}
                    className="p-3 bg-[#181c26] border border-[#222836] rounded-xl flex items-center justify-between"
                  >
                    <div>
                      <span className="text-sm font-medium text-white">{model}</span>
                      <div className="text-xs text-google-gray-400 mt-0.5">
                        {detail.tokens.toLocaleString()} tokens ({(detail.share * 100).toFixed(1)}% share)
                      </div>
                    </div>
                    <div className="text-right">
                      <span className="text-sm font-bold text-white font-mono">
                        ${detail.cost.toFixed(2)}
                      </span>
                    </div>
                  </div>
                ))
              )}
            </div>
          </div>
        </div>

        {/* Footer */}
        <div className="p-4 bg-[#181c26] border-t border-[#222836] flex items-center justify-between">
          {onOpenDashboard ? (
            <button
              onClick={() => onOpenDashboard(user.user_id)}
              className="px-3 py-1.5 rounded-lg bg-emerald-500/10 hover:bg-emerald-500/20 text-emerald-400 border border-emerald-500/30 text-xs font-semibold flex items-center space-x-1.5 transition-colors"
            >
              <span>Open Personal Dashboard →</span>
            </button>
          ) : (
            <div />
          )}
          <button
            onClick={onClose}
            className="px-4 py-2 rounded-lg bg-google-blue hover:bg-google-blue/90 text-white text-xs font-semibold shadow transition-colors"
          >
            Close
          </button>
        </div>
      </div>
    </div>
  );
};
