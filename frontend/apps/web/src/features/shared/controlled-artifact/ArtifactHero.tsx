import type React from "react";
import { Fragment } from "react";
import { Icon } from "../../../components/ui/Icon";
import type { BreadcrumbItem } from "./types";
import styles from "./ArtifactHero.module.css";

interface ArtifactHeroProps {
  breadcrumb: BreadcrumbItem[];
  docCard?: React.ReactNode;
  badges?: React.ReactNode;
  title: string;
  subtitle?: React.ReactNode;
  actions?: React.ReactNode;
}

export function ArtifactHero({
  breadcrumb,
  docCard,
  badges,
  title,
  subtitle,
  actions,
}: ArtifactHeroProps) {
  return (
    <header className={styles.hero}>
      <div className={styles.heroOverlay} />

      <nav aria-label="Breadcrumb" className={styles.breadcrumb}>
        {breadcrumb.map((item, index) => {
          const isLast = index === breadcrumb.length - 1;

          return (
            <Fragment key={`${item.label}-${index}`}>
              {item.href && !isLast ? (
                <a href={item.href} className={styles.breadcrumbLink}>
                  {item.label}
                </a>
              ) : (
                <span>{item.label}</span>
              )}
              {!isLast && (
                <Icon
                  name="chevron"
                  size={10}
                  className={styles.breadcrumbSep}
                />
              )}
            </Fragment>
          );
        })}
      </nav>

      <div className={styles.heroGrid}>
        <div className={styles.heroDocCard}>{docCard}</div>

        <div className={styles.heroContent}>
          <div className={styles.heroBadges}>{badges}</div>
          <h1 className={styles.heroTitle}>{title}</h1>
          <div className={styles.heroSubtitle}>{subtitle}</div>
          <div className={styles.heroActions}>{actions}</div>
        </div>
      </div>
    </header>
  );
}
