import React from 'react';
import { Layers, DollarSign, ShieldCheck, Zap, Activity, User, Sparkles } from 'lucide-react';
import { ThemeToggle } from './ThemeToggle';

interface NavbarProps {
  activeTab: 'overview' | 'costs' | 'licenses' | 'setup' | 'my-consumption';
  setActiveTab: (tab: 'overview' | 'costs' | 'licenses' | 'setup' | 'my-consumption') => void;
  days: number;
  setDays: (days: number) => void;
  onOpenChangelog?: () => void;
}

export const Navbar: React.FC<NavbarProps> = ({
  activeTab,
  setActiveTab,
  days,
  setDays,
  onOpenChangelog,
}) => {
  return (
    <header className="border-b border-[#202530] bg-[#14171f]/80 backdrop-blur-md sticky top-0 z-40">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex items-center justify-between h-16 gap-2 sm:gap-4">
          {/* Brand */}
          <div className="flex items-center space-x-2.5 sm:space-x-3 flex-shrink-0">
            <div className="h-9 w-9 rounded-lg bg-google-blue/15 flex items-center justify-center text-google-blue border border-google-blue/30 shadow-sm shadow-google-blue/20 flex-shrink-0">
              <Zap className="h-5 w-5" />
            </div>
            <div className="flex flex-col justify-center min-w-0">
              <div className="flex items-center space-x-2">
                <span className="font-bold text-white tracking-tight text-base sm:text-lg whitespace-nowrap leading-tight">
                  Usage and Spend
                </span>
              </div>
              <p className="text-[11px] text-google-gray-500 hidden sm:block whitespace-nowrap leading-tight">
                Gemini platform &amp; Antigravity
              </p>
            </div>
          </div>

          {/* Navigation Tabs */}
          <nav
            aria-label="Main Navigation"
            className="flex items-center space-x-1 sm:space-x-1.5 overflow-x-auto no-scrollbar py-1 min-w-0"
          >
            <button
              onClick={() => setActiveTab('overview')}
              className={`flex items-center space-x-1.5 px-2.5 sm:px-3 py-1.5 rounded-md text-xs sm:text-sm font-medium whitespace-nowrap flex-shrink-0 transition-colors ${
                activeTab === 'overview'
                  ? 'bg-[#1e2330] text-white border border-[#2d3548]'
                  : 'text-google-gray-400 hover:text-white hover:bg-[#1a1e27]'
              }`}
            >
              <Layers className="h-4 w-4 flex-shrink-0" />
              <span>Overview</span>
            </button>

            <button
              onClick={() => setActiveTab('my-consumption')}
              className={`flex items-center space-x-1.5 px-2.5 sm:px-3 py-1.5 rounded-md text-xs sm:text-sm font-medium whitespace-nowrap flex-shrink-0 transition-colors ${
                activeTab === 'my-consumption'
                  ? 'bg-[#1e2330] text-emerald-400 border border-emerald-500/30 shadow-sm shadow-emerald-500/10'
                  : 'text-google-gray-400 hover:text-white hover:bg-[#1a1e27]'
              }`}
            >
              <User className="h-4 w-4 flex-shrink-0" />
              <span>My Consumption</span>
            </button>

            <button
              onClick={() => setActiveTab('costs')}
              className={`flex items-center space-x-1.5 px-2.5 sm:px-3 py-1.5 rounded-md text-xs sm:text-sm font-medium whitespace-nowrap flex-shrink-0 transition-colors ${
                activeTab === 'costs'
                  ? 'bg-[#1e2330] text-white border border-[#2d3548]'
                  : 'text-google-gray-400 hover:text-white hover:bg-[#1a1e27]'
              }`}
            >
              <DollarSign className="h-4 w-4 flex-shrink-0" />
              <span>Attributed Costs</span>
            </button>

            <button
              onClick={() => setActiveTab('licenses')}
              className={`flex items-center space-x-1.5 px-2.5 sm:px-3 py-1.5 rounded-md text-xs sm:text-sm font-medium whitespace-nowrap flex-shrink-0 transition-colors ${
                activeTab === 'licenses'
                  ? 'bg-[#1e2330] text-white border border-[#2d3548]'
                  : 'text-google-gray-400 hover:text-white hover:bg-[#1a1e27]'
              }`}
            >
              <ShieldCheck className="h-4 w-4 flex-shrink-0" />
              <span>Seat Governance</span>
            </button>

            <button
              onClick={() => setActiveTab('setup')}
              className={`flex items-center space-x-1.5 px-2.5 sm:px-3 py-1.5 rounded-md text-xs sm:text-sm font-medium whitespace-nowrap flex-shrink-0 transition-colors ${
                activeTab === 'setup'
                  ? 'bg-[#1e2330] text-white border border-[#2d3548]'
                  : 'text-google-gray-400 hover:text-white hover:bg-[#1a1e27]'
              }`}
            >
              <Activity className="h-4 w-4 flex-shrink-0" />
              <span>Setup & Health</span>
            </button>
          </nav>

          {/* Right Controls: Time Window & Theme Selector */}
          <div className="flex items-center space-x-1.5 sm:space-x-2 flex-shrink-0">
            {/* Time Window Selector */}
            <div className="flex items-center space-x-1 bg-[#1a1e28] p-1 rounded-lg border border-[#282f40] flex-shrink-0">
              {[7, 14, 30].map((d) => (
                <button
                  key={d}
                  onClick={() => setDays(d)}
                  className={`text-xs px-2 sm:px-2.5 py-1 rounded font-medium whitespace-nowrap transition-all ${
                    days === d
                      ? 'bg-google-blue text-white shadow'
                      : 'text-google-gray-400 hover:text-white'
                  }`}
                >
                  {d}d
                </button>
              ))}
            </div>

            {/* What's new Changelog Trigger */}
            {onOpenChangelog && (
              <button
                type="button"
                onClick={onOpenChangelog}
                className="flex items-center space-x-1.5 bg-[#1a1e28] hover:bg-[#202530] border border-[#282f40] rounded-lg px-2.5 py-1.5 text-xs font-medium text-amber-400 hover:text-amber-300 transition-all shadow-sm focus:outline-none focus:ring-1 focus:ring-amber-400 whitespace-nowrap flex-shrink-0"
                title="What's new"
                aria-label="What's new"
              >
                <Sparkles className="h-4 w-4 flex-shrink-0" />
                <span className="hidden xl:inline font-medium text-google-gray-300">What's new</span>
              </button>
            )}

            {/* Theme Switcher Button */}
            <ThemeToggle />
          </div>
        </div>
      </div>
    </header>
  );
};
