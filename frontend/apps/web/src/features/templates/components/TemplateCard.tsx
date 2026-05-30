import { Avatar } from "../../../components/ui/Avatar";
import { CodeChip } from "../../../components/ui/CodeChip";
import { StatusPill, type DocumentStatus } from "../../../components/ui/StatusPill";
import { MiniDocPreview } from "./MiniDocPreview";
import styles from "./TemplateCard.module.css";

type TemplateCardProps = {
  title: string;
  // ADR 0013: always present — published revision, else latest working revision (REV00 for drafts).
  version: string;
  status: Extract<DocumentStatus, "published" | "draft" | "archived">;
  author: string;
  updated: string;
  onClick?: () => void;
};

export function TemplateCard({ title, version, status, author, updated, onClick }: TemplateCardProps) {
  const firstName = author.split(" ")[0];

  return (
    <div
      className={styles.card}
      role="button"
      tabIndex={0}
      onClick={onClick}
      onKeyDown={(e) => {
        if ((e.key === "Enter" || e.key === " ") && onClick) {
          e.preventDefault();
          onClick();
        }
      }}
      aria-label={`Abrir template ${title}`}
    >
      <div className={styles.previewArea}>
        <MiniDocPreview />
        <div className={styles.badges}>
          <StatusPill status={status} />
          <CodeChip>{version}</CodeChip>
        </div>
      </div>
      <div className={styles.body}>
        <div className={styles.title}>{title}</div>
        <div className={styles.divider} />
        <div className={styles.footer}>
          <Avatar name={author} size="sm" />
          <span className={styles.author}>{firstName}</span>
          <span className={styles.time}>{updated}</span>
        </div>
      </div>
    </div>
  );
}
