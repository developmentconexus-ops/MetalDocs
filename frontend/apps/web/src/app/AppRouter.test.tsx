import type { RouteObject } from 'react-router-dom';
import { describe, expect, it } from 'vitest';
import { dashboardRoutes } from '../features/dashboard/routes';

/**
 * Root index-route invariant (M0/F0.2, preserved through F0.3).
 *
 * The protected app must mount exactly ONE index route at the root layout level so `/` resolves
 * unambiguously to the Dashboard. React-Router treats two sibling index routes as a conflict; the
 * pre-M0 router declared two (Dashboard + the dead Operations stub). F0.2 removed the Operations
 * index; F0.3 deleted the Operations/Audit feature folders outright. Dashboard is now the sole
 * carrier of a root-level index route — this guard fails if a second root index is (re)introduced.
 *
 * Asserted on the source route-config array (not the instantiated `router`) — instantiating
 * `createBrowserRouter` would start a real navigation, which is a side effect unit tests must not
 * trigger. `dashboardRoutes` is the only array that contributes a root-level index entry to
 * `AppRouter`'s protected children.
 */

const rootIndexCarriers: ReadonlyArray<readonly RouteObject[]> = [dashboardRoutes];

describe('root index route', () => {
  it('declares exactly one root-level index route', () => {
    const rootIndexRoutes = rootIndexCarriers.flat().filter((route) => route.index === true);
    expect(rootIndexRoutes).toHaveLength(1);
  });

  it('the sole root index route is Dashboard', () => {
    expect(dashboardRoutes.filter((route) => route.index === true)).toHaveLength(1);
  });
});
