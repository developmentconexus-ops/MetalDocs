/* global React, MD */
const { Icon: I4, Logo: L4, Avatar: A4, Sidebar: S4 } = MD;
const { useState: us4, useEffect: ue4 } = React;

// =====================================================================
// PUBLICADO v4 — A landing page DEDICATED to the document.
// We do NOT render the document body. We render INFORMATION about it.
// Editorial / sober · wine palette · much white · large typography.
// =====================================================================

const D4 = {
  code: 'PR-EHS-014',
  title: 'Procedimento de Bloqueio e Etiquetagem (LOTO)',
  subtitle: 'Critérios técnicos e operacionais para o isolamento controlado de fontes de energia perigosa antes de qualquer intervenção em equipamentos industriais.',
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
  typeShort: 'Procedimento',
  signoffs: [
    { name: 'Renata Souza', role: 'Coord. SSMA', when: '11 mar · 18:04', stage: 'Revisão técnica' },
    { name: 'Marcos Lima', role: 'Gerente Industrial', when: '12 mar · 09:12', stage: 'Aprovação gerencial' },
    { name: 'Rafael Bittencourt', role: 'Diretor de Operações', when: '12 mar · 14:30', stage: 'Anuência de diretoria' },
  ],
  versions: [
    { v: 'v3.2', when: '12 mar 2026', author: 'C. Mendes', current: true, summary: 'Inclusão de procedimento para painéis CCM-04. Atualização da matriz de risco para fontes hidráulicas.' },
    { v: 'v3.1', when: '08 jan 2026', author: 'C. Mendes', summary: 'Correção de typo no item 4.3 (responsabilidades).' },
    { v: 'v3.0', when: '14 nov 2025', author: 'R. Souza', summary: 'Revisão geral pós-auditoria interna. Novas etiquetas padronizadas.' },
    { v: 'v2.4', when: '02 jul 2025', author: 'R. Souza', summary: 'Última versão da revisão anterior.' },
  ],
  related: [
    { code: 'IT-EHS-021', title: 'Inspeção de cadeados e travas',  type: 'Instrução' },
    { code: 'FR-EHS-008', title: 'Ficha de bloqueio individual',   type: 'Formulário' },
    { code: 'PR-MAN-103', title: 'Manutenção preventiva CCM-04',   type: 'Procedimento' },
  ],
  comments: [
    { who: 'Marcos Lima', when: '13 mar', text: 'Equipes de Pomerode já alinhadas. Treinamento iniciado para CCM-04 na próxima segunda.' },
    { who: 'Renata Souza', when: '14 mar', text: 'Adicionei FR-EHS-008 como anexo de referência para os supervisores de turno.' },
  ],
};

// ===== ISO seal (compact tooltip) =====
function ISOSeal4({ hash, hashFull }) {
  const [open, setOpen] = us4(false);
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
          animation: 'sealPulse4 2.4s ease-in-out infinite' }}/>
        ISO 9001 · {hash}
      </span>
      {open && (
        <div style={{
          position: 'absolute', top: 'calc(100% + 6px)', left: 0, zIndex: 50,
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
      <style>{`@keyframes sealPulse4 { 0%,100% { opacity: 1; transform: scale(1);} 50% { opacity: .55; transform: scale(.85);} }`}</style>
    </span>
  );
}

// ===== Symbolic doc card (no real preview) =====
// "Capa de relatório": área color band + giant code + type
function DocCard({ size = 'lg', tilt = false }) {
  const dims = {
    sm: { w: 120, h: 160, codeSize: 14, titleSize: 9, padding: 14 },
    md: { w: 180, h: 240, codeSize: 20, titleSize: 11, padding: 18 },
    lg: { w: 240, h: 320, codeSize: 26, titleSize: 13, padding: 24 },
  }[size];
  const transform = tilt ? 'perspective(1200px) rotateY(-12deg) rotateX(4deg)' : 'none';
  return (
    <div style={{
      width: dims.w, height: dims.h,
      background: '#fff',
      border: '1px solid var(--border)',
      borderRadius: 4,
      boxShadow: tilt
        ? '24px 24px 60px rgba(74,33,33,0.18), 4px 4px 16px rgba(74,33,33,0.10)'
        : '0 12px 32px rgba(74,33,33,0.10), 0 2px 6px rgba(74,33,33,0.06)',
      transform,
      transformStyle: 'preserve-3d',
      display: 'flex', flexDirection: 'column',
      position: 'relative', overflow: 'hidden',
      flexShrink: 0,
    }}>
      {/* Top color band (area) */}
      <div style={{
        height: dims.h * 0.18,
        background: 'linear-gradient(135deg, var(--brand) 0%, var(--brand-deep) 100%)',
        borderBottom: '1px solid var(--brand-deep)',
        display: 'flex', alignItems: 'center', padding: `0 ${dims.padding}px`,
        color: '#fff', fontFamily: 'var(--font-mono)', fontSize: dims.titleSize - 1,
        letterSpacing: '0.10em', textTransform: 'uppercase', fontWeight: 500,
      }}>
        {D4.areaShort}
      </div>
      {/* Body */}
      <div style={{ flex: 1, padding: dims.padding, display: 'flex', flexDirection: 'column' }}>
        <div style={{
          fontFamily: 'var(--font-mono)', fontSize: dims.codeSize,
          color: 'var(--brand-deep)', fontWeight: 600, letterSpacing: '-0.01em',
          marginBottom: 6,
        }}>{D4.code}</div>
        <div style={{
          fontSize: dims.titleSize, color: 'var(--text-soft)',
          letterSpacing: '0.04em', textTransform: 'uppercase', fontWeight: 500,
          fontFamily: 'var(--font-mono)',
        }}>{D4.typeShort}</div>
        <div style={{ flex: 1 }}/>
        <div style={{
          height: 1, background: 'var(--border)', marginBottom: 10,
        }}/>
        <div style={{ fontSize: dims.titleSize - 1, color: 'var(--text-muted)',
          fontFamily: 'var(--font-mono)', letterSpacing: '0.04em' }}>
          {D4.version}
        </div>
      </div>
    </div>
  );
}

// ===== Action buttons (composable) =====
function PrimaryView({ size = 'lg', obsolete }) {
  const h = size === 'lg' ? 44 : 36;
  return (
    <button className={`btn btn-primary ${size === 'lg' ? 'btn-lg' : ''}`}
      disabled={obsolete}
      style={{
        height: h, padding: size === 'lg' ? '0 22px' : '0 16px',
        fontSize: size === 'lg' ? 14 : 13, fontWeight: 600,
        opacity: obsolete ? 0.4 : 1,
      }}>
      <I4 name="eye" size={16}/>Visualizar documento
    </button>
  );
}
function SecondaryActions({ obsolete, compact }) {
  return (
    <>
      <button className={`btn ${compact ? 'btn-sm' : ''}`} disabled={obsolete} style={{ opacity: obsolete ? 0.4 : 1 }}>
        <I4 name="download" size={13}/>Baixar PDF
      </button>
      <button className={`btn ${compact ? 'btn-sm' : ''}`} disabled={obsolete} style={{ opacity: obsolete ? 0.4 : 1 }}>
        <I4 name="edit" size={13}/>Iniciar revisão
      </button>
      <button className={`btn btn-ghost ${compact ? 'btn-sm' : ''}`} disabled={obsolete} style={{ opacity: obsolete ? 0.4 : 1 }}>
        <I4 name="link" size={13}/>Copiar link
      </button>
    </>
  );
}

// ===== Breadcrumb (always present) =====
function Breadcrumb() {
  return (
    <div style={{ display: 'flex', alignItems: 'center', gap: 8,
      fontSize: 11, color: 'var(--text-muted)', fontFamily: 'var(--font-mono)',
      letterSpacing: '0.04em', textTransform: 'uppercase' }}>
      <a href="#" style={{ color: 'inherit', textDecoration: 'none' }}>Biblioteca</a>
      <I4 name="chevron-right" size={10}/>
      <a href="#" style={{ color: 'inherit', textDecoration: 'none' }}>{D4.areaShort}</a>
      <I4 name="chevron-right" size={10}/>
      <a href="#" style={{ color: 'inherit', textDecoration: 'none' }}>Procedimentos</a>
      <I4 name="chevron-right" size={10}/>
      <span style={{ color: 'var(--text)' }}>{D4.code}</span>
    </div>
  );
}

// =====================================================================
// HERO VARIATIONS
// =====================================================================

// HERO B — Pure typography. Title gigante editorial, no thumbnail.
function HeroB({ obsolete }) {
  return (
    <header style={{
      padding: '32px 56px 40px',
      background: 'var(--surface)',
      borderBottom: '1px solid var(--border)',
      position: 'relative',
    }}>
      <Breadcrumb/>

      <div style={{ display: 'flex', alignItems: 'flex-start', gap: 24, marginTop: 18 }}>
        <div style={{ flex: 1, minWidth: 0 }}>
          {/* Tag row */}
          <div style={{ display: 'flex', alignItems: 'center', gap: 10, marginBottom: 14 }}>
            <span style={{
              display: 'inline-flex', alignItems: 'center',
              height: 26, padding: '0 12px', borderRadius: 'var(--r-pill)',
              background: 'var(--brand-pale)', color: 'var(--brand)',
              fontSize: 12, fontWeight: 600, fontFamily: 'var(--font-mono)',
              letterSpacing: '0.04em',
            }}>{D4.code}</span>
            <span style={{ fontSize: 11, color: 'var(--text-muted)',
              fontFamily: 'var(--font-mono)', letterSpacing: '0.06em', textTransform: 'uppercase' }}>
              {D4.typeShort} · {D4.area}
            </span>
            {!obsolete && (
              <span style={{
                display: 'inline-flex', alignItems: 'center', gap: 5,
                height: 22, padding: '0 10px', fontSize: 11, fontWeight: 500,
                background: 'var(--success-bg)', color: 'var(--success)',
                borderRadius: 'var(--r-pill)', fontFamily: 'var(--font-mono)',
                letterSpacing: '0.04em', textTransform: 'uppercase',
              }}>
                <span style={{ width: 5, height: 5, background: 'currentColor', borderRadius: '50%' }}/>
                {D4.version} · vigente
              </span>
            )}
          </div>

          {/* Title */}
          <h1 style={{
            fontSize: 44, lineHeight: 1.05, letterSpacing: '-0.025em',
            margin: '0 0 14px', fontWeight: 600, color: 'var(--text)',
            maxWidth: 820, textWrap: 'balance',
          }}>{D4.title}</h1>

          {/* Subtitle */}
          <p style={{
            fontSize: 16, lineHeight: 1.55, color: 'var(--text-soft)',
            margin: '0 0 24px', maxWidth: 720,
          }}>{D4.subtitle}</p>

          {/* Actions row */}
          <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
            <PrimaryView size="lg" obsolete={obsolete}/>
            <SecondaryActions obsolete={obsolete}/>
          </div>
        </div>
      </div>
    </header>
  );
}

// HERO C — Tilted 3D card + title beside it.
function HeroC({ obsolete }) {
  return (
    <header style={{
      padding: '40px 56px 48px',
      background: 'linear-gradient(180deg, var(--brand-pale) 0%, var(--surface) 100%)',
      borderBottom: '1px solid var(--border)',
      position: 'relative', overflow: 'hidden',
    }}>
      {/* Subtle wine grid texture */}
      <div style={{
        position: 'absolute', inset: 0, opacity: 0.04, pointerEvents: 'none',
        backgroundImage: 'linear-gradient(var(--brand) 1px, transparent 1px), linear-gradient(90deg, var(--brand) 1px, transparent 1px)',
        backgroundSize: '32px 32px',
      }}/>
      <div style={{ position: 'relative', marginBottom: 24 }}>
        <Breadcrumb/>
      </div>

      <div style={{
        position: 'relative', display: 'grid',
        gridTemplateColumns: '300px 1fr',
        gap: 56, alignItems: 'center',
      }}>
        {/* 3D doc card */}
        <div style={{ display: 'flex', justifyContent: 'center' }}>
          <DocCard size="lg" tilt/>
        </div>

        {/* Right: title + actions */}
        <div>
          <div style={{ display: 'flex', alignItems: 'center', gap: 10, marginBottom: 16 }}>
            <span style={{
              display: 'inline-flex', alignItems: 'center',
              height: 26, padding: '0 12px', borderRadius: 'var(--r-pill)',
              background: 'var(--surface)', color: 'var(--brand)',
              border: '1px solid var(--brand-pale)',
              fontSize: 12, fontWeight: 600, fontFamily: 'var(--font-mono)',
              letterSpacing: '0.04em',
            }}>{D4.code}</span>
            {!obsolete && (
              <span style={{
                display: 'inline-flex', alignItems: 'center', gap: 5,
                height: 22, padding: '0 10px', fontSize: 11, fontWeight: 500,
                background: 'var(--success-bg)', color: 'var(--success)',
                borderRadius: 'var(--r-pill)', fontFamily: 'var(--font-mono)',
                letterSpacing: '0.04em', textTransform: 'uppercase',
              }}>
                <span style={{ width: 5, height: 5, background: 'currentColor', borderRadius: '50%' }}/>
                {D4.version} · vigente
              </span>
            )}
          </div>

          <h1 style={{
            fontSize: 38, lineHeight: 1.1, letterSpacing: '-0.025em',
            margin: '0 0 14px', fontWeight: 600, color: 'var(--text)',
            maxWidth: 620, textWrap: 'balance',
          }}>{D4.title}</h1>

          <p style={{
            fontSize: 15, lineHeight: 1.55, color: 'var(--text-soft)',
            margin: '0 0 22px', maxWidth: 560,
          }}>{D4.subtitle}</p>

          <div style={{ display: 'flex', alignItems: 'center', gap: 10, flexWrap: 'wrap' }}>
            <PrimaryView size="lg" obsolete={obsolete}/>
            <SecondaryActions obsolete={obsolete}/>
          </div>
        </div>
      </div>
    </header>
  );
}

// HERO D — Notion-style: medium thumbnail left + title/metadata right.
function HeroD({ obsolete }) {
  return (
    <header style={{
      padding: '32px 56px 32px',
      background: 'var(--surface)',
      borderBottom: '1px solid var(--border)',
    }}>
      <Breadcrumb/>

      <div style={{
        display: 'grid', gridTemplateColumns: '200px 1fr auto',
        gap: 32, alignItems: 'flex-start', marginTop: 22,
      }}>
        {/* Thumbnail */}
        <DocCard size="md"/>

        {/* Center: title block + inline metadata */}
        <div style={{ minWidth: 0, paddingTop: 6 }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 12 }}>
            <span style={{
              display: 'inline-flex', alignItems: 'center',
              height: 22, padding: '0 10px', borderRadius: 'var(--r-pill)',
              background: 'var(--brand-pale)', color: 'var(--brand)',
              fontSize: 11, fontWeight: 600, fontFamily: 'var(--font-mono)',
              letterSpacing: '0.04em',
            }}>{D4.code}</span>
            <span style={{ fontSize: 11, color: 'var(--text-muted)',
              fontFamily: 'var(--font-mono)', letterSpacing: '0.06em', textTransform: 'uppercase' }}>
              {D4.typeShort}
            </span>
            {!obsolete && (
              <span style={{
                display: 'inline-flex', alignItems: 'center', gap: 5,
                height: 22, padding: '0 10px', fontSize: 11, fontWeight: 500,
                background: 'var(--success-bg)', color: 'var(--success)',
                borderRadius: 'var(--r-pill)', fontFamily: 'var(--font-mono)',
                letterSpacing: '0.04em', textTransform: 'uppercase',
              }}>
                <span style={{ width: 5, height: 5, background: 'currentColor', borderRadius: '50%' }}/>
                {D4.version} · vigente
              </span>
            )}
          </div>

          <h1 style={{
            fontSize: 32, lineHeight: 1.1, letterSpacing: '-0.02em',
            margin: '0 0 12px', fontWeight: 600, color: 'var(--text)',
            textWrap: 'balance',
          }}>{D4.title}</h1>

          <p style={{
            fontSize: 14, lineHeight: 1.55, color: 'var(--text-soft)',
            margin: '0 0 16px', maxWidth: 640,
          }}>{D4.subtitle}</p>

          {/* Inline meta strip */}
          <div style={{ display: 'flex', alignItems: 'center', gap: 14, flexWrap: 'wrap',
            fontSize: 12, color: 'var(--text-muted)' }}>
            <span style={{ display: 'inline-flex', alignItems: 'center', gap: 6 }}>
              <A4 name={D4.publishedBy} size={20}/>
              <span style={{ color: 'var(--text)' }}>{D4.publishedBy}</span>
              <span>·</span>
              <span>{D4.publishedByRole}</span>
            </span>
            <span>·</span>
            <span>publicado <span style={{ color: 'var(--text)' }}>{D4.publishedAt}</span></span>
            <span>·</span>
            <span>vigente até <span style={{ color: 'var(--text)' }}>{D4.reviewIn}</span></span>
            <span>·</span>
            <span>{D4.pageCount} páginas</span>
          </div>
        </div>

        {/* Right: action stack */}
        <div style={{ display: 'flex', flexDirection: 'column', gap: 8, minWidth: 220 }}>
          <PrimaryView size="lg" obsolete={obsolete}/>
          <button className="btn" disabled={obsolete} style={{ opacity: obsolete ? 0.4 : 1, justifyContent: 'center' }}>
            <I4 name="download" size={13}/>Baixar PDF
          </button>
          <div style={{ display: 'flex', gap: 6 }}>
            <button className="btn btn-sm" disabled={obsolete} style={{ flex: 1, opacity: obsolete ? 0.4 : 1, justifyContent: 'center' }}>
              <I4 name="edit" size={12}/>Revisar
            </button>
            <button className="btn btn-sm btn-ghost" disabled={obsolete} style={{ flex: 1, opacity: obsolete ? 0.4 : 1, justifyContent: 'center' }}>
              <I4 name="link" size={12}/>Link
            </button>
          </div>
        </div>
      </div>
    </header>
  );
}

// =====================================================================
// SECTION COMPONENTS (shared between A and B page heights)
// =====================================================================

function SectionTitle({ kicker, title, action }) {
  return (
    <div style={{ display: 'flex', alignItems: 'flex-end', justifyContent: 'space-between',
      marginBottom: 14, paddingBottom: 10, borderBottom: '1px solid var(--border)' }}>
      <div>
        <div style={{ fontSize: 10.5, fontFamily: 'var(--font-mono)',
          letterSpacing: '0.10em', textTransform: 'uppercase',
          color: 'var(--text-muted)', marginBottom: 4 }}>{kicker}</div>
        <h2 style={{ fontSize: 18, lineHeight: 1.2, margin: 0,
          fontWeight: 600, letterSpacing: '-0.01em', color: 'var(--text)' }}>{title}</h2>
      </div>
      {action}
    </div>
  );
}

function MetaGrid() {
  const rows = [
    ['Tipo de documento', D4.type],
    ['Área responsável', D4.area],
    ['Código', D4.code, 'mono'],
    ['Versão atual', D4.version, 'mono'],
    ['Publicado em', D4.publishedAtFull],
    ['Vigente desde', D4.effectiveAt],
    ['Próxima revisão', D4.reviewIn],
    ['Páginas · tamanho', `${D4.pageCount} páginas · ${D4.size}`],
  ];
  return (
    <section>
      <SectionTitle kicker="01 · Identificação" title="Metadados estruturados"/>
      <div style={{
        display: 'grid', gridTemplateColumns: 'repeat(2, 1fr)',
        rowGap: 0, columnGap: 36,
      }}>
        {rows.map(([label, value, mono]) => (
          <div key={label} style={{
            display: 'flex', justifyContent: 'space-between', alignItems: 'baseline',
            padding: '13px 0', borderBottom: '1px solid var(--border)',
          }}>
            <span style={{ fontSize: 12, color: 'var(--text-muted)' }}>{label}</span>
            <span style={{ fontSize: 13, color: 'var(--text)',
              fontFamily: mono ? 'var(--font-mono)' : 'inherit',
              fontWeight: mono ? 500 : 400 }}>{value}</span>
          </div>
        ))}
      </div>
    </section>
  );
}

function SignoffsSection() {
  return (
    <section>
      <SectionTitle kicker="02 · Cadeia de aprovação" title="Sign-offs desta versão"/>
      <div style={{ display: 'flex', flexDirection: 'column' }}>
        {D4.signoffs.map((s, i) => (
          <div key={s.name} style={{
            display: 'grid', gridTemplateColumns: '32px 1fr auto',
            gap: 16, alignItems: 'center',
            padding: '16px 0',
            borderBottom: i < D4.signoffs.length - 1 ? '1px solid var(--border)' : 'none',
          }}>
            <div style={{ position: 'relative', display: 'flex', justifyContent: 'center' }}>
              <div style={{
                width: 28, height: 28, borderRadius: '50%',
                background: 'var(--success-bg)', color: 'var(--success)',
                display: 'flex', alignItems: 'center', justifyContent: 'center',
                border: '1px solid color-mix(in oklab, var(--success) 30%, transparent)',
              }}>
                <I4 name="check" size={14}/>
              </div>
              {i < D4.signoffs.length - 1 && (
                <span style={{ position: 'absolute', top: 30, bottom: -16, left: '50%',
                  width: 1, background: 'var(--border)', transform: 'translateX(-50%)' }}/>
              )}
            </div>
            <div>
              <div style={{ fontSize: 10.5, fontFamily: 'var(--font-mono)',
                letterSpacing: '0.06em', textTransform: 'uppercase',
                color: 'var(--text-muted)', marginBottom: 3 }}>{s.stage}</div>
              <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
                <A4 name={s.name} size={22}/>
                <span style={{ fontSize: 14, color: 'var(--text)', fontWeight: 500 }}>{s.name}</span>
                <span style={{ fontSize: 12, color: 'var(--text-muted)' }}>{s.role}</span>
              </div>
            </div>
            <span className="mono" style={{ fontSize: 11, color: 'var(--text-muted)' }}>{s.when}</span>
          </div>
        ))}
      </div>
    </section>
  );
}

function ReadsSection() {
  return (
    <section>
      <SectionTitle
        kicker="03 · Cobertura de leitura"
        title="Quem leu esta versão"
        action={
          <a href="#" style={{ fontSize: 12, color: 'var(--brand)', textDecoration: 'none',
            display: 'inline-flex', alignItems: 'center', gap: 4 }}>
            Abrir Fanout completo<I4 name="chevron-right" size={11}/>
          </a>
        }
      />
      <div style={{ display: 'flex', alignItems: 'center', gap: 20 }}>
        <div style={{ flex: 1 }}>
          <div style={{ display: 'flex', alignItems: 'baseline', gap: 8, marginBottom: 8 }}>
            <span style={{ fontSize: 32, fontWeight: 600, letterSpacing: '-0.02em',
              color: 'var(--text)', fontFamily: 'var(--font-display)' }}>84.5%</span>
            <span style={{ fontSize: 13, color: 'var(--text-muted)' }}>142 de 168 alvos</span>
          </div>
          <div style={{ height: 6, background: 'var(--surface-2)', borderRadius: 999, overflow: 'hidden' }}>
            <div style={{ width: '84.5%', height: '100%', background: 'var(--brand)', borderRadius: 999 }}/>
          </div>
          <div style={{ display: 'flex', justifyContent: 'space-between', marginTop: 8,
            fontSize: 11, color: 'var(--text-muted)' }}>
            <span>26 pendentes · 4 com mais de 7 dias</span>
            <span>vence em 8 dias</span>
          </div>
        </div>
      </div>
    </section>
  );
}

function VersionsSection() {
  return (
    <section>
      <SectionTitle kicker="04 · Linhagem" title="Histórico de versões"/>
      <div style={{ display: 'flex', flexDirection: 'column' }}>
        {D4.versions.map((v, i) => (
          <div key={v.v} style={{
            display: 'grid', gridTemplateColumns: '64px 1fr 120px',
            gap: 16, alignItems: 'flex-start',
            padding: '14px 0',
            borderBottom: i < D4.versions.length - 1 ? '1px solid var(--border)' : 'none',
          }}>
            <span className="mono" style={{
              fontSize: 13, fontWeight: v.current ? 600 : 500,
              color: v.current ? 'var(--brand)' : 'var(--text)',
              display: 'inline-flex', alignItems: 'center', gap: 6,
            }}>
              {v.current && <span style={{ width: 6, height: 6, background: 'var(--brand)', borderRadius: '50%' }}/>}
              {v.v}
            </span>
            <div>
              <div style={{ fontSize: 13, color: 'var(--text)', marginBottom: 3, lineHeight: 1.5 }}>{v.summary}</div>
              <div style={{ fontSize: 11, color: 'var(--text-muted)' }}>{v.author}</div>
            </div>
            <span className="mono" style={{ fontSize: 11, color: 'var(--text-muted)', justifySelf: 'end' }}>{v.when}</span>
          </div>
        ))}
      </div>
    </section>
  );
}

function RelatedSection() {
  return (
    <section>
      <SectionTitle kicker="05 · Referências" title="Documentos relacionados"/>
      <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
        {D4.related.map(r => (
          <a key={r.code} href="#" style={{
            display: 'grid', gridTemplateColumns: 'auto 1fr auto auto',
            gap: 14, alignItems: 'center',
            padding: '12px 14px', borderRadius: 6,
            border: '1px solid var(--border)', background: 'var(--surface)',
            textDecoration: 'none', color: 'var(--text)',
            transition: 'all 120ms',
          }} onMouseEnter={e => {
            e.currentTarget.style.background = 'var(--surface-2)';
            e.currentTarget.style.borderColor = 'var(--border-strong)';
          }} onMouseLeave={e => {
            e.currentTarget.style.background = 'var(--surface)';
            e.currentTarget.style.borderColor = 'var(--border)';
          }}>
            <span style={{
              width: 32, height: 32, background: 'var(--brand-pale)',
              color: 'var(--brand)', borderRadius: 4,
              display: 'inline-flex', alignItems: 'center', justifyContent: 'center',
            }}>
              <I4 name="document" size={15}/>
            </span>
            <div>
              <div style={{ fontSize: 13, color: 'var(--text)', fontWeight: 500 }}>{r.title}</div>
              <div className="mono" style={{ fontSize: 11, color: 'var(--text-muted)' }}>{r.code} · {r.type}</div>
            </div>
            <span style={{
              fontSize: 10.5, fontFamily: 'var(--font-mono)', color: 'var(--text-muted)',
              letterSpacing: '0.04em', textTransform: 'uppercase',
            }}>vincula a esta IT</span>
            <I4 name="chevron-right" size={13} style={{ color: 'var(--text-muted)' }}/>
          </a>
        ))}
      </div>
    </section>
  );
}

function AuditSection() {
  return (
    <section>
      <SectionTitle kicker="06 · Auditoria" title="Evidência de imutabilidade"/>
      <div style={{
        background: 'var(--surface)', border: '1px solid var(--border)',
        borderRadius: 8, padding: 20,
        display: 'grid', gridTemplateColumns: '1fr auto', gap: 20, alignItems: 'center',
      }}>
        <div>
          <div style={{ fontSize: 13, color: 'var(--text)', lineHeight: 1.55, marginBottom: 12, maxWidth: 540 }}>
            Este documento foi congelado no momento da publicação. A cadeia de hash SHA-256
            está íntegra e disponível para auditoria conforme cláusula 7.5.3 da ISO 9001.
          </div>
          <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
            <ISOSeal4 hash={D4.hash} hashFull={D4.hashFull}/>
            <span style={{ fontSize: 11, color: 'var(--text-muted)' }}>
              Selado em {D4.publishedAtFull}
            </span>
          </div>
        </div>
        <button className="btn btn-sm">
          <I4 name="download" size={12}/>Exportar evidência
        </button>
      </div>
    </section>
  );
}

function CommentsSection() {
  return (
    <section>
      <SectionTitle kicker="07 · Discussão interna" title="Comentários"/>
      <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
        {D4.comments.map((c, i) => (
          <div key={i} style={{
            display: 'grid', gridTemplateColumns: '32px 1fr',
            gap: 12,
            padding: '14px 0',
            borderBottom: '1px solid var(--border)',
          }}>
            <A4 name={c.who} size={32}/>
            <div>
              <div style={{ display: 'flex', alignItems: 'baseline', gap: 8, marginBottom: 4 }}>
                <span style={{ fontSize: 13, color: 'var(--text)', fontWeight: 500 }}>{c.who}</span>
                <span style={{ fontSize: 11, color: 'var(--text-muted)' }}>{c.when}</span>
              </div>
              <div style={{ fontSize: 13, color: 'var(--text-soft)', lineHeight: 1.55 }}>{c.text}</div>
            </div>
          </div>
        ))}
        <button className="btn btn-sm btn-ghost" style={{ alignSelf: 'flex-start', marginTop: 4 }}>
          <I4 name="plus" size={12}/>Adicionar comentário
        </button>
      </div>
    </section>
  );
}

// =====================================================================
// OBSOLETE OVERLAY — Whole page goes pale, big stamp diagonal.
// =====================================================================
function ObsoleteOverlay() {
  return (
    <div style={{
      position: 'absolute', inset: 0, pointerEvents: 'none', zIndex: 30,
      display: 'flex', alignItems: 'flex-start', justifyContent: 'center',
      paddingTop: 80,
    }}>
      <div style={{
        transform: 'rotate(-8deg)',
        border: '6px solid var(--danger)', color: 'var(--danger)',
        padding: '14px 36px',
        fontFamily: 'var(--font-mono)', fontSize: 36, fontWeight: 700,
        letterSpacing: '0.18em', textTransform: 'uppercase',
        background: 'rgba(255,255,255,0.65)',
        borderRadius: 4,
        boxShadow: '0 8px 32px rgba(200,54,74,0.20)',
      }}>OBSOLETO</div>
    </div>
  );
}

// =====================================================================
// PAGE LAYOUTS
// =====================================================================

// PAGE A — fits in one screen. Hero + sections in 2-col grid below.
function PageA({ Hero, obsolete }) {
  return (
    <div className="palette-wine" style={{
      height: 980, display: 'flex', background: 'var(--bg)', position: 'relative',
      filter: obsolete ? 'grayscale(0.65)' : 'none',
      opacity: obsolete ? 0.85 : 1,
    }}>
      <S4 variant="rail" active="documents"/>
      <div style={{ flex: 1, overflow: 'hidden', display: 'flex', flexDirection: 'column', position: 'relative' }}>
        <Hero obsolete={obsolete}/>

        <div style={{ flex: 1, overflow: 'auto', padding: '28px 56px' }}>
          {/* 2-column dense grid */}
          <div style={{ display: 'grid', gridTemplateColumns: '1.1fr 1fr', gap: 36 }}>
            <div style={{ display: 'flex', flexDirection: 'column', gap: 28 }}>
              <MetaGrid/>
              <SignoffsSection/>
            </div>
            <div style={{ display: 'flex', flexDirection: 'column', gap: 28 }}>
              <ReadsSection/>
              <VersionsSection/>
              <RelatedSection/>
            </div>
          </div>
        </div>

        {obsolete && <ObsoleteOverlay/>}
      </div>
    </div>
  );
}

// PAGE B — generous scroll, single column.
function PageB({ Hero, obsolete }) {
  return (
    <div className="palette-wine" style={{
      height: 980, display: 'flex', background: 'var(--bg)', position: 'relative',
      filter: obsolete ? 'grayscale(0.65)' : 'none',
      opacity: obsolete ? 0.85 : 1,
    }}>
      <S4 variant="rail" active="documents"/>
      <div style={{ flex: 1, overflow: 'hidden', display: 'flex', flexDirection: 'column', position: 'relative' }}>
        <div style={{ flex: 1, overflow: 'auto' }}>
          <Hero obsolete={obsolete}/>

          <div style={{ padding: '40px 56px 80px', maxWidth: 920, margin: '0 auto' }}>
            <div style={{ display: 'flex', flexDirection: 'column', gap: 48 }}>
              <MetaGrid/>
              <SignoffsSection/>
              <ReadsSection/>
              <VersionsSection/>
              <RelatedSection/>
              <AuditSection/>
              <CommentsSection/>
            </div>
          </div>
        </div>

        {obsolete && <ObsoleteOverlay/>}
      </div>
    </div>
  );
}

// =====================================================================
// EXPORTS — 6 variants total: B/C/D heroes × A/B page heights
// =====================================================================
const PubB_A = ({ obsolete }) => <PageA Hero={HeroB} obsolete={obsolete}/>;
const PubB_B = ({ obsolete }) => <PageB Hero={HeroB} obsolete={obsolete}/>;
const PubC_A = ({ obsolete }) => <PageA Hero={HeroC} obsolete={obsolete}/>;
const PubC_B = ({ obsolete }) => <PageB Hero={HeroC} obsolete={obsolete}/>;
const PubD_A = ({ obsolete }) => <PageA Hero={HeroD} obsolete={obsolete}/>;
const PubD_B = ({ obsolete }) => <PageB Hero={HeroD} obsolete={obsolete}/>;

window.MD_ONDA1_V4 = { PubB_A, PubB_B, PubC_A, PubC_B, PubD_A, PubD_B };
