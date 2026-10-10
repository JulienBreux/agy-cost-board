import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { Footer } from './Footer';

describe('Footer Component', () => {
  it('renders default footer links and labels correctly', () => {
    render(<Footer />);

    // GitHub repository link
    const githubLink = screen.getByRole('link', { name: /github repository/i });
    expect(githubLink).toHaveAttribute('href', 'https://github.com/JulienBreux/agy-cost-board');
    expect(githubLink).toHaveAttribute('target', '_blank');
    expect(githubLink).toHaveAttribute('rel', 'noopener noreferrer');

    // Project name and default version
    expect(screen.getByText('AGY Cost Board')).toBeInTheDocument();
    const versionLink = screen.getByRole('link', { name: /v0\.4\.0/i });
    expect(versionLink).toHaveAttribute(
      'href',
      'https://github.com/JulienBreux/agy-cost-board/releases/tag/v0.4.0'
    );

    // Author profile link
    const authorLink = screen.getByRole('link', { name: 'Julien Breux' });
    expect(authorLink).toHaveAttribute('href', 'https://github.com/JulienBreux');
    expect(authorLink).toHaveAttribute('target', '_blank');
    expect(authorLink).toHaveAttribute('rel', 'noopener noreferrer');

    // Heart emoji and text
    expect(screen.getByText('With')).toBeInTheDocument();
    expect(screen.getByText('❤️')).toBeInTheDocument();
    expect(screen.getByText('by')).toBeInTheDocument();
  });

  it('renders custom props when provided', () => {
    render(
      <Footer
        projectName="CustomApp"
        version="v2.5.0"
        githubRepoUrl="https://github.com/example/customapp"
        releasesUrl="https://github.com/example/customapp/releases"
        authorName="Jane Doe"
        authorUrl="https://github.com/janedoe"
      />
    );

    expect(screen.getByText('CustomApp')).toBeInTheDocument();

    const versionLink = screen.getByRole('link', { name: /v2\.5\.0/i });
    expect(versionLink).toHaveAttribute('href', 'https://github.com/example/customapp/releases');

    const githubLink = screen.getByRole('link', { name: /github repository/i });
    expect(githubLink).toHaveAttribute('href', 'https://github.com/example/customapp');

    const authorLink = screen.getByRole('link', { name: 'Jane Doe' });
    expect(authorLink).toHaveAttribute('href', 'https://github.com/janedoe');
  });

  it('handles dirty git versions by stripping suffix for release tag URL', () => {
    render(<Footer version="v0.4.0-dirty" />);
    const versionLink = screen.getByRole('link', { name: /v0\.4\.0-dirty/i });
    expect(versionLink).toBeInTheDocument();
    expect(versionLink).toHaveAttribute(
      'href',
      'https://github.com/JulienBreux/agy-cost-board/releases/tag/v0.4.0'
    );
  });

  it('links to releases root for non-tag dev version', () => {
    render(<Footer version="dev" />);
    const versionLink = screen.getByRole('link', { name: 'dev' });
    expect(versionLink).toHaveAttribute(
      'href',
      'https://github.com/JulienBreux/agy-cost-board/releases'
    );
  });
});
