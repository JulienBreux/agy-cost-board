import React from 'react';
import { Layers, DollarSign, ShieldCheck, Zap } from 'lucide-react';

interface NavbarProps {
  activeTab: 'overview' | 'costs' | 'licenses';
  setActiveTab: (tab: 'overview' | 'costs' | 'licenses') => void;
  days: number;
  setDays: (days: number) => void;
}

export const Navbar: React.FC<NavbarProps> = ({
  activeTab,
  setActiveTab,
  days,
  setDays,
}) => {
  return (
    <header className="border-b border-[#202530] bg-[#14171f]/80 backdrop-blur-md sticky top-0 z-40">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex items-center justify-between h-16">
          {/* Brand */}
          <div className="flex items-center space-x-3">
            <div className="h-9 w-9 rounded-lg bg-google-blue/15 flex items-center justify-center text-google-blue border border-google-blue/30 shadow-sm shadow-google-blue/20">
              <Zap className="h-5 w-5" />
            </div>
            <div>
              <div className="flex items-center space-x-2">
                <span className="font-bold text-white tracking-tight text-base sm:text-lg">AGY GE Board</span>
                <span className="text-[10px] font-semibold bg-google-blue/20 text-google-blue border border-google-blue/30 px-1.5 py-0.5 rounded tracking-wide uppercase">
                  Cloud Run
                </span>
              </div>
              <p className="text-[11px] text-google-gray-500 hidden sm:block">
                Antigravity & Gemini Enterprise Cost Attribution
              </p>
            </div>
          </div>

          {/* Navigation Tabs */}
          <nav className="flex space-x-1 sm:space-x-2">
            <button
              onClick={() => setActiveTab('overview')}
              className={`flex items-center space-x-1.5 px-3 py-1.5 rounded-md text-xs sm:text-sm font-medium transition-colors ${
                activeTab === 'overview'
                  ? 'bg-[#1e2330] text-white border border-[#2d3548]'
                  : 'text-google-gray-400 hover:text-white hover:bg-[#1a1e27]'
              }`}
            >
              <Layers className="h-4 w-4" />
              <span>Overview</span>
            </button>

            <button
              onClick={() => setActiveTab('costs')}
              className={`flex items-center space-x-1.5 px-3 py-1.5 rounded-md text-xs sm:text-sm font-medium transition-colors ${
                activeTab === 'costs'
                  ? 'bg-[#1e2330] text-white border border-[#2d3548]'
                  : 'text-google-gray-400 hover:text-white hover:bg-[#1a1e27]'
              }`}
            >
              <DollarSign className="h-4 w-4" />
              <span>Attributed Costs</span>
            </button>

            <button
              onClick={() => setActiveTab('licenses')}
              className={`flex items-center space-x-1.5 px-3 py-1.5 rounded-md text-xs sm:text-sm font-medium transition-colors ${
                activeTab === 'licenses'
                  ? 'bg-[#1e2330] text-white border border-[#2d3548]'
                  : 'text-google-gray-400 hover:text-white hover:bg-[#1a1e27]'
              }`}
            >
              <ShieldCheck className="h-4 w-4" />
              <span>Seat Governance</span>
            </button>
          </nav>

          {/* Time Window Selector */}
          <div className="flex items-center space-x-1 bg-[#1a1e28] p-1 rounded-lg border border-[#282f40]">
            {[7, 14, 30].map((d) => (
              <button
                key={d}
                onClick={() => setDays(d)}
                className={`text-xs px-2.5 py-1 rounded font-medium transition-all ${
                  days === d
                    ? 'bg-google-blue text-white shadow'
                    : 'text-google-gray-400 hover:text-white'
                }`}
              >
                {d}d
              </button>
            ))}
          </div>
        </div>
      </div>
    </header>
  );
};
