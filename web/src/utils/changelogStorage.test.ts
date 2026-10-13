import { describe, it, expect, beforeEach, vi } from 'vitest';
import {
  getLastSeenVersion,
  setLastSeenVersion,
  hasNewVersion,
  parseSemver,
  CHANGELOG_COOKIE_NAME,
  CHANGELOG_STORAGE_KEY,
} from './changelogStorage';

describe('changelogStorage', () => {
  beforeEach(() => {
    // Clear cookies
    document.cookie.split(';').forEach((cookie) => {
      const eqPos = cookie.indexOf('=');
      const name = eqPos > -1 ? cookie.substr(0, eqPos).trim() : cookie.trim();
      document.cookie = `${name}=;expires=Thu, 01 Jan 1970 00:00:00 GMT;path=/`;
    });
    // Clear localStorage
    localStorage.clear();
  });

  describe('parseSemver', () => {
    it('parses standard semver strings with and without v prefix', () => {
      expect(parseSemver('v1.2.3')).toEqual([1, 2, 3]);
      expect(parseSemver('1.2.3')).toEqual([1, 2, 3]);
      expect(parseSemver('v0.5.1')).toEqual([0, 5, 1]);
    });

    it('parses semver with pre-release or build suffixes', () => {
      expect(parseSemver('v0.5.1-dirty')).toEqual([0, 5, 1]);
      expect(parseSemver('v0.5.1-1-gabcdef')).toEqual([0, 5, 1]);
      expect(parseSemver('1.0.0-rc.1')).toEqual([1, 0, 0]);
    });

    it('returns null for empty or non-semver strings', () => {
      expect(parseSemver('')).toBeNull();
      expect(parseSemver('dev')).toBeNull();
      expect(parseSemver('alpha')).toBeNull();
    });
  });

  describe('hasNewVersion', () => {
    it('returns true if no lastSeenVersion is set', () => {
      expect(hasNewVersion('v0.5.1', null)).toBe(true);
      expect(hasNewVersion('v0.5.1', '')).toBe(true);
    });

    it('returns false if currentVersion is invalid or empty', () => {
      expect(hasNewVersion('', 'v0.5.0')).toBe(false);
      expect(hasNewVersion('dev', 'v0.5.0')).toBe(false);
    });

    it('returns false if currentVersion equals lastSeenVersion', () => {
      expect(hasNewVersion('v0.5.1', 'v0.5.1')).toBe(false);
      expect(hasNewVersion('0.5.1', 'v0.5.1')).toBe(false);
    });

    it('returns true if currentVersion is strictly newer than lastSeenVersion', () => {
      expect(hasNewVersion('v0.5.1', 'v0.5.0')).toBe(true);
      expect(hasNewVersion('v0.6.0', 'v0.5.9')).toBe(true);
      expect(hasNewVersion('v1.0.0', 'v0.9.9')).toBe(true);
      expect(hasNewVersion('v0.5.1-dirty', 'v0.5.0')).toBe(true);
    });

    it('returns false if currentVersion is older than lastSeenVersion', () => {
      expect(hasNewVersion('v0.5.0', 'v0.5.1')).toBe(false);
      expect(hasNewVersion('v0.4.9', 'v0.5.0')).toBe(false);
    });
  });

  describe('getLastSeenVersion', () => {
    it('returns null when neither cookie nor localStorage is set', () => {
      expect(getLastSeenVersion()).toBeNull();
    });

    it('reads version from cookie when present', () => {
      document.cookie = `${CHANGELOG_COOKIE_NAME}=v0.5.0; path=/`;
      expect(getLastSeenVersion()).toBe('v0.5.0');
    });

    it('falls back to localStorage when cookie is absent', () => {
      localStorage.setItem(CHANGELOG_STORAGE_KEY, 'v0.4.0');
      expect(getLastSeenVersion()).toBe('v0.4.0');
    });

    it('prioritizes cookie over localStorage if both exist', () => {
      document.cookie = `${CHANGELOG_COOKIE_NAME}=v0.5.1; path=/`;
      localStorage.setItem(CHANGELOG_STORAGE_KEY, 'v0.5.0');
      expect(getLastSeenVersion()).toBe('v0.5.1');
    });
  });

  describe('setLastSeenVersion', () => {
    it('writes to both cookie and localStorage', () => {
      setLastSeenVersion('v0.5.1');
      expect(document.cookie).toContain(`${CHANGELOG_COOKIE_NAME}=v0.5.1`);
      expect(localStorage.getItem(CHANGELOG_STORAGE_KEY)).toBe('v0.5.1');
    });

    it('handles localStorage errors gracefully without throwing', () => {
      const setItemSpy = vi.spyOn(Storage.prototype, 'setItem').mockImplementation(() => {
        throw new Error('QuotaExceeded');
      });
      expect(() => setLastSeenVersion('v0.5.1')).not.toThrow();
      expect(document.cookie).toContain(`${CHANGELOG_COOKIE_NAME}=v0.5.1`);
      setItemSpy.mockRestore();
    });
  });
});
