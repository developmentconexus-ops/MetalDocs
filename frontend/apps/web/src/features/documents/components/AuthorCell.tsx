import { Avatar } from '../../../components/ui/Avatar';
import styles from './AuthorCell.module.css';

// Curated muted palette — saturated enough to read on white, desaturated
// enough to not fight the rest of the UI. Color is derived from the name
// hash so each author gets a stable identity across sessions.
const AVATAR_COLORS = [
  '#3f8382', '#3f5c83', '#834c3f', '#6b3f83', '#3f6b52',
  '#836a3f', '#3f7183', '#83433f', '#4a3f83', '#3f833f',
] as const;

function avatarColor(name: string): string {
  let hash = 0;
  for (let i = 0; i < name.length; i++) {
    hash = name.charCodeAt(i) + ((hash << 5) - hash);
  }
  return AVATAR_COLORS[Math.abs(hash) % AVATAR_COLORS.length];
}

function firstName(raw: string): string {
  const head = raw.split(/[\s\-_.]+/).filter(Boolean)[0] ?? raw;
  return head.charAt(0).toUpperCase() + head.slice(1);
}

type AuthorCellProps = {
  /** Username/display name from the backend (e.g. `author-test`, `Marina Silva`). */
  name: string | null | undefined;
};

export function AuthorCell({ name }: AuthorCellProps) {
  if (!name) {
    return <span className={styles.cell}>–</span>;
  }
  return (
    <span className={styles.cell}>
      <Avatar name={name} size="sm" color={avatarColor(name)} />
      <span className={styles.firstName}>{firstName(name)}</span>
    </span>
  );
}
