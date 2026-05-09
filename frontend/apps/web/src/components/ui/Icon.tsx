import type { CSSProperties } from 'react';

export type IconName =
  | 'home' | 'docs' | 'library' | 'registry' | 'template' | 'inbox'
  | 'workflow' | 'taxonomy' | 'users' | 'audit' | 'bell' | 'search'
  | 'plus' | 'chevron' | 'chevdown' | 'filter' | 'list' | 'lock'
  | 'check' | 'x' | 'download' | 'upload' | 'history' | 'link'
  | 'cog' | 'arrow' | 'sparkle' | 'more' | 'eye' | 'edit'
  | 'calendar' | 'clock' | 'shield' | 'mail' | 'chevron-right' | 'chevron-left';

const PATHS: Record<IconName, React.ReactNode> = {
  home:     <><path d="M3 10l7-6 7 6"/><path d="M5 9v8h10V9"/></>,
  docs:     <><path d="M5 3h7l4 4v13H5z"/><path d="M12 3v4h4"/></>,
  library:  <><path d="M3 4h4v14H3zM9 4h4v14H9z"/><path d="M15 5l3 1-3 13"/></>,
  registry: <><rect x="3" y="3" width="14" height="14" rx="1"/><path d="M3 7h14M3 11h14M3 15h14M7 3v14"/></>,
  template: <><rect x="3" y="3" width="14" height="14" rx="1"/><path d="M3 8h14M8 8v9"/></>,
  inbox:    <><path d="M3 12l3-8h8l3 8v5H3z"/><path d="M3 12h4l1 2h4l1-2h4"/></>,
  workflow: <><circle cx="5" cy="5" r="2"/><circle cx="15" cy="15" r="2"/><circle cx="15" cy="5" r="2"/><path d="M7 5h6M5 7v6a2 2 0 002 2h6"/></>,
  taxonomy: <><path d="M10 3v4M10 9v8M5 13h10M3 17h4M13 17h4"/><circle cx="10" cy="3" r="1.5"/></>,
  users:    <><circle cx="7" cy="7" r="3"/><path d="M2 17c0-3 2-5 5-5s5 2 5 5"/><circle cx="14" cy="8" r="2"/><path d="M11 17c0-2 1-3 3-3s4 1 4 3"/></>,
  audit:    <><path d="M5 3h7l4 4v13H5z"/><path d="M8 11l2 2 4-4"/></>,
  bell:     <><path d="M5 8a5 5 0 0110 0v4l1.5 2h-13L5 12V8z"/><path d="M8 16a2 2 0 004 0"/></>,
  search:   <><circle cx="9" cy="9" r="5"/><path d="M13 13l4 4"/></>,
  plus:     <><path d="M10 4v12M4 10h12"/></>,
  chevron:  <><path d="M7 4l6 6-6 6"/></>,
  chevdown: <><path d="M5 8l5 5 5-5"/></>,
  filter:   <><path d="M3 5h14M6 10h8M8 15h4"/></>,
  list:     <><path d="M3 5h14M3 10h14M3 15h14"/></>,
  lock:     <><rect x="4" y="9" width="12" height="9" rx="1"/><path d="M7 9V6a3 3 0 016 0v3"/></>,
  check:    <><path d="M4 10l4 4 8-8"/></>,
  x:        <><path d="M5 5l10 10M15 5L5 15"/></>,
  download: <><path d="M10 3v10M5 9l5 5 5-5M3 17h14"/></>,
  upload:   <><path d="M10 14V4M5 9l5-5 5 5M3 17h14"/></>,
  history:  <><path d="M3 10a7 7 0 1 0 2-5"/><path d="M3 3v4h4M10 6v5l3 2"/></>,
  link:     <><path d="M8 12l4-4M7 8h-2a3 3 0 000 6h2M13 12h2a3 3 0 000-6h-2"/></>,
  cog:      <><circle cx="10" cy="10" r="2.5"/><path d="M10 3v2M10 15v2M3 10h2M15 10h2M5 5l1.5 1.5M14 14l-1.5-1.5M5 15l1.5-1.5M14 6l-1.5 1.5"/></>,
  arrow:    <><path d="M5 10h10M11 6l4 4-4 4"/></>,
  sparkle:  <><path d="M10 3l1.5 4 4 1.5-4 1.5L10 14l-1.5-4L4 8.5l4-1.5z"/></>,
  more:     <><circle cx="5" cy="10" r="1.5"/><circle cx="10" cy="10" r="1.5"/><circle cx="15" cy="10" r="1.5"/></>,
  eye:      <><circle cx="10" cy="10" r="3"/><path d="M2 10s3-6 8-6 8 6 8 6-3 6-8 6-8-6-8-6z"/></>,
  edit:          <><path d="M4 16l1-4L13 4l3 3-8 8-4 1z"/><path d="M13 4l3 3"/></>,
  calendar:      <><rect x="3" y="4" width="14" height="14" rx="1"/><path d="M3 9h14M7 3v2M13 3v2"/></>,
  clock:         <><circle cx="10" cy="10" r="7"/><path d="M10 6v5l3 2"/></>,
  shield:        <><path d="M10 3l7 3v5c0 4-3 6.5-7 8-4-1.5-7-4-7-8V6l7-3z"/></>,
  mail:          <><rect x="2" y="5" width="16" height="12" rx="1"/><path d="M2 6l8 6 8-6"/></>,
  'chevron-right': <><path d="M7 4l6 6-6 6"/></>,
  'chevron-left':  <><path d="M13 4l-6 6 6 6"/></>,
};

type IconProps = {
  name: IconName;
  size?: number;
  className?: string;
  style?: CSSProperties;
};

export function Icon({ name, size = 16, className, style }: IconProps) {
  return (
    <svg
      width={size}
      height={size}
      viewBox="0 0 20 20"
      fill="none"
      stroke="currentColor"
      strokeWidth="1.5"
      strokeLinecap="round"
      strokeLinejoin="round"
      className={className}
      style={style}
      aria-hidden="true"
    >
      {PATHS[name]}
    </svg>
  );
}
