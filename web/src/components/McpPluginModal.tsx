import React, { useEffect, useState } from 'react';
import {
  Cpu,
  Check,
  Copy,
  ShieldCheck,
  Terminal,
  Globe,
  X,
  Wrench,
  Sparkles,
} from 'lucide-react';
import { LicenseGovernance, OverviewMetrics } from '../api';
import { formatCurrency } from '../utils/currency';

export interface McpPluginModalProps {
  isOpen: boolean;
  onClose: () => void;
  mcpEnabled: boolean;
  overview: OverviewMetrics | null;
  governance: LicenseGovernance | null;
}

const MCP_TOOLS = [
  {
    name: 'get_overview_metrics',
    args: 'days?: number',
    description:
      'High-level spend, token consumption, active users, seat utilization %, and model breakdown.',
  },
  {
    name: 'get_license_status',
    args: 'days?: number',
    description:
      'Gemini Enterprise seat quota, assigned/active/dormant seats, and reclaimable license savings.',
  },
  {
    name: 'get_users_costs',
    args: 'days?: number, model?: string',
    description:
      'Proportional per-user, per-model, per-day cost attribution rows reconciled with GCP Billing.',
  },
  {
    name: 'get_user_dashboard',
    args: 'user_id: string, days?: number, monthly_budget?: number',
    description:
      'Personal developer burn rate, projected month-end spend, budget status, and FinOps recommendations.',
  },
];

export const McpPluginModal: React.FC<McpPluginModalProps> = ({
  isOpen,
  onClose,
  mcpEnabled,
  overview,
  governance,
}) => {
  const [transport, setTransport] = useState<'http' | 'stdio'>('http');
  const [copied, setCopied] = useState(false);

  useEffect(() => {
    if (!isOpen) return;

    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        onClose();
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => {
      window.removeEventListener('keydown', handleKeyDown);
    };
  }, [isOpen, onClose]);

  if (!isOpen) return null;

  const origin =
    typeof window !== 'undefined' && window.location.origin
      ? window.location.origin
      : 'http://localhost:8080';

  const httpConfigSnippet = JSON.stringify(
    {
      mcpServers: {
        'agy-cost-board': {
          url: `${origin}/mcp`,
        },
      },
    },
    null,
    2
  );

  const stdioConfigSnippet = JSON.stringify(
    {
      mcpServers: {
        'agy-cost-board': {
          command: 'agy-cost-board',
          args: ['mcp', '--demo'],
          env: {
            AGY_COST_BOARD_MCP_ENABLED: 'true',
          },
        },
      },
    },
    null,
    2
  );

  const activeSnippet = transport === 'http' ? httpConfigSnippet : stdioConfigSnippet;

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(activeSnippet);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch {
      // ignore clipboard errors in headless environments
    }
  };

  const seatQuota = governance?.seat_quota ?? overview?.seat_quota ?? 10;
  const activeSeats = governance?.active_seats ?? overview?.active_users_count ?? 0;
  const dormantSeats = governance?.dormant_seats ?? 0;
  const utilizationPct = Math.round(
    governance?.utilization_pct ?? overview?.utilization_pct ?? 0
  );
  const currency = governance?.currency ?? overview?.currency ?? 'USD';
  const totalCost = overview?.total_billed_cost ?? 0;
  const estimatedSavings =
    governance?.estimated_monthly_savings ?? overview?.estimated_monthly_savings ?? 0;

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/75 backdrop-blur-sm animate-in fade-in duration-200"
      data-testid="mcp-modal-backdrop"
      onClick={(e) => {
        if (e.target === e.currentTarget) {
          onClose();
        }
      }}
    >
      <div
        role="dialog"
        aria-modal="true"
        aria-labelledby="mcp-modal-title"
        className="relative w-full max-w-2xl max-h-[88vh] flex flex-col bg-[#161a23] border border-[#232936] text-white rounded-2xl shadow-2xl overflow-hidden animate-in zoom-in-95 duration-200"
        onClick={(e) => e.stopPropagation()}
      >
        {/* Header */}
        <div className="px-6 pt-5 pb-4 border-b border-[#232936] flex items-start justify-between gap-4">
          <div className="flex items-start space-x-3">
            <div className="p-2.5 rounded-xl bg-google-blue/15 text-google-blue border border-google-blue/30 flex-shrink-0">
              <Cpu className="w-5 h-5" />
            </div>
            <div>
              <div className="flex flex-wrap items-center gap-2">
                <h2
                  id="mcp-modal-title"
                  className="text-base sm:text-lg font-bold text-white tracking-tight"
                >
                  Antigravity Plugin &amp; MCP Server
                </h2>
                <span
                  data-testid="mcp-status-badge"
                  className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-semibold border ${
                    mcpEnabled
                      ? 'bg-emerald-500/15 text-emerald-300 border-emerald-500/30'
                      : 'bg-rose-500/15 text-rose-300 border-rose-500/30'
                  }`}
                >
                  <span
                    className={`w-1.5 h-1.5 rounded-full mr-1.5 ${
                      mcpEnabled ? 'bg-emerald-400' : 'bg-rose-400'
                    }`}
                  />
                  {mcpEnabled ? 'MCP Enabled' : 'MCP Disabled'}
                </span>
              </div>
              <p className="mt-1 text-xs text-google-gray-400">
                Expose live BigQuery cost attribution &amp; Gemini Enterprise seat quotas directly to Antigravity IDE and CLI agents.
              </p>
            </div>
          </div>

          <button
            type="button"
            onClick={onClose}
            aria-label="Close MCP modal"
            className="p-1.5 rounded-lg text-google-gray-400 hover:text-white hover:bg-[#202532] transition-colors"
          >
            <X className="w-4 h-4" />
          </button>
        </div>

        {/* Body */}
        <div className="overflow-y-auto px-6 py-5 space-y-5 flex-1 text-xs sm:text-sm">
          {/* Live Quota Widget Preview */}
          <div className="bg-[#11141c] border border-[#222836] rounded-xl p-4 space-y-3">
            <div className="flex items-center justify-between">
              <div className="flex items-center space-x-2">
                <ShieldCheck className="w-4 h-4 text-emerald-400" />
                <span className="text-xs font-bold uppercase tracking-wider text-google-gray-300">
                  Live Antigravity Quota &amp; Spend Telemetry
                </span>
              </div>
              <span className="text-[11px] font-mono text-google-gray-400">
                Protocol 2025-03-26
              </span>
            </div>

            <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
              <div className="bg-[#161a24] border border-[#232938] rounded-lg p-3">
                <div className="text-[11px] text-google-gray-400">Active Seat Quota</div>
                <div className="mt-1 flex items-baseline justify-between">
                  <span className="text-base font-bold text-white">
                    {activeSeats} / {seatQuota}
                  </span>
                  <span className="text-xs font-semibold text-emerald-400">
                    {utilizationPct}%
                  </span>
                </div>
                <div className="mt-2 w-full h-1.5 bg-[#232938] rounded-full overflow-hidden">
                  <div
                    className="h-full bg-emerald-400 rounded-full transition-all"
                    style={{ width: `${Math.min(100, utilizationPct)}%` }}
                  />
                </div>
              </div>

              <div className="bg-[#161a24] border border-[#232938] rounded-lg p-3">
                <div className="text-[11px] text-google-gray-400">Dormant Seats</div>
                <div className="mt-1 flex items-baseline justify-between">
                  <span className="text-base font-bold text-amber-400">
                    {dormantSeats} seats
                  </span>
                  <span className="text-[11px] text-google-gray-400">
                    Save {formatCurrency(estimatedSavings, currency)}/mo
                  </span>
                </div>
              </div>

              <div className="bg-[#161a24] border border-[#232938] rounded-lg p-3">
                <div className="text-[11px] text-google-gray-400">Attributed Spend (30d)</div>
                <div className="mt-1 flex items-baseline justify-between">
                  <span className="text-base font-bold text-google-blue">
                    {formatCurrency(totalCost, currency)}
                  </span>
                  <span className="text-[11px] text-google-gray-400">Reconciled</span>
                </div>
              </div>
            </div>
          </div>

          {/* Available MCP Tools */}
          <div className="space-y-2.5">
            <div className="flex items-center space-x-2">
              <Wrench className="w-4 h-4 text-google-blue" />
              <h3 className="text-xs font-bold uppercase tracking-wider text-google-gray-300">
                Exposed MCP Tools (4)
              </h3>
            </div>
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-2.5">
              {MCP_TOOLS.map((tool) => (
                <div
                  key={tool.name}
                  className="bg-[#12151e] border border-[#222836] rounded-lg p-3 space-y-1"
                >
                  <div className="flex items-center justify-between gap-2">
                    <code className="text-xs font-mono font-bold text-emerald-400">
                      {tool.name}
                    </code>
                  </div>
                  <div className="text-[11px] font-mono text-google-gray-500">
                    ({tool.args})
                  </div>
                  <p className="text-xs text-google-gray-300 leading-relaxed">
                    {tool.description}
                  </p>
                </div>
              ))}
            </div>
          </div>

          {/* Antigravity MCP Configuration Snippet */}
          <div className="space-y-2.5">
            <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2">
              <div className="flex items-center space-x-2">
                <Sparkles className="w-4 h-4 text-amber-400" />
                <h3 className="text-xs font-bold uppercase tracking-wider text-google-gray-300">
                  Antigravity Connection (`mcp_config.json`)
                </h3>
              </div>

              <div className="flex items-center space-x-1 bg-[#11141c] p-1 rounded-lg border border-[#222836]">
                <button
                  type="button"
                  onClick={() => setTransport('http')}
                  className={`flex items-center space-x-1 px-2.5 py-1 rounded text-xs font-medium transition-colors ${
                    transport === 'http'
                      ? 'bg-google-blue text-white'
                      : 'text-google-gray-400 hover:text-white'
                  }`}
                >
                  <Globe className="w-3.5 h-3.5" />
                  <span>Streamable HTTP</span>
                </button>
                <button
                  type="button"
                  onClick={() => setTransport('stdio')}
                  className={`flex items-center space-x-1 px-2.5 py-1 rounded text-xs font-medium transition-colors ${
                    transport === 'stdio'
                      ? 'bg-google-blue text-white'
                      : 'text-google-gray-400 hover:text-white'
                  }`}
                >
                  <Terminal className="w-3.5 h-3.5" />
                  <span>Stdio CLI</span>
                </button>
              </div>
            </div>

            <div className="relative bg-[#0d1017] border border-[#222836] rounded-lg p-3.5 font-mono text-xs text-google-gray-200 overflow-x-auto">
              <button
                type="button"
                onClick={handleCopy}
                aria-label="Copy MCP config"
                className="absolute top-2.5 right-2.5 flex items-center space-x-1 px-2.5 py-1 text-xs font-sans font-medium bg-[#1a1e28] hover:bg-[#242a38] text-google-gray-200 hover:text-white border border-[#2d3548] rounded transition-colors"
              >
                {copied ? (
                  <>
                    <Check className="w-3.5 h-3.5 text-emerald-400" />
                    <span className="text-emerald-400">Copied</span>
                  </>
                ) : (
                  <>
                    <Copy className="w-3.5 h-3.5" />
                    <span>Copy JSON</span>
                  </>
                )}
              </button>
              <pre
                data-testid="mcp-config-snippet"
                className="text-emerald-400/90 leading-relaxed pr-24"
              >
                {activeSnippet}
              </pre>
            </div>

            <div className="bg-[#12151e] border border-[#222836] rounded-lg p-3 text-xs text-google-gray-400 space-y-1">
              <div>
                <strong className="text-white">Native Antigravity Plugin:</strong> Bundled at{' '}
                <code className="text-google-blue font-mono">
                  .agents/plugins/agy-cost-board/
                </code>{' '}
                with preconfigured <code className="font-mono">cost-and-quota-advisor</code> skill and rules.
              </div>
              <div>
                <strong className="text-white">Runtime Toggle:</strong> Set{' '}
                <code className="text-amber-300 font-mono">AGY_COST_BOARD_MCP_ENABLED=false</code> (or{' '}
                <code className="font-mono">--mcp-enabled=false</code>) to disable the MCP server on HTTP and stdio.
              </div>
            </div>
          </div>
        </div>

        {/* Footer */}
        <div className="px-6 py-3.5 border-t border-[#232936] bg-[#131720]/80 flex justify-end items-center">
          <button
            type="button"
            onClick={onClose}
            className="px-4 py-1.5 rounded-xl text-xs sm:text-sm font-semibold bg-[#1a73e8] hover:bg-[#1557b0] text-white transition-all"
          >
            Done
          </button>
        </div>
      </div>
    </div>
  );
};
