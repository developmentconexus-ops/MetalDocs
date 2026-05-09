import type React from 'react';
import { Fragment } from 'react';
import { Icon } from '../../../components/ui/Icon';
import styles from './DocumentHero.module.css';

interface DocumentHeroProps {
  breadcrumbItems: Array<{ label: string; href?: string }>;
  docCard: React.ReactNode;
  badges: React.ReactNode;
  title: string;
  subtitle: React.ReactNode;
  actions: React.ReactNode;
}

export function DocumentHero({
  breadcrumbItems,
  docCard,
  badges,
  title,
  subtitle,
  actions,
}: DocumentHeroProps) {
  return (
    <header className={styles.hero}>
      <div className={styles.heroOverlay} />

      <nav aria-label="Breadcrumb" className={styles.breadcrumb}>
        {breadcrumbItems.map((item, index) => {
          const isLast = index === breadcrumbItems.length - 1;

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
