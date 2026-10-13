export const CHANGELOG_COOKIE_NAME = 'agy_last_seen_version';
export const CHANGELOG_STORAGE_KEY = 'agy_last_seen_version';

/**
 * Extracts numeric [major, minor, patch] from a semver-like string.
 * Strips 'v' prefix and build/pre-release suffixes (e.g., v0.5.1-dirty -> [0, 5, 1]).
 */
export function parseSemver(version: string): [number, number, number] | null {
  if (!version || typeof version !== 'string') {
    return null;
  }
  const match = /^v?(\d+)\.(\d+)\.(\d+)/.exec(version.trim());
  if (!match) {
    return null;
  }
  return [parseInt(match[1], 10), parseInt(match[2], 10), parseInt(match[3], 10)];
}

/**
 * Evaluates whether currentVersion is strictly newer than lastSeenVersion.
 * Returns true if no lastSeenVersion has been recorded and currentVersion is valid.
 */
export function hasNewVersion(currentVersion: string, lastSeenVersion: string | null | undefined): boolean {
  const current = parseSemver(currentVersion);
  if (!current) {
    return false;
  }
  if (!lastSeenVersion) {
    return true;
  }
  const last = parseSemver(lastSeenVersion);
  if (!last) {
    return true;
  }

  // Compare Major
  if (current[0] > last[0]) return true;
  if (current[0] < last[0]) return false;

  // Compare Minor
  if (current[1] > last[1]) return true;
  if (current[1] < last[1]) return false;

  // Compare Patch
  return current[2] > last[2];
}

/**
 * Retrieves the last-seen version from browser cookies, falling back to localStorage.
 */
export function getLastSeenVersion(): string | null {
  try {
    if (typeof document !== 'undefined' && document.cookie) {
      const match = document.cookie
        .split(';')
        .map((c) => c.trim())
        .find((c) => c.startsWith(`${CHANGELOG_COOKIE_NAME}=`));
      if (match) {
        const val = decodeURIComponent(match.substring(CHANGELOG_COOKIE_NAME.length + 1).trim());
        if (val) {
          return val;
        }
      }
    }
  } catch (err) {
    console.debug('Failed to read changelog cookie:', err);
  }

  try {
    if (typeof localStorage !== 'undefined') {
      const val = localStorage.getItem(CHANGELOG_STORAGE_KEY);
      if (val) {
        return val;
      }
    }
  } catch (err) {
    console.debug('Failed to read changelog from localStorage:', err);
  }

  return null;
}

/**
 * Persists the last-seen version into browser cookies (1 year TTL) and localStorage.
 */
export function setLastSeenVersion(version: string): void {
  if (!version) return;

  try {
    if (typeof document !== 'undefined') {
      const maxAgeSeconds = 365 * 24 * 60 * 60; // 1 year
      document.cookie = `${CHANGELOG_COOKIE_NAME}=${encodeURIComponent(version)}; Max-Age=${maxAgeSeconds}; Path=/; SameSite=Lax`;
    }
  } catch (err) {
    console.debug('Failed to write changelog cookie:', err);
  }

  try {
    if (typeof localStorage !== 'undefined') {
      localStorage.setItem(CHANGELOG_STORAGE_KEY, version);
    }
  } catch (err) {
    console.debug('Failed to write changelog to localStorage:', err);
  }
}
