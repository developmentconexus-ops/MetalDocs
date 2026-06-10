import { useCallback, useEffect, useMemo, useState } from "react";
import { useSearchParams } from "react-router-dom";
import AuditEventsTable from "../components/AuditEventsTable";
import AuditExportButton from "../components/AuditExportButton";
import AuditFilterBar, {
  type AuditDatePreset,
  type AuditFilterValue,
} from "../components/AuditFilterBar";
import type { AuditExportFilter } from "../mutations/useExportAuditMutation";
import {
  useAuditEventsQuery,
  type AuditEventsQueryParams,
} from "../queries/useAuditEventsQuery";
import styles from "./AuditTab.module.css";

const PAGE_SIZE = 50;

const PRESET_VALUES: ReadonlyArray<AuditDatePreset> = [
  "24h",
  "7d",
  "30d",
  "90d",
  "custom",
];

const PRESET_MS: Readonly<Record<Exclude<AuditDatePreset, "custom">, number>> = {
  "24h": 24 * 60 * 60 * 1000,
  "7d": 7 * 24 * 60 * 60 * 1000,
  "30d": 30 * 24 * 60 * 60 * 1000,
  "90d": 90 * 24 * 60 * 60 * 1000,
};

// Rolling presets recompute their `occurredAfter` against the live `now` —
// refreshed every minute via `windowKey` so a long-idle tab does not freeze
// the filter window. 1-minute granularity is adequate for an audit timeline;
// finer ticks would just thrash the query cache.
const ROLLING_WINDOW_TICK_MS = 60_000;

function readPreset(raw: string | null): AuditDatePreset | undefined {
  if (!raw) return undefined;
  return PRESET_VALUES.includes(raw as AuditDatePreset)
    ? (raw as AuditDatePreset)
    : undefined;
}

export function resolveDateWindow(
  preset: AuditDatePreset | undefined,
  occurredAfter: string | undefined,
  occurredBefore: string | undefined,
  now: number = Date.now(),
): { occurredAfter?: string; occurredBefore?: string } {
  if (preset === "custom") {
    return { occurredAfter, occurredBefore };
  }
  if (preset) {
    const ms = PRESET_MS[preset];
    return { occurredAfter: new Date(now - ms).toISOString() };
  }
  return {};
}

export default function AuditTab() {
  const [searchParams, setSearchParams] = useSearchParams();

  const filterValue: AuditFilterValue = useMemo(
    () => ({
      actorId: searchParams.get("actorId") ?? undefined,
      action: searchParams.get("action") ?? undefined,
      resourceType: searchParams.get("resourceType") ?? undefined,
      resourceId: searchParams.get("resourceId") ?? undefined,
      datePreset: readPreset(searchParams.get("datePreset")),
      occurredAfter: searchParams.get("occurredAfter") ?? undefined,
      occurredBefore: searchParams.get("occurredBefore") ?? undefined,
      q: searchParams.get("q") ?? undefined,
    }),
    [searchParams],
  );

  const isRollingPreset =
    !!filterValue.datePreset && filterValue.datePreset !== "custom";

  const [windowKey, setWindowKey] = useState<number>(() => Date.now());
  useEffect(() => {
    if (!isRollingPreset) return;
    const id = setInterval(
      () => setWindowKey(Date.now()),
      ROLLING_WINDOW_TICK_MS,
    );
    return () => clearInterval(id);
  }, [isRollingPreset]);

  const dateWindow = useMemo(
    () =>
      resolveDateWindow(
        filterValue.datePreset,
        filterValue.occurredAfter,
        filterValue.occurredBefore,
        windowKey,
      ),
    [
      filterValue.datePreset,
      filterValue.occurredAfter,
      filterValue.occurredBefore,
      windowKey,
    ],
  );

  // Wire object: the keys here go straight onto GET /audit/events, and the
  // handler reads snake_case (actor_id/resource_type/…). The camelCase UI
  // filterValue is mapped to those wire keys here. `satisfies` (not a `:`
  // annotation) excess-checks the literal, so any future camel-key drift is a
  // tsc error rather than a silently-ignored param (the bug this rename fixed).
  const queryParams = useMemo(
    () =>
      ({
        limit: PAGE_SIZE,
        actor_id: filterValue.actorId,
        action: filterValue.action,
        resource_type: filterValue.resourceType,
        resource_id: filterValue.resourceId,
        occurred_after: dateWindow.occurredAfter,
        occurred_before: dateWindow.occurredBefore,
        q: filterValue.q,
      }) satisfies AuditEventsQueryParams,
    [
      filterValue.actorId,
      filterValue.action,
      filterValue.resourceType,
      filterValue.resourceId,
      filterValue.q,
      dateWindow.occurredAfter,
      dateWindow.occurredBefore,
    ],
  );

  const eventsQuery = useAuditEventsQuery(queryParams);

  const events = useMemo(
    () => eventsQuery.data?.pages.flatMap((p) => p.items) ?? [],
    [eventsQuery.data],
  );

  const lastPage = eventsQuery.data?.pages.at(-1);
  const hasMore = lastPage?.hasMore ?? false;

  const handleFilterChange = useCallback(
    (next: AuditFilterValue) => {
      const params = new URLSearchParams(searchParams);
      const apply = (key: string, val: string | undefined) => {
        if (val) params.set(key, val);
        else params.delete(key);
      };
      apply("actorId", next.actorId);
      apply("action", next.action);
      apply("resourceType", next.resourceType);
      apply("resourceId", next.resourceId);
      apply("datePreset", next.datePreset);
      apply("occurredAfter", next.occurredAfter);
      apply("occurredBefore", next.occurredBefore);
      apply("q", next.q);
      setSearchParams(params, { replace: true });
    },
    [searchParams, setSearchParams],
  );

  const handleLoadMore = useCallback(() => {
    if (eventsQuery.hasNextPage && !eventsQuery.isFetchingNextPage) {
      eventsQuery.fetchNextPage();
    }
  }, [eventsQuery]);

  // Export POST body filter — serialized into the request, so it must carry the
  // snake_case keys the handler decodes (filter.actor_id/resource_type/…). Same
  // camel→snake mapping + `satisfies` guard as queryParams above.
  const exportFilter = useMemo(
    () =>
      ({
        actor_id: filterValue.actorId,
        action: filterValue.action,
        resource_type: filterValue.resourceType,
        resource_id: filterValue.resourceId,
        occurred_after: dateWindow.occurredAfter,
        occurred_before: dateWindow.occurredBefore,
        q: filterValue.q,
      }) satisfies AuditExportFilter,
    [filterValue, dateWindow.occurredAfter, dateWindow.occurredBefore],
  );

  const isInitialLoading = eventsQuery.isLoading;

  return (
    <section
      className={styles.tab}
      data-testid="admin-audit-tab"
      aria-labelledby="admin-audit-heading"
    >
      <div className={styles.headerRow}>
        <div className={styles.titleBlock}>
          <h2 id="admin-audit-heading">Auditoria</h2>
          <p className={styles.lede}>
            Trilha append-only de ações no workspace. Filtros gravam URL para
            compartilhamento e investigação reproduzível.
          </p>
        </div>
        <AuditExportButton filter={exportFilter} />
      </div>

      <div className={styles.toolbar} role="region" aria-label="Filtros">
        <AuditFilterBar value={filterValue} onChange={handleFilterChange} />
      </div>

      <AuditEventsTable
        events={events}
        isLoading={isInitialLoading}
        error={eventsQuery.error}
        onRetry={() => eventsQuery.refetch()}
        hasMore={hasMore}
        isFetchingMore={eventsQuery.isFetchingNextPage}
        onLoadMore={handleLoadMore}
      />
    </section>
  );
}
