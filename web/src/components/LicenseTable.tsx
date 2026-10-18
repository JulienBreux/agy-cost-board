import React from 'react';
import { ShieldCheck, AlertCircle, Clock, Trash2, CheckCircle2, Globe, Sparkles } from 'lucide-react';
import { LicenseGovernance, LicenseSeat } from '../api';
import { formatCurrency } from '../utils/currency';
import { WeeklyAllowancePoolCard } from './WeeklyAllowancePoolCard';

interface LicenseTableProps {
  governance: LicenseGovernance;
  onSelectUser: (userId: string) => void;
}

const renderTierBadge = (seat: LicenseSeat) => {
  const tier = (seat.tier || 'standard').toLowerCase();
  const isPlus = tier === 'plus';
  return (
    <div className="flex items-center gap-1.5 flex-wrap">
      <span
        className={`px-2 py-0.5 rounded text-[11px] font-semibold border ${
          isPlus
            ? 'bg-purple-500/15 text-purple-300 border-purple-500/30'
            : 'bg-blue-500/15 text-blue-300 border-blue-500/30'
        }`}
      >
        {isPlus ? 'Plus ($3.75/wk)' : 'Standard ($2.50/wk)'}
      </span>
      {seat.is_trial && (
        <span
          className="px-2 py-0.5 rounded text-[10px] font-semibold bg-amber-500/15 text-amber-300 border border-amber-500/30"
          title={seat.trial_ends_at ? `Ends ${seat.trial_ends_at}` : 'Free Trial'}
        >
          Free Trial
        </span>
      )}
    </div>
  );
};

const renderLocationBadge = (location?: string) => {
  const loc = (location || 'global').toUpperCase();
  return (
    <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded text-[11px] font-mono bg-slate-800 text-slate-300 border border-slate-700">
      <Globe className="w-3 h-3 text-slate-400" />
      {loc}
    </span>
  );
};

export const LicenseTable: React.FC<LicenseTableProps> = ({ governance, onSelectUser }) => {
  const dormantUsers = Array.isArray(governance?.dormant_users) ? governance.dormant_users : [];
  const allSeats = Array.isArray(governance?.all_seats) ? governance.all_seats : [];
  const seatQuota = governance?.seat_quota > 0 ? governance.seat_quota : 1;
  const standardSeats = Number(governance?.standard_seats ?? 0);
  const plusSeats = Number(governance?.plus_seats ?? 0);
  const trialSeats = Number(governance?.trial_seats ?? 0);
  const locations = Array.isArray(governance?.locations) && governance.locations.length > 0
    ? governance.locations
    : ['global', 'eu', 'us'];

  return (
    <div className="space-y-6">
      {/* Weekly Allowance Pool & What-If Simulator */}
      {governance?.weekly_pool && (
        <WeeklyAllowancePoolCard
          pool={governance.weekly_pool}
          currency={governance.currency || 'USD'}
        />
      )}

      {/* Governance Summary Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        {/* Active Seats */}
        <div className="bg-[#14171f] border border-[#222836] rounded-xl p-5 shadow-sm">
          <div className="flex items-center space-x-2 text-google-green mb-2">
            <CheckCircle2 className="h-5 w-5" />
            <h4 className="text-sm font-semibold">Active Seats</h4>
          </div>
          <div className="text-3xl font-bold text-white">{governance.active_seats}</div>
          <p className="text-xs text-google-gray-400 mt-1">
            Actively consuming Gemini &amp; Antigravity tokens in window
          </p>
        </div>

        {/* Multi-Tier & Trial Distribution */}
        <div className="bg-[#14171f] border border-[#222836] rounded-xl p-5 shadow-sm">
          <div className="flex items-center space-x-2 text-purple-400 mb-2">
            <Sparkles className="h-5 w-5" />
            <h4 className="text-sm font-semibold">Multi-Tier &amp; Regions</h4>
          </div>
          <div className="flex items-baseline gap-3">
            <span className="text-2xl font-bold text-white">{standardSeats} Std</span>
            <span className="text-2xl font-bold text-purple-300">{plusSeats} Plus</span>
          </div>
          <div className="flex items-center gap-2 mt-2 flex-wrap">
            {trialSeats > 0 && (
              <span className="px-2 py-0.5 rounded text-[11px] font-medium bg-amber-500/15 text-amber-300 border border-amber-500/30">
                {trialSeats} Free Trial
              </span>
            )}
            <span className="text-xs text-google-gray-400 font-mono">
              Regions: {locations.map((l) => l.toUpperCase()).join(', ')}
            </span>
          </div>
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
            {formatCurrency(governance.estimated_monthly_savings, governance.currency)}/mo
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
            {governance.assigned_seats} assigned of {governance.seat_quota} licensed seats ({Number(governance.utilization_pct ?? 0).toFixed(1)}%)
          </span>
        </div>

        <div className="h-3 w-full bg-[#1e2330] rounded-full overflow-hidden flex">
          <div
            className="h-full bg-google-green transition-all"
            style={{ width: `${Math.min((governance.active_seats / seatQuota) * 100, 100)}%` }}
            title="Active Seats"
          />
          <div
            className="h-full bg-google-red transition-all"
            style={{ width: `${Math.min((governance.dormant_seats / seatQuota) * 100, 100)}%` }}
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
            {dormantUsers.length} seats at risk
          </span>
        </div>

        <div className="overflow-x-auto">
          <table className="w-full text-left text-xs sm:text-sm">
            <thead className="bg-[#181c26] text-google-gray-400 text-[11px] uppercase tracking-wider border-b border-[#222836]">
              <tr>
                <th className="py-3 px-4 font-semibold">User</th>
                <th className="py-3 px-4 font-semibold">Tier &amp; Allowance</th>
                <th className="py-3 px-4 font-semibold">Location</th>
                <th className="py-3 px-4 font-semibold">Status</th>
                <th className="py-3 px-4 font-semibold">Last Recorded Activity</th>
                <th className="py-3 px-4 font-semibold text-right">Window Tokens</th>
                <th className="py-3 px-4 text-center">Action</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-[#1e2330]">
              {dormantUsers.length === 0 ? (
                <tr>
                  <td colSpan={7} className="py-8 text-center text-google-gray-500">
                    No dormant licenses detected. All assigned seats are active!
                  </td>
                </tr>
              ) : (
                dormantUsers.map((seat) => (
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
                    <td className="py-3 px-4">{renderTierBadge(seat)}</td>
                    <td className="py-3 px-4">{renderLocationBadge(seat.location)}</td>
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
                      {Number(seat.total_tokens_in_window ?? 0).toLocaleString()}
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

      {/* Full Multi-Tier & Multi-Location Seat Roster */}
      {allSeats.length > 0 && (
        <div className="bg-[#14171f] border border-[#222836] rounded-xl overflow-hidden shadow-sm">
          <div className="p-4 border-b border-[#222836] flex items-center justify-between">
            <div>
              <h3 className="text-sm font-semibold text-white">Assigned Seats Roster (Multi-Tier &amp; Multi-Location Discovery)</h3>
              <p className="text-xs text-google-gray-400">
                Discovered Gemini Enterprise subscriptions across global, eu, and us regions
              </p>
            </div>
            <span className="text-xs bg-indigo-500/15 text-indigo-300 border border-indigo-500/30 px-2.5 py-1 rounded font-medium">
              {allSeats.length} assigned seats
            </span>
          </div>

          <div className="overflow-x-auto">
            <table className="w-full text-left text-xs sm:text-sm">
              <thead className="bg-[#181c26] text-google-gray-400 text-[11px] uppercase tracking-wider border-b border-[#222836]">
                <tr>
                  <th className="py-3 px-4 font-semibold">User</th>
                  <th className="py-3 px-4 font-semibold">Tier &amp; Weekly Allowance</th>
                  <th className="py-3 px-4 font-semibold">Location</th>
                  <th className="py-3 px-4 font-semibold">Status</th>
                  <th className="py-3 px-4 font-semibold text-right">Window Tokens</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-[#1e2330]">
                {allSeats.map((seat) => (
                  <tr
                    key={`roster-${seat.user_id}`}
                    className="hover:bg-[#1a1e28]/70 transition-colors cursor-pointer"
                    onClick={() => onSelectUser(seat.user_id)}
                  >
                    <td className="py-3 px-4 font-medium text-white">{seat.user_id}</td>
                    <td className="py-3 px-4">{renderTierBadge(seat)}</td>
                    <td className="py-3 px-4">{renderLocationBadge(seat.location)}</td>
                    <td className="py-3 px-4">
                      <span
                        className={`px-2 py-0.5 rounded text-[11px] font-semibold border ${
                          seat.status === 'active'
                            ? 'bg-emerald-500/15 text-emerald-300 border-emerald-500/30'
                            : seat.status === 'at_risk'
                            ? 'bg-amber-500/15 text-amber-300 border-amber-500/30'
                            : 'bg-google-red/15 text-google-red border-google-red/20'
                        }`}
                      >
                        {seat.status.toUpperCase()}
                      </span>
                    </td>
                    <td className="py-3 px-4 text-right font-mono text-google-gray-300">
                      {Number(seat.total_tokens_in_window ?? 0).toLocaleString()}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}
    </div>
  );
};
