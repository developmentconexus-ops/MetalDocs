/* global React, MD */
const { Icon: I3, Logo: L3, Avatar: A3, Sidebar: S3, Toolbar: T3 } = MD;
const { useState: us3, useEffect: ue3, useMemo: um3 } = React;

// =====================================================================
// PUBLICADO v3 — The page IS the document.
// Reference: Notion file page · Google Docs view · GitHub release page
// Doc body is the hero. Metadata is companion. Version history is demoted.
// =====================================================================

const DOC3 = {
  code: 'PR-EHS-014',
  title: 'Procedimento de Bloqueio e Etiquetagem (LOTO)',
  subtitle: 'Procedimento operacional de segurança para isolamento de fontes de energia perigosa antes de manutenção em equipamentos industriais.',
  version: 'v3.2',
  publishedAt: '12 mar 2026',
  publishedAtFull: '12 mar 2026 · 14:32',
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
  type: 'Procedimento operacional',
  reads: { total: 142, target: 168 },
  signoffs: [
    { name: 'Renata Souza', role: 'Coord. SSMA', when: '11 mar' },
    { name: 'Marcos Lima', role: 'Gerente Industrial', when: '12 mar' },
    { name: 'Diretoria',   role: 'Anuente', when: '12 mar' },
  ],
  versions: [
    { v: 'v3.2', when: '12 mar 2026', author: 'C. Mendes', current: true },
    { v: 'v3.1', when: '08 jan 2026', author: 'C. Mendes' },
    { v: 'v3.0', when: '14 nov 2025', author: 'R. Souza' },
    { v: 'v2.4', when: '02 jul 2025', author: 'R. Souza' },
  ],
  tags: ['LOTO', 'Bloqueio', 'Manutenção', 'Energia perigosa', 'NR-12'],
  related: [
    { code: 'IT-EHS-021', title: 'Inspeção de cadeados e travas',  type: 'Instrução' },
    { code: 'FR-EHS-008', title: 'Ficha de bloqueio individual',   type: 'Formulário' },
    { code: 'PR-MAN-103', title: 'Manutenção preventiva CCM-04',   type: 'Procedimento' },
  ],
};

const PERMS3 = {
  full:    { view: true, download: true, revise: true, share: true, label: 'Editor · acesso total' },
  reader:  { view: true, download: false, revise: false, share: true, label: 'Leitor · sem download' },
  visitor: { view: true, download: false, revise: false, share: false, label: 'Visitante · só visualiza' },
};

// ---------- ISO seal compact ----------
function ISOSeal3({ hash, hashFull }) {
  const [open, setOpen] = us3(false);
  return (
    <span style={{ position: 'relative', display: 'inline-flex' }}
      onMouseEnter={() => setOpen(true)} onMouseLeave={() => setOpen(false)}>
      <span style={{
        display: 'inline-flex', alignItems: 'center', gap: 6,
        height: 22, padding: '0 9px',
        fontFamily: 'var(--font-mono)', fontSize: 10.5, letterSpacing: '0.04em',
        background: 'var(--success-bg)', color: 'var(--success)',
        border: '1px solid color-mix(in oklab, var(--success) 30%, transparent)',
        borderRadius: 'var(--r-pill)', cursor: 'help',
      }}>
        <span style={{ width: 6, height: 6, borderRadius: '50%', background: 'currentColor',
          animation: 'sealPulse3 2.4s ease-in-out infinite' }}/>
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
      <style>{`@keyframes sealPulse3 { 0%,100% { opacity: 1; transform: scale(1);} 50% { opacity: .55; transform: scale(.85);} }`}</style>
    </span>
  );
}

// ---------- Permission switcher (compact, light) ----------
function PermSwitcher3({ perm, setPerm }) {
  return (
    <div style={{
      display: 'inline-flex', alignItems: 'center', gap: 4,
      padding: '3px 6px 3px 10px',
      background: 'var(--surface)', border: '1px solid var(--border)',
      borderRadius: 'var(--r-pill)', fontSize: 10.5,
      color: 'var(--text-muted)', fontFamily: 'var(--font-mono)', letterSpacing: '0.04em',
    }}>
      <span>VOCÊ:</span>
      {Object.keys(PERMS3).map(k => (
        <button key={k} onClick={() => setPerm(k)} style={{
          height: 20, padding: '0 8px', border: 'none', cursor: 'pointer',
          background: perm === k ? 'var(--brand)' : 'transparent',
          color: perm === k ? '#fff' : 'inherit',
          borderRadius: 'var(--r-pill)', fontSize: 10.5,
          fontFamily: 'inherit', letterSpacing: 'inherit',
          textTransform: 'uppercase', fontWeight: perm === k ? 600 : 400,
        }}>{k}</button>
      ))}
    </div>
  );
}

// ---------- Document preview (the page IS the document) ----------
// A realistic-looking rendered first page. Editorial placeholder typography.
function DocPaper({ pageIdx = 0, total = 14 }) {
  return (
    <div style={{
      width: 720, minHeight: 1020,
      background: '#fff', boxShadow: '0 24px 60px rgba(74,33,33,0.18), 0 4px 12px rgba(74,33,33,0.08)',
      border: '1px solid rgba(0,0,0,0.06)',
      padding: '64px 72px',
      fontFamily: 'var(--font-sans)', color: '#1a0e0e',
      position: 'relative',
    }}>
      {/* Header strip */}
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center',
        paddingBottom: 14, borderBottom: '1px solid #e6dcdc', marginBottom: 36,
        fontSize: 9.5, fontFamily: 'var(--font-mono)', color: '#8a7575', letterSpacing: '0.06em', textTransform: 'uppercase' }}>
        <span>{DOC3.code} · {DOC3.version}</span>
        <span>Confidencial · Uso interno</span>
        <span>Pág. {pageIdx + 1} / {total}</span>
      </div>

      {/* Title block */}
      <div style={{ marginBottom: 32 }}>
        <div style={{ fontSize: 10, color: '#6b1f2a', letterSpacing: '0.10em', textTransform: 'uppercase',
          fontFamily: 'var(--font-mono)', fontWeight: 500, marginBottom: 10 }}>
          Procedimento Operacional · SSMA
        </div>
        <h1 style={{ fontSize: 28, lineHeight: 1.15, letterSpacing: '-0.02em',
          margin: 0, color: '#1a0e0e', fontWeight: 600 }}>
          Procedimento de Bloqueio e<br/>Etiquetagem (LOTO)
        </h1>
        <div style={{ marginTop: 14, fontSize: 12, color: '#4a3434', lineHeight: 1.6, maxWidth: 560 }}>
          Define os critérios técnicos e operacionais para isolamento controlado de fontes
          de energia perigosa antes de qualquer intervenção em equipamentos industriais,
          visando à proteção de pessoas, instalações e continuidade operacional.
        </div>
      </div>

      {/* Section 1 */}
      <div style={{ marginBottom: 28 }}>
        <h2 style={{ fontSize: 13, fontFamily: 'var(--font-mono)', letterSpacing: '0.04em',
          color: '#6b1f2a', textTransform: 'uppercase', margin: '0 0 12px', fontWeight: 600 }}>
          1. Objetivo
        </h2>
        <p style={{ fontSize: 12, lineHeight: 1.7, color: '#1a0e0e', margin: '0 0 8px' }}>
          Estabelecer procedimentos seguros para o controle de energias perigosas
          (elétrica, mecânica, hidráulica, pneumática, química e térmica) durante
          atividades de manutenção, reparo, ajuste ou inspeção em máquinas e
          equipamentos, prevenindo a sua liberação inesperada.
        </p>
      </div>

      {/* Section 2 */}
      <div style={{ marginBottom: 28 }}>
        <h2 style={{ fontSize: 13, fontFamily: 'var(--font-mono)', letterSpacing: '0.04em',
          color: '#6b1f2a', textTransform: 'uppercase', margin: '0 0 12px', fontWeight: 600 }}>
          2. Aplicação
        </h2>
        <p style={{ fontSize: 12, lineHeight: 1.7, color: '#1a0e0e', margin: '0 0 8px' }}>
          Aplica-se a todos os colaboradores próprios e terceiros que executem
          atividades de manutenção em equipamentos industriais nas plantas
          Pomerode, Joinville e Manaus. É de cumprimento obrigatório.
        </p>
      </div>

      {/* Section 3 with table preview */}
      <div style={{ marginBottom: 28 }}>
        <h2 style={{ fontSize: 13, fontFamily: 'var(--font-mono)', letterSpacing: '0.04em',
          color: '#6b1f2a', textTransform: 'uppercase', margin: '0 0 12px', fontWeight: 600 }}>
          3. Definições
        </h2>
        <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: 11.5 }}>
          <tbody>
            {[
              ['LOTO', 'Lockout/Tagout. Procedimento de bloqueio e etiquetagem.'],
              ['Cadeado individual', 'Dispositivo de bloqueio com chave única, sob posse exclusiva do executante.'],
              ['Etiqueta', 'Identificação visual com nome, data e motivo do bloqueio.'],
              ['CCM', 'Centro de Controle de Motores. Painel de partida e proteção.'],
            ].map(([term, def]) => (
              <tr key={term} style={{ borderBottom: '1px solid #f0e9e9' }}>
                <td style={{ padding: '8px 0', verticalAlign: 'top', width: 140,
                  fontFamily: 'var(--font-mono)', fontWeight: 600, color: '#6b1f2a' }}>{term}</td>
                <td style={{ padding: '8px 0', verticalAlign: 'top', color: '#1a0e0e', lineHeight: 1.6 }}>{def}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {/* Footer of doc page */}
      <div style={{ position: 'absolute', left: 72, right: 72, bottom: 36,
        display: 'flex', justifyContent: 'space-between', fontSize: 9, color: '#b3a0a0',
        fontFamily: 'var(--font-mono)', letterSpacing: '0.04em', textTransform: 'uppercase',
        paddingTop: 10, borderTop: '1px solid #f0e9e9' }}>
        <span>{DOC3.code} · {DOC3.version} · publicado 12 mar 2026</span>
        <span style={{ display: 'inline-flex', alignItems: 'center', gap: 5 }}>
          <span style={{ width: 4, height: 4, background: 'var(--success)', borderRadius: '50%' }}/>
          Vigente
        </span>
      </div>
    </div>
  );
}

// ---------- Doc preview controls (sticky toolbar above paper) ----------
function PaperToolbar({ page, total, onPage, zoom, setZoom, perm }) {
  const p = PERMS3[perm];
  return (
    <div style={{
      position: 'sticky', top: 0, zIndex: 10,
      display: 'flex', alignItems: 'center', justifyContent: 'space-between',
      padding: '10px 16px', background: 'var(--surface)',
      borderBottom: '1px solid var(--border)',
      fontSize: 12,
    }}>
      <div style={{ display: 'flex', alignItems: 'center', gap: 14 }}>
        <button className="btn btn-sm btn-ghost" onClick={() => onPage(Math.max(0, page - 1))} disabled={page === 0}>
          <I3 name="chevron-left" size={13}/>
        </button>
        <span className="mono" style={{ fontSize: 11, color: 'var(--text-muted)' }}>
          Pág. <span style={{ color: 'var(--text)', fontWeight: 600 }}>{page + 1}</span> / {total}
        </span>
        <button className="btn btn-sm btn-ghost" onClick={() => onPage(Math.min(total - 1, page + 1))} disabled={page === total - 1}>
          <I3 name="chevron-right" size={13}/>
        </button>
        <span style={{ width: 1, height: 16, background: 'var(--border)' }}/>
        <button className="btn btn-sm btn-ghost" onClick={() => setZoom(Math.max(0.5, zoom - 0.1))}>−</button>
        <span className="mono" style={{ fontSize: 11, minWidth: 38, textAlign: 'center', color: 'var(--text-muted)' }}>
          {Math.round(zoom * 100)}%
        </span>
        <button className="btn btn-sm btn-ghost" onClick={() => setZoom(Math.min(1.4, zoom + 0.1))}>+</button>
        <button className="btn btn-sm btn-ghost" onClick={() => setZoom(0.85)}>Ajustar</button>
      </div>
      <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
        <button className="btn btn-sm btn-ghost"><I3 name="search" size={13}/></button>
        <button className="btn btn-sm btn-ghost"><I3 name="copy" size={13}/></button>
        <button className="btn btn-sm btn-ghost"><I3 name="more" size={13}/></button>
      </div>
    </div>
  );
}

// ---------- Sidebar metadata blocks ----------
function MetaRow({ label, children, mono }) {
  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 3 }}>
      <span className="kicker" style={{ fontSize: 9.5 }}>{label}</span>
      <span style={{ fontSize: 13, color: 'var(--text)',
        fontFamily: mono ? 'var(--font-mono)' : 'inherit' }}>{children}</span>
    </div>
  );
}

// =====================================================================
// MAIN: PublicadoV3
// =====================================================================
function PublicadoV3() {
  const [perm, setPerm] = us3('full');
  const [page, setPage] = us3(0);
  const [zoom, setZoom] = us3(0.85);
  const [showVersions, setShowVersions] = us3(false);
  const p = PERMS3[perm];
  const obsolete = false; // could be a tweak

  return (
    <div className="palette-wine" style={{ height: 980, display: 'flex', background: 'var(--bg)' }}>
      <S3 variant="rail" active="documents"/>

      <div style={{ flex: 1, display: 'flex', flexDirection: 'column', overflow: 'hidden' }}>

        {/* SLIM HEADER — title + breadcrumb + primary actions */}
        <div style={{
          padding: '14px 32px 14px 32px',
          background: 'var(--surface)',
          borderBottom: '1px solid var(--border)',
          display: 'grid', gridTemplateColumns: '1fr auto', gap: 24, alignItems: 'center',
          flexShrink: 0,
        }}>
          {/* Left: breadcrumb + title */}
          <div style={{ minWidth: 0 }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 4,
              fontSize: 11, color: 'var(--text-muted)', fontFamily: 'var(--font-mono)',
              letterSpacing: '0.04em', textTransform: 'uppercase' }}>
              <a href="#" style={{ color: 'inherit', textDecoration: 'none' }}>Biblioteca</a>
              <I3 name="chevron-right" size={10}/>
              <a href="#" style={{ color: 'inherit', textDecoration: 'none' }}>{DOC3.areaShort}</a>
              <I3 name="chevron-right" size={10}/>
              <a href="#" style={{ color: 'inherit', textDecoration: 'none' }}>Procedimentos</a>
              <I3 name="chevron-right" size={10}/>
              <span style={{ color: 'var(--text)' }}>{DOC3.code}</span>
            </div>
            <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
              <span style={{
                display: 'inline-flex', alignItems: 'center', gap: 6,
                height: 24, padding: '0 10px', borderRadius: 'var(--r-pill)',
                background: 'var(--brand-pale)', color: 'var(--brand)',
                fontSize: 11, fontWeight: 600, fontFamily: 'var(--font-mono)',
                letterSpacing: '0.04em',
              }}>
                {DOC3.code}
              </span>
              <h1 style={{ fontSize: 18, lineHeight: 1.2, margin: 0,
                fontWeight: 600, color: 'var(--text)', letterSpacing: '-0.01em',
                whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis', maxWidth: 540 }}>
                {DOC3.title}
              </h1>
              <span style={{
                display: 'inline-flex', alignItems: 'center', gap: 5,
                height: 22, padding: '0 8px', fontSize: 11, fontWeight: 500,
                background: 'var(--success-bg)', color: 'var(--success)',
                borderRadius: 'var(--r-pill)', fontFamily: 'var(--font-mono)',
                letterSpacing: '0.04em', textTransform: 'uppercase',
              }}>
                <span style={{ width: 5, height: 5, background: 'currentColor', borderRadius: '50%' }}/>
                {DOC3.version} · vigente
              </span>
            </div>
          </div>

          {/* Right: primary actions */}
          <div style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
            <PermSwitcher3 perm={perm} setPerm={setPerm}/>
            <span style={{ width: 1, height: 24, background: 'var(--border)', margin: '0 6px' }}/>
            <button className="btn btn-sm btn-ghost" disabled={!p.share}>
              <I3 name="link" size={13}/>Copiar link
            </button>
            <button className="btn btn-sm btn-ghost" disabled={!p.revise}>
              <I3 name="edit" size={13}/>Revisar
            </button>
            <button className="btn btn-sm btn-ghost"><I3 name="more" size={13}/></button>
            <button className="btn btn-sm btn-primary" disabled={!p.download} style={{ height: 30 }}>
              <I3 name="download" size={13}/>Baixar PDF
            </button>
          </div>
        </div>

        {/* MAIN BODY: document preview + metadata sidebar */}
        <div style={{ flex: 1, display: 'grid', gridTemplateColumns: '1fr 320px',
          overflow: 'hidden', minHeight: 0 }}>

          {/* DOCUMENT VIEWER (the protagonist) */}
          <div style={{ display: 'flex', flexDirection: 'column', overflow: 'hidden',
            background: 'var(--surface-3)', borderRight: '1px solid var(--border)' }}>

            <PaperToolbar page={page} total={DOC3.pageCount} onPage={setPage}
              zoom={zoom} setZoom={setZoom} perm={perm}/>

            {/* Paper area (scrollable) */}
            <div style={{ flex: 1, overflow: 'auto', padding: '32px 0',
              display: 'flex', justifyContent: 'center', alignItems: 'flex-start' }}>
              <div style={{ transform: `scale(${zoom})`, transformOrigin: 'top center',
                transition: 'transform 200ms ease' }}>
                <DocPaper pageIdx={page} total={DOC3.pageCount}/>
              </div>
            </div>
          </div>

          {/* METADATA RAIL (right) */}
          <aside style={{ overflow: 'auto', background: 'var(--surface)',
            display: 'flex', flexDirection: 'column' }}>

            {/* Identity */}
            <section style={{ padding: '20px 22px', borderBottom: '1px solid var(--border)' }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 14 }}>
                <div style={{ fontSize: 11, fontFamily: 'var(--font-mono)',
                  letterSpacing: '0.06em', textTransform: 'uppercase',
                  color: 'var(--text-muted)' }}>Identificação</div>
                <span style={{ flex: 1, height: 1, background: 'var(--border)' }}/>
              </div>

              <div style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>
                <MetaRow label="Tipo">{DOC3.type}</MetaRow>
                <MetaRow label="Área responsável">{DOC3.area}</MetaRow>
                <MetaRow label="Código" mono>{DOC3.code}</MetaRow>
                <MetaRow label="Páginas">{DOC3.pageCount} · {DOC3.size}</MetaRow>
              </div>
            </section>

            {/* Vigência & ciclo */}
            <section style={{ padding: '20px 22px', borderBottom: '1px solid var(--border)' }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 14 }}>
                <div style={{ fontSize: 11, fontFamily: 'var(--font-mono)',
                  letterSpacing: '0.06em', textTransform: 'uppercase',
                  color: 'var(--text-muted)' }}>Vigência</div>
                <span style={{ flex: 1, height: 1, background: 'var(--border)' }}/>
              </div>
              <div style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>
                <MetaRow label="Versão atual" mono>{DOC3.version}</MetaRow>
                <MetaRow label="Publicado em">{DOC3.publishedAtFull}</MetaRow>
                <MetaRow label="Vigente desde">{DOC3.effectiveAt}</MetaRow>
                <MetaRow label="Próxima revisão">{DOC3.reviewIn}</MetaRow>
              </div>
            </section>

            {/* Owner & sign-offs */}
            <section style={{ padding: '20px 22px', borderBottom: '1px solid var(--border)' }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 14 }}>
                <div style={{ fontSize: 11, fontFamily: 'var(--font-mono)',
                  letterSpacing: '0.06em', textTransform: 'uppercase',
                  color: 'var(--text-muted)' }}>Responsabilidade</div>
                <span style={{ flex: 1, height: 1, background: 'var(--border)' }}/>
              </div>

              <div style={{ display: 'flex', alignItems: 'center', gap: 10, marginBottom: 14 }}>
                <A3 name={DOC3.publishedBy} size={32}/>
                <div>
                  <div style={{ fontSize: 13, fontWeight: 500, color: 'var(--text)' }}>{DOC3.publishedBy}</div>
                  <div style={{ fontSize: 11, color: 'var(--text-muted)' }}>{DOC3.publishedByRole} · publicou esta versão</div>
                </div>
              </div>

              <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
                {DOC3.signoffs.map(s => (
                  <div key={s.name} style={{ display: 'flex', alignItems: 'center', gap: 10,
                    padding: '8px 10px', background: 'var(--surface-2)',
                    border: '1px solid var(--border)', borderRadius: 6 }}>
                    <I3 name="check" size={12}/>
                    <div style={{ flex: 1, minWidth: 0 }}>
                      <div style={{ fontSize: 12, fontWeight: 500, color: 'var(--text)',
                        whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }}>{s.name}</div>
                      <div style={{ fontSize: 10.5, color: 'var(--text-muted)' }}>{s.role}</div>
                    </div>
                    <span className="mono" style={{ fontSize: 10, color: 'var(--text-muted)' }}>{s.when}</span>
                  </div>
                ))}
              </div>
            </section>

            {/* Tags */}
            <section style={{ padding: '20px 22px', borderBottom: '1px solid var(--border)' }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 12 }}>
                <div style={{ fontSize: 11, fontFamily: 'var(--font-mono)',
                  letterSpacing: '0.06em', textTransform: 'uppercase',
                  color: 'var(--text-muted)' }}>Tags</div>
                <span style={{ flex: 1, height: 1, background: 'var(--border)' }}/>
              </div>
              <div style={{ display: 'flex', flexWrap: 'wrap', gap: 5 }}>
                {DOC3.tags.map(t => (
                  <span key={t} style={{
                    display: 'inline-flex', alignItems: 'center',
                    height: 22, padding: '0 9px', fontSize: 11,
                    background: 'var(--surface-2)', color: 'var(--text-soft)',
                    border: '1px solid var(--border)', borderRadius: 'var(--r-pill)',
                  }}>{t}</span>
                ))}
              </div>
            </section>

            {/* Related docs */}
            <section style={{ padding: '20px 22px', borderBottom: '1px solid var(--border)' }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 12 }}>
                <div style={{ fontSize: 11, fontFamily: 'var(--font-mono)',
                  letterSpacing: '0.06em', textTransform: 'uppercase',
                  color: 'var(--text-muted)' }}>Documentos relacionados</div>
                <span style={{ flex: 1, height: 1, background: 'var(--border)' }}/>
              </div>
              <div style={{ display: 'flex', flexDirection: 'column', gap: 4 }}>
                {DOC3.related.map(r => (
                  <a key={r.code} href="#" style={{
                    display: 'flex', alignItems: 'center', gap: 10,
                    padding: '8px 10px', borderRadius: 6,
                    color: 'var(--text)', textDecoration: 'none',
                    transition: 'background 120ms',
                  }} onMouseEnter={e => e.currentTarget.style.background = 'var(--surface-2)'}
                     onMouseLeave={e => e.currentTarget.style.background = 'transparent'}>
                    <I3 name="document" size={14} style={{ color: 'var(--text-muted)' }}/>
                    <div style={{ flex: 1, minWidth: 0 }}>
                      <div style={{ fontSize: 12, fontWeight: 500, color: 'var(--text)',
                        whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }}>{r.title}</div>
                      <div className="mono" style={{ fontSize: 10, color: 'var(--text-muted)' }}>
                        {r.code} · {r.type}
                      </div>
                    </div>
                    <I3 name="chevron-right" size={11} style={{ color: 'var(--text-faint)' }}/>
                  </a>
                ))}
              </div>
            </section>

            {/* History collapsed by default */}
            <section style={{ padding: '20px 22px' }}>
              <button onClick={() => setShowVersions(v => !v)} style={{
                width: '100%', display: 'flex', alignItems: 'center', gap: 8,
                padding: 0, background: 'transparent', border: 'none',
                cursor: 'pointer', textAlign: 'left',
              }}>
                <div style={{ fontSize: 11, fontFamily: 'var(--font-mono)',
                  letterSpacing: '0.06em', textTransform: 'uppercase',
                  color: 'var(--text-muted)' }}>
                  Histórico de versões ({DOC3.versions.length})
                </div>
                <span style={{ flex: 1, height: 1, background: 'var(--border)' }}/>
                <I3 name={showVersions ? 'chevron-up' : 'chevron-down'} size={12}
                  style={{ color: 'var(--text-muted)' }}/>
              </button>

              {showVersions && (
                <div style={{ marginTop: 14, display: 'flex', flexDirection: 'column', gap: 0 }}>
                  {DOC3.versions.map((v, i) => (
                    <div key={v.v} style={{
                      display: 'flex', alignItems: 'center', gap: 10,
                      padding: '10px 0',
                      borderBottom: i < DOC3.versions.length - 1 ? '1px solid var(--border)' : 'none',
                    }}>
                      <span className="mono" style={{
                        fontSize: 11, fontWeight: v.current ? 600 : 400,
                        color: v.current ? 'var(--brand)' : 'var(--text-muted)',
                        minWidth: 32,
                      }}>{v.v}</span>
                      <div style={{ flex: 1, minWidth: 0 }}>
                        <div style={{ fontSize: 12, color: 'var(--text)' }}>{v.when}</div>
                        <div style={{ fontSize: 10.5, color: 'var(--text-muted)' }}>{v.author}</div>
                      </div>
                      {v.current ? (
                        <span style={{
                          fontSize: 9.5, fontFamily: 'var(--font-mono)',
                          letterSpacing: '0.04em', color: 'var(--success)',
                          textTransform: 'uppercase', fontWeight: 600,
                        }}>Atual</span>
                      ) : (
                        <button className="btn btn-sm btn-ghost" style={{ height: 22, fontSize: 11, padding: '0 8px' }}>
                          ver
                        </button>
                      )}
                    </div>
                  ))}
                </div>
              )}
            </section>

            {/* Compliance footer */}
            <section style={{ padding: '16px 22px', background: 'var(--surface-2)',
              borderTop: '1px solid var(--border)', marginTop: 'auto' }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: 10, marginBottom: 8 }}>
                <ISOSeal3 hash={DOC3.hash} hashFull={DOC3.hashFull}/>
              </div>
              <div style={{ fontSize: 10.5, color: 'var(--text-muted)', lineHeight: 1.5 }}>
                Cadeia de hash íntegra. Documento congelado conforme cl. 7.5.3.
              </div>
            </section>
          </aside>
        </div>
      </div>
    </div>
  );
}

window.MD_ONDA1_V3 = { PublicadoV3 };
