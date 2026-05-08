import { Avatar } from "../../../components/ui/Avatar";
import { CodeChip } from "../../../components/ui/CodeChip";
import { StatusPill, type DocumentStatus } from "../../../components/ui/StatusPill";
import { MiniDocPreview } from "./MiniDocPreview";
import styles from "./TemplateCard.module.css";

type TemplateCardProps = {
  title: string;
  version: string;
  status: Extract<DocumentStatus, "published" | "draft" | "archived">;
  author: string;
  updated: string;
};

export function TemplateCard({ title, version, status, author, updated }: TemplateCardProps) {
  const firstName = author.split(" ")[0];

  return (
    <div className={styles.card}>
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
