import React from 'react';

export interface FooterProps {
  projectName?: string;
  version?: string;
  githubRepoUrl?: string;
  releasesUrl?: string;
  authorName?: string;
  authorUrl?: string;
}

export const Footer: React.FC<FooterProps> = ({
  projectName = 'AGY Cost Board',
  version = 'v0.4.0',
  githubRepoUrl = 'https://github.com/JulienBreux/agy-cost-board',
  releasesUrl,
  authorName = 'Julien Breux',
  authorUrl = 'https://github.com/JulienBreux',
}) => {
  // Strip git dirty flag suffix (-dirty) for GitHub release tag URL
  const cleanTag = version.replace(/-dirty$/, '');
  const isSemanticTag = /^v?\d+\.\d+\.\d+/.test(cleanTag);
  const effectiveReleasesUrl =
    releasesUrl ||
    (isSemanticTag ? `${githubRepoUrl}/releases/tag/${cleanTag}` : `${githubRepoUrl}/releases`);

  return (
    <footer className="border-t border-[#1e2330] bg-[#0c0e14] py-4 text-xs sm:text-sm text-google-gray-400">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 flex flex-col sm:flex-row items-center justify-between gap-3 sm:gap-0">
        {/* Left: GitHub icon & link • Project Name • Version */}
        <div className="flex items-center space-x-2">
          <a
            href={githubRepoUrl}
            target="_blank"
            rel="noopener noreferrer"
            className="flex items-center space-x-1.5 text-google-gray-300 hover:text-white transition-colors group"
            aria-label="GitHub repository"
          >
            <svg
              className="h-4 w-4 fill-current text-google-gray-400 group-hover:text-white transition-colors"
              viewBox="0 0 24 24"
              aria-hidden="true"
            >
              <path
                fillRule="evenodd"
                clipRule="evenodd"
                d="M12 2C6.477 2 2 6.484 2 12.017c0 4.425 2.865 8.18 6.839 9.504.5.092.682-.217.682-.483 0-.237-.008-.868-.013-1.703-2.782.605-3.369-1.343-3.369-1.343-.454-1.158-1.11-1.466-1.11-1.466-.908-.62.069-.608.069-.608 1.003.07 1.53 1.032 1.53 1.032.892 1.53 2.341 1.088 2.91.832.092-.647.35-1.088.636-1.338-2.22-.253-4.555-1.113-4.555-4.951 0-1.093.39-1.988 1.029-2.688-.103-.253-.446-1.272.098-2.65 0 0 .84-.27 2.75 1.026A9.564 9.564 0 0112 6.844c.85.004 1.705.115 2.504.337 1.909-1.296 2.747-1.027 2.747-1.027.546 1.379.202 2.398.1 2.651.64.7 1.028 1.595 1.028 2.688 0 3.848-2.339 4.695-4.566 4.943.359.309.678.92.678 1.855 0 1.338-.012 2.419-.012 2.747 0 .268.18.58.688.482A10.019 10.019 0 0022 12.017C22 6.484 17.522 2 12 2z"
              />
            </svg>
            <span className="font-medium">GitHub</span>
          </a>

          <span className="text-google-gray-600 select-none">•</span>

          <span className="text-google-gray-300 font-medium">
            {projectName}
          </span>

          <a
            href={effectiveReleasesUrl}
            target="_blank"
            rel="noopener noreferrer"
            className="font-mono text-google-blue hover:text-blue-400 hover:underline transition-colors font-medium ml-1"
            title={`View release ${version}`}
          >
            {version}
          </a>
        </div>

        {/* Right: With ❤️ by Julien Breux */}
        <div className="flex items-center text-google-gray-400">
          <span>With</span>
          <span className="mx-1 text-red-500 select-none" role="img" aria-label="love">
            ❤️
          </span>
          <span>by</span>
          <a
            href={authorUrl}
            target="_blank"
            rel="noopener noreferrer"
            className="ml-1 font-semibold text-white hover:text-google-blue transition-colors underline-offset-2 hover:underline"
          >
            {authorName}
          </a>
        </div>
      </div>
    </footer>
  );
};
