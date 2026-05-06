import styles from './Avatar.module.css';

type AvatarSize = 'sm' | 'md' | 'lg';

type AvatarProps = {
  name: string;
  size?: AvatarSize;
  /** Optional background override (e.g. hashed-from-name color). Falls back to --brand. */
  color?: string;
};

function initials(name: string): string {
  // Split on whitespace, hyphen, underscore, dot — so "author-test" → "AT", "Marina Silva" → "MS".
  const parts = name.trim().split(/[\s\-_.]+/).filter(Boolean);
  if (parts.length === 0) return '?';
  if (parts.length === 1) return parts[0].slice(0, 2).toUpperCase();
  return (parts[0][0] + parts[parts.length - 1][0]).toUpperCase();
}

const sizeClass: Record<AvatarSize, string> = {
  sm: `${styles.avatar} ${styles.sm}`,
  md: styles.avatar,
  lg: `${styles.avatar} ${styles.lg}`,
};

export function Avatar({ name, size = 'md', color }: AvatarProps) {
  return (
    <span
      className={sizeClass[size]}
      title={name}
      aria-label={name}
      style={color ? { background: color } : undefined}
    >
      {initials(name)}
    </span>
  );
}
