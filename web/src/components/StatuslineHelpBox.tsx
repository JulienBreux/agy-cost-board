import React, { useState } from 'react';
import { Terminal, ExternalLink, Copy, Check, ChevronDown, ChevronUp, Sparkles } from 'lucide-react';

interface StatuslineHelpBoxProps {
  userId?: string;
  days?: number;
  currentSpend?: number;
  currency?: string;
}

export const StatuslineHelpBox: React.FC<StatuslineHelpBoxProps> = ({
  userId,
  days = 30,
  currentSpend,
  currency = '$',
}) => {
  const [isCollapsed, setIsCollapsed] = useState(() => {
    try {
      return localStorage.getItem('agy_statusline_help_collapsed') === 'true';
    } catch {
      return false;
    }
  });

  const [copiedCmd, setCopiedCmd] = useState(false);
  const [copiedUrl, setCopiedUrl] = useState(false);

  const toggleCollapse = () => {
    const next = !isCollapsed;
    setIsCollapsed(next);
    try {
      localStorage.setItem('agy_statusline_help_collapsed', String(next));
    } catch {
      // Ignore localStorage write failures
    }
  };

  const origin =
    typeof window !== 'undefined' && window.location.origin
      ? window.location.origin
      : 'https://<your-server>';

  const userQueryParam = userId ? `?user=${encodeURIComponent(userId)}&days=${days}&ttl=300` : `?days=${days}&ttl=300`;
  const scriptUrl = `${origin}/statusline.sh${userQueryParam}`;
  const installCmd = `curl -sS "${scriptUrl}" -o ~/.gemini/antigravity-cli/statusline.sh && chmod +x ~/.gemini/antigravity-cli/statusline.sh`;

  const copyToClipboard = async (text: string, type: 'cmd' | 'url') => {
    try {
      await navigator.clipboard.writeText(text);
      if (type === 'cmd') {
        setCopiedCmd(true);
        setTimeout(() => setCopiedCmd(false), 2000);
      } else {
        setCopiedUrl(true);
        setTimeout(() => setCopiedUrl(false), 2000);
      }
    } catch (err) {
      console.error('Failed to copy to clipboard', err);
    }
  };

  const currencySymbol = currency === 'EUR' ? '€' : currency === 'USD' ? '$' : currency;
  const displayCost =
    typeof currentSpend === 'number' && currentSpend > 0
      ? `${currencySymbol}${Math.round(currentSpend)}`
      : `${currencySymbol}18`;

  return (
    <div
      data-testid="statusline-help-box"
      className="bg-[#14171f] border border-[#222836] rounded-xl p-5 shadow-sm transition-all"
    >
      {/* Header & Toggle */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
        <div className="flex items-center space-x-3">
          <div className="p-2.5 rounded-lg bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 flex-shrink-0">
            <Terminal className="w-5 h-5" />
          </div>
          <div>
            <div className="flex items-center space-x-2">
              <h3 className="text-sm sm:text-base font-semibold text-white tracking-tight">
                Antigravity CLI Status Bar Integration
              </h3>
              <span className="hidden sm:inline-flex items-center space-x-1 px-2 py-0.5 text-[10px] font-semibold rounded-full bg-emerald-500/15 text-emerald-300 border border-emerald-500/30">
                <Sparkles className="w-3 h-3 mr-0.5" />
                Live Spend in Statusline
              </span>
            </div>
            <p className="text-xs text-google-gray-400 mt-0.5">
              Display your real-time spend directly inside the agy status bar with zero terminal latency
            </p>
          </div>
        </div>

        <button
          type="button"
          onClick={toggleCollapse}
          aria-label={isCollapsed ? 'Expand statusline help' : 'Collapse statusline help'}
          className="self-end sm:self-center flex items-center space-x-1.5 px-3 py-1.5 text-xs font-medium text-google-gray-300 hover:text-white bg-[#1a1e27] hover:bg-[#202530] border border-[#222836] rounded-lg transition-colors"
        >
          <span>{isCollapsed ? 'Show Instructions' : 'Hide Instructions'}</span>
          {isCollapsed ? <ChevronDown className="w-4 h-4" /> : <ChevronUp className="w-4 h-4" />}
        </button>
      </div>

      {/* Collapsible Content */}
      {!isCollapsed && (
        <div className="mt-5 space-y-5 border-t border-[#1f2430] pt-4">
          {/* Quick Install Command */}
          <div className="space-y-2">
            <div className="flex items-center justify-between">
              <label className="text-xs font-medium text-google-gray-300">
                One-Line Quick Install
              </label>
              <div className="flex items-center space-x-2">
                <a
                  href={scriptUrl}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="inline-flex items-center space-x-1 text-xs text-google-blue hover:underline"
                >
                  <span>Direct Script Link</span>
                  <ExternalLink className="w-3 h-3" />
                </a>
                <span className="text-google-gray-600">|</span>
                <button
                  type="button"
                  onClick={() => copyToClipboard(scriptUrl, 'url')}
                  className="text-xs text-google-gray-400 hover:text-white transition-colors"
                >
                  {copiedUrl ? 'URL Copied!' : 'Copy URL'}
                </button>
              </div>
            </div>

            <div className="relative group bg-[#0d1017] border border-[#222836] rounded-lg p-3 font-mono text-xs text-google-gray-200 overflow-x-auto flex items-center justify-between gap-3">
              <code className="whitespace-pre flex-1 text-emerald-400/90">{installCmd}</code>
              <button
                type="button"
                onClick={() => copyToClipboard(installCmd, 'cmd')}
                aria-label="Copy install command"
                className="shrink-0 flex items-center space-x-1 px-2.5 py-1 text-xs font-medium bg-[#1a1e28] hover:bg-[#242a38] text-google-gray-200 hover:text-white border border-[#2d3548] rounded transition-colors"
              >
                {copiedCmd ? (
                  <>
                    <Check className="w-3.5 h-3.5 text-emerald-400" />
                    <span className="text-emerald-400 font-semibold">Copied!</span>
                  </>
                ) : (
                  <>
                    <Copy className="w-3.5 h-3.5" />
                    <span>Copy</span>
                  </>
                )}
              </button>
            </div>
          </div>

          {/* Configuration Note */}
          <div className="bg-[#181c26] border border-[#222836] rounded-lg p-3 text-xs text-google-gray-300 flex flex-col sm:flex-row sm:items-center justify-between gap-2">
            <div>
              <span className="font-semibold text-white">Antigravity Configuration: </span>
              Ensure your <code className="text-google-blue bg-[#101218] px-1.5 py-0.5 rounded border border-[#222836]">~/.gemini/antigravity-cli/config.json</code> has:
            </div>
            <code className="bg-[#101218] text-google-gray-200 px-2 py-1 rounded border border-[#222836] font-mono text-[11px] whitespace-nowrap">
              {'"statusline": "~/.gemini/antigravity-cli/statusline.sh"'}
            </code>
          </div>

          {/* Example of Rendering */}
          <div className="space-y-2">
            <div className="flex items-center justify-between">
              <span className="text-xs font-medium text-google-gray-300">
                Example of Rendering in Antigravity Status Bar
              </span>
              <span className="text-[11px] text-google-gray-400">
                Terminal preview (asynchronous background cache)
              </span>
            </div>

            {/* Mock Terminal Bar Window */}
            <div className="bg-[#090b10] border border-[#1e2330] rounded-lg overflow-hidden shadow-inner font-mono text-xs select-none">
              {/* Terminal Window Header Bar */}
              <div className="bg-[#10131a] border-b border-[#1b202c] px-3 py-1.5 flex items-center space-x-2">
                <div className="flex space-x-1.5">
                  <div className="w-2.5 h-2.5 rounded-full bg-rose-500/80" />
                  <div className="w-2.5 h-2.5 rounded-full bg-amber-500/80" />
                  <div className="w-2.5 h-2.5 rounded-full bg-emerald-500/80" />
                </div>
                <span className="text-[11px] text-google-gray-400 font-sans pl-2">
                  Antigravity CLI (agy)
                </span>
              </div>

              {/* Terminal Screen Body */}
              <div className="p-3.5 space-y-2 overflow-x-auto">
                <div className="text-google-gray-400 text-[11px] flex items-center space-x-1.5">
                  <span className="text-google-gray-400">&gt;</span>
                  <span>Accept-edits mode: file edits auto-approved (shift+tab to cycle)</span>
                </div>

                {/* Statusline Row */}
                <div className="flex items-center space-x-1.5 whitespace-nowrap text-xs pt-1 border-t border-[#161a24]">
                  {/* Status Indicator */}
                  <span className="text-emerald-400 font-bold flex items-center">
                    <span className="text-[10px] mr-1">●</span>READY
                  </span>

                  <span className="text-google-gray-400">/</span>

                  {/* Active Foundation Model */}
                  <span className="text-fuchsia-400 font-medium">Gemini 3.8 Flash (High)</span>

                  <span className="text-google-gray-400">|</span>

                  {/* Context Window Stats */}
                  <span className="text-google-gray-400">ctx</span>
                  <span className="text-white font-medium">▧............</span>
                  <span className="text-google-gray-400">6.0%</span>

                  <span className="text-google-gray-400">·</span>

                  {/* Artifacts, Subagents, Tasks, Sandbox */}
                  <span className="text-google-gray-400">artifacts 0</span>
                  <span className="text-google-gray-400">·</span>
                  <span className="text-google-gray-400">subagents 0</span>
                  <span className="text-google-gray-400">·</span>
                  <span className="text-google-gray-400">tasks 0</span>
                  <span className="text-google-gray-400">·</span>
                  <span className="text-google-gray-400">sandbox off</span>

                  <span className="text-google-gray-400">·</span>

                  {/* Attributed Spend Highlight */}
                  <span className="inline-flex items-center px-1.5 py-0.5 rounded bg-emerald-500/15 text-emerald-300 font-bold border border-emerald-500/30 text-xs tracking-tight shadow-sm">
                    cost {displayCost}
                  </span>
                </div>
              </div>
            </div>

            <p className="text-[11px] text-google-gray-400 italic">
              Spend is cached locally and refreshed in the background every 5 minutes. If offline or unreachable, the CLI continues executing with zero latency.
            </p>
          </div>
        </div>
      )}
    </div>
  );
};
