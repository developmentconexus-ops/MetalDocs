/* global React, MD */
const { Icon, Logo, Avatar, Sidebar, Toolbar } = MD;
const { useState, useEffect, useRef } = React;

// ===== Shared mock data =====
const DOC = {
  code: 'PR-EHS-014',
  title: 'Procedimento de Bloqueio e Etiquetagem (LOTO)',
  version: 'v3.2',
  publishedAt: '12 mar 2026 · 14:32',
  publishedBy: 'Carolina Mendes',
  area: 'Saúde, Segurança & Meio Ambiente',
  hash: 'a7f3c9e1d8b4',
  hashFull: 'a7f3c9e1d8b4f6e2c1a9b3d7e8f4c2a1b9d6e3f7a8c4b2d1e9f6a3c7b8d4e2f1',
  pageCount: 14,
  effectiveAt: '15 mar 2026',
  reviewIn: '2027-03-15',
  signoffs: [
    { name: 'Renata Souza', role: 'Coord. SSMA', when: '11 mar · 18:04' },
    { name: 'Marcos Lima', role: 'Gerente Industrial', when: '12 mar · 09:12' },
    { name: 'Diretoria', role: 'Anuente', when: '12 mar · 14:30' },
  ],
  versions: [
    { v: 'v3.2', when: '12 mar 2026', author: 'C. Mendes', current: true, summary: 'Inclusão de procedimento para painéis CCM-04. Atualização da matriz de risco.' },
    { v: 'v3.1', when: '08 jan 2026', author: 'C. Mendes', summary: 'Correção de typo no item 4.3.' },
    { v: 'v3.0', when: '14 nov 2025', author: 'R. Souza', summary: 'Revisão geral pós-auditoria interna. Novas etiquetas padronizadas.' },
    { v: 'v2.4', when: '02 jul 2025', author: 'R. Souza', summary: 'Última versão da revisão anterior.' },
    { v: 'v2.3', when: '18 mar 2025', author: 'A. Tavares', summary: '—' },
  ],
};

const FANOUT = {
  totalTargets: 248,
  read: 182,
  acknowledged: 156,
  pending: 66,
  overdue: 12,
  deadline: '22 mar 2026',
  daysLeft: 8,
  byArea: [
    { area: 'Manutenção', total: 64, read: 58, ack: 52, color: 1.00 },
    { area: 'Produção · Linha A', total: 42, read: 38, ack: 34, color: 0.95 },
    { area: 'Produção · Linha B', total: 38, read: 22, ack: 18, color: 0.55 },
    { area: 'SSMA', total: 18, read: 18, ack: 18, color: 1.00 },
    { area: 'Engenharia', total: 24, read: 19, ack: 16, color: 0.78 },
    { area: 'Logística', total: 28, read: 14, ack: 10, color: 0.46 },
    { area: 'Qualidade', total: 16, read: 8, ack: 5, color: 0.35 },
    { area: 'Administrativo', total: 18, read: 5, ack: 3, color: 0.18 },
  ],
  recentReads: [
    { who: 'Daniela Prado', area: 'Manutenção', when: 'há 4 min', status: 'ack' },
    { who: 'João Vitor Souza', area: 'Produção · A', when: 'há 12 min', status: 'read' },
    { who: 'Eduardo Castro', area: 'Engenharia', when: 'há 28 min', status: 'ack' },
    { who: 'Marina Albuquerque', area: 'Manutenção', when: 'há 1h', status: 'read' },
    { who: 'Pedro Henrique Lima', area: 'Logística', when: 'há 2h', status: 'ack' },
  ],
  pendingPeople: [
    { who: 'Roberto Carvalho', area: 'Logística', last: 'nunca abriu', overdue: false },
    { who: 'Ana Paula Mendes', area: 'Administrativo', last: 'nunca abriu', overdue: true },
    { who: 'Thiago Rocha', area: 'Qualidade', last: '08 mar (apenas leu)', overdue: false },
    { who: 'Felipe Andrade', area: 'Produção · B', last: 'há 6 dias', overdue: false },
    { who: 'Beatriz Moreira', area: 'Administrativo', last: 'nunca abriu', overdue: true },
    { who: 'Gabriel Costa', area: 'Logística', last: 'nunca abriu', overdue: true },
  ],
};

// ============================================================
// Hooks
// ============================================================
function useCounter(target, duration = 1200) {
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
// Shared atoms
// ============================================================
function ISOSeal({ hash, hashFull, pulsing = true }) {
  const [open, setOpen] = useState(false);
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
        background: 'var(--success-bg)', color: 'var(--success)',
        border: '1px solid var(--success)', borderRadius: 'var(--r-pill)',
        cursor: 'help',
      }}>
        <span style={{
          width: 6, height: 6, borderRadius: '50%', background: 'currentColor',
          animation: pulsing ? 'sealPulse 2.4s ease-in-out infinite' : 'none',
        }}/>
        ISO 9001 · {hash}
      </span>
      {open && (
        <div style={{
          position: 'absolute', top: 'calc(100% + 6px)', right: 0, zIndex: 50,
          padding: '10px 12px', background: 'var(--text)', color: 'var(--surface)',
          borderRadius: 6, fontSize: 11, lineHeight: 1.5, width: 320,
          boxShadow: 'var(--shadow-2)',
        }}>
          <div style={{ fontFamily: 'var(--font-mono)', fontSize: 9.5, opacity: 0.7, marginBottom: 4 }}>SHA-256 (12 + ...)</div>
          <div style={{ fontFamily: 'var(--font-mono)', fontSize: 11, wordBreak: 'break-all', opacity: 0.95 }}>{hashFull}</div>
          <div style={{ marginTop: 8, paddingTop: 8, borderTop: '1px solid rgba(255,255,255,0.12)', fontSize: 10.5, opacity: 0.75 }}>
            Documento congelado. Cadeia de hash íntegra · Cl. 7.5.3
          </div>
        </div>
      )}
      <style>{`@keyframes sealPulse { 0%,100% { opacity: 1; transform: scale(1);} 50% { opacity: .55; transform: scale(.85);} }`}</style>
    </span>
  );
}

function MetaRow({ label, value, mono }) {
  return (
    <div style={{ display: 'grid', gridTemplateColumns: '110px 1fr', gap: 12, padding: '8px 0', borderBottom: '1px solid var(--border)', fontSize: 12.5 }}>
      <div className="kicker" style={{ alignSelf: 'center' }}>{label}</div>
      <div style={{ color: 'var(--text)', fontFamily: mono ? 'var(--font-mono)' : 'inherit', fontSize: mono ? 11 : 12.5 }}>{value}</div>
    </div>
  );
}

function SkeletonNumber({ ready, value, suffix = '', size = 36 }) {
  const n = useCounter(ready ? value : 0, 1400);
  if (!ready) {
    return (
      <div style={{
        height: size + 4, width: size * 2,
        background: 'linear-gradient(90deg, var(--surface-2) 0%, var(--surface-3) 50%, var(--surface-2) 100%)',
        backgroundSize: '200% 100%', animation: 'sk 1.4s linear infinite',
        borderRadius: 4,
      }}/>
    );
  }
  return (
    <div className="num" style={{ fontSize: size, lineHeight: 1, fontWeight: 600, letterSpacing: '-0.03em', fontFamily: 'var(--font-display)' }}>
      {n}{suffix}
    </div>
  );
}

function DocBody() {
  return (
    <div style={{ fontSize: 13, lineHeight: 1.7, color: 'var(--text-soft)' }}>
      <h3 style={{ fontSize: 14, fontWeight: 600, color: 'var(--text)', margin: '0 0 8px' }}>1. Objetivo</h3>
      <p style={{ margin: '0 0 18px' }}>Estabelecer o procedimento padrão para bloqueio e etiquetagem (LOTO — <em>Lockout/Tagout</em>) de fontes de energia perigosa em equipamentos da planta industrial, garantindo segurança durante manutenção, limpeza ou ajustes.</p>
      <h3 style={{ fontSize: 14, fontWeight: 600, color: 'var(--text)', margin: '0 0 8px' }}>2. Aplicação</h3>
      <p style={{ margin: '0 0 18px' }}>Aplica-se a todos os colaboradores próprios e terceiros que executem qualquer atividade em equipamentos onde a liberação inesperada de energia possa causar lesão ou dano material.</p>
      <h3 style={{ fontSize: 14, fontWeight: 600, color: 'var(--text)', margin: '0 0 8px' }}>3. Definições</h3>
      <p style={{ margin: '0 0 12px' }}>Para os efeitos deste procedimento, considera-se:</p>
      <ul style={{ margin: '0 0 18px', paddingLeft: 18 }}>
        <li style={{ marginBottom: 4 }}><strong style={{ color: 'var(--text)' }}>Bloqueio (Lockout):</strong> dispositivo físico (cadeado, tag) que impede a operação do equipamento.</li>
        <li style={{ marginBottom: 4 }}><strong style={{ color: 'var(--text)' }}>Etiqueta (Tagout):</strong> aviso visual de não-operação, fixado ao dispositivo de bloqueio.</li>
        <li><strong style={{ color: 'var(--text)' }}>Energia perigosa:</strong> elétrica, mecânica, hidráulica, pneumática, térmica, química ou gravitacional.</li>
      </ul>
      <h3 style={{ fontSize: 14, fontWeight: 600, color: 'var(--text)', margin: '0 0 8px' }}>4. Responsabilidades</h3>
      <p style={{ margin: 0 }}>O <strong style={{ color: 'var(--text)' }}>colaborador autorizado</strong> deve aplicar e remover seu próprio dispositivo de bloqueio. O <strong style={{ color: 'var(--text)' }}>supervisor</strong> garante a aplicação do procedimento e mantém o registro no sistema MetalDocs.</p>
    </div>
  );
}

function ActionsCluster({ flat = false }) {
  const baseBtn = flat ? { background: 'transparent', border: 'none' } : {};
  return (
    <>
      <button className="btn btn-primary"><Icon name="download" size={14}/>Baixar PDF assinado</button>
      <button className="btn" style={baseBtn}><Icon name="link" size={14}/>Copiar link</button>
      <button className="btn" style={baseBtn}><Icon name="more" size={14}/>Mais</button>
    </>
  );
}

// ============================================================
// VARIATION A — A4 PAPER (publicado)
// ============================================================
function PublicadoA() {
  return (
    <div className="palette-wine" style={{ height: 980, display: 'flex', background: 'var(--bg)' }}>
      <Sidebar variant="rail" active="documents"/>
      <div style={{ flex: 1, display: 'flex', flexDirection: 'column', minWidth: 0 }}>
        {/* Top crumb bar */}
        <div style={{ height: 44, padding: '0 22px', borderBottom: '1px solid var(--border)', background: 'var(--surface)', display: 'flex', alignItems: 'center', gap: 12 }}>
          <button className="btn btn-ghost btn-sm"><Icon name="chevron-left" size={13}/>Voltar</button>
          <div style={{ width: 1, height: 16, background: 'var(--border)' }}/>
          <span className="caption">Documentos · SSMA · Procedimentos</span>
          <span style={{ marginLeft: 'auto' }}><ISOSeal hash={DOC.hash} hashFull={DOC.hashFull}/></span>
        </div>
        <div style={{ flex: 1, display: 'grid', gridTemplateColumns: '1fr 320px', minHeight: 0 }}>
          {/* A4 paper viewport */}
          <div style={{ overflow: 'auto', padding: '32px 0', background: 'var(--bg)', display: 'flex', justifyContent: 'center' }}>
            <div style={{ width: 720, background: 'var(--surface)', boxShadow: '0 1px 3px rgba(0,0,0,0.06), 0 12px 32px rgba(74,33,33,0.10)', border: '1px solid var(--border)' }}>
              {/* Paper header */}
              <div style={{ borderBottom: '2px solid var(--brand)', padding: '32px 56px 20px', display: 'grid', gridTemplateColumns: '1fr auto', alignItems: 'end', gap: 24 }}>
                <div>
                  <div className="kicker" style={{ marginBottom: 6 }}>{DOC.code} · {DOC.version}</div>
                  <h1 style={{ fontSize: 24, lineHeight: 1.2, letterSpacing: '-0.02em', margin: 0, color: 'var(--text)', fontWeight: 600 }}>{DOC.title}</h1>
                </div>
                <Logo size={32}/>
              </div>
              {/* Paper body */}
              <div style={{ padding: '28px 56px 40px' }}>
                <DocBody/>
              </div>
              {/* Paper footer */}
              <div style={{ padding: '14px 56px', borderTop: '1px solid var(--border)', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                <span className="tiny mono">{DOC.code} · {DOC.version}</span>
                <span className="tiny mono">1 / {DOC.pageCount}</span>
              </div>
            </div>
          </div>
          {/* Right panel */}
          <aside style={{ borderLeft: '1px solid var(--border)', background: 'var(--surface)', padding: 22, overflow: 'auto', display: 'flex', flexDirection: 'column', gap: 18 }}>
            <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
              <ActionsCluster/>
            </div>
            <div>
              <div className="kicker" style={{ marginBottom: 8 }}>Identificação</div>
              <MetaRow label="Versão" value={<><strong>{DOC.version}</strong> · publicada</>}/>
              <MetaRow label="Publicada" value={DOC.publishedAt}/>
              <MetaRow label="Por" value={DOC.publishedBy}/>
              <MetaRow label="Vigência" value={`${DOC.effectiveAt} → 15 mar 2027`}/>
              <MetaRow label="Hash" value={DOC.hash + '...'} mono/>
            </div>
            <div>
              <div className="kicker" style={{ marginBottom: 8 }}>Aprovações</div>
              {DOC.signoffs.map((s, i) => (
                <div key={i} style={{ display: 'flex', alignItems: 'center', gap: 10, padding: '10px 0', borderBottom: i < DOC.signoffs.length - 1 ? '1px solid var(--border)' : 'none' }}>
                  <Avatar name={s.name} size={28}/>
                  <div style={{ flex: 1, minWidth: 0 }}>
                    <div style={{ fontSize: 12.5, fontWeight: 500 }}>{s.name}</div>
                    <div className="caption" style={{ fontSize: 11 }}>{s.role}</div>
                  </div>
                  <div className="tiny mono" style={{ fontSize: 10.5 }}>{s.when}</div>
                </div>
              ))}
            </div>
          </aside>
        </div>
      </div>
    </div>
  );
}

// ============================================================
// VARIATION B — NOTION/LINEAR (publicado)
// ============================================================
function PublicadoB() {
  const [revOpen, setRevOpen] = useState(false);
  return (
    <div className="palette-wine" style={{ height: 980, display: 'flex', background: 'var(--surface)' }}>
      <Sidebar variant="rail" active="documents"/>
      <div style={{ flex: 1, display: 'flex', flexDirection: 'column', minWidth: 0 }}>
        {/* Slim header */}
        <div style={{ height: 40, padding: '0 22px', borderBottom: '1px solid var(--border)', display: 'flex', alignItems: 'center', gap: 10, background: 'var(--surface)' }}>
          <Icon name="chevron-left" size={14}/>
          <span className="caption">SSMA / Procedimentos</span>
          <span style={{ marginLeft: 'auto', display: 'flex', gap: 6, alignItems: 'center' }}>
            <ISOSeal hash={DOC.hash} hashFull={DOC.hashFull}/>
            <button className="btn btn-sm btn-ghost"><Icon name="link" size={12}/>Compartilhar</button>
            <button className="btn btn-sm btn-primary"><Icon name="download" size={12}/>PDF</button>
          </span>
        </div>
        <div style={{ flex: 1, display: 'grid', gridTemplateColumns: '1fr 280px', minHeight: 0 }}>
          {/* Editor-like body */}
          <div style={{ overflow: 'auto', padding: '40px 80px 80px' }}>
            <div style={{ maxWidth: 720, margin: '0 auto' }}>
              <div style={{ display: 'flex', gap: 8, marginBottom: 14 }}>
                <span className="pill pill-frozen"><span className="dot"/>Publicado</span>
                <span className="pill"><span style={{ width: 5, height: 5, borderRadius: '50%', background: 'var(--brand)', display: 'inline-block', marginRight: 4 }}/>{DOC.code}</span>
                <span className="pill mono">{DOC.version}</span>
              </div>
              <h1 style={{ fontSize: 38, lineHeight: 1.1, fontWeight: 600, letterSpacing: '-0.025em', margin: '0 0 14px', color: 'var(--text)' }}>{DOC.title}</h1>
              <div style={{ display: 'flex', alignItems: 'center', gap: 10, color: 'var(--text-muted)', fontSize: 13, marginBottom: 32, paddingBottom: 14, borderBottom: '1px solid var(--border)' }}>
                <Avatar name={DOC.publishedBy} size={20}/>
                <span>{DOC.publishedBy}</span>
                <span style={{ width: 3, height: 3, borderRadius: '50%', background: 'var(--text-faint)' }}/>
                <span>publicado em {DOC.publishedAt}</span>
                <span style={{ width: 3, height: 3, borderRadius: '50%', background: 'var(--text-faint)' }}/>
                <span>vigente até 15 mar 2027</span>
              </div>
              <DocBody/>
              <div style={{ marginTop: 40, paddingTop: 24, borderTop: '1px solid var(--border)', fontSize: 12, color: 'var(--text-muted)', display: 'flex', justifyContent: 'space-between' }}>
                <span>Página 1 de {DOC.pageCount}</span>
                <span className="mono">{DOC.code} · {DOC.version} · {DOC.hash}</span>
              </div>
            </div>
          </div>
          {/* Slim right rail */}
          <aside style={{ borderLeft: '1px solid var(--border)', background: 'var(--surface-2)', padding: '22px 18px', overflow: 'auto', display: 'flex', flexDirection: 'column', gap: 22 }}>
            <button className="btn btn-primary" style={{ width: '100%', justifyContent: 'center' }}><Icon name="edit" size={14}/>Iniciar revisão (v3.3)</button>
            <div>
              <div className="kicker" style={{ marginBottom: 8 }}>Histórico</div>
              <div style={{ display: 'flex', flexDirection: 'column' }}>
                {DOC.versions.slice(0, revOpen ? DOC.versions.length : 3).map((v, i) => (
                  <div key={i} style={{ display: 'flex', alignItems: 'flex-start', gap: 10, padding: '10px 0', borderBottom: '1px solid var(--border)' }}>
                    <div style={{ marginTop: 4, width: 7, height: 7, borderRadius: '50%', background: v.current ? 'var(--brand)' : 'var(--border-strong)', flexShrink: 0 }}/>
                    <div style={{ flex: 1, minWidth: 0 }}>
                      <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: 12 }}>
                        <span className="mono" style={{ fontWeight: v.current ? 600 : 400, color: v.current ? 'var(--brand)' : 'var(--text)' }}>{v.v}</span>
                        <span className="tiny" style={{ fontSize: 10.5 }}>{v.when}</span>
                      </div>
                      <div className="tiny" style={{ fontSize: 10.5, marginTop: 2 }}>{v.author}</div>
                    </div>
                  </div>
                ))}
                <button className="btn btn-ghost btn-sm" style={{ marginTop: 6, justifyContent: 'flex-start', padding: 0 }} onClick={() => setRevOpen(o => !o)}>
                  {revOpen ? '— Recolher' : `+ Ver todas (${DOC.versions.length})`}
                </button>
              </div>
            </div>
            <div>
              <div className="kicker" style={{ marginBottom: 8 }}>Aprovaram</div>
              <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
                {DOC.signoffs.map((s, i) => (
                  <div key={i} style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                    <Avatar name={s.name} size={22}/>
                    <div style={{ flex: 1, minWidth: 0, fontSize: 11.5 }}>
                      <div style={{ fontWeight: 500 }}>{s.name}</div>
                      <div className="tiny" style={{ fontSize: 10 }}>{s.role}</div>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          </aside>
        </div>
      </div>
    </div>
  );
}

// ============================================================
// VARIATION C — HERO + TIMELINE (publicado)
// ============================================================
function PublicadoC() {
  return (
    <div className="palette-wine" style={{ height: 980, display: 'flex', background: 'var(--bg)' }}>
      <Sidebar variant="rail" active="documents"/>
      <div style={{ flex: 1, overflow: 'auto', display: 'flex', flexDirection: 'column' }}>
        {/* HERO */}
        <div style={{
          padding: '40px 48px 32px', background: 'var(--brand-deep)', color: '#f9f0f0', position: 'relative', overflow: 'hidden',
        }}>
          <div style={{
            position: 'absolute', inset: 0,
            backgroundImage: 'radial-gradient(circle at 20% 0%, rgba(200,54,74,0.25), transparent 60%), linear-gradient(180deg, transparent 0%, rgba(0,0,0,0.18) 100%)',
            pointerEvents: 'none',
          }}/>
          <div style={{ position: 'relative', display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: 24 }}>
            <button className="btn btn-ghost btn-sm" style={{ color: '#e8d6d6', borderColor: 'rgba(255,255,255,0.16)' }}><Icon name="chevron-left" size={13}/>Voltar</button>
            <ISOSeal hash={DOC.hash} hashFull={DOC.hashFull}/>
          </div>
          <div style={{ position: 'relative', display: 'grid', gridTemplateColumns: '1fr 320px', gap: 40, alignItems: 'flex-end' }}>
            <div>
              <div className="kicker mono" style={{ color: '#c8364a', marginBottom: 12, letterSpacing: '0.12em' }}>{DOC.area.toUpperCase()} · {DOC.code}</div>
              <h1 style={{ fontSize: 44, lineHeight: 1.05, fontWeight: 600, letterSpacing: '-0.03em', margin: '0 0 16px', color: '#fff', maxWidth: 760 }}>{DOC.title}</h1>
              <div style={{ display: 'flex', gap: 24, alignItems: 'center', fontSize: 13, color: 'rgba(255,255,255,0.75)' }}>
                <span><span className="mono" style={{ color: '#fff', fontWeight: 600 }}>{DOC.version}</span> publicada {DOC.publishedAt}</span>
                <span style={{ width: 1, height: 14, background: 'rgba(255,255,255,0.2)' }}/>
                <span>por {DOC.publishedBy}</span>
                <span style={{ width: 1, height: 14, background: 'rgba(255,255,255,0.2)' }}/>
                <span>vigente até 15 mar 2027</span>
              </div>
            </div>
            <div style={{ display: 'flex', flexDirection: 'column', gap: 8, justifySelf: 'end' }}>
              <button className="btn btn-lg" style={{ background: '#fff', color: 'var(--brand-deep)', border: 'none', minWidth: 220, justifyContent: 'center' }}><Icon name="download" size={15}/>Baixar PDF assinado</button>
              <div style={{ display: 'flex', gap: 6 }}>
                <button className="btn btn-ghost" style={{ flex: 1, color: '#e8d6d6', borderColor: 'rgba(255,255,255,0.16)', justifyContent: 'center' }}><Icon name="link" size={13}/>Link</button>
                <button className="btn btn-ghost" style={{ flex: 1, color: '#e8d6d6', borderColor: 'rgba(255,255,255,0.16)', justifyContent: 'center' }}><Icon name="edit" size={13}/>Revisar</button>
              </div>
            </div>
          </div>
        </div>
        {/* Git-style version timeline */}
        <div style={{ padding: '20px 48px', borderBottom: '1px solid var(--border)', background: 'var(--surface)' }}>
          <div className="kicker" style={{ marginBottom: 12 }}>Linha do tempo de versões</div>
          <div style={{ position: 'relative', display: 'flex', alignItems: 'flex-start', justifyContent: 'space-between', paddingTop: 8 }}>
            <div style={{ position: 'absolute', top: 14, left: 8, right: 8, height: 2, background: 'var(--border)' }}/>
            {DOC.versions.slice().reverse().map((v, i) => (
              <div key={i} style={{ position: 'relative', display: 'flex', flexDirection: 'column', alignItems: 'center', minWidth: 90, gap: 6 }}>
                <div style={{ width: 14, height: 14, borderRadius: '50%', background: v.current ? 'var(--brand)' : 'var(--surface)', border: `2px solid ${v.current ? 'var(--brand-deep)' : 'var(--border-strong)'}`, boxShadow: v.current ? '0 0 0 4px var(--brand-pale)' : 'none' }}/>
                <span className="mono" style={{ fontSize: 11.5, fontWeight: v.current ? 600 : 400, color: v.current ? 'var(--brand)' : 'var(--text)' }}>{v.v}</span>
                <span className="tiny" style={{ fontSize: 10 }}>{v.when}</span>
                <span className="tiny" style={{ fontSize: 9.5, textAlign: 'center', maxWidth: 110, lineHeight: 1.3 }}>{v.author}</span>
              </div>
            ))}
          </div>
        </div>
        {/* Body in column */}
        <div style={{ display: 'grid', gridTemplateColumns: '1fr 320px', gap: 40, padding: '32px 48px 48px' }}>
          <div style={{ maxWidth: 720 }}>
            <DocBody/>
          </div>
          <aside style={{ position: 'sticky', top: 0, alignSelf: 'flex-start', display: 'flex', flexDirection: 'column', gap: 22 }}>
            <div className="card" style={{ padding: 16 }}>
              <div className="kicker" style={{ marginBottom: 10 }}>Aprovações</div>
              {DOC.signoffs.map((s, i) => (
                <div key={i} style={{ display: 'flex', alignItems: 'center', gap: 10, padding: '8px 0', borderBottom: i < DOC.signoffs.length - 1 ? '1px solid var(--border)' : 'none' }}>
                  <Avatar name={s.name} size={26}/>
                  <div style={{ flex: 1, minWidth: 0, fontSize: 12 }}>
                    <div style={{ fontWeight: 500 }}>{s.name}</div>
                    <div className="tiny" style={{ fontSize: 10.5 }}>{s.role} · {s.when}</div>
                  </div>
                  <Icon name="check" size={14}/>
                </div>
              ))}
            </div>
          </aside>
        </div>
      </div>
    </div>
  );
}

// ============================================================
// VARIATION A — Fanout COVERAGE DASHBOARD
// ============================================================
function FanoutA() {
  const [ready, setReady] = useState(false);
  useEffect(() => { const t = setTimeout(() => setReady(true), 300); return () => clearTimeout(t); }, []);
  const pct = Math.round((FANOUT.acknowledged / FANOUT.totalTargets) * 100);
  const animatedPct = useCounter(ready ? pct : 0, 1600);

  return (
    <div className="palette-wine" style={{ height: 980, display: 'flex', background: 'var(--bg)' }}>
      <Sidebar variant="rail" active="documents"/>
      <div style={{ flex: 1, overflow: 'auto' }}>
        <Toolbar
          breadcrumb={['SSMA', DOC.code, 'Distribuição']}
          title="Cobertura de leitura"
          actions={[
            { label: 'Lembrete em massa', icon: 'mail' },
            { label: 'Exportar', icon: 'download' },
          ]}
        />
        <div style={{ padding: '0 32px 40px', display: 'flex', flexDirection: 'column', gap: 24 }}>
          {/* Top metrics row */}
          <div style={{ display: 'grid', gridTemplateColumns: '1.6fr 1fr 1fr 1fr', gap: 18 }}>
            {/* Big donut */}
            <div className="card" style={{ padding: 24, display: 'flex', alignItems: 'center', gap: 28 }}>
              <div style={{ position: 'relative', width: 152, height: 152 }}>
                <svg viewBox="0 0 100 100" style={{ width: '100%', height: '100%', transform: 'rotate(-90deg)' }}>
                  <circle cx="50" cy="50" r="42" fill="none" stroke="var(--surface-3)" strokeWidth="8"/>
                  <circle cx="50" cy="50" r="42" fill="none" stroke="var(--brand)" strokeWidth="8" strokeLinecap="round"
                    strokeDasharray={2 * Math.PI * 42}
                    strokeDashoffset={2 * Math.PI * 42 * (1 - (ready ? pct / 100 : 0))}
                    style={{ transition: 'stroke-dashoffset 1.6s cubic-bezier(.2,.8,.2,1)' }}/>
                </svg>
                <div style={{ position: 'absolute', inset: 0, display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center' }}>
                  <div className="num" style={{ fontSize: 38, lineHeight: 1, fontWeight: 600, fontFamily: 'var(--font-display)', letterSpacing: '-0.02em' }}>{animatedPct}<span style={{ fontSize: 18, color: 'var(--text-muted)' }}>%</span></div>
                  <div className="kicker" style={{ marginTop: 6 }}>Aprovaram leitura</div>
                </div>
              </div>
              <div style={{ flex: 1 }}>
                <div className="kicker" style={{ marginBottom: 4 }}>Cobertura geral</div>
                <div style={{ fontSize: 15, lineHeight: 1.45, color: 'var(--text-soft)' }}>
                  <strong style={{ color: 'var(--text)' }}>{FANOUT.acknowledged} de {FANOUT.totalTargets}</strong> destinatários reconheceram a leitura. <span style={{ color: 'var(--warning)', fontWeight: 500 }}>{FANOUT.overdue} em atraso.</span>
                </div>
                <div style={{ marginTop: 14, display: 'flex', gap: 20, fontSize: 12 }}>
                  <span style={{ display: 'inline-flex', alignItems: 'center', gap: 6 }}><span style={{ width: 8, height: 8, borderRadius: 2, background: 'var(--brand)' }}/>Ack</span>
                  <span style={{ display: 'inline-flex', alignItems: 'center', gap: 6 }}><span style={{ width: 8, height: 8, borderRadius: 2, background: 'var(--brand-soft)' }}/>Apenas leu</span>
                  <span style={{ display: 'inline-flex', alignItems: 'center', gap: 6 }}><span style={{ width: 8, height: 8, borderRadius: 2, background: 'var(--surface-3)', border: '1px solid var(--border-strong)' }}/>Pendente</span>
                </div>
              </div>
            </div>
            {/* 3 metric cards */}
            {[
              { kicker: 'Total', value: FANOUT.totalTargets },
              { kicker: 'Pendentes', value: FANOUT.pending, accent: true },
              { kicker: 'Em atraso', value: FANOUT.overdue, danger: true },
            ].map((m, i) => (
              <div key={i} className="card" style={{ padding: 20, display: 'flex', flexDirection: 'column', justifyContent: 'space-between' }}>
                <div className="kicker">{m.kicker}</div>
                <SkeletonNumber ready={ready} value={m.value} size={42}/>
                <div className="caption" style={{ fontSize: 11.5, color: m.danger ? 'var(--danger)' : m.accent ? 'var(--warning)' : 'var(--text-muted)' }}>
                  {i === 0 && `Vence ${FANOUT.deadline} · ${FANOUT.daysLeft}d`}
                  {i === 1 && '6 áreas com pendência'}
                  {i === 2 && '+2 desde ontem'}
                </div>
              </div>
            ))}
          </div>

          {/* Coverage by area + recent feed */}
          <div style={{ display: 'grid', gridTemplateColumns: '1.6fr 1fr', gap: 18 }}>
            <div className="card" style={{ padding: 24 }}>
              <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 16 }}>
                <div>
                  <div className="h3">Cobertura por área</div>
                  <div className="caption">Ordenado por pendência</div>
                </div>
                <button className="btn btn-sm btn-ghost"><Icon name="filter" size={12}/>Filtrar</button>
              </div>
              <div style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>
                {FANOUT.byArea.slice().sort((a, b) => a.color - b.color).map((row, i) => {
                  const pctRow = ready ? Math.round(row.color * 100) : 0;
                  const tone = row.color > 0.85 ? 'var(--success)' : row.color > 0.6 ? 'var(--brand-soft)' : row.color > 0.4 ? 'var(--warning)' : 'var(--danger)';
                  return (
                    <div key={i} style={{ display: 'grid', gridTemplateColumns: '180px 1fr 90px', gap: 14, alignItems: 'center' }}>
                      <div>
                        <div style={{ fontSize: 13, fontWeight: 500 }}>{row.area}</div>
                        <div className="tiny" style={{ fontSize: 10.5 }}>{row.ack}/{row.total}</div>
                      </div>
                      <div style={{ position: 'relative', height: 8, background: 'var(--surface-3)', borderRadius: 999 }}>
                        <div style={{
                          position: 'absolute', left: 0, top: 0, bottom: 0, width: `${pctRow}%`, background: tone, borderRadius: 999,
                          transition: 'width 1.6s cubic-bezier(.2,.8,.2,1)',
                        }}/>
                      </div>
                      <div className="num mono" style={{ fontSize: 12, textAlign: 'right', color: tone, fontWeight: 600 }}>{pctRow}%</div>
                    </div>
                  );
                })}
              </div>
            </div>
            <div className="card" style={{ padding: 24, display: 'flex', flexDirection: 'column' }}>
              <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 14 }}>
                <div>
                  <div className="h3">Atividade ao vivo</div>
                  <div className="caption">Linha do tempo de leituras</div>
                </div>
                <span style={{ display: 'inline-flex', alignItems: 'center', gap: 5, fontSize: 10.5, color: 'var(--success)' }}>
                  <span style={{ width: 6, height: 6, borderRadius: '50%', background: 'var(--success)', animation: 'sealPulse 1.6s infinite' }}/>
                  AO VIVO
                </span>
              </div>
              <div style={{ display: 'flex', flexDirection: 'column' }}>
                {FANOUT.recentReads.map((r, i) => (
                  <div key={i} style={{ display: 'flex', gap: 10, padding: '10px 0', borderBottom: i < FANOUT.recentReads.length - 1 ? '1px solid var(--border)' : 'none' }}>
                    <Avatar name={r.who} size={28}/>
                    <div style={{ flex: 1, minWidth: 0 }}>
                      <div style={{ fontSize: 12.5, fontWeight: 500 }}>{r.who}</div>
                      <div className="tiny" style={{ fontSize: 10.5 }}>{r.area} · {r.when}</div>
                    </div>
                    <span className={`pill ${r.status === 'ack' ? 'pill-frozen' : 'pill-approved'}`} style={{ alignSelf: 'center' }}>
                      <span className="dot"/>{r.status === 'ack' ? 'ack' : 'leu'}
                    </span>
                  </div>
                ))}
              </div>
            </div>
          </div>
        </div>
      </div>
      <style>{`@keyframes sk { 0% { background-position: 200% 0; } 100% { background-position: -200% 0; } }`}</style>
    </div>
  );
}

// ============================================================
// VARIATION B — Fanout MAILCHIMP-STYLE LIST
// ============================================================
function FanoutB() {
  const [tab, setTab] = useState('pending');
  return (
    <div className="palette-wine" style={{ height: 980, display: 'flex', background: 'var(--bg)' }}>
      <Sidebar variant="rail" active="documents"/>
      <div style={{ flex: 1, overflow: 'auto' }}>
        <Toolbar
          breadcrumb={['SSMA', DOC.code, 'Distribuição']}
          title="Destinatários"
          actions={[
            { label: 'Lembrete em massa', icon: 'mail', primary: true },
            { label: 'Exportar', icon: 'download' },
          ]}
        />
        <div style={{ padding: '0 32px 40px', display: 'flex', flexDirection: 'column', gap: 18 }}>
          {/* Stat strip */}
          <div className="card" style={{ padding: '18px 24px', display: 'grid', gridTemplateColumns: 'repeat(5, 1fr)', gap: 32 }}>
            {[
              { k: 'Cobertura', v: '63%', sub: `${FANOUT.acknowledged}/${FANOUT.totalTargets} ack`, tone: 'var(--brand)' },
              { k: 'Leram', v: FANOUT.read, sub: `${Math.round(FANOUT.read/FANOUT.totalTargets*100)}% do total`, tone: 'var(--text)' },
              { k: 'Reconheceram', v: FANOUT.acknowledged, sub: 'aceitaram leitura', tone: 'var(--success)' },
              { k: 'Pendentes', v: FANOUT.pending, sub: '6 áreas afetadas', tone: 'var(--warning)' },
              { k: 'Em atraso', v: FANOUT.overdue, sub: '+2 desde ontem', tone: 'var(--danger)' },
            ].map((m, i) => (
              <div key={i}>
                <div className="kicker">{m.k}</div>
                <div className="num" style={{ fontSize: 28, fontWeight: 600, fontFamily: 'var(--font-display)', letterSpacing: '-0.02em', color: m.tone, lineHeight: 1.1, marginTop: 4 }}>{m.v}</div>
                <div className="tiny" style={{ fontSize: 10.5 }}>{m.sub}</div>
              </div>
            ))}
          </div>
          {/* Tabs + filters row */}
          <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
            <div style={{ display: 'inline-flex', background: 'var(--surface)', border: '1px solid var(--border)', borderRadius: 6, padding: 3 }}>
              {[
                { id: 'pending', label: 'Pendentes', n: FANOUT.pending },
                { id: 'overdue', label: 'Em atraso', n: FANOUT.overdue },
                { id: 'acked', label: 'Concluído', n: FANOUT.acknowledged },
                { id: 'all', label: 'Todos', n: FANOUT.totalTargets },
              ].map(t => (
                <button key={t.id} onClick={() => setTab(t.id)} style={{
                  height: 28, padding: '0 12px', border: 'none', cursor: 'pointer',
                  background: tab === t.id ? 'var(--surface-3)' : 'transparent',
                  borderRadius: 4, fontSize: 12, fontWeight: tab === t.id ? 500 : 400,
                  color: tab === t.id ? 'var(--text)' : 'var(--text-muted)',
                  display: 'inline-flex', alignItems: 'center', gap: 6,
                }}>
                  {t.label}
                  <span className="mono tiny" style={{ fontSize: 10, padding: '1px 5px', background: tab === t.id ? 'var(--brand-pale)' : 'var(--surface-3)', color: tab === t.id ? 'var(--brand)' : 'var(--text-muted)', borderRadius: 3 }}>{t.n}</span>
                </button>
              ))}
            </div>
            <div className="input" style={{ flex: 1, display: 'inline-flex', alignItems: 'center', gap: 8, color: 'var(--text-muted)' }}>
              <Icon name="search" size={13}/>Filtrar destinatários…
            </div>
            <button className="btn btn-sm"><Icon name="filter" size={12}/>Área</button>
          </div>
          {/* Big list */}
          <div className="card" style={{ overflow: 'hidden' }}>
            <div style={{ display: 'grid', gridTemplateColumns: '32px 1fr 200px 200px 140px 80px', gap: 12, padding: '10px 18px', background: 'var(--surface-2)', borderBottom: '1px solid var(--border)', fontSize: 10.5, color: 'var(--text-muted)', textTransform: 'uppercase', letterSpacing: '0.05em', fontWeight: 500 }}>
              <span><input type="checkbox" style={{ accentColor: 'var(--brand)' }}/></span>
              <span>Destinatário</span>
              <span>Área</span>
              <span>Última atividade</span>
              <span>Status</span>
              <span></span>
            </div>
            {FANOUT.pendingPeople.map((p, i) => (
              <div key={i} style={{
                display: 'grid', gridTemplateColumns: '32px 1fr 200px 200px 140px 80px', gap: 12,
                padding: '12px 18px', borderBottom: i < FANOUT.pendingPeople.length - 1 ? '1px solid var(--border)' : 'none',
                alignItems: 'center', fontSize: 13,
              }}>
                <span><input type="checkbox" style={{ accentColor: 'var(--brand)' }}/></span>
                <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
                  <Avatar name={p.who} size={28}/>
                  <span style={{ fontWeight: 500 }}>{p.who}</span>
                </div>
                <span className="caption" style={{ fontSize: 12 }}>{p.area}</span>
                <span className="caption" style={{ fontSize: 12, color: p.last.includes('nunca') ? 'var(--danger)' : 'var(--text-muted)' }}>{p.last}</span>
                <span className={`pill ${p.overdue ? 'pill-rejected' : 'pill-draft'}`}><span className="dot"/>{p.overdue ? 'em atraso' : 'pendente'}</span>
                <button className="btn btn-sm btn-ghost"><Icon name="mail" size={11}/>Lembrar</button>
              </div>
            ))}
            <div style={{ padding: '10px 18px', background: 'var(--surface-2)', borderTop: '1px solid var(--border)', display: 'flex', justifyContent: 'space-between', alignItems: 'center', fontSize: 12, color: 'var(--text-muted)' }}>
              <span>Mostrando 6 de {FANOUT.pending} pendentes</span>
              <span style={{ display: 'inline-flex', gap: 4 }}>
                <button className="btn btn-sm btn-ghost"><Icon name="chevron-left" size={12}/></button>
                <button className="btn btn-sm btn-ghost"><Icon name="chevron-right" size={12}/></button>
              </span>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

// ============================================================
// VARIATION C — Fanout MAP / BUBBLES
// ============================================================
function FanoutC() {
  const [hover, setHover] = useState(null);
  // Layout bubbles in a treemap-ish grid with sizes ~ total
  const total = FANOUT.byArea.reduce((s, a) => s + a.total, 0);
  const bubbles = FANOUT.byArea.map(a => ({ ...a, w: Math.sqrt(a.total / total) * 100 }));

  return (
    <div className="palette-wine" style={{ height: 980, display: 'flex', background: 'var(--bg)' }}>
      <Sidebar variant="rail" active="documents"/>
      <div style={{ flex: 1, overflow: 'auto' }}>
        <Toolbar
          breadcrumb={['SSMA', DOC.code, 'Distribuição']}
          title="Mapa de cobertura por área"
          actions={[
            { label: 'Lembrete em massa', icon: 'mail' },
            { label: 'Exportar', icon: 'download' },
          ]}
        />
        <div style={{ padding: '0 32px 40px', display: 'grid', gridTemplateColumns: '1fr 320px', gap: 18 }}>
          <div className="card" style={{ padding: 28, position: 'relative', minHeight: 540 }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: 20 }}>
              <div>
                <div className="kicker">Cada bolha = uma área · tamanho = nº de destinatários · cor = % cobertura</div>
              </div>
              <div style={{ display: 'flex', alignItems: 'center', gap: 10, fontSize: 11.5, color: 'var(--text-muted)' }}>
                <span style={{ display: 'inline-flex', alignItems: 'center', gap: 4 }}><span style={{ width: 12, height: 12, borderRadius: '50%', background: 'var(--success)' }}/>≥ 85%</span>
                <span style={{ display: 'inline-flex', alignItems: 'center', gap: 4 }}><span style={{ width: 12, height: 12, borderRadius: '50%', background: 'var(--brand-soft)' }}/>60–85%</span>
                <span style={{ display: 'inline-flex', alignItems: 'center', gap: 4 }}><span style={{ width: 12, height: 12, borderRadius: '50%', background: 'var(--warning)' }}/>40–60%</span>
                <span style={{ display: 'inline-flex', alignItems: 'center', gap: 4 }}><span style={{ width: 12, height: 12, borderRadius: '50%', background: 'var(--danger)' }}/>&lt; 40%</span>
              </div>
            </div>
            <div style={{ display: 'flex', flexWrap: 'wrap', alignItems: 'center', justifyContent: 'center', gap: 18, padding: '20px 0', minHeight: 460 }}>
              {bubbles.map((b, i) => {
                const size = 80 + b.w * 4;
                const tone = b.color > 0.85 ? 'var(--success)' : b.color > 0.6 ? 'var(--brand-soft)' : b.color > 0.4 ? 'var(--warning)' : 'var(--danger)';
                const isHovered = hover === i;
                return (
                  <div key={i}
                    onMouseEnter={() => setHover(i)}
                    onMouseLeave={() => setHover(null)}
                    style={{
                      width: size, height: size, borderRadius: '50%',
                      background: tone, color: '#fff',
                      display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center',
                      boxShadow: isHovered ? `0 0 0 6px ${tone}22, 0 8px 24px rgba(0,0,0,0.18)` : '0 4px 12px rgba(0,0,0,0.12)',
                      transition: 'all .25s cubic-bezier(.2,.8,.2,1)',
                      cursor: 'pointer', textAlign: 'center', padding: 8,
                    }}>
                    <div className="num" style={{ fontSize: size > 130 ? 28 : 22, fontWeight: 600, lineHeight: 1, fontFamily: 'var(--font-display)' }}>{Math.round(b.color * 100)}%</div>
                    <div style={{ fontSize: size > 130 ? 11.5 : 10, marginTop: 4, opacity: 0.92, lineHeight: 1.2, fontWeight: 500, maxWidth: '90%' }}>{b.area}</div>
                    <div style={{ fontSize: 9, marginTop: 2, opacity: 0.7, fontFamily: 'var(--font-mono)' }}>{b.ack}/{b.total}</div>
                  </div>
                );
              })}
            </div>
          </div>
          <aside style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
            <div className="card" style={{ padding: 18 }}>
              <div className="kicker" style={{ marginBottom: 8 }}>Resumo</div>
              <div className="num" style={{ fontSize: 44, fontWeight: 600, fontFamily: 'var(--font-display)', letterSpacing: '-0.025em', lineHeight: 1, color: 'var(--brand)' }}>63%</div>
              <div className="caption" style={{ marginTop: 4 }}>{FANOUT.acknowledged}/{FANOUT.totalTargets} reconheceram</div>
              <div style={{ marginTop: 14, padding: '10px 12px', background: 'var(--warning-bg)', color: 'var(--warning)', borderRadius: 6, fontSize: 11.5, lineHeight: 1.4 }}>
                <strong>Atenção:</strong> Administrativo e Logística com cobertura abaixo de 50%. Recomenda-se lembrete focado.
              </div>
            </div>
            <div className="card" style={{ padding: 18 }}>
              <div className="kicker" style={{ marginBottom: 10 }}>Áreas críticas</div>
              {FANOUT.byArea.filter(a => a.color < 0.5).sort((a,b)=>a.color-b.color).map((a, i) => (
                <div key={i} style={{ display: 'flex', justifyContent: 'space-between', padding: '8px 0', borderBottom: '1px solid var(--border)', fontSize: 12.5 }}>
                  <span style={{ fontWeight: 500 }}>{a.area}</span>
                  <span className="mono" style={{ color: 'var(--danger)' }}>{Math.round(a.color*100)}%</span>
                </div>
              ))}
              <button className="btn btn-sm btn-primary" style={{ width: '100%', justifyContent: 'center', marginTop: 12 }}><Icon name="mail" size={12}/>Lembrar áreas críticas</button>
            </div>
          </aside>
        </div>
      </div>
    </div>
  );
}

// Export
window.MD_ONDA1 = { PublicadoA, PublicadoB, PublicadoC, FanoutA, FanoutB, FanoutC };
