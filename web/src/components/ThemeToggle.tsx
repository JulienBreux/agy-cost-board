import React, { useState, useRef, useEffect } from 'react';
import { Sun, Moon, Monitor, Check, ChevronDown } from 'lucide-react';
import { useTheme, ThemeMode } from '../context/ThemeContext';

export const ThemeToggle: React.FC = () => {
  const { theme, effectiveTheme, setTheme } = useTheme();
  const [isOpen, setIsOpen] = useState(false);
  const menuRef = useRef<HTMLDivElement>(null);

  // Close dropdown on outside click
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (menuRef.current && !menuRef.current.contains(event.target as Node)) {
        setIsOpen(false);
      }
    };

    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === 'Escape') {
        setIsOpen(false);
      }
    };

    document.addEventListener('mousedown', handleClickOutside);
    document.addEventListener('keydown', handleKeyDown);
    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
      document.removeEventListener('keydown', handleKeyDown);
    };
  }, []);

  const getThemeIcon = (mode: ThemeMode) => {
    switch (mode) {
      case 'light':
        return <Sun className="h-4 w-4 text-amber-500" />;
      case 'dark':
        return <Moon className="h-4 w-4 text-google-blue" />;
      case 'auto':
        return <Monitor className="h-4 w-4 text-emerald-400" />;
    }
  };

  const getThemeLabel = (mode: ThemeMode) => {
    switch (mode) {
      case 'light':
        return 'Light';
      case 'dark':
        return 'Dark';
      case 'auto':
        return 'Auto';
    }
  };

  return (
    <div className="relative inline-block text-left" ref={menuRef}>
      <button
        type="button"
        data-testid="theme-toggle-btn"
        aria-label={`Theme: ${getThemeLabel(theme)}. Switch to light, dark, or auto.`}
        aria-haspopup="true"
        aria-expanded={isOpen}
        title="Switch theme (light, dark, auto)"
        onClick={() => setIsOpen((prev) => !prev)}
        className="flex items-center space-x-1.5 bg-[#1a1e28] hover:bg-[#202530] border border-[#282f40] rounded-lg px-2.5 py-1.5 text-xs font-medium text-google-gray-300 hover:text-white transition-all shadow-sm focus:outline-none focus:ring-1 focus:ring-google-blue"
      >
        <span className="flex-shrink-0">{getThemeIcon(theme)}</span>
        <span className="hidden sm:inline font-medium">
          {getThemeLabel(theme)}
        </span>
        <ChevronDown
          className={`h-3 w-3 text-google-gray-400 transition-transform duration-200 ${
            isOpen ? 'rotate-180 text-white' : ''
          }`}
        />
      </button>

      {isOpen && (
        <div
          role="menu"
          aria-orientation="vertical"
          aria-labelledby="theme-toggle-btn"
          className="absolute right-0 mt-1.5 w-44 rounded-xl bg-[#14171f] border border-[#282f40] shadow-2xl py-1 z-50 animate-fade-in focus:outline-none"
        >
          <div className="px-3 py-1.5 text-[10px] font-semibold uppercase tracking-wider text-google-gray-500 border-b border-[#202530] mb-1">
            Appearance
          </div>

          <button
            type="button"
            role="menuitem"
            data-testid="theme-option-light"
            onClick={() => {
              setTheme('light');
              setIsOpen(false);
            }}
            className={`w-full flex items-center justify-between px-3 py-2 text-xs transition-colors text-left ${
              theme === 'light'
                ? 'bg-google-blue/15 text-google-blue font-semibold'
                : 'text-google-gray-300 hover:bg-[#1f2430] hover:text-white'
            }`}
          >
            <div className="flex items-center space-x-2.5">
              <Sun className="h-4 w-4 text-amber-500 flex-shrink-0" />
              <div>
                <div className="font-medium">Light</div>
                <div className="text-[10px] text-google-gray-500">Bright background</div>
              </div>
            </div>
            {theme === 'light' && <Check className="h-4 w-4 text-google-blue ml-2 shrink-0" />}
          </button>

          <button
            type="button"
            role="menuitem"
            data-testid="theme-option-dark"
            onClick={() => {
              setTheme('dark');
              setIsOpen(false);
            }}
            className={`w-full flex items-center justify-between px-3 py-2 text-xs transition-colors text-left ${
              theme === 'dark'
                ? 'bg-google-blue/15 text-google-blue font-semibold'
                : 'text-google-gray-300 hover:bg-[#1f2430] hover:text-white'
            }`}
          >
            <div className="flex items-center space-x-2.5">
              <Moon className="h-4 w-4 text-google-blue flex-shrink-0" />
              <div>
                <div className="font-medium">Dark</div>
                <div className="text-[10px] text-google-gray-500">Dark FinOps styling</div>
              </div>
            </div>
            {theme === 'dark' && <Check className="h-4 w-4 text-google-blue ml-2 shrink-0" />}
          </button>

          <button
            type="button"
            role="menuitem"
            data-testid="theme-option-auto"
            onClick={() => {
              setTheme('auto');
              setIsOpen(false);
            }}
            className={`w-full flex items-center justify-between px-3 py-2 text-xs transition-colors text-left ${
              theme === 'auto'
                ? 'bg-google-blue/15 text-google-blue font-semibold'
                : 'text-google-gray-300 hover:bg-[#1f2430] hover:text-white'
            }`}
          >
            <div className="flex items-center space-x-2.5">
              <Monitor className="h-4 w-4 text-emerald-400 flex-shrink-0" />
              <div>
                <div className="font-medium">Auto</div>
                <div className="text-[10px] text-google-gray-500">
                  System ({effectiveTheme})
                </div>
              </div>
            </div>
            {theme === 'auto' && <Check className="h-4 w-4 text-google-blue ml-2 shrink-0" />}
          </button>
        </div>
      )}
    </div>
  );
};
