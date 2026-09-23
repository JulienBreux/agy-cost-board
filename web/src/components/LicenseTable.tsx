import React from 'react';
import { ShieldCheck, AlertCircle, Clock, Trash2, CheckCircle2 } from 'lucide-react';
import { LicenseGovernance } from '../api';

interface LicenseTableProps {
  governance: LicenseGovernance;
  onSelectUser: (userId: string) => void;
}

export const LicenseTable: React.FC<LicenseTableProps> = ({ governance, onSelectUser }) => {
  return (
    <div className="space-y-6">
      {/* Governance Summary Grid */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        {/* Active Seats */}
        <div className="bg-[#14171f] border border-[#222836] rounded-xl p-5 shadow-sm">
          <div className="flex items-center space-x-2 text-google-green mb-2">
            <CheckCircle2 className="h-5 w-5" />
            <h4 className="text-sm font-semibold">Active Seats</h4>
          </div>
          <div className="text-3xl font-bold text-white">{governance.active_seats}</div>
          <p className="text-xs text-google-gray-400 mt-1">
            Actively consuming Gemini & Antigravity tokens in window
          </p>
        </div>

        {/* Dormant Seats */}
        <div className="bg-[#14171f] border border-[#222836] rounded-xl p-5 shadow-sm">
          <div className="flex items-center space-x-2 text-google-red mb-2">
            <AlertCircle className="h-5 w-5" />
            <h4 className="text-sm font-semibold">Dormant Licenses</h4>
          </div>
          <div className="text-3xl font-bold text-google-red">{governance.dormant_seats}</div>
          <p className="text-xs text-google-gray-400 mt-1">
            0 tokens consumed for &gt;30 consecutive days
          </p>
        </div>

        {/* Seat Reclamation Savings */}
        <div className="bg-[#14171f] border border-[#222836] rounded-xl p-5 shadow-sm">
          <div className="flex items-center space-x-2 text-google-yellow mb-2">
            <ShieldCheck className="h-5 w-5" />
            <h4 className="text-sm font-semibold">Reclamation Savings</h4>
          </div>
          <div className="text-3xl font-bold text-google-yellow">
            ${governance.estimated_monthly_savings.toFixed(2)}/mo
          </div>
          <p className="text-xs text-google-gray-400 mt-1">
            At $45.00 standard Gemini Enterprise seat list price
          </p>
        </div>
      </div>

      {/* Quota Gauge Card */}
      <div className="bg-[#14171f] border border-[#222836] rounded-xl p-5 shadow-sm">
        <div className="flex items-center justify-between mb-2">
          <span className="text-sm font-semibold text-white">License Quota Consumption</span>
          <span className="text-xs font-mono text-google-gray-400">
            {governance.assigned_seats} assigned of {governance.seat_quota} licensed seats ({governance.utilization_pct.toFixed(1)}%)
          </span>
        </div>

        <div className="h-3 w-full bg-[#1e2330] rounded-full overflow-hidden flex">
          <div
            className="h-full bg-google-green transition-all"
            style={{ width: `${(governance.active_seats / governance.seat_quota) * 100}%` }}
            title="Active Seats"
          />
          <div
            className="h-full bg-google-red transition-all"
            style={{ width: `${(governance.dormant_seats / governance.seat_quota) * 100}%` }}
            title="Dormant Seats"
          />
        </div>
        <div className="flex items-center space-x-4 mt-3 text-xs text-google-gray-400">
          <div className="flex items-center space-x-1.5">
            <span className="w-2.5 h-2.5 rounded-full bg-google-green" />
            <span>Active ({governance.active_seats})</span>
          </div>
          <div className="flex items-center space-x-1.5">
            <span className="w-2.5 h-2.5 rounded-full bg-google-red" />
            <span>Dormant ({governance.dormant_seats})</span>
          </div>
          <div className="flex items-center space-x-1.5">
            <span className="w-2.5 h-2.5 rounded-full bg-[#1e2330]" />
            <span>Available ({Math.max(governance.seat_quota - governance.assigned_seats, 0)})</span>
          </div>
        </div>
      </div>

      {/* Dormant Seats Table */}
      <div className="bg-[#14171f] border border-[#222836] rounded-xl overflow-hidden shadow-sm">
        <div className="p-4 border-b border-[#222836] flex items-center justify-between">
          <div>
            <h3 className="text-sm font-semibold text-white">Dormant Licenses Recommended for Reclamation</h3>
            <p className="text-xs text-google-gray-400">Users with zero inference activity during the selected lookback period</p>
          </div>
          <span className="text-xs bg-google-red/10 text-google-red border border-google-red/20 px-2.5 py-1 rounded font-medium">
            {governance.dormant_users.length} seats at risk
          </span>
        </div>

        <div className="overflow-x-auto">
          <table className="w-full text-left text-xs sm:text-sm">
            <thead className="bg-[#181c26] text-google-gray-400 text-[11px] uppercase tracking-wider border-b border-[#222836]">
              <tr>
                <th className="py-3 px-4 font-semibold">User</th>
                <th className="py-3 px-4 font-semibold">Status</th>
                <th className="py-3 px-4 font-semibold">Last Recorded Activity</th>
                <th className="py-3 px-4 font-semibold text-right">Window Tokens</th>
                <th className="py-3 px-4 text-center">Action</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-[#1e2330]">
              {governance.dormant_users.length === 0 ? (
                <tr>
                  <td colSpan={5} className="py-8 text-center text-google-gray-500">
                    No dormant licenses detected. All assigned seats are active!
                  </td>
                </tr>
              ) : (
                governance.dormant_users.map((seat) => (
                  <tr
                    key={seat.user_id}
                    className="hover:bg-[#1a1e28]/70 transition-colors cursor-pointer"
                    onClick={() => onSelectUser(seat.user_id)}
                  >
                    <td className="py-3 px-4 font-medium text-white flex items-center space-x-2">
                      <span className="w-6 h-6 rounded-full bg-google-red/20 text-google-red flex items-center justify-center text-xs font-bold uppercase">
                        {seat.user_id[0]}
                      </span>
                      <span>{seat.user_id}</span>
                    </td>
                    <td className="py-3 px-4">
                      <span className="px-2 py-0.5 rounded text-[11px] font-semibold bg-google-red/15 text-google-red border border-google-red/20">
                        {seat.status.toUpperCase()}
                      </span>
                    </td>
                    <td className="py-3 px-4 text-google-gray-400 flex items-center space-x-1.5">
                      <Clock className="h-3.5 w-3.5 text-google-gray-500" />
                      <span>{seat.last_activity ? new Date(seat.last_activity).toLocaleDateString() : 'Never'}</span>
                    </td>
                    <td className="py-3 px-4 text-right font-mono text-google-gray-300">
                      {seat.total_tokens_in_window.toLocaleString()}
                    </td>
                    <td className="py-3 px-4 text-center">
                      <button
                        onClick={(e) => {
                          e.stopPropagation();
                          alert(`Reclaim recommendation logged for ${seat.user_id}. Release seat in Google Workspace Admin console.`);
                        }}
                        className="inline-flex items-center space-x-1 px-2.5 py-1 rounded text-xs font-medium bg-[#1e2330] hover:bg-google-red/20 text-google-gray-300 hover:text-google-red border border-[#2d3548] hover:border-google-red/40 transition-colors"
                      >
                        <Trash2 className="h-3 w-3" />
                        <span>Reclaim</span>
                      </button>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
};
