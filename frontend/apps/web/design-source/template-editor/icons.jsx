/* global React */
// Lucide-style line icons. stroke-width 1.75, round caps, 24px viewBox.
// Size default 16; pass size={18} for rails.

const Ic = ({ children, size = 16, className = "", strokeWidth = 1.75 }) => (
  <svg
    width={size} height={size}
    viewBox="0 0 24 24"
    fill="none"
    stroke="currentColor"
    strokeWidth={strokeWidth}
    strokeLinecap="round"
    strokeLinejoin="round"
    className={className}
    aria-hidden="true"
  >{children}</svg>
);

const Icons = {
  file:         (p) => <Ic {...p}><path d="M14 3H7a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h10a2 2 0 0 0 2-2V8z"/><path d="M14 3v5h5"/></Ic>,
  menu:         (p) => <Ic {...p}><path d="M3 6h18M3 12h18M3 18h18"/></Ic>,
  send:         (p) => <Ic {...p}><path d="M22 2 11 13"/><path d="M22 2 15 22l-4-9-9-4z"/></Ic>,
  share:        (p) => <Ic {...p}><circle cx="18" cy="5" r="3"/><circle cx="6" cy="12" r="3"/><circle cx="18" cy="19" r="3"/><path d="m8.6 13.5 6.8 4M15.4 6.5l-6.8 4"/></Ic>,
  history:      (p) => <Ic {...p}><path d="M3 12a9 9 0 1 0 3-6.7L3 8"/><path d="M3 3v5h5"/><path d="M12 7v5l3 2"/></Ic>,
  check:        (p) => <Ic {...p}><path d="M20 6 9 17l-5-5"/></Ic>,
  chevronDown:  (p) => <Ic {...p}><path d="m6 9 6 6 6-6"/></Ic>,

  // rails
  blocks:       (p) => <Ic {...p}><rect x="3" y="3" width="7" height="7" rx="1"/><rect x="14" y="3" width="7" height="7" rx="1"/><rect x="3" y="14" width="7" height="7" rx="1"/><rect x="14" y="14" width="7" height="7" rx="1"/></Ic>,
  layoutGrid:   (p) => <Ic {...p}><rect x="3" y="3" width="7" height="7" rx="1"/><rect x="14" y="3" width="7" height="7" rx="1"/><rect x="3" y="14" width="7" height="7" rx="1"/><rect x="14" y="14" width="7" height="7" rx="1"/></Ic>,
  image:        (p) => <Ic {...p}><rect x="3" y="3" width="18" height="18" rx="2"/><circle cx="9" cy="9" r="2"/><path d="m21 15-5-5L5 21"/></Ic>,
  listTree:     (p) => <Ic {...p}><path d="M21 12h-8"/><path d="M21 6h-8"/><path d="M21 18h-8"/><path d="M3 6v12a2 2 0 0 0 2 2h3"/><path d="M3 6a2 2 0 0 1 2-2h3"/></Ic>,
  search:       (p) => <Ic {...p}><circle cx="11" cy="11" r="7"/><path d="m21 21-4.3-4.3"/></Ic>,
  panelRight:   (p) => <Ic {...p}><rect x="3" y="3" width="18" height="18" rx="2"/><path d="M15 3v18"/></Ic>,
  braces:       (p) => <Ic {...p}><path d="M8 3H7a2 2 0 0 0-2 2v5a2 2 0 0 1-2 2 2 2 0 0 1 2 2v5a2 2 0 0 0 2 2h1"/><path d="M16 21h1a2 2 0 0 0 2-2v-5a2 2 0 0 1 2-2 2 2 0 0 1-2-2V5a2 2 0 0 0-2-2h-1"/></Ic>,
  message:      (p) => <Ic {...p}><path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/></Ic>,
  gitBranch:    (p) => <Ic {...p}><circle cx="6" cy="3" r="2"/><circle cx="6" cy="15" r="2"/><circle cx="18" cy="9" r="2"/><path d="M6 5v8"/><path d="M18 11v1a2 2 0 0 1-2 2H8"/></Ic>,

  // toolbar
  undo:         (p) => <Ic {...p}><path d="M3 7v6h6"/><path d="M21 17a9 9 0 0 0-15-6.7L3 13"/></Ic>,
  redo:         (p) => <Ic {...p}><path d="M21 7v6h-6"/><path d="M3 17a9 9 0 0 1 15-6.7L21 13"/></Ic>,
  bold:         (p) => <Ic {...p}><path d="M6 4h8a4 4 0 0 1 0 8H6z"/><path d="M6 12h9a4 4 0 0 1 0 8H6z"/></Ic>,
  italic:       (p) => <Ic {...p}><line x1="19" y1="4" x2="10" y2="4"/><line x1="14" y1="20" x2="5" y2="20"/><line x1="15" y1="4" x2="9" y2="20"/></Ic>,
  underline:    (p) => <Ic {...p}><path d="M6 4v7a6 6 0 0 0 12 0V4"/><line x1="4" y1="21" x2="20" y2="21"/></Ic>,
  strike:       (p) => <Ic {...p}><path d="M16 4H9a3 3 0 0 0-2.83 4"/><path d="M14 12a4 4 0 0 1 0 8H6"/><line x1="4" y1="12" x2="20" y2="12"/></Ic>,
  textColor:    (p) => <Ic {...p}><path d="M4 20h16"/><path d="m6 16 6-12 6 12"/><path d="M8 12h8"/></Ic>,
  highlight:    (p) => <Ic {...p}><path d="m9 11-6 6v3h3l6-6"/><path d="m12 8 7-7 3 3-7 7"/><path d="m15 5 3 3"/></Ic>,
  alignLeft:    (p) => <Ic {...p}><line x1="3" y1="6" x2="21" y2="6"/><line x1="3" y1="12" x2="15" y2="12"/><line x1="3" y1="18" x2="18" y2="18"/></Ic>,
  alignCenter:  (p) => <Ic {...p}><line x1="3" y1="6" x2="21" y2="6"/><line x1="6" y1="12" x2="18" y2="12"/><line x1="4" y1="18" x2="20" y2="18"/></Ic>,
  alignRight:   (p) => <Ic {...p}><line x1="3" y1="6" x2="21" y2="6"/><line x1="9" y1="12" x2="21" y2="12"/><line x1="6" y1="18" x2="21" y2="18"/></Ic>,
  alignJustify: (p) => <Ic {...p}><line x1="3" y1="6" x2="21" y2="6"/><line x1="3" y1="12" x2="21" y2="12"/><line x1="3" y1="18" x2="21" y2="18"/></Ic>,
  listBullet:   (p) => <Ic {...p}><line x1="8" y1="6" x2="21" y2="6"/><line x1="8" y1="12" x2="21" y2="12"/><line x1="8" y1="18" x2="21" y2="18"/><circle cx="4" cy="6" r="1"/><circle cx="4" cy="12" r="1"/><circle cx="4" cy="18" r="1"/></Ic>,
  listNumber:   (p) => <Ic {...p}><line x1="10" y1="6" x2="21" y2="6"/><line x1="10" y1="12" x2="21" y2="12"/><line x1="10" y1="18" x2="21" y2="18"/><path d="M4 6h1v4"/><path d="M4 10h2"/><path d="M6 18H4c0-1 2-2 2-3s-1-1.5-2-1"/></Ic>,
  outdent:      (p) => <Ic {...p}><line x1="21" y1="6" x2="11" y2="6"/><line x1="21" y1="12" x2="11" y2="12"/><line x1="21" y1="18" x2="11" y2="18"/><path d="m7 8-4 4 4 4"/></Ic>,
  indent:       (p) => <Ic {...p}><line x1="21" y1="6" x2="11" y2="6"/><line x1="21" y1="12" x2="11" y2="12"/><line x1="21" y1="18" x2="11" y2="18"/><path d="m3 8 4 4-4 4"/></Ic>,
  link:         (p) => <Ic {...p}><path d="M10 13a5 5 0 0 0 7 0l3-3a5 5 0 0 0-7-7l-1 1"/><path d="M14 11a5 5 0 0 0-7 0l-3 3a5 5 0 0 0 7 7l1-1"/></Ic>,
  table:        (p) => <Ic {...p}><rect x="3" y="3" width="18" height="18" rx="2"/><path d="M3 9h18"/><path d="M3 15h18"/><path d="M9 3v18"/><path d="M15 3v18"/></Ic>,
  imageSm:      (p) => <Ic {...p}><rect x="3" y="3" width="18" height="18" rx="2"/><circle cx="9" cy="9" r="2"/><path d="m21 15-5-5L5 21"/></Ic>,
  plusSquare:   (p) => <Ic {...p}><rect x="3" y="3" width="18" height="18" rx="2"/><path d="M12 8v8"/><path d="M8 12h8"/></Ic>,

  // blocks panel
  heading:      (p) => <Ic {...p}><path d="M6 4v16"/><path d="M18 4v16"/><path d="M6 12h12"/></Ic>,
  columns:      (p) => <Ic {...p}><rect x="3" y="3" width="7" height="18" rx="1"/><rect x="14" y="3" width="7" height="18" rx="1"/></Ic>,
  mediaDrop:    (p) => <Ic {...p}><rect x="3" y="3" width="18" height="18" rx="2"/><path d="M3 14h6l2 3h2l2-3h6"/></Ic>,
  signature:    (p) => <Ic {...p}><path d="M3 17c2-4 6-8 9-4s5 4 9-1"/><path d="M3 21h18"/></Ic>,
  quote:        (p) => <Ic {...p}><path d="M3 21c0-4 2-8 5-10v4H5v6z"/><path d="M13 21c0-4 2-8 5-10v4h-3v6z"/></Ic>,
  divider:      (p) => <Ic {...p}><line x1="3" y1="12" x2="21" y2="12"/><line x1="3" y1="6" x2="5" y2="6"/><line x1="19" y1="6" x2="21" y2="6"/><line x1="3" y1="18" x2="5" y2="18"/><line x1="19" y1="18" x2="21" y2="18"/></Ic>,
  gridThree:    (p) => <Ic {...p}><rect x="3" y="3" width="5" height="18" rx="1"/><rect x="9.5" y="3" width="5" height="18" rx="1"/><rect x="16" y="3" width="5" height="18" rx="1"/></Ic>,
  stamp:        (p) => <Ic {...p}><path d="M5 20h14"/><path d="M8 17h8l-1-3 1-3a4 4 0 1 0-8 0l1 3z"/></Ic>,
  variable:     (p) => <Ic {...p}><path d="M8 3v18"/><path d="M16 3v18"/><path d="M5 7s2-2 3-2"/><path d="M19 7s-2-2-3-2"/></Ic>,
  close:        (p) => <Ic {...p}><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></Ic>,
  arrowLeft:    (p) => <Ic {...p}><line x1="19" y1="12" x2="5" y2="12"/><path d="m12 19-7-7 7-7"/></Ic>,
  plus:         (p) => <Ic {...p}><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></Ic>,
  corners:      (p) => <Ic {...p}><path d="M21 8V5a2 2 0 0 0-2-2h-3"/><path d="M3 16v3a2 2 0 0 0 2 2h3"/><path d="M3 8V5a2 2 0 0 1 2-2h3"/><path d="M21 16v3a2 2 0 0 1-2 2h-3"/></Ic>,
};

window.Icons = Icons;
