import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { Footer, releaseUrlFor } from './Footer';

const REPO = 'https://github.com/JulienBreux/agy-cost-board';

describe('Footer Component', () => {
  it('renders default footer links and labels correctly', () => {
    render(<Footer version="v0.5.0" />);

    // GitHub repository link
    const githubLink = screen.getByRole('link', { name: /github repository/i });
    expect(githubLink).toHaveAttribute('href', REPO);
    expect(githubLink).toHaveAttribute('target', '_blank');
    expect(githubLink).toHaveAttribute('rel', 'noopener noreferrer');

    // Project name and version
    expect(screen.getByText('AGY Cost Board')).toBeInTheDocument();
    const versionLink = screen.getByRole('link', { name: /v0\.5\.0/i });
    expect(versionLink).toHaveAttribute('href', `${REPO}/releases/tag/v0.5.0`);

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

  it('hides the version link until a version is known', () => {
    render(<Footer />);
    expect(screen.queryByTitle(/view release/i)).not.toBeInTheDocument();
    expect(screen.getByText('AGY Cost Board')).toBeInTheDocument();
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

  it('displays the raw version but links to the clean release tag', () => {
    render(<Footer version="v0.4.0-dirty" />);
    const versionLink = screen.getByRole('link', { name: 'v0.4.0-dirty' });
    expect(versionLink).toHaveAttribute('href', `${REPO}/releases/tag/v0.4.0`);
  });

  it('calls onOpenChangelog when version link is clicked with an onOpenChangelog handler', () => {
    const onOpenChangelog = vi.fn();
    render(<Footer version="v0.5.0" onOpenChangelog={onOpenChangelog} />);

    const versionLink = screen.getByRole('link', { name: /v0\.5\.0/i });
    fireEvent.click(versionLink);

    expect(onOpenChangelog).toHaveBeenCalledTimes(1);
  });
});

describe('releaseUrlFor', () => {
  it.each([
    ['v0.5.0', `${REPO}/releases/tag/v0.5.0`],
    ['0.5.0', `${REPO}/releases/tag/v0.5.0`], // goreleaser {{.Version}} without "v"
    ['v0.5.0-dirty', `${REPO}/releases/tag/v0.5.0`],
    ['v0.5.0-3-gabc1234', `${REPO}/releases/tag/v0.5.0`], // git describe between tags
    ['v0.5.0-3-gabc1234-dirty', `${REPO}/releases/tag/v0.5.0`],
    ['dev', `${REPO}/releases`],
    ['4d56be3', `${REPO}/releases`], // git describe --always without tags
    ['', `${REPO}/releases`],
  ])('maps %j to %s', (version, expected) => {
    expect(releaseUrlFor(REPO, version)).toBe(expected);
  });
});
