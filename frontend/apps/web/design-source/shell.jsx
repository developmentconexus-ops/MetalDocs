/* global React */
const { useState } = React;

// Inline icons — minimal, 1.5 stroke
const Icon = ({ name, size = 16 }) => {
  const paths = {
    home: <><path d="M3 10l7-6 7 6"/><path d="M5 9v8h10V9"/></>,
    docs: <><path d="M5 3h7l4 4v13H5z"/><path d="M12 3v4h4"/></>,
    library: <><path d="M3 4h4v14H3zM9 4h4v14H9z"/><path d="M15 5l3 1-3 13"/></>,
    registry: <><rect x="3" y="3" width="14" height="14" rx="1"/><path d="M3 7h14M3 11h14M3 15h14M7 3v14"/></>,
    template: <><rect x="3" y="3" width="14" height="14" rx="1"/><path d="M3 8h14M8 8v9"/></>,
    inbox: <><path d="M3 12l3-8h8l3 8v5H3z"/><path d="M3 12h4l1 2h4l1-2h4"/></>,
    workflow: <><circle cx="5" cy="5" r="2"/><circle cx="15" cy="15" r="2"/><circle cx="15" cy="5" r="2"/><path d="M7 5h6M5 7v6a2 2 0 002 2h6"/></>,
    taxonomy: <><path d="M10 3v4M10 9v8M5 13h10M3 17h4M13 17h4"/><circle cx="10" cy="3" r="1.5"/></>,
    users: <><circle cx="7" cy="7" r="3"/><path d="M2 17c0-3 2-5 5-5s5 2 5 5"/><circle cx="14" cy="8" r="2"/><path d="M11 17c0-2 1-3 3-3s4 1 4 3"/></>,
    audit: <><path d="M5 3h7l4 4v13H5z"/><path d="M8 11l2 2 4-4"/></>,
    bell: <><path d="M5 8a5 5 0 0110 0v4l1.5 2h-13L5 12V8z"/><path d="M8 16a2 2 0 004 0"/></>,
    search: <><circle cx="9" cy="9" r="5"/><path d="M13 13l4 4"/></>,
    plus: <><path d="M10 4v12M4 10h12"/></>,
    chevron: <><path d="M7 4l6 6-6 6"/></>,
    chevdown: <><path d="M5 8l5 5 5-5"/></>,
    filter: <><path d="M3 5h14M6 10h8M8 15h4"/></>,
    grid: <><rect x="3" y="3" width="6" height="6"/><rect x="11" y="3" width="6" height="6"/><rect x="3" y="11" width="6" height="6"/><rect x="11" y="11" width="6" height="6"/></>,
    list: <><path d="M3 5h14M3 10h14M3 15h14"/></>,
    more: <><circle cx="4" cy="10" r="1.5"/><circle cx="10" cy="10" r="1.5"/><circle cx="16" cy="10" r="1.5"/></>,
    lock: <><rect x="4" y="9" width="12" height="9" rx="1"/><path d="M7 9V6a3 3 0 016 0v3"/></>,
    check: <><path d="M4 10l4 4 8-8"/></>,
    x: <><path d="M5 5l10 10M15 5L5 15"/></>,
    download: <><path d="M10 3v10M5 9l5 5 5-5M3 17h14"/></>,
    upload: <><path d="M10 14V4M5 9l5-5 5 5M3 17h14"/></>,
    history: <><path d="M3 10a7 7 0 1 0 2-5"/><path d="M3 3v4h4M10 6v5l3 2"/></>,
    link: <><path d="M8 12l4-4M7 8h-2a3 3 0 000 6h2M13 12h2a3 3 0 000-6h-2"/></>,
    cog: <><circle cx="10" cy="10" r="2.5"/><path d="M10 3v2M10 15v2M3 10h2M15 10h2M5 5l1.5 1.5M14 14l-1.5-1.5M5 15l1.5-1.5M14 6l-1.5 1.5"/></>,
    arrow: <><path d="M5 10h10M11 6l4 4-4 4"/></>,
    sparkle: <><path d="M10 3l1.5 4 4 1.5-4 1.5L10 14l-1.5-4L4 8.5l4-1.5z"/></>,
  };
  return (
    <svg width={size} height={size} viewBox="0 0 20 20" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" style={{ flexShrink: 0 }}>
      {paths[name]}
    </svg>
  );
};

// Status pill
const Status = ({ s }) => {
  const map = {
    draft: { c: 'pill-draft', t: 'Rascunho' },
    review: { c: 'pill-review', t: 'Em revisão' },
    approved: { c: 'pill-approved', t: 'Aprovado' },
    frozen: { c: 'pill-frozen', t: 'Frozen' },
    rejected: { c: 'pill-rejected', t: 'Rejeitado' },
    archived: { c: 'pill-archived', t: 'Arquivado' },
  };
  const v = map[s] || map.draft;
  return <span className={`pill ${v.c}`}><span className="dot"/>{v.t}</span>;
};

// Avatar
const Avatar = ({ name, size = 'md' }) => {
  const initials = name.split(' ').map(p => p[0]).slice(0, 2).join('').toUpperCase();
  const cls = size === 'sm' ? 'avatar-sm' : size === 'lg' ? 'avatar-lg' : '';
  // generate a stable color
  const hue = name.charCodeAt(0) * 7 % 360;
  return <span className={`avatar ${cls}`} style={{ background: `hsl(${hue}, 35%, 38%)` }}>{initials}</span>;
};

const Logo = ({ wordmark = true }) => (
  <span className="logo">
    <span className="logo-mark"/>
    {wordmark && <span style={{ fontSize: 15 }}>MetalDocs</span>}
  </span>
);

// Sidebar — two-tier: rail + section panel
const Sidebar = ({ active = 'library', section = 'documents' }) => {
  const railItems = [
    { id: 'home', icon: 'home', label: 'Início' },
    { id: 'documents', icon: 'docs', label: 'Documentos' },
    { id: 'registry', icon: 'registry', label: 'Registro' },
    { id: 'templates', icon: 'template', label: 'Templates' },
    { id: 'approvals', icon: 'inbox', label: 'Aprovações', badge: 4 },
    { id: 'taxonomy', icon: 'taxonomy', label: 'Taxonomia' },
    { id: 'audit', icon: 'audit', label: 'Auditoria' },
    { id: 'admin', icon: 'users', label: 'Admin' },
  ];

  const sections = {
    documents: {
      title: 'Documentos',
      groups: [
        { label: 'Visões', items: [
          { id: 'library', label: 'Biblioteca', count: '2.847' },
          { id: 'my-docs', label: 'Meus documentos', count: '38' },
          { id: 'recent', label: 'Recentes', count: '12' },
          { id: 'shared', label: 'Compartilhados' },
        ]},
        { label: 'Estado', items: [
          { id: 'drafts', label: 'Rascunhos', count: '23', dot: 'draft' },
          { id: 'review', label: 'Em revisão', count: '8', dot: 'review' },
          { id: 'frozen', label: 'Frozen', count: '2.612', dot: 'frozen' },
          { id: 'archived', label: 'Arquivados', count: '187', dot: 'archived' },
        ]},
        { label: 'Áreas', items: [
          { id: 'rh', label: 'RH', count: '412' },
          { id: 'qua', label: 'Qualidade', count: '861' },
          { id: 'prod', label: 'Produção', count: '1.024' },
          { id: 'ti', label: 'TI & Segurança', count: '294' },
          { id: 'fin', label: 'Financeiro', count: '156' },
        ]},
      ]
    }
  };

  const sec = sections[section] || sections.documents;

  return (
    <aside style={{ display: 'flex', height: '100%', borderRight: '1px solid var(--border)' }}>
      {/* RAIL */}
      <nav style={{ width: 56, background: 'var(--rail)', display: 'flex', flexDirection: 'column', alignItems: 'center', padding: '12px 0', gap: 4 }}>
        <div style={{ width: 28, height: 28, borderRadius: 6, background: 'var(--brand)', marginBottom: 8, display: 'grid', placeItems: 'center' }}>
          <span style={{ width: 14, height: 14, background: 'white', borderRadius: 2, display: 'block', position: 'relative' }}>
            <span style={{ position: 'absolute', left: 2, right: 2, top: 3, height: 1, background: 'var(--brand)' }}/>
            <span style={{ position: 'absolute', left: 2, right: 2, top: 6, height: 1, background: 'var(--brand)' }}/>
            <span style={{ position: 'absolute', left: 2, right: 2, top: 9, height: 1, background: 'var(--brand)' }}/>
          </span>
        </div>
        {railItems.map(it => (
          <button key={it.id} title={it.label} style={{
            position: 'relative',
            width: 40, height: 36,
            border: 'none', background: 'transparent', cursor: 'pointer',
            display: 'grid', placeItems: 'center',
            color: it.id === section ? 'white' : 'var(--rail-text-muted)',
            borderRadius: 6,
            ...(it.id === section ? { background: 'rgba(255,255,255,0.07)' } : {}),
          }}>
            <Icon name={it.icon} size={18}/>
            {it.id === section && <span style={{ position: 'absolute', left: -12, top: 6, bottom: 6, width: 2, background: 'var(--rail-active)', borderRadius: '0 2px 2px 0' }}/>}
            {it.badge && <span style={{ position: 'absolute', top: 4, right: 4, minWidth: 14, height: 14, borderRadius: 7, background: 'var(--accent)', color: 'white', fontSize: 9, fontWeight: 600, display: 'grid', placeItems: 'center', padding: '0 3px' }}>{it.badge}</span>}
          </button>
        ))}
        <div className="spacer"/>
        <button style={{ width: 40, height: 36, border: 'none', background: 'transparent', cursor: 'pointer', color: 'var(--rail-text-muted)', display: 'grid', placeItems: 'center' }}>
          <Icon name="cog" size={18}/>
        </button>
        <div style={{ marginTop: 6 }}><Avatar name="Marina S" size="sm"/></div>
      </nav>

      {/* SECTION PANEL */}
      <div style={{ width: 224, background: 'var(--surface-2)', display: 'flex', flexDirection: 'column', padding: '12px 0' }}>
        <div style={{ padding: '0 14px 10px', display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
          <div className="h3" style={{ fontSize: 14 }}>{sec.title}</div>
          <button className="btn btn-icon btn-sm btn-ghost" title="Novo"><Icon name="plus" size={14}/></button>
        </div>
        <div style={{ padding: '0 10px' }}>
          <div style={{ position: 'relative' }}>
            <span style={{ position: 'absolute', left: 8, top: '50%', transform: 'translateY(-50%)', color: 'var(--text-muted)' }}><Icon name="search" size={13}/></span>
            <input className="input" placeholder="Buscar..." style={{ height: 28, paddingLeft: 26, fontSize: 12, background: 'var(--surface)' }}/>
            <span className="kbd" style={{ position: 'absolute', right: 6, top: '50%', transform: 'translateY(-50%)' }}>⌘K</span>
          </div>
        </div>
        <div style={{ padding: '14px 0', overflow: 'auto', flex: 1 }}>
          {sec.groups.map((g, gi) => (
            <div key={gi} style={{ marginBottom: 14 }}>
              <div className="kicker" style={{ padding: '0 16px 6px', fontSize: 10 }}>{g.label}</div>
              {g.items.map(it => {
                const isActive = it.id === active;
                return (
                  <button key={it.id} className="row" style={{
                    display: 'flex', alignItems: 'center', gap: 8,
                    width: 'calc(100% - 8px)', margin: '0 4px',
                    padding: '5px 12px',
                    border: 'none', background: isActive ? 'var(--surface)' : 'transparent',
                    borderRadius: 5,
                    cursor: 'pointer', textAlign: 'left',
                    color: isActive ? 'var(--text)' : 'var(--text-soft)',
                    fontSize: 13, fontWeight: isActive ? 500 : 400,
                    boxShadow: isActive ? 'var(--shadow-1)' : 'none',
                  }}>
                    {it.dot && <span style={{ width: 6, height: 6, borderRadius: '50%', background: it.dot === 'draft' ? 'var(--text-muted)' : it.dot === 'review' ? 'var(--warning)' : it.dot === 'frozen' ? 'var(--success)' : 'var(--text-faint)' }}/>}
                    <span>{it.label}</span>
                    <span className="spacer"/>
                    {it.count && <span className="num" style={{ fontSize: 11, color: 'var(--text-muted)', fontFamily: 'var(--font-mono)' }}>{it.count}</span>}
                  </button>
                );
              })}
            </div>
          ))}
        </div>
        <div style={{ padding: '10px 14px', borderTop: '1px solid var(--border)', display: 'flex', alignItems: 'center', gap: 8, fontSize: 11, color: 'var(--text-muted)' }}>
          <span style={{ width: 6, height: 6, borderRadius: '50%', background: 'var(--success)' }}/>
          <span>Sistema operacional</span>
          <span className="spacer"/>
          <span className="mono">v3.4.2</span>
        </div>
      </div>
    </aside>
  );
};

// Top toolbar — search + actions
const Toolbar = ({ title, breadcrumb, actions = true }) => (
  <header style={{ height: 52, borderBottom: '1px solid var(--border)', background: 'var(--surface)', display: 'flex', alignItems: 'center', padding: '0 20px', gap: 16, flexShrink: 0 }}>
    <div style={{ display: 'flex', alignItems: 'center', gap: 8, fontSize: 13, color: 'var(--text-muted)' }}>
      {breadcrumb && breadcrumb.map((b, i) => (
        <React.Fragment key={i}>
          {i > 0 && <span style={{ color: 'var(--text-faint)' }}>/</span>}
          <span style={{ color: i === breadcrumb.length - 1 ? 'var(--text)' : 'var(--text-muted)', fontWeight: i === breadcrumb.length - 1 ? 500 : 400 }}>{b}</span>
        </React.Fragment>
      ))}
      {!breadcrumb && <span style={{ color: 'var(--text)', fontWeight: 500, fontSize: 14 }}>{title}</span>}
    </div>
    <div className="spacer"/>
    <div style={{ position: 'relative', width: 320 }}>
      <span style={{ position: 'absolute', left: 9, top: '50%', transform: 'translateY(-50%)', color: 'var(--text-muted)' }}><Icon name="search" size={14}/></span>
      <input className="input" placeholder="Buscar documentos, códigos, autores…" style={{ paddingLeft: 30, paddingRight: 50 }}/>
      <span className="kbd" style={{ position: 'absolute', right: 8, top: '50%', transform: 'translateY(-50%)' }}>⌘K</span>
    </div>
    {actions && (
      <>
        <button className="btn btn-icon btn-ghost" title="Notificações" style={{ position: 'relative' }}>
          <Icon name="bell" size={16}/>
          <span style={{ position: 'absolute', top: 6, right: 6, width: 6, height: 6, borderRadius: '50%', background: 'var(--accent)' }}/>
        </button>
        <div className="divider-v" style={{ height: 24 }}/>
        <button className="btn btn-primary"><Icon name="plus" size={14}/>Novo documento</button>
      </>
    )}
  </header>
);

window.MD = { Icon, Status, Avatar, Logo, Sidebar, Toolbar };

// Rail-only sidebar (just the 56px icon column, no filter panel)
const Rail = ({ section = 'documents' }) => {
  const items = [
    { id: 'home', icon: 'home', label: 'Início' },
    { id: 'documents', icon: 'docs', label: 'Documentos' },
    { id: 'registry', icon: 'registry', label: 'Registro' },
    { id: 'templates', icon: 'template', label: 'Templates' },
    { id: 'approvals', icon: 'inbox', label: 'Aprovações', badge: 4 },
    { id: 'taxonomy', icon: 'taxonomy', label: 'Taxonomia' },
    { id: 'audit', icon: 'audit', label: 'Auditoria' },
    { id: 'admin', icon: 'users', label: 'Admin' },
  ];
  return (
    <nav style={{ width: 56, background: 'var(--rail)', display: 'flex', flexDirection: 'column', alignItems: 'center', padding: '12px 0', gap: 4, borderRight: '1px solid var(--border)', flexShrink: 0 }}>
      <div style={{ width: 28, height: 28, borderRadius: 6, background: 'var(--brand)', marginBottom: 8, display: 'grid', placeItems: 'center' }}>
        <span style={{ width: 14, height: 14, background: 'white', borderRadius: 2, display: 'block', position: 'relative' }}>
          <span style={{ position: 'absolute', left: 2, right: 2, top: 3, height: 1, background: 'var(--brand)' }}/>
          <span style={{ position: 'absolute', left: 2, right: 2, top: 6, height: 1, background: 'var(--brand)' }}/>
          <span style={{ position: 'absolute', left: 2, right: 2, top: 9, height: 1, background: 'var(--brand)' }}/>
        </span>
      </div>
      {items.map(it => (
        <button key={it.id} title={it.label} style={{
          position: 'relative', width: 40, height: 36, border: 'none', background: it.id === section ? 'rgba(255,255,255,0.07)' : 'transparent',
          cursor: 'pointer', display: 'grid', placeItems: 'center', color: it.id === section ? 'white' : 'var(--rail-text-muted)', borderRadius: 6,
        }}>
          <Icon name={it.icon} size={18}/>
          {it.id === section && <span style={{ position: 'absolute', left: -12, top: 6, bottom: 6, width: 2, background: 'var(--rail-active)', borderRadius: '0 2px 2px 0' }}/>}
          {it.badge && <span style={{ position: 'absolute', top: 4, right: 4, minWidth: 14, height: 14, borderRadius: 7, background: 'var(--accent)', color: 'white', fontSize: 9, fontWeight: 600, display: 'grid', placeItems: 'center', padding: '0 3px' }}>{it.badge}</span>}
        </button>
      ))}
      <span style={{ flex: 1 }}/>
      <button style={{ width: 40, height: 36, border: 'none', background: 'transparent', cursor: 'pointer', color: 'var(--rail-text-muted)', display: 'grid', placeItems: 'center' }}>
        <Icon name="cog" size={18}/>
      </button>
      <div style={{ marginTop: 6 }}><Avatar name="Marina S" size="sm"/></div>
    </nav>
  );
};

window.MD.Rail = Rail;
