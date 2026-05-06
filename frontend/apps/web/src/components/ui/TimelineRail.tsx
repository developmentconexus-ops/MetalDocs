import type { ReactNode } from "react";
import styles from "./TimelineRail.module.css";

export type TimelineRailItem = {
  id: string;
  title: ReactNode;
  subtitle?: ReactNode;
  aside?: ReactNode;
  active?: boolean;
  onClick?: () => void;
};

type TimelineRailProps = {
  items: TimelineRailItem[];
  emptyState?: ReactNode;
  accent?: "blue" | "gold" | "green";
  ariaLabel?: string;
  variant?: "card" | "flat";
};

function railClassName(accent: TimelineRailProps["accent"]) {
  switch (accent) {
    case "gold":
      return `${styles.root} ${styles.rootGold}`;
    case "green":
      return `${styles.root} ${styles.rootGreen}`;
    case "blue":
    default:
      return `${styles.root} ${styles.rootBlue}`;
  }
}

export function TimelineRail({
  items,
  emptyState = "Sem itens para exibir.",
  accent = "blue",
  ariaLabel = "Timeline",
  variant = "card",
}: TimelineRailProps) {
  if (items.length === 0) {
    return <div className={styles.emptyState}>{emptyState}</div>;
  }

  return (
    <div className={railClassName(accent)} aria-label={ariaLabel}>
      {items.map((item, index) => {
        const isLast = index === items.length - 1;
        const isActive = item.active ?? index === 0;
        const shellClassName = `${styles.itemShell}${isLast ? ` ${styles.itemShellLast}` : ""}`;
        const markerClassName = `${styles.marker}${isActive ? ` ${styles.markerActive}` : ""}`;

        const cardContent = (
          <>
            <div className={styles.itemContent}>
              <strong className={styles.itemTitle}>{item.title}</strong>
              {item.subtitle ? <small className={styles.itemSubtitle}>{item.subtitle}</small> : null}
            </div>
            {item.aside ? <span className={styles.itemAside}>{item.aside}</span> : null}
          </>
        );

        const flatContent = (
          <>
            <strong className={styles.itemFlatTitle}>{item.title}</strong>
            {item.subtitle ? <span className={styles.itemFlatSubtitle}>{item.subtitle}</span> : null}
            {item.aside ? <span className={styles.itemFlatAside}>{item.aside}</span> : null}
          </>
        );

        return (
          <div key={item.id} className={shellClassName}>
            <span className={markerClassName} aria-hidden="true" />
            {variant === "flat" ? (
              item.onClick ? (
                <button type="button" className={styles.itemFlatButton} onClick={item.onClick}>
                  {flatContent}
                </button>
              ) : (
                <div className={styles.itemFlat}>{flatContent}</div>
              )
            ) : item.onClick ? (
              <button type="button" className={styles.itemCardButton} onClick={item.onClick}>
                {cardContent}
              </button>
            ) : (
              <div className={styles.itemCardStatic}>{cardContent}</div>
            )}
          </div>
        );
      })}
    </div>
  );
}
