type AvatarSize = 'sm' | 'md' | 'lg';

type AvatarProps = {
  name: string;
  size?: AvatarSize;
};

function initials(name: string): string {
  const parts = name.trim().split(/\s+/);
  if (parts.length === 1) return parts[0].slice(0, 2).toUpperCase();
  return (parts[0][0] + parts[parts.length - 1][0]).toUpperCase();
}

const sizeClass: Record<AvatarSize, string> = {
  sm: 'avatar avatar-sm',
  md: 'avatar',
  lg: 'avatar avatar-lg',
};

export function Avatar({ name, size = 'md' }: AvatarProps) {
  return (
    <span className={sizeClass[size]} title={name} aria-label={name}>
      {initials(name)}
    </span>
  );
}
