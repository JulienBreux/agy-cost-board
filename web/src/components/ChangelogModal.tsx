import React, { useEffect, useRef } from 'react';
import { Sparkles } from 'lucide-react';
import { ChangelogRelease, CHANGELOG_DATA } from '../data/changelog';

export interface ChangelogModalProps {
  isOpen: boolean;
  onClose: () => void;
  releases?: ChangelogRelease[];
  projectName?: string;
}

export const ChangelogModal: React.FC<ChangelogModalProps> = ({
  isOpen,
  onClose,
  releases = CHANGELOG_DATA,
  projectName = 'AGY Cost Board',
}) => {
  const modalRef = useRef<HTMLDivElement>(null);

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

  // Prevent background scroll when modal is open
  useEffect(() => {
    if (isOpen) {
      document.body.style.overflow = 'hidden';
    } else {
      document.body.style.overflow = '';
    }
    return () => {
      document.body.style.overflow = '';
    };
  }, [isOpen]);

  if (!isOpen) return null;

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/75 backdrop-blur-sm animate-in fade-in duration-200"
      data-testid="changelog-backdrop"
      onClick={(e) => {
        if (e.target === e.currentTarget) {
          onClose();
        }
      }}
    >
      <div
        ref={modalRef}
        role="dialog"
        aria-modal="true"
        aria-labelledby="changelog-modal-title"
        className="relative w-full max-w-xl max-h-[85vh] flex flex-col bg-[#161a23] border border-[#232936] text-white rounded-2xl shadow-2xl overflow-hidden animate-in zoom-in-95 duration-200"
        onClick={(e) => e.stopPropagation()}
      >
        {/* Header matching screenshot */}
        <div className="px-6 pt-6 pb-4 border-b border-[#232936]">
          <div className="flex items-center space-x-2 text-amber-400">
            <Sparkles className="w-5 h-5 fill-amber-400" aria-hidden="true" />
            <h2
              id="changelog-modal-title"
              className="text-lg font-bold text-white tracking-tight"
            >
              What's new in {projectName}
            </h2>
          </div>
          <p className="mt-1 text-xs sm:text-sm text-google-gray-400">
            Updates to {projectName} — scroll for earlier versions.
          </p>
        </div>

        {/* Scrollable Changelog List */}
        <div className="overflow-y-auto px-6 py-5 space-y-6 flex-1 text-sm text-google-gray-300 divide-y divide-[#232936]/60">
          {releases.map((release, releaseIdx) => {
            const isLatest = releaseIdx === 0;

            return (
              <div
                key={release.version}
                className={releaseIdx > 0 ? 'pt-6' : ''}
              >
                {/* Version badge & Date */}
                <div className="flex items-center space-x-2.5 mb-2">
                  {isLatest ? (
                    <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-mono font-medium bg-[#1a73e8] text-white shadow-sm">
                      {release.version}
                    </span>
                  ) : (
                    <span className="inline-flex items-center text-xs font-mono font-semibold text-google-gray-400">
                      {release.version}
                    </span>
                  )}
                  <span className="text-xs text-google-gray-400 font-normal">
                    {release.date}
                  </span>
                </div>

                {/* Release headline */}
                <h3 className="text-sm font-semibold text-white mb-2 leading-snug">
                  {release.title}
                </h3>

                {/* Bullets */}
                <ul className="space-y-2 list-disc pl-4 text-xs sm:text-sm text-google-gray-300 marker:text-google-gray-500">
                  {release.items.map((item, itemIdx) => (
                    <li key={itemIdx} className="leading-relaxed">
                      {item.title ? (
                        <>
                          <strong className="font-semibold text-white">
                            {item.title}:
                          </strong>{' '}
                          <span>{item.description}</span>
                        </>
                      ) : (
                        <span>{item.description}</span>
                      )}
                    </li>
                  ))}
                </ul>
              </div>
            );
          })}
        </div>

        {/* Footer with "Got it" button */}
        <div className="px-6 py-4 border-t border-[#232936] bg-[#131720]/80 backdrop-blur-sm flex justify-end items-center">
          <button
            type="button"
            onClick={onClose}
            className="px-5 py-2 rounded-xl text-sm font-semibold bg-[#1a73e8] hover:bg-[#1557b0] active:scale-95 text-white shadow-sm transition-all focus:outline-none focus:ring-2 focus:ring-google-blue focus:ring-offset-2 focus:ring-offset-[#161a23]"
          >
            Got it
          </button>
        </div>
      </div>
    </div>
  );
};
