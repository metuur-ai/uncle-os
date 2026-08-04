import { TabType } from './types';
import { LESSONS } from './lessons';

/**
 * Hash-based routing.
 *
 * The site is a static build served from a project subpath on GitHub Pages, so
 * History API paths (/uncle-os/cli) would 404 on refresh or deep-link. The hash
 * survives a hard reload without any server rewrite.
 *
 * Routes are namespaced as `#/<tab>` so in-page anchors (`#main` from the skip
 * link) are never mistaken for navigation.
 */

const ROUTE_PREFIX = '#/';

const isTab = (value: string): value is TabType => LESSONS.some((l) => l.id === value);

/** The hash a tab should be reachable at. Home is the bare root route. */
export const hashForTab = (tab: TabType): string =>
  tab === 'home' ? ROUTE_PREFIX : `${ROUTE_PREFIX}${tab}`;

/**
 * The tab named by the current URL, or null when the hash is not a route —
 * an in-page anchor or an unknown id leaves the current view untouched.
 */
export const tabFromHash = (hash: string = window.location.hash): TabType | null => {
  if (hash === '' || hash === '#') return 'home';
  if (!hash.startsWith(ROUTE_PREFIX)) return null;

  const id = hash.slice(ROUTE_PREFIX.length);
  if (id === '') return 'home';
  return isTab(id) ? id : null;
};
