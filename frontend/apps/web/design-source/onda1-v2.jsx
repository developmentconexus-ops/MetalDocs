/* global React, MD */
const { Icon, Logo, Avatar, Sidebar, Toolbar } = MD;
const { useState, useEffect, useMemo } = React;

// ===== Shared mock data (refined) =====
const DOC2 = {
  code: 'PR-EHS-014',
  title: 'Procedimento de Bloqueio e Etiquetagem (LOTO)',
  version: 'v3.2',
  publishedAt: '12 mar 2026 · 14:32',
  publishedBy: 'Carolina Mendes',
  publishedByRole: 'Coord. SSMA',
  area: 'Saúde, Segurança & Meio Ambiente',
  areaShort: 'SSMA',
  hash: 'a7f3c9e1d8b4',
  hashFull: 'a7f3c9e1d8b4f6e2c1a9b3d7e8f4c2a1b9d6e3f7a8c4b2d1e9f6a3c7b8d4e2f1',
  pageCount: 14,
  effectiveAt: '15 mar 2026',
  reviewIn: '15 mar 2027',
  size: '2.4 MB',
  type: 'Procedimento',
  signoffs: [
    { name: 'Renata Souza', role: 'Coord. SSMA', when: '11 mar · 18:04' },
    { name: 'Marcos Lima', role: 'Gerente Industrial', when: '12 mar · 09:12' },
    { name: 'Diretoria', role: 'Anuente', when: '12 mar · 14:30' },
  ],
  versions: [
    { v: 'v3.2', when: '12 mar 2026', author: 'C. Mendes', current: true, summary: 'Inclusão de procedimento para painéis CCM-04. Atualização da matriz de risco para fontes hidráulicas.' },
    { v: 'v3.1', when: '08 jan 2026', author: 'C. Mendes', summary: 'Correção de typo no item 4.3 (responsabilidades).' },
    { v: 'v3.0', when: '14 nov 2025', author: 'R. Souza', summary: 'Revisão geral pós-auditoria interna. Novas etiquetas padronizadas para todas as plantas.' },
    { v: 'v2.4', when: '02 jul 2025', author: 'R. Souza', summary: 'Última versão da revisão anterior.' },
    { v: 'v2.3', when: '18 mar 2025', author: 'A. Tavares', summary: 'Ajustes de redação após feedback de Engenharia.' },
  ],
};

// ============================================================
// Hooks
// ============================================================
function useCounter2(target, duration = 1200) {
  const [n, setN] = useState(0);
  useEffect(() => {
    let raf, start;
    const tick = (t) => {
      if (!start) start = t;
      const p = Math.min(1, (t - start) / duration);
      setN(Math.round(target * (1 - Math.pow(1 - p, 3))));
      if (p < 1) raf = requestAnimationFrame(tick);
    };
    raf = requestAnimationFrame(tick);
    return () => cancelAnimationFrame(raf);
  }, [target, duration]);
  return n;
}

// ============================================================
// ISO seal (compact, with hash tooltip)
// ============================================================
function ISOSeal2({ hash, hashFull, dark }) {
  const [open, setOpen] = useState(false);
  const tone = dark
    ? { bg: 'rgba(56,178,134,0.16)', color: '#7dd3a8', border: 'rgba(125,211,168,0.35)' }
    : { bg: 'var(--success-bg)', color: 'var(--success)', border: 'var(--success)' };
  return (
    <span
      style={{ position: 'relative', display: 'inline-flex' }}
      onMouseEnter={() => setOpen(true)}
      onMouseLeave={() => setOpen(false)}
    >
      <span style={{
        display: 'inline-flex', alignItems: 'center', gap: 6,
        height: 22, padding: '0 9px',
        fontFamily: 'var(--font-mono)', fontSize: 10.5, letterSpacing: '0.04em',
        background: tone.bg, color: tone.color,
        border: `1px solid ${tone.border}`, borderRadius: 'var(--r-pill)',
        cursor: 'help',
      }}>
        <span style={{
          width: 6, height: 6, borderRadius: '50%', background: 'currentColor',
          animation: 'sealPulse2 2.4s ease-in-out infinite',
        }}/>
        ISO 9001 · {hash}
      </span>
      {open && (
        <div style={{
          position: 'absolute', top: 'calc(100% + 6px)', right: 0, zIndex: 50,
          padding: '10px 12px', background: '#1a0e0e', color: '#fff',
          borderRadius: 6, fontSize: 11, lineHeight: 1.5, width: 320,
          boxShadow: '0 12px 32px rgba(0,0,0,0.3)',
        }}>
          <div style={{ fontFamily: 'var(--font-mono)', fontSize: 9.5, opacity: 0.7, marginBottom: 4 }}>SHA-256</div>
          <div style={{ fontFamily: 'var(--font-mono)', fontSize: 11, wordBreak: 'break-all', opacity: 0.95 }}>{hashFull}</div>
          <div style={{ marginTop: 8, paddingTop: 8, borderTop: '1px solid rgba(255,255,255,0.12)', fontSize: 10.5, opacity: 0.75 }}>
            Documento congelado · cadeia de hash íntegra · Cl. 7.5.3
          </div>
        </div>
      )}
      <style>{`@keyframes sealPulse2 { 0%,100% { opacity: 1; transform: scale(1);} 50% { opacity: .55; transform: scale(.85);} }`}</style>
    </span>
  );
}

// ============================================================
// PERMISSION CONTEXT — controls action availability
// ============================================================
const PERMISSION_PRESETS = {
  full:    { view: true, download: true, revise: true, share: true, label: 'Editor · acesso total' },
  reader:  { view: true, download: false, revise: false, share: true, label: 'Leitor · sem download' },
  visitor: { view: true, download: false, revise: false, share: false, label: 'Visitante · só visualiza' },
};

// Permission switcher (mock — for design review only)
function PermSwitcher({ perm, setPerm, dark }) {
  return (
    <div style={{
      display: 'inline-flex', alignItems: 'center', gap: 6,
      padding: '4px 6px 4px 10px',
      background: dark ? 'rgba(255,255,255,0.06)' : 'var(--surface-2)',
      border: `1px solid ${dark ? 'rgba(255,255,255,0.12)' : 'var(--border)'}`,
      borderRadius: 'var(--r-pill)', fontSize: 10.5,
      color: dark ? 'rgba(255,255,255,0.7)' : 'var(--text-muted)',
      fontFamily: 'var(--font-mono)', letterSpacing: '0.04em',
    }}>
      <span>VOCÊ:</span>
      {Object.keys(PERMISSION_PRESETS).map(k => (
        <button key={k} onClick={() => setPerm(k)} style={{
          height: 20, padding: '0 8px', border: 'none', cursor: 'pointer',
          background: perm === k ? (dark ? 'rgba(255,255,255,0.16)' : 'var(--surface)') : 'transparent',
          color: perm === k ? (dark ? '#fff' : 'var(--text)') : 'inherit',
          borderRadius: 'var(--r-pill)', fontSize: 10.5,
          fontFamily: 'inherit', letterSpacing: 'inherit',
          textTransform: 'uppercase', fontWeight: perm === k ? 600 : 400,
        }}>{k}</button>
      ))}
    </div>
  );
}

// ============================================================
// PUBLICADO v2 — Hero landing (no doc body, permission-aware)
// ============================================================
function PublicadoV2() {
  const [perm, setPerm] = useState('full');
  const p = PERMISSION_PRESETS[perm];
  const [showHash, setShowHash] = useState(false);

  return (
    <div className="palette-wine" style={{ height: 980, display: 'flex', background: 'var(--bg)' }}>
      <Sidebar variant="rail" active="documents"/>
      <div style={{ flex: 1, overflow: 'auto', display: 'flex', flexDirection: 'column' }}>
        {/* HERO */}
        <div style={{
          padding: '20px 48px 36px',
          background: 'linear-gradient(135deg, #2a1418 0%, #3e1018 60%, #4a1822 100%)',
          color: '#f9f0f0', position: 'relative', overflow: 'hidden',
        }}>
          {/* Texture */}
          <div style={{
            position: 'absolute', inset: 0,
            backgroundImage: 'radial-gradient(circle at 15% 0%, rgba(200,54,74,0.20), transparent 50%), radial-gradient(circle at 85% 100%, rgba(139,46,58,0.25), transparent 50%)',
            pointerEvents: 'none',
          }}/>
          {/* Grid line texture */}
          <div style={{
            position: 'absolute', inset: 0, opacity: 0.04,
            backgroundImage: 'linear-gradient(rgba(255,255,255,0.5) 1px, transparent 1px), linear-gradient(90deg, rgba(255,255,255,0.5) 1px, transparent 1px)',
            backgroundSize: '32px 32px',
            pointerEvents: 'none',
          }}/>

          {/* Top bar */}
          <div style={{ position: 'relative', display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 28 }}>
            <button className="btn btn-ghost btn-sm" style={{ color: '#e8d6d6', borderColor: 'rgba(255,255,255,0.16)' }}>
              <Icon name="chevron-left" size={13}/>Voltar para Biblioteca
            </button>
            <div style={{ display: 'flex', gap: 10, alignItems: 'center' }}>
              <PermSwitcher perm={perm} setPerm={setPerm} dark/>
              <ISOSeal2 hash={DOC2.hash} hashFull={DOC2.hashFull} dark/>
            </div>
          </div>

          {/* Hero content */}
          <div style={{ position: 'relative', display: 'grid', gridTemplateColumns: '1fr 360px', gap: 48, alignItems: 'flex-end' }}>
            <div>
              <div style={{ display: 'flex', alignItems: 'center', gap: 10, marginBottom: 14 }}>
                <span style={{
                  display: 'inline-flex', alignItems: 'center', gap: 6,
                  fontFamily: 'var(--font-mono)', fontSize: 11, letterSpacing: '0.08em',
                  color: '#c8364a', textTransform: 'uppercase', fontWeight: 500,
                }}>
                  <span style={{ width: 6, height: 6, background: '#c8364a', borderRadius: 1 }}/>
                  {DOC2.areaShort} · {DOC2.code}
                </span>
                <span style={{ width: 1, height: 12, background: 'rgba(255,255,255,0.15)' }}/>
                <span style={{ fontSize: 11, color: 'rgba(255,255,255,0.55)', letterSpacing: '0.04em', textTransform: 'uppercase' }}>
                  {DOC2.type}
                </span>
              </div>
              <h1 style={{
                fontSize: 46, lineHeight: 1.05, fontWeight: 600,
                letterSpacing: '-0.03em', margin: '0 0 20px',
                color: '#fff', maxWidth: 760,
              }}>{DOC2.title}</h1>
              <div style={{ display: 'flex', alignItems: 'center', gap: 14, fontSize: 13, color: 'rgba(255,255,255,0.7)' }}>
                <Avatar name={DOC2.publishedBy} size={28}/>
                <div>
                  <div style={{ color: '#fff', fontSize: 13, fontWeight: 500 }}>{DOC2.publishedBy}</div>
                  <div style={{ fontSize: 11, color: 'rgba(255,255,255,0.55)' }}>{DOC2.publishedByRole}</div>
                </div>
                <span style={{ width: 1, height: 24, background: 'rgba(255,255,255,0.12)', marginLeft: 4 }}/>
                <div>
                  <div className="kicker" style={{ color: 'rgba(255,255,255,0.5)', fontSize: 9.5 }}>Versão atual</div>
                  <div className="mono" style={{ fontSize: 14, color: '#fff', fontWeight: 600 }}>{DOC2.version}</div>
                </div>
                <div>
                  <div className="kicker" style={{ color: 'rgba(255,255,255,0.5)', fontSize: 9.5 }}>Publicada em</div>
                  <div style={{ fontSize: 13, color: '#fff' }}>{DOC2.publishedAt}</div>
                </div>
                <div>
                  <div className="kicker" style={{ color: 'rgba(255,255,255,0.5)', fontSize: 9.5 }}>Vigente até</div>
                  <div style={{ fontSize: 13, color: '#fff' }}>{DOC2.reviewIn}</div>
                </div>
              </div>
            </div>

            {/* Action stack */}
            <div style={{ display: 'flex', flexDirection: 'column', gap: 8, justifySelf: 'end', minWidth: 280 }}>
              {/* Primary: visualizar — always available */}
              <button className="btn btn-lg" style={{
                background: '#fff', color: 'var(--brand-deep)', border: 'none',
                justifyContent: 'center', height: 44, fontSize: 14, fontWeight: 600,
              }}>
                <Icon name="eye" size={16}/>Visualizar documento
              </button>

              {/* Secondary actions in a 2x2 grid */}
              <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 6 }}>
                <PermAction enabled={p.download} label="Baixar PDF" icon="download" hint="sem permissão"/>
                <PermAction enabled={p.revise} label="Revisar (v3.3)" icon="edit" hint="só editores"/>
                <PermAction enabled={p.share} label="Copiar link" icon="link" hint="—"/>
                <PermAction enabled label="Mais" icon="more"/>
              </div>

              {/* Permission summary */}
              <div style={{
                marginTop: 6, padding: '8px 10px',
                background: 'rgba(0,0,0,0.18)', borderRadius: 6,
                fontSize: 10.5, color: 'rgba(255,255,255,0.6)',
                display: 'flex', alignItems: 'center', gap: 6,
              }}>
                <Icon name="shield" size={12}/>
                <span>{p.label}</span>
              </div>
            </div>
          </div>
        </div>

        {/* MAIN — version timeline + audit panel + signoffs */}
        <div style={{ padding: '36px 48px 48px', display: 'grid', gridTemplateColumns: '1fr 340px', gap: 32 }}>
          {/* Left: history + identification */}
          <div style={{ display: 'flex', flexDirection: 'column', gap: 28 }}>
            {/* Version timeline (Git-like) */}
            <section>
              <div style={{ display: 'flex', alignItems: 'baseline', justifyContent: 'space-between', marginBottom: 14 }}>
                <div>
                  <div className="kicker">Histórico de versões</div>
                  <h2 className="h2" style={{ marginTop: 4 }}>Linhagem do documento</h2>
                </div>
                <button className="btn btn-sm btn-ghost"><Icon name="filter" size={12}/>Comparar versões</button>
              </div>

              {/* Horizontal git-style timeline */}
              <div style={{ position: 'relative', padding: '20px 0 4px' }}>
                <div style={{ position: 'absolute', top: 28, left: 16, right: 16, height: 2, background: 'var(--border)' }}/>
                <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                  {DOC2.versions.slice().reverse().map((v, i) => (
                    <div key={i} style={{ position: 'relative', display: 'flex', flexDirection: 'column', alignItems: 'center', minWidth: 100, gap: 6 }}>
                      <div style={{
                        width: 16, height: 16, borderRadius: '50%',
                        background: v.current ? 'var(--brand)' : 'var(--surface)',
                        border: `2px solid ${v.current ? 'var(--brand-deep)' : 'var(--border-strong)'}`,
                        boxShadow: v.current ? '0 0 0 5px var(--brand-pale)' : 'none',
                        zIndex: 1,
                      }}/>
                      <span className="mono" style={{ fontSize: 12.5, fontWeight: v.current ? 600 : 400, color: v.current ? 'var(--brand)' : 'var(--text)' }}>{v.v}</span>
                      <span className="tiny" style={{ fontSize: 10 }}>{v.when}</span>
                    </div>
                  ))}
                </div>
              </div>

              {/* Detailed list below */}
              <div style={{ marginTop: 16, display: 'flex', flexDirection: 'column' }}>
                {DOC2.versions.map((v, i) => (
                  <div key={i} style={{
                    display: 'grid', gridTemplateColumns: '60px 110px 1fr 140px 60px', gap: 14,
                    padding: '12px 0', borderTop: '1px solid var(--border)',
                    alignItems: 'center',
                    background: v.current ? 'var(--brand-pale)' : 'transparent',
                    margin: v.current ? '0 -12px' : 0,
                    paddingLeft: v.current ? 12 : 0, paddingRight: v.current ? 12 : 0,
                    borderRadius: v.current ? 6 : 0,
                  }}>
                    <span className="mono" style={{ fontSize: 12.5, fontWeight: v.current ? 600 : 500, color: v.current ? 'var(--brand)' : 'var(--text)' }}>{v.v}</span>
                    <span className="caption" style={{ fontSize: 12 }}>{v.when}</span>
                    <span style={{ fontSize: 12.5, color: 'var(--text-soft)', lineHeight: 1.4 }}>{v.summary}</span>
                    <div style={{ display: 'flex', alignItems: 'center', gap: 6, fontSize: 12 }}>
                      <Avatar name={v.author} size={20}/><span>{v.author}</span>
                    </div>
                    <span style={{ display: 'inline-flex', justifyContent: 'flex-end', gap: 4 }}>
                      {v.current ? <span className="pill pill-frozen"><span className="dot"/>atual</span> : <button className="btn btn-sm btn-ghost"><Icon name="eye" size={11}/></button>}
                    </span>
                  </div>
                ))}
              </div>
            </section>

            {/* Identification card */}
            <section>
              <div className="kicker" style={{ marginBottom: 12 }}>Identificação & Evidência</div>
              <div className="card" style={{ padding: 0, overflow: 'hidden' }}>
                <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: 0 }}>
                  {[
                    ['Código', DOC2.code, true],
                    ['Tipo', DOC2.type, false],
                    ['Área', DOC2.area, false],
                    ['Páginas', `${DOC2.pageCount} pp · ${DOC2.size}`, false],
                    ['Vigente desde', DOC2.effectiveAt, false],
                    ['Próxima revisão', DOC2.reviewIn, false],
                  ].map((row, i) => (
                    <div key={i} style={{
                      padding: '14px 18px',
                      borderRight: i % 3 !== 2 ? '1px solid var(--border)' : 'none',
                      borderBottom: i < 3 ? '1px solid var(--border)' : 'none',
                    }}>
                      <div className="kicker" style={{ marginBottom: 4 }}>{row[0]}</div>
                      <div style={{ fontSize: 13, fontWeight: row[2] ? 500 : 400, fontFamily: row[2] ? 'var(--font-mono)' : 'inherit' }}>{row[1]}</div>
                    </div>
                  ))}
                </div>
                {/* Hash strip */}
                <div style={{
                  borderTop: '1px solid var(--border)',
                  padding: '12px 18px', background: 'var(--surface-2)',
                  display: 'flex', alignItems: 'center', gap: 10,
                  fontFamily: 'var(--font-mono)', fontSize: 11,
                }}>
                  <Icon name="check" size={13}/>
                  <span className="kicker" style={{ fontSize: 9.5 }}>SHA-256</span>
                  <span style={{ flex: 1, color: 'var(--text-soft)', wordBreak: 'break-all' }}>
                    {showHash ? DOC2.hashFull : DOC2.hashFull.slice(0, 24) + '…'}
                  </span>
                  <button className="btn btn-sm btn-ghost" onClick={() => setShowHash(s => !s)}>
                    <Icon name={showHash ? 'eye-off' : 'eye'} size={11}/>
                  </button>
                  <button className="btn btn-sm btn-ghost"><Icon name="copy" size={11}/></button>
                </div>
              </div>
            </section>
          </div>

          {/* Right column: signoffs + activity */}
          <aside style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
            <div className="card" style={{ padding: 18 }}>
              <div className="kicker" style={{ marginBottom: 12 }}>Aprovações</div>
              {DOC2.signoffs.map((s, i) => (
                <div key={i} style={{ display: 'flex', alignItems: 'center', gap: 10, padding: '10px 0', borderBottom: i < DOC2.signoffs.length - 1 ? '1px solid var(--border)' : 'none' }}>
                  <Avatar name={s.name} size={28}/>
                  <div style={{ flex: 1, minWidth: 0 }}>
                    <div style={{ fontSize: 12.5, fontWeight: 500 }}>{s.name}</div>
                    <div className="tiny" style={{ fontSize: 10.5 }}>{s.role}</div>
                  </div>
                  <div style={{ textAlign: 'right' }}>
                    <Icon name="check" size={14}/>
                    <div className="tiny mono" style={{ fontSize: 10 }}>{s.when}</div>
                  </div>
                </div>
              ))}
            </div>

            <div className="card" style={{ padding: 18 }}>
              <div style={{ display: 'flex', alignItems: 'baseline', justifyContent: 'space-between', marginBottom: 12 }}>
                <div className="kicker">Distribuição</div>
                <button className="btn btn-sm btn-ghost" style={{ padding: '0 4px' }}>Ver tudo →</button>
              </div>
              <div style={{ display: 'flex', alignItems: 'baseline', gap: 8, marginBottom: 8 }}>
                <span className="num" style={{ fontSize: 32, fontFamily: 'var(--font-display)', fontWeight: 600, letterSpacing: '-0.02em', color: 'var(--brand)' }}>63%</span>
                <span className="caption">cobertura</span>
              </div>
              <div style={{ height: 6, background: 'var(--surface-3)', borderRadius: 999, overflow: 'hidden', marginBottom: 8 }}>
                <div style={{ width: '63%', height: '100%', background: 'var(--brand)', borderRadius: 999 }}/>
              </div>
              <div className="caption" style={{ fontSize: 11.5 }}>156 de 248 destinatários reconheceram · vence em 8 dias</div>
            </div>

            <div className="card" style={{ padding: 18 }}>
              <div className="kicker" style={{ marginBottom: 10 }}>Atividade recente</div>
              {[
                { who: 'Daniela Prado', what: 'reconheceu leitura', when: 'há 4 min' },
                { who: 'Marina A.', what: 'leu o documento', when: 'há 1h' },
                { who: 'C. Mendes', what: 'publicou esta versão', when: '12 mar' },
              ].map((a, i) => (
                <div key={i} style={{ display: 'flex', gap: 8, padding: '8px 0', borderBottom: i < 2 ? '1px solid var(--border)' : 'none' }}>
                  <Avatar name={a.who} size={22}/>
                  <div style={{ flex: 1, minWidth: 0, fontSize: 11.5, lineHeight: 1.4 }}>
                    <div><strong>{a.who}</strong> {a.what}</div>
                    <div className="tiny" style={{ fontSize: 10 }}>{a.when}</div>
                  </div>
                </div>
              ))}
            </div>
          </aside>
        </div>
      </div>
    </div>
  );
}

function PermAction({ enabled, label, icon, hint }) {
  const [hover, setHover] = useState(false);
  return (
    <div style={{ position: 'relative' }} onMouseEnter={() => setHover(true)} onMouseLeave={() => setHover(false)}>
      <button
        className="btn"
        disabled={!enabled}
        style={{
          width: '100%', justifyContent: 'center', height: 38,
          background: enabled ? 'rgba(255,255,255,0.08)' : 'rgba(255,255,255,0.03)',
          border: `1px solid ${enabled ? 'rgba(255,255,255,0.18)' : 'rgba(255,255,255,0.08)'}`,
          color: enabled ? '#fff' : 'rgba(255,255,255,0.35)',
          cursor: enabled ? 'pointer' : 'not-allowed',
          fontSize: 12.5,
        }}
      >
        {!enabled && <Icon name="lock" size={11}/>}
        {enabled && <Icon name={icon} size={13}/>}
        {label}
      </button>
      {!enabled && hover && hint && (
        <div style={{
          position: 'absolute', top: '100%', left: '50%', transform: 'translateX(-50%) translateY(4px)',
          background: '#1a0e0e', color: '#fff', padding: '5px 9px', borderRadius: 4,
          fontSize: 10.5, whiteSpace: 'nowrap', zIndex: 30,
          boxShadow: '0 4px 12px rgba(0,0,0,0.3)',
        }}>{hint}</div>
      )}
    </div>
  );
}

// ============================================================
// FANOUT v2 — Combined dashboard + table
// ============================================================
const FANOUT2 = {
  totalTargets: 248,
  read: 182,
  acknowledged: 156,
  pending: 66,
  overdue: 12,
  deadline: '22 mar 2026',
  daysLeft: 8,
  // Daily reads over the last 8 days (since publish)
  dailyReads: [
    { day: '12/03', cumAck: 22 },
    { day: '13/03', cumAck: 58 },
    { day: '14/03', cumAck: 84 },
    { day: '15/03', cumAck: 102 },
    { day: '16/03', cumAck: 118 },
    { day: '17/03', cumAck: 132 },
    { day: '18/03', cumAck: 144 },
    { day: '19/03', cumAck: 156 },
  ],
  byArea: [
    { area: 'Manutenção',          total: 64, read: 58, ack: 52 },
    { area: 'Produção · Linha A',  total: 42, read: 38, ack: 34 },
    { area: 'SSMA',                total: 18, read: 18, ack: 18 },
    { area: 'Engenharia',          total: 24, read: 19, ack: 16 },
    { area: 'Produção · Linha B',  total: 38, read: 22, ack: 18 },
    { area: 'Logística',           total: 28, read: 14, ack: 10 },
    { area: 'Qualidade',           total: 16, read: 8,  ack: 5  },
    { area: 'Administrativo',      total: 18, read: 5,  ack: 3  },
  ],
  people: [
    { who: 'Roberto Carvalho',  area: 'Logística',       last: 'nunca abriu',           status: 'pending', overdue: false, role: 'Op. Empilhadeira' },
    { who: 'Ana Paula Mendes',  area: 'Administrativo',  last: 'nunca abriu',           status: 'pending', overdue: true,  role: 'Analista RH' },
    { who: 'Thiago Rocha',      area: 'Qualidade',       last: '08 mar (apenas leu)',   status: 'read',    overdue: false, role: 'Insp. Qualidade' },
    { who: 'Felipe Andrade',    area: 'Produção · B',    last: 'há 6 dias',             status: 'pending', overdue: false, role: 'Op. Multifuncional' },
    { who: 'Beatriz Moreira',   area: 'Administrativo',  last: 'nunca abriu',           status: 'pending', overdue: true,  role: 'Aux. Administrativo' },
    { who: 'Gabriel Costa',     area: 'Logística',       last: 'nunca abriu',           status: 'pending', overdue: true,  role: 'Conferente' },
    { who: 'Daniela Prado',     area: 'Manutenção',      last: 'há 4 min',              status: 'ack',     overdue: false, role: 'Téc. Mecânica' },
    { who: 'Eduardo Castro',    area: 'Engenharia',      last: 'há 28 min',             status: 'ack',     overdue: false, role: 'Eng. Processo' },
    { who: 'Marina Albuquerque',area: 'Manutenção',      last: 'há 1h',                 status: 'read',    overdue: false, role: 'Téc. Elétrica' },
  ],
};

function FanoutV2() {
  const [ready, setReady] = useState(false);
  const [tab, setTab] = useState('pending');
  useEffect(() => { const t = setTimeout(() => setReady(true), 200); return () => clearTimeout(t); }, []);

  const pct = Math.round((FANOUT2.acknowledged / FANOUT2.totalTargets) * 100);
  const pctReached = Math.round((FANOUT2.read / FANOUT2.totalTargets) * 100);
  const animatedPct = useCounter2(ready ? pct : 0, 1600);

  // Sort areas by coverage asc — surfaces problems
  const areas = useMemo(() => FANOUT2.byArea.map(a => ({
    ...a,
    pct: a.ack / a.total,
    pctRead: a.read / a.total,
  })).sort((x, y) => x.pct - y.pct), []);

  // Filter table by tab
  const filtered = useMemo(() => {
    if (tab === 'pending') return FANOUT2.people.filter(p => p.status === 'pending');
    if (tab === 'overdue') return FANOUT2.people.filter(p => p.overdue);
    if (tab === 'read') return FANOUT2.people.filter(p => p.status === 'read');
    if (tab === 'acked') return FANOUT2.people.filter(p => p.status === 'ack');
    return FANOUT2.people;
  }, [tab]);

  const tabCounts = {
    pending: FANOUT2.people.filter(p => p.status === 'pending').length,
    overdue: FANOUT2.people.filter(p => p.overdue).length,
    read: FANOUT2.people.filter(p => p.status === 'read').length,
    acked: FANOUT2.people.filter(p => p.status === 'ack').length,
    all: FANOUT2.people.length,
  };

  return (
    <div className="palette-wine" style={{ height: 980, display: 'flex', background: 'var(--bg)' }}>
      <Sidebar variant="rail" active="documents"/>
      <div style={{ flex: 1, overflow: 'auto' }}>
        {/* Toolbar */}
        <div style={{ padding: '18px 32px 0' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 10, marginBottom: 6, fontSize: 12, color: 'var(--text-muted)' }}>
            <button className="btn btn-ghost btn-sm" style={{ padding: 0 }}><Icon name="chevron-left" size={13}/>SSMA</button>
            <span>/</span><span className="mono">{DOC2.code}</span><span>/</span><span style={{ color: 'var(--text)' }}>Distribuição</span>
          </div>
          <div style={{ display: 'flex', alignItems: 'flex-end', justifyContent: 'space-between', marginBottom: 18 }}>
            <div>
              <h1 className="display-1" style={{ fontSize: 28, marginBottom: 4 }}>Cobertura de leitura</h1>
              <div className="caption">{DOC2.title} · vence em <strong style={{ color: 'var(--warning)' }}>{FANOUT2.daysLeft} dias</strong> ({FANOUT2.deadline})</div>
            </div>
            <div style={{ display: 'flex', gap: 6 }}>
              <button className="btn"><Icon name="download" size={13}/>Exportar relatório</button>
              <button className="btn btn-primary"><Icon name="mail" size={13}/>Lembrete em massa</button>
            </div>
          </div>
        </div>

        <div style={{ padding: '0 32px 40px', display: 'flex', flexDirection: 'column', gap: 18 }}>
          {/* TOP ROW — donut + KPIs + sparkline trend */}
          <div style={{ display: 'grid', gridTemplateColumns: '320px 1fr', gap: 18 }}>
            {/* Donut card */}
            <div className="card" style={{ padding: 22, display: 'flex', flexDirection: 'column', gap: 16 }}>
              <div className="kicker">Cobertura geral</div>
              <div style={{ display: 'flex', alignItems: 'center', gap: 18 }}>
                <div style={{ position: 'relative', width: 130, height: 130, flexShrink: 0 }}>
                  <svg viewBox="0 0 100 100" style={{ width: '100%', height: '100%', transform: 'rotate(-90deg)' }}>
                    <circle cx="50" cy="50" r="42" fill="none" stroke="var(--surface-3)" strokeWidth="9"/>
                    {/* Read ring (lighter, behind) */}
                    <circle cx="50" cy="50" r="42" fill="none" stroke="var(--brand-soft)" strokeWidth="9" strokeLinecap="round" opacity="0.5"
                      strokeDasharray={2 * Math.PI * 42}
                      strokeDashoffset={2 * Math.PI * 42 * (1 - (ready ? pctReached / 100 : 0))}
                      style={{ transition: 'stroke-dashoffset 1.6s cubic-bezier(.2,.8,.2,1)' }}/>
                    {/* Ack ring (front) */}
                    <circle cx="50" cy="50" r="42" fill="none" stroke="var(--brand)" strokeWidth="9" strokeLinecap="round"
                      strokeDasharray={2 * Math.PI * 42}
                      strokeDashoffset={2 * Math.PI * 42 * (1 - (ready ? pct / 100 : 0))}
                      style={{ transition: 'stroke-dashoffset 1.6s cubic-bezier(.2,.8,.2,1) .15s' }}/>
                  </svg>
                  <div style={{ position: 'absolute', inset: 0, display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center' }}>
                    <div className="num" style={{ fontSize: 32, lineHeight: 1, fontWeight: 600, fontFamily: 'var(--font-display)', letterSpacing: '-0.02em' }}>{animatedPct}<span style={{ fontSize: 16, color: 'var(--text-muted)' }}>%</span></div>
                    <div className="kicker" style={{ marginTop: 4, fontSize: 9.5 }}>ack</div>
                  </div>
                </div>
                <div style={{ flex: 1, display: 'flex', flexDirection: 'column', gap: 10 }}>
                  <LegendRow color="var(--brand)" label="Reconheceu" value={FANOUT2.acknowledged} total={FANOUT2.totalTargets}/>
                  <LegendRow color="var(--brand-soft)" label="Apenas leu" value={FANOUT2.read - FANOUT2.acknowledged} total={FANOUT2.totalTargets} faded/>
                  <LegendRow color="var(--surface-3)" label="Pendente" value={FANOUT2.pending} total={FANOUT2.totalTargets}/>
                </div>
              </div>
            </div>

            {/* KPI strip + trend chart */}
            <div className="card" style={{ padding: 0, display: 'grid', gridTemplateColumns: '1fr 1fr 1fr', gap: 0, overflow: 'hidden' }}>
              {/* KPI 1 */}
              <KPICell ready={ready} kicker="Total alvos" value={FANOUT2.totalTargets} sub="248 pessoas em 8 áreas"/>
              {/* KPI 2 */}
              <KPICell ready={ready} kicker="Pendentes" value={FANOUT2.pending} sub="6 áreas com pendência" tone="warning" trend="-4 vs. ontem"/>
              {/* KPI 3 */}
              <KPICell ready={ready} kicker="Em atraso" value={FANOUT2.overdue} sub="precisam de lembrete" tone="danger" trend="+2 vs. ontem"/>
              {/* Trend area — full width below */}
              <div style={{ gridColumn: '1 / -1', borderTop: '1px solid var(--border)', padding: '18px 22px' }}>
                <div style={{ display: 'flex', alignItems: 'baseline', justifyContent: 'space-between', marginBottom: 8 }}>
                  <div>
                    <div className="kicker">Acumulado de reconhecimentos</div>
                    <div className="tiny" style={{ marginTop: 2 }}>desde a publicação · 8 dias</div>
                  </div>
                  <div style={{ display: 'flex', gap: 14, fontSize: 11, color: 'var(--text-muted)' }}>
                    <span style={{ display: 'inline-flex', alignItems: 'center', gap: 4 }}><span style={{ width: 16, height: 2, background: 'var(--brand)' }}/>Reconhecimentos</span>
                    <span style={{ display: 'inline-flex', alignItems: 'center', gap: 4 }}><span style={{ width: 16, height: 0, borderTop: '2px dashed var(--text-faint)' }}/>Meta (228 / 92%)</span>
                  </div>
                </div>
                <Sparkline ready={ready}/>
              </div>
            </div>
          </div>

          {/* MIDDLE — area ranking with bullet bars */}
          <div className="card" style={{ padding: 24 }}>
            <div style={{ display: 'flex', alignItems: 'baseline', justifyContent: 'space-between', marginBottom: 18 }}>
              <div>
                <div className="kicker">Cobertura por área</div>
                <h2 className="h2" style={{ marginTop: 4 }}>Onde está a pendência</h2>
              </div>
              <div style={{ display: 'flex', gap: 16, fontSize: 11, color: 'var(--text-muted)' }}>
                <span style={{ display: 'inline-flex', alignItems: 'center', gap: 5 }}><span style={{ width: 10, height: 10, background: 'var(--brand)', borderRadius: 2 }}/>Ack</span>
                <span style={{ display: 'inline-flex', alignItems: 'center', gap: 5 }}><span style={{ width: 10, height: 10, background: 'var(--brand-soft)', opacity: 0.4, borderRadius: 2 }}/>Apenas leu</span>
                <span style={{ display: 'inline-flex', alignItems: 'center', gap: 5 }}><span style={{ width: 10, height: 10, border: '1px dashed var(--text-faint)', borderRadius: 2 }}/>Pendente</span>
              </div>
            </div>
            <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
              {areas.map((a, i) => (
                <AreaRow key={i} area={a} ready={ready}/>
              ))}
            </div>
          </div>

          {/* BOTTOM — table */}
          <div className="card" style={{ overflow: 'hidden' }}>
            <div style={{ padding: '16px 22px', borderBottom: '1px solid var(--border)', display: 'flex', alignItems: 'center', gap: 12 }}>
              <div>
                <div className="kicker">Destinatários</div>
                <h2 className="h2" style={{ marginTop: 4, fontSize: 18 }}>Lista detalhada</h2>
              </div>
              <div style={{ display: 'inline-flex', background: 'var(--surface-2)', border: '1px solid var(--border)', borderRadius: 6, padding: 3, marginLeft: 'auto' }}>
                {[
                  { id: 'pending', label: 'Pendentes', n: tabCounts.pending },
                  { id: 'overdue', label: 'Em atraso', n: tabCounts.overdue, tone: 'var(--danger)' },
                  { id: 'read', label: 'Apenas leram', n: tabCounts.read },
                  { id: 'acked', label: 'Concluído', n: tabCounts.acked },
                  { id: 'all', label: 'Todos', n: tabCounts.all },
                ].map(t => (
                  <button key={t.id} onClick={() => setTab(t.id)} style={{
                    height: 28, padding: '0 12px', border: 'none', cursor: 'pointer',
                    background: tab === t.id ? 'var(--surface)' : 'transparent',
                    boxShadow: tab === t.id ? 'var(--shadow-1)' : 'none',
                    borderRadius: 4, fontSize: 12, fontWeight: tab === t.id ? 500 : 400,
                    color: tab === t.id ? 'var(--text)' : 'var(--text-muted)',
                    display: 'inline-flex', alignItems: 'center', gap: 6,
                  }}>
                    {t.label}
                    <span className="mono" style={{ fontSize: 10, padding: '1px 5px', background: tab === t.id ? (t.tone || 'var(--brand-pale)') : 'var(--surface-3)', color: tab === t.id ? (t.tone ? '#fff' : 'var(--brand)') : 'var(--text-muted)', borderRadius: 3 }}>{t.n}</span>
                  </button>
                ))}
              </div>
              <div className="input" style={{ width: 240, display: 'inline-flex', alignItems: 'center', gap: 8, color: 'var(--text-muted)' }}>
                <Icon name="search" size={13}/>Filtrar…
              </div>
            </div>

            <div style={{ display: 'grid', gridTemplateColumns: '32px 1.5fr 1fr 1fr 0.7fr 130px 80px', gap: 12, padding: '10px 22px', background: 'var(--surface-2)', borderBottom: '1px solid var(--border)', fontSize: 10.5, color: 'var(--text-muted)', textTransform: 'uppercase', letterSpacing: '0.05em', fontWeight: 500 }}>
              <span><input type="checkbox" style={{ accentColor: 'var(--brand)' }}/></span>
              <span>Destinatário</span>
              <span>Função</span>
              <span>Área</span>
              <span></span>
              <span>Última atividade</span>
              <span></span>
            </div>
            {filtered.map((p, i) => (
              <PersonRow key={i} p={p} last={i === filtered.length - 1}/>
            ))}
            <div style={{ padding: '12px 22px', background: 'var(--surface-2)', borderTop: '1px solid var(--border)', display: 'flex', justifyContent: 'space-between', alignItems: 'center', fontSize: 12, color: 'var(--text-muted)' }}>
              <span>Mostrando {filtered.length} de {FANOUT2.totalTargets}</span>
              <span style={{ display: 'inline-flex', gap: 4 }}>
                <button className="btn btn-sm btn-ghost"><Icon name="chevron-left" size={12}/></button>
                <button className="btn btn-sm btn-ghost"><Icon name="chevron-right" size={12}/></button>
              </span>
            </div>
          </div>
        </div>
      </div>
      <style>{`@keyframes sk2 { 0% { background-position: 200% 0; } 100% { background-position: -200% 0; } }`}</style>
    </div>
  );
}

// --- Small components used by FanoutV2 ---

function LegendRow({ color, label, value, total, faded }) {
  return (
    <div style={{ display: 'flex', alignItems: 'center', gap: 8, fontSize: 12.5 }}>
      <span style={{ width: 10, height: 10, background: color, borderRadius: 2, opacity: faded ? 0.5 : 1, flexShrink: 0 }}/>
      <span style={{ flex: 1, color: 'var(--text-soft)' }}>{label}</span>
      <span className="mono num" style={{ fontWeight: 500 }}>{value}</span>
      <span className="tiny mono" style={{ minWidth: 36, textAlign: 'right' }}>{Math.round(value/total*100)}%</span>
    </div>
  );
}

function KPICell({ ready, kicker, value, sub, tone, trend }) {
  const n = useCounter2(ready ? value : 0, 1400);
  const color = tone === 'warning' ? 'var(--warning)' : tone === 'danger' ? 'var(--danger)' : 'var(--text)';
  return (
    <div style={{ padding: '20px 22px', borderRight: '1px solid var(--border)', display: 'flex', flexDirection: 'column', gap: 6 }}>
      <div className="kicker">{kicker}</div>
      {ready ? (
        <div className="num" style={{ fontSize: 38, fontWeight: 600, fontFamily: 'var(--font-display)', letterSpacing: '-0.02em', color, lineHeight: 1 }}>{n}</div>
      ) : (
        <div style={{ height: 38, width: 64, background: 'linear-gradient(90deg, var(--surface-2), var(--surface-3), var(--surface-2))', backgroundSize: '200% 100%', animation: 'sk2 1.4s linear infinite', borderRadius: 4 }}/>
      )}
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'baseline' }}>
        <span className="tiny" style={{ fontSize: 11 }}>{sub}</span>
        {trend && <span className="tiny mono" style={{ fontSize: 10, color }}>{trend}</span>}
      </div>
    </div>
  );
}

function Sparkline({ ready }) {
  const data = FANOUT2.dailyReads;
  const w = 1, h = 1;
  const max = 248; // total targets, so y is comparable to goal
  const step = 100 / (data.length - 1);
  const points = data.map((d, i) => `${i * step},${100 - (d.cumAck / max) * 100}`);
  const path = `M ${points.join(' L ')}`;
  const area = `M 0,100 L ${points.join(' L ')} L 100,100 Z`;
  const goal = (228 / max) * 100; // 92% of total

  return (
    <div style={{ position: 'relative', height: 110 }}>
      <svg viewBox="0 0 100 100" preserveAspectRatio="none" style={{ width: '100%', height: '100%' }}>
        {/* Goal dashed line */}
        <line x1="0" x2="100" y1={100 - goal} y2={100 - goal} stroke="var(--text-faint)" strokeWidth="0.4" strokeDasharray="1.5,1.5" vectorEffect="non-scaling-stroke"/>
        {/* Gradient under the line */}
        <defs>
          <linearGradient id="sparkfill" x1="0" x2="0" y1="0" y2="1">
            <stop offset="0%" stopColor="var(--brand)" stopOpacity="0.22"/>
            <stop offset="100%" stopColor="var(--brand)" stopOpacity="0"/>
          </linearGradient>
        </defs>
        <path d={area} fill="url(#sparkfill)" style={{
          opacity: ready ? 1 : 0,
          transition: 'opacity .8s ease .4s',
        }}/>
        <path d={path} fill="none" stroke="var(--brand)" strokeWidth="0.7" strokeLinecap="round" strokeLinejoin="round" vectorEffect="non-scaling-stroke" style={{
          strokeDasharray: 200, strokeDashoffset: ready ? 0 : 200,
          transition: 'stroke-dashoffset 1.6s cubic-bezier(.2,.8,.2,1)',
        }}/>
        {/* Endpoint dot */}
        <circle cx={points[points.length-1].split(',')[0]} cy={points[points.length-1].split(',')[1]} r="1.2" fill="var(--brand)" vectorEffect="non-scaling-stroke" style={{ opacity: ready ? 1 : 0, transition: 'opacity .4s ease 1.2s' }}/>
      </svg>
      {/* X-axis labels */}
      <div style={{ display: 'flex', justifyContent: 'space-between', position: 'absolute', left: 0, right: 0, bottom: -4, fontSize: 10, color: 'var(--text-muted)', fontFamily: 'var(--font-mono)' }}>
        {data.map((d, i) => <span key={i}>{i % 2 === 0 ? d.day : ''}</span>)}
      </div>
    </div>
  );
}

function AreaRow({ area, ready }) {
  const ackPct = ready ? area.pct * 100 : 0;
  const readPct = ready ? area.pctRead * 100 : 0;
  const isCritical = area.pct < 0.5;

  return (
    <div style={{ display: 'grid', gridTemplateColumns: '180px 1fr 90px 60px', gap: 14, alignItems: 'center' }}>
      <div>
        <div style={{ fontSize: 13, fontWeight: 500, display: 'flex', alignItems: 'center', gap: 6 }}>
          {isCritical && <span style={{ width: 5, height: 5, borderRadius: '50%', background: 'var(--danger)' }}/>}
          {area.area}
        </div>
        <div className="tiny" style={{ fontSize: 10.5 }}>{area.ack} ack · {area.read} leram · {area.total} alvos</div>
      </div>
      <div style={{ position: 'relative', height: 22, background: 'var(--surface-3)', borderRadius: 4, overflow: 'hidden' }}>
        {/* Read background bar (lighter) */}
        <div style={{
          position: 'absolute', left: 0, top: 0, bottom: 0,
          width: `${readPct}%`, background: 'var(--brand-soft)', opacity: 0.35,
          transition: 'width 1.6s cubic-bezier(.2,.8,.2,1) .1s',
        }}/>
        {/* Ack bar (front) */}
        <div style={{
          position: 'absolute', left: 0, top: 0, bottom: 0,
          width: `${ackPct}%`, background: 'var(--brand)',
          transition: 'width 1.6s cubic-bezier(.2,.8,.2,1) .25s',
        }}/>
        {/* Goal marker at 92% */}
        <div style={{
          position: 'absolute', left: '92%', top: -2, bottom: -2, width: 1,
          borderLeft: '1px dashed var(--text-faint)',
        }}/>
      </div>
      <div className="num mono" style={{ fontSize: 13, textAlign: 'right', fontWeight: 600, color: isCritical ? 'var(--danger)' : 'var(--text)' }}>
        {Math.round(area.pct * 100)}%
      </div>
      <div style={{ textAlign: 'right' }}>
        <button className="btn btn-sm btn-ghost"><Icon name="mail" size={11}/></button>
      </div>
    </div>
  );
}

function PersonRow({ p, last }) {
  const [hover, setHover] = useState(false);
  const tone = p.status === 'ack' ? 'pill-frozen' : p.status === 'read' ? 'pill-approved' : p.overdue ? 'pill-rejected' : 'pill-draft';
  const label = p.status === 'ack' ? 'reconheceu' : p.status === 'read' ? 'apenas leu' : p.overdue ? 'em atraso' : 'pendente';
  return (
    <div
      onMouseEnter={() => setHover(true)}
      onMouseLeave={() => setHover(false)}
      style={{
        display: 'grid', gridTemplateColumns: '32px 1.5fr 1fr 1fr 0.7fr 130px 80px', gap: 12,
        padding: '12px 22px', borderBottom: last ? 'none' : '1px solid var(--border)',
        alignItems: 'center', fontSize: 13, position: 'relative',
        background: hover ? 'var(--surface-2)' : 'transparent', transition: 'background .15s',
      }}>
      <span><input type="checkbox" style={{ accentColor: 'var(--brand)' }}/></span>
      <div style={{ display: 'flex', alignItems: 'center', gap: 10, position: 'relative' }}>
        <Avatar name={p.who} size={28}/>
        <span style={{ fontWeight: 500 }}>{p.who}</span>
        {hover && (
          <div style={{
            position: 'absolute', top: '100%', left: 38, marginTop: 6, zIndex: 30,
            background: '#1a0e0e', color: '#fff', padding: '8px 12px', borderRadius: 6,
            fontSize: 11.5, lineHeight: 1.5, whiteSpace: 'nowrap',
            boxShadow: '0 8px 20px rgba(0,0,0,0.3)',
          }}>
            <div style={{ fontWeight: 600 }}>{p.who}</div>
            <div style={{ opacity: 0.7 }}>{p.role} · {p.area}</div>
            <div style={{ opacity: 0.7, fontFamily: 'var(--font-mono)', fontSize: 10.5, marginTop: 4 }}>{p.last}</div>
          </div>
        )}
      </div>
      <span className="caption" style={{ fontSize: 12 }}>{p.role}</span>
      <span className="caption" style={{ fontSize: 12 }}>{p.area}</span>
      <span/>
      <span className="caption" style={{ fontSize: 11.5, color: p.last.includes('nunca') ? 'var(--danger)' : 'var(--text-muted)' }}>{p.last}</span>
      <div style={{ display: 'flex', justifyContent: 'flex-end', gap: 4 }}>
        <span className={`pill ${tone}`}><span className="dot"/>{label}</span>
      </div>
    </div>
  );
}

window.MD_ONDA1_V2 = { PublicadoV2, FanoutV2 };
