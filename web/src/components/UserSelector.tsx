import React, { useState, useRef, useEffect } from 'react';
import { User, ChevronDown, Search, Check, RotateCcw } from 'lucide-react';

interface UserSelectorProps {
  currentUser: string;
  selectedUser: string;
  availableUsers: string[];
  onSelectUser: (userId: string) => void;
}

export const UserSelector: React.FC<UserSelectorProps> = ({
  currentUser,
  selectedUser,
  availableUsers,
  onSelectUser,
}) => {
  const [isOpen, setIsOpen] = useState(false);
  const [searchTerm, setSearchTerm] = useState('');
  const dropdownRef = useRef<HTMLDivElement>(null);

  const isSelf = currentUser && selectedUser && currentUser.toLowerCase() === selectedUser.toLowerCase();

  // Combine and sort unique users, putting current user first
  const uniqueUsers = Array.from(new Set([currentUser, ...availableUsers].filter(Boolean)));
  const filteredUsers = uniqueUsers.filter((u) =>
    u.toLowerCase().includes(searchTerm.toLowerCase())
  );

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setIsOpen(false);
      }
    };
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  return (
    <div className="relative" ref={dropdownRef}>
      <div className="flex items-center space-x-2">
        <div className="flex items-center space-x-2 bg-[#1a1e27] border border-[#202530] rounded-lg px-3 py-1.5 shadow-sm">
          <div className="w-6 h-6 rounded-full bg-google-blue/20 text-google-blue flex items-center justify-center font-medium text-xs">
            <User className="w-3.5 h-3.5" />
          </div>
          <div className="flex items-center space-x-1.5">
            <span className="text-sm font-medium text-white max-w-[200px] sm:max-w-xs truncate">
              {selectedUser}
            </span>
            {isSelf && (
              <span className="px-1.5 py-0.5 text-[10px] font-semibold tracking-wide uppercase bg-google-blue/15 text-google-blue border border-google-blue/30 rounded">
                You
              </span>
            )}
          </div>
          <button
            type="button"
            aria-label="Switch User"
            onClick={() => setIsOpen((prev) => !prev)}
            className="ml-1 p-1 text-google-gray-400 hover:text-white rounded hover:bg-[#202530] transition-colors"
          >
            <ChevronDown className={`w-4 h-4 transition-transform duration-200 ${isOpen ? 'rotate-180' : ''}`} />
          </button>
        </div>

        {!isSelf && currentUser && (
          <button
            type="button"
            onClick={() => onSelectUser(currentUser)}
            aria-label="View My Dashboard"
            className="flex items-center space-x-1 px-2.5 py-1.5 text-xs font-medium text-google-gray-300 hover:text-white bg-[#1a1e27] hover:bg-[#202530] border border-[#202530] rounded-lg transition-colors"
            title="Reset to your personal dashboard"
          >
            <RotateCcw className="w-3.5 h-3.5 text-google-blue" />
            <span className="hidden sm:inline">View My Dashboard</span>
          </button>
        )}
      </div>

      {isOpen && (
        <div className="absolute left-0 mt-2 w-72 sm:w-80 bg-[#161a22] border border-[#2d3548] rounded-xl shadow-2xl z-50 overflow-hidden animate-in fade-in zoom-in-95 duration-100">
          <div className="p-2 border-b border-[#202530]">
            <div className="relative">
              <Search className="w-4 h-4 absolute left-2.5 top-1/2 -translate-y-1/2 text-google-gray-400" />
              <input
                type="text"
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                placeholder="Search developer email..."
                autoFocus
                className="w-full bg-[#111318] border border-[#202530] rounded-lg pl-8 pr-3 py-1.5 text-xs text-white placeholder-google-gray-500 focus:outline-none focus:border-google-blue"
              />
            </div>
          </div>

          <div className="max-h-60 overflow-y-auto py-1">
            {filteredUsers.length === 0 ? (
              <div className="px-4 py-3 text-xs text-google-gray-400 text-center">
                No developers found
              </div>
            ) : (
              filteredUsers.map((user) => {
                const isCurrentActive = user === selectedUser;
                const isUserSelf = currentUser && user.toLowerCase() === currentUser.toLowerCase();

                return (
                  <button
                    key={user}
                    type="button"
                    onClick={() => {
                      onSelectUser(user);
                      setIsOpen(false);
                      setSearchTerm('');
                    }}
                    className={`w-full flex items-center justify-between px-3 py-2 text-xs text-left transition-colors ${
                      isCurrentActive
                        ? 'bg-google-blue/15 text-google-blue'
                        : 'text-google-gray-300 hover:bg-[#1f2430] hover:text-white'
                    }`}
                  >
                    <div className="flex items-center space-x-2 truncate">
                      <div className="w-5 h-5 rounded-full bg-[#202530] flex items-center justify-center text-[10px] text-google-gray-300">
                        {user.charAt(0).toUpperCase()}
                      </div>
                      <span className="truncate">{user}</span>
                      {isUserSelf && (
                        <span className="px-1 text-[9px] uppercase font-semibold bg-google-blue/20 text-google-blue rounded">
                          You
                        </span>
                      )}
                    </div>
                    {isCurrentActive && <Check className="w-3.5 h-3.5 text-google-blue shrink-0 ml-2" />}
                  </button>
                );
              })
            )}
          </div>
        </div>
      )}
    </div>
  );
};
