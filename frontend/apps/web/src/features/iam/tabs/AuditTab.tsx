import { useCallback, useEffect, useMemo, useState } from "react";
import { useSearchParams } from "react-router-dom";
import AuditEventsTable from "../components/AuditEventsTable";
import AuditExportButton from "../components/AuditExportButton";
import AuditFilterBar, {
  type AuditDatePreset,
  type AuditFilterValue,
} from "../components/AuditFilterBar";
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

  const queryParams: AuditEventsQueryParams = useMemo(
    () => ({
      limit: PAGE_SIZE,
      actorId: filterValue.actorId,
      action: filterValue.action,
      resourceType: filterValue.resourceType,
      resourceId: filterValue.resourceId,
      occurredAfter: dateWindow.occurredAfter,
      occurredBefore: dateWindow.occurredBefore,
      q: filterValue.q,
    }),
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

  const exportFilter = useMemo(
    () => ({
      actorId: filterValue.actorId,
      action: filterValue.action,
      resourceType: filterValue.resourceType,
      resourceId: filterValue.resourceId,
      occurredAfter: dateWindow.occurredAfter,
      occurredBefore: dateWindow.occurredBefore,
      q: filterValue.q,
    }),
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
