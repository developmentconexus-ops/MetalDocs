/* global React, MD */
const { Icon: I5, Avatar: A5, Sidebar: S5 } = MD;
const { useState: us5, useEffect: ue5, useRef: ur5 } = React;

// =====================================================================
// PUBLICADO v5 — Hero C refined (smaller card) + premium structured body
// References: Linear issue · Vercel deployment · Stripe dashboard
// =====================================================================

const D5 = {
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
  { name: 'Rafael Bittencourt', role: 'Diretor de Operações', when: '12 mar · 14:30', stage: 'Anuência de diretoria' }],

  versions: [
  { v: 'v3.2', when: '12 mar 2026', author: 'C. Mendes', current: true,
    tag: 'minor', summary: 'Inclusão de procedimento para painéis CCM-04. Atualização da matriz de risco para fontes hidráulicas.',
    changes: { added: 4, modified: 12, removed: 1 } },
  { v: 'v3.1', when: '08 jan 2026', author: 'C. Mendes',
    tag: 'patch', summary: 'Correção de typo no item 4.3 (responsabilidades).',
    changes: { added: 0, modified: 2, removed: 0 } },
  { v: 'v3.0', when: '14 nov 2025', author: 'R. Souza',
    tag: 'major', summary: 'Revisão geral pós-auditoria interna. Novas etiquetas padronizadas para todas as plantas.',
    changes: { added: 18, modified: 22, removed: 6 } },
  { v: 'v2.4', when: '02 jul 2025', author: 'R. Souza',
    tag: 'minor', summary: 'Última versão da revisão anterior.',
    changes: { added: 3, modified: 8, removed: 0 } },
  { v: 'v2.3', when: '18 mar 2025', author: 'A. Tavares',
    tag: 'patch', summary: 'Ajustes de redação após feedback de Engenharia.',
    changes: { added: 0, modified: 5, removed: 0 } }],

  related: [
  { code: 'IT-EHS-021', title: 'Inspeção de cadeados e travas', type: 'Instrução', rel: 'referenciado por' },
  { code: 'FR-EHS-008', title: 'Ficha de bloqueio individual', type: 'Formulário', rel: 'anexo obrigatório' },
  { code: 'PR-MAN-103', title: 'Manutenção preventiva CCM-04', type: 'Procedimento', rel: 'invoca este PR' }],

  comments: [
  { who: 'Marcos Lima', role: 'Gerente Industrial', when: '13 mar · 09:42',
    text: 'Equipes de Pomerode já alinhadas. Treinamento iniciado para CCM-04 na próxima segunda.' },
  { who: 'Renata Souza', role: 'Coord. SSMA', when: '14 mar · 11:08',
    text: 'Adicionei FR-EHS-008 como anexo de referência para os supervisores de turno.' }]

};

// ---------- ISO seal ----------
function ISOSeal5({ hash, hashFull }) {
  const [open, setOpen] = us5(false);
  return (
    <span style={{ position: 'relative', display: 'inline-flex' }}
    onMouseEnter={() => setOpen(true)} onMouseLeave={() => setOpen(false)}>
      <span style={{
        display: 'inline-flex', alignItems: 'center', gap: 6,
        height: 22, padding: '0 9px',
        fontFamily: 'var(--font-mono)', fontSize: 10.5, letterSpacing: '0.04em',
        background: 'var(--success-bg)', color: 'var(--success)',
        border: '1px solid color-mix(in oklab, var(--success) 30%, transparent)',
        borderRadius: 'var(--r-pill)', cursor: 'help'
      }}>
        <span style={{ width: 6, height: 6, borderRadius: '50%', background: 'currentColor',
          animation: 'sealPulse5 2.4s ease-in-out infinite' }} />
        {hash}
      </span>
      {open &&
      <div style={{
        position: 'absolute', bottom: 'calc(100% + 6px)', left: 0, zIndex: 50,
        padding: '10px 12px', background: '#1a0e0e', color: '#fff',
        borderRadius: 6, fontSize: 11, lineHeight: 1.5, width: 320,
        boxShadow: '0 12px 32px rgba(0,0,0,0.3)'
      }}>
          <div style={{ fontFamily: 'var(--font-mono)', fontSize: 9.5, opacity: 0.7, marginBottom: 4 }}>SHA-256</div>
          <div style={{ fontFamily: 'var(--font-mono)', fontSize: 11, wordBreak: 'break-all', opacity: 0.95 }}>{hashFull}</div>
        </div>
      }
      <style>{`@keyframes sealPulse5 { 0%,100% { opacity: 1; transform: scale(1);} 50% { opacity: .55; transform: scale(.85);} }`}</style>
    </span>);

}

// ---------- Smaller doc card (tilted) ----------
function DocCardSmall({ tilt = true }) {
  return (
    <div style={{
      width: 168, height: 224,
      background: '#fff',
      border: '1px solid var(--border)',
      borderRadius: 4,
      boxShadow: tilt ?
      '20px 20px 48px rgba(74,33,33,0.16), 4px 4px 14px rgba(74,33,33,0.10)' :
      '0 12px 32px rgba(74,33,33,0.10)',
      transform: tilt ? 'perspective(1200px) rotateY(-12deg) rotateX(4deg)' : 'none',
      transformStyle: 'preserve-3d',
      display: 'flex', flexDirection: 'column',
      flexShrink: 0
    }}>
      <div style={{
        height: 36,
        background: 'linear-gradient(135deg, var(--brand) 0%, var(--brand-deep) 100%)',
        display: 'flex', alignItems: 'center', padding: '0 14px',
        color: '#fff', fontFamily: 'var(--font-mono)', fontSize: 9.5,
        letterSpacing: '0.10em', textTransform: 'uppercase', fontWeight: 500
      }}>{D5.areaShort}</div>
      <div style={{ flex: 1, padding: 14, display: 'flex', flexDirection: 'column' }}>
        <div style={{
          fontFamily: 'var(--font-mono)', fontSize: 18,
          color: 'var(--brand-deep)', fontWeight: 600, letterSpacing: '-0.01em',
          marginBottom: 4
        }}>{D5.code}</div>
        <div style={{
          fontSize: 9.5, color: 'var(--text-soft)',
          letterSpacing: '0.06em', textTransform: 'uppercase', fontWeight: 500,
          fontFamily: 'var(--font-mono)'
        }}>{D5.typeShort}</div>
        <div style={{ flex: 1 }} />
        <div style={{ height: 1, background: 'var(--border)', marginBottom: 8 }} />
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <span style={{ fontSize: 9.5, color: 'var(--text-muted)',
            fontFamily: 'var(--font-mono)', letterSpacing: '0.04em' }}>{D5.version}</span>
          <span style={{ width: 5, height: 5, borderRadius: '50%', background: 'var(--success)' }} />
        </div>
      </div>
    </div>);

}

// ---------- Hero C (refined, smaller card) ----------
function HeroC5({ obsolete }) {
  return (
    <header style={{
      padding: '32px 56px 40px',
      background: 'linear-gradient(180deg, var(--brand-pale) 0%, var(--surface) 100%)',
      borderBottom: '1px solid var(--border)',
      position: 'relative', overflow: 'hidden'
    }}>
      <div style={{
        position: 'absolute', inset: 0, opacity: 0.04, pointerEvents: 'none',
        backgroundImage: 'linear-gradient(var(--brand) 1px, transparent 1px), linear-gradient(90deg, var(--brand) 1px, transparent 1px)',
        backgroundSize: '32px 32px'
      }} />
      <div style={{ position: 'relative', marginBottom: 24,
        display: 'flex', alignItems: 'center', gap: 8,
        fontSize: 11, color: 'var(--text-muted)', fontFamily: 'var(--font-mono)',
        letterSpacing: '0.04em', textTransform: 'uppercase' }}>
        <a href="#" style={{ color: 'inherit', textDecoration: 'none' }}>Biblioteca</a>
        <I5 name="chevron-right" size={10} />
        <a href="#" style={{ color: 'inherit', textDecoration: 'none' }}>{D5.areaShort}</a>
        <I5 name="chevron-right" size={10} />
        <a href="#" style={{ color: 'inherit', textDecoration: 'none' }}>Procedimentos</a>
        <I5 name="chevron-right" size={10} />
        <span style={{ color: 'var(--text)' }}>{D5.code}</span>
      </div>

      <div style={{
        position: 'relative', display: 'grid',
        gridTemplateColumns: '210px 1fr',
        gap: 40, alignItems: 'center'
      }}>
        <div style={{ display: 'flex', justifyContent: 'center' }}>
          <DocCardSmall tilt />
        </div>

        <div>
          <div style={{ display: 'flex', alignItems: 'center', gap: 10, marginBottom: 14 }}>
            <span style={{
              display: 'inline-flex', alignItems: 'center',
              height: 24, padding: '0 11px', borderRadius: 'var(--r-pill)',
              background: 'var(--surface)', color: 'var(--brand)',
              border: '1px solid var(--brand-pale)',
              fontSize: 11.5, fontWeight: 600, fontFamily: 'var(--font-mono)',
              letterSpacing: '0.04em'
            }}>{D5.code}</span>
            {!obsolete &&
            <span style={{
              display: 'inline-flex', alignItems: 'center', gap: 5,
              height: 22, padding: '0 10px', fontSize: 11, fontWeight: 500,
              background: 'var(--success-bg)', color: 'var(--success)',
              borderRadius: 'var(--r-pill)', fontFamily: 'var(--font-mono)',
              letterSpacing: '0.04em', textTransform: 'uppercase'
            }}>
                <span style={{ width: 5, height: 5, background: 'currentColor', borderRadius: '50%' }} />
                {D5.version} · vigente
              </span>
            }
            <span style={{ fontSize: 11, color: 'var(--text-muted)',
              fontFamily: 'var(--font-mono)', letterSpacing: '0.06em', textTransform: 'uppercase' }}>
              {D5.typeShort}
            </span>
          </div>

          <h1 style={{
            fontSize: 36, lineHeight: 1.1, letterSpacing: '-0.025em',
            margin: '0 0 12px', fontWeight: 600, color: 'var(--text)',
            maxWidth: 640, textWrap: 'balance'
          }}>{D5.title}</h1>

          <p style={{
            fontSize: 14.5, lineHeight: 1.55, color: 'var(--text-soft)',
            margin: '0 0 22px', maxWidth: 580
          }}>{D5.subtitle}</p>

          <div style={{ display: 'flex', alignItems: 'center', gap: 8, flexWrap: 'wrap' }}>
            <button className="btn btn-primary btn-lg" disabled={obsolete}
            style={{ height: 42, padding: '0 20px', fontSize: 14, fontWeight: 600 }}>
              <I5 name="eye" size={15} />Visualizar documento
            </button>
            <button className="btn" disabled={obsolete}>
              <I5 name="download" size={13} />Baixar PDF
            </button>
            <button className="btn" disabled={obsolete}>
              <I5 name="edit" size={13} />Iniciar revisão
            </button>
            <button className="btn btn-ghost" disabled={obsolete}>
              <I5 name="link" size={13} />Copiar link
            </button>
          </div>
        </div>
      </div>
    </header>);

}

// ---------- KPI strip (Stripe-style overview row) ----------
function KPIStrip() {
  const kpis = [
  { label: 'Versão atual', value: D5.version, mono: true, hint: 'desde 12 mar' },
  { label: 'Cobertura', value: '84.5%', sub: '142/168 leram',
    bar: 84.5, tone: 'brand' },
  { label: 'Próxima revisão', value: '15 mar 2027', sub: 'em 364 dias' },
  { label: 'Páginas', value: D5.pageCount, sub: D5.size, mono: true }];

  return (
    <div style={{
      display: 'grid', gridTemplateColumns: 'repeat(4, 1fr)',
      background: 'var(--surface)',
      border: '1px solid var(--border)',
      borderRadius: 8,
      overflow: 'hidden'
    }}>
      {kpis.map((k, i) =>
      <div key={k.label} style={{
        padding: '18px 22px',
        borderRight: i < kpis.length - 1 ? '1px solid var(--border)' : 'none'
      }}>
          <div className="kicker" style={{ fontSize: 9.5, marginBottom: 8 }}>{k.label}</div>
          <div style={{
          fontSize: 22, fontWeight: 600, letterSpacing: '-0.02em',
          color: 'var(--text)', lineHeight: 1.1,
          fontFamily: k.mono ? 'var(--font-mono)' : 'var(--font-display)',
          marginBottom: k.bar !== undefined ? 8 : 4
        }}>{k.value}</div>
          {k.bar !== undefined &&
        <div style={{ height: 4, background: 'var(--surface-3)', borderRadius: 999, overflow: 'hidden', marginBottom: 6 }}>
              <div style={{ width: `${k.bar}%`, height: '100%', background: 'var(--brand)' }} />
            </div>
        }
          {(k.sub || k.hint) &&
        <div style={{ fontSize: 11, color: 'var(--text-muted)' }}>{k.sub || k.hint}</div>
        }
        </div>
      )}
    </div>);

}

// ---------- Section title (with optional aside link) ----------
function ST({ kicker, title, aside }) {
  return (
    <div style={{ display: 'flex', alignItems: 'flex-end', justifyContent: 'space-between',
      marginBottom: 18 }}>
      <div>
        <div className="kicker" style={{ fontSize: 9.5, marginBottom: 4 }}>{kicker}</div>
        <h2 style={{ fontSize: 17, lineHeight: 1.2, margin: 0,
          fontWeight: 600, letterSpacing: '-0.01em', color: 'var(--text)' }}>{title}</h2>
      </div>
      {aside}
    </div>);

}

// ---------- About card (subtitle + key facts) ----------
function AboutCard() {
  const facts = [
  { icon: 'briefcase', label: 'Tipo', value: D5.type },
  { icon: 'layers', label: 'Área', value: D5.area },
  { icon: 'calendar', label: 'Vigente desde', value: D5.effectiveAt },
  { icon: 'clock', label: 'Próxima revisão', value: D5.reviewIn },
  { icon: 'document', label: 'Tamanho', value: `${D5.pageCount} páginas · ${D5.size}` },
  { icon: 'shield', label: 'Selo ISO', value: D5.hash, mono: true }];

  return (
    <div style={{
      background: 'var(--surface)',
      border: '1px solid var(--border)',
      borderRadius: 8,
      overflow: 'hidden'
    }}>
      {/* Top: owner banner */}
      <div style={{
        padding: '18px 22px',
        borderBottom: '1px solid var(--border)',
        background: 'var(--surface-2)',
        display: 'flex', alignItems: 'center', gap: 12
      }}>
        <A5 name={D5.publishedBy} size={36} />
        <div style={{ flex: 1 }}>
          <div style={{ fontSize: 13, color: 'var(--text)', fontWeight: 500 }}>{D5.publishedBy}</div>
          <div style={{ fontSize: 11, color: 'var(--text-muted)' }}>{D5.publishedByRole} · publicou em {D5.publishedAtFull}</div>
        </div>
        <button className="btn btn-sm btn-ghost"><I5 name="mail" size={12} />Contatar</button>
      </div>

      {/* Facts grid */}
      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr' }}>
        {facts.map((f, i) =>
        <div key={f.label} style={{
          padding: '14px 22px',
          display: 'flex', alignItems: 'flex-start', gap: 12,
          borderRight: i % 2 === 0 ? '1px solid var(--border)' : 'none',
          borderBottom: i < facts.length - 2 ? '1px solid var(--border)' : 'none'
        }}>
            <div style={{
            width: 28, height: 28, borderRadius: 6,
            background: 'var(--brand-pale)', color: 'var(--brand)',
            display: 'flex', alignItems: 'center', justifyContent: 'center',
            flexShrink: 0
          }}>
              <I5 name={f.icon} size={14} />
            </div>
            <div style={{ minWidth: 0, flex: 1 }}>
              <div className="kicker" style={{ fontSize: 9, marginBottom: 2 }}>{f.label}</div>
              <div style={{
              fontSize: 12.5, color: 'var(--text)', fontWeight: 500,
              fontFamily: f.mono ? 'var(--font-mono)' : 'inherit',
              whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis'
            }}>{f.value}</div>
            </div>
          </div>
        )}
      </div>
    </div>);

}

// ---------- Sign-off pipeline (horizontal stepper) ----------
function SignoffPipeline() {
  return (
    <div style={{
      background: 'var(--surface)',
      border: '1px solid var(--border)',
      borderRadius: 8,
      padding: '24px 24px 22px'
    }}>
      <div style={{
        position: 'relative',
        display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)',
        gap: 0
      }}>
        {/* Connector line */}
        <div style={{
          position: 'absolute', top: 18, left: '16.66%', right: '16.66%',
          height: 2, background: 'var(--success)',
          borderRadius: 1
        }} />

        {D5.signoffs.map((s, i) =>
        <div key={s.name} style={{
          position: 'relative',
          display: 'flex', flexDirection: 'column', alignItems: 'center',
          textAlign: 'center', padding: '0 8px'
        }}>
            {/* Pin */}
            <div style={{
            width: 36, height: 36, borderRadius: '50%',
            background: 'var(--success)', color: '#fff',
            display: 'flex', alignItems: 'center', justifyContent: 'center',
            border: '4px solid var(--surface)',
            boxShadow: '0 2px 8px rgba(26,107,53,0.30)',
            marginBottom: 14,
            zIndex: 2
          }}>
              <I5 name="check" size={16} />
            </div>
            <div className="kicker" style={{ fontSize: 9, marginBottom: 6 }}>
              {s.stage}
            </div>
            <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 4 }}>
              <A5 name={s.name} size={22} />
              <span style={{ fontSize: 13, color: 'var(--text)', fontWeight: 500 }}>{s.name}</span>
            </div>
            <div style={{ fontSize: 11, color: 'var(--text-muted)', marginBottom: 4 }}>{s.role}</div>
            <div className="mono" style={{ fontSize: 10.5, color: 'var(--text-faint)' }}>{s.when}</div>
          </div>
        )}
      </div>
    </div>);

}

// ---------- Version timeline (horizontal pins, animated) ----------
function VersionTimeline() {
  const versions = [...D5.versions].reverse();
  const currentIdx = versions.findIndex((v) => v.current);
  const [hovered, setHovered] = us5(currentIdx >= 0 ? currentIdx : versions.length - 1);
  const tagColor = (tag) => ({
    major: { bg: 'var(--brand)', text: '#fff' },
    minor: { bg: 'var(--info-bg)', text: 'var(--info)' },
    patch: { bg: 'var(--surface-3)', text: 'var(--text-muted)' }
  })[tag];

  return (
    <div style={{
      background: 'var(--surface)',
      border: '1px solid var(--border)',
      borderRadius: 8,
      padding: '28px 28px 24px'
    }}>
      {/* Track */}
      <div style={{ position: 'relative', height: 56, marginBottom: 20 }}>
        {/* Line */}
        <div style={{
          position: 'absolute', left: 8, right: 8, top: 27,
          height: 2,
          background: 'linear-gradient(90deg, var(--surface-3) 0%, var(--brand-pale) 60%, var(--brand) 100%)',
          borderRadius: 1
        }} />
        {/* Pins */}
        <div style={{
          position: 'absolute', inset: 0,
          display: 'grid', gridTemplateColumns: `repeat(${versions.length}, 1fr)`
        }}>
          {versions.map((v, i) => {
            const active = i === hovered;
            const tone = tagColor(v.tag);
            return (
              <button key={v.v}
              onMouseEnter={() => setHovered(i)}
              onFocus={() => setHovered(i)}
              style={{
                position: 'relative',
                background: 'transparent', border: 'none', cursor: 'pointer',
                padding: 0,
                display: 'flex', flexDirection: 'column', alignItems: 'center'
              }}>
                {/* Date label above */}
                <div style={{
                  fontSize: 9.5, fontFamily: 'var(--font-mono)',
                  color: active ? 'var(--text)' : 'var(--text-muted)',
                  letterSpacing: '0.04em', textTransform: 'uppercase',
                  marginBottom: 6, fontWeight: active ? 600 : 400,
                  transition: 'all 200ms ease'
                }}>{v.when.split(' ')[0]} {v.when.split(' ')[1]}</div>

                {/* Pin */}
                <div style={{
                  position: 'relative',
                  width: active ? 22 : 14, height: active ? 22 : 14,
                  borderRadius: '50%',
                  background: v.current ? 'var(--brand)' : 'var(--surface)',
                  border: `2px solid ${v.current ? 'var(--brand-deep)' : 'var(--border-strong)'}`,
                  boxShadow: active ? '0 4px 14px rgba(107,31,42,0.30)' : 'none',
                  transition: 'all 200ms cubic-bezier(.34,1.56,.64,1)',
                  zIndex: 2
                }}>
                  {v.current &&
                  <span style={{
                    position: 'absolute', inset: -6, borderRadius: '50%',
                    border: '2px solid var(--brand)',
                    animation: 'pinPulse5 2s ease-out infinite'
                  }} />
                  }
                </div>

                {/* Version label */}
                <div className="mono" style={{
                  fontSize: 11, fontWeight: 600,
                  color: v.current ? 'var(--brand)' : active ? 'var(--text)' : 'var(--text-muted)',
                  marginTop: 6,
                  transition: 'all 200ms ease'
                }}>{v.v}</div>
              </button>);

          })}
        </div>
      </div>

      {/* Detail panel (animated swap) */}
      <div key={hovered} style={{
        animation: 'fadeUp5 240ms ease',
        padding: '16px 18px',
        background: 'var(--surface-2)',
        border: '1px solid var(--border)',
        borderRadius: 6,
        display: 'grid', gridTemplateColumns: '1fr auto', gap: 20, alignItems: 'flex-start'
      }}>
        <div>
          <div style={{ display: 'flex', alignItems: 'center', gap: 10, marginBottom: 8 }}>
            <span className="mono" style={{
              fontSize: 13, fontWeight: 600, color: 'var(--text)'
            }}>{versions[hovered].v}</span>
            {(() => {
              const t = tagColor(versions[hovered].tag);
              return (
                <span style={{
                  height: 18, padding: '0 7px', borderRadius: 'var(--r-pill)',
                  fontSize: 9.5, fontFamily: 'var(--font-mono)',
                  letterSpacing: '0.06em', textTransform: 'uppercase', fontWeight: 600,
                  background: t.bg, color: t.text,
                  display: 'inline-flex', alignItems: 'center'
                }}>{versions[hovered].tag}</span>);

            })()}
            {versions[hovered].current &&
            <span style={{
              height: 18, padding: '0 8px', borderRadius: 'var(--r-pill)',
              fontSize: 9.5, fontFamily: 'var(--font-mono)',
              letterSpacing: '0.06em', textTransform: 'uppercase', fontWeight: 600,
              background: 'var(--success-bg)', color: 'var(--success)',
              display: 'inline-flex', alignItems: 'center'
            }}>Atual</span>
            }
            <span className="mono" style={{ fontSize: 10.5, color: 'var(--text-muted)' }}>
              · {versions[hovered].when} · {versions[hovered].author}
            </span>
          </div>
          <div style={{ fontSize: 13, color: 'var(--text)', lineHeight: 1.55 }}>
            {versions[hovered].summary}
          </div>
          {/* Diff stats */}
          <div style={{ display: 'flex', gap: 14, marginTop: 12, fontSize: 11,
            fontFamily: 'var(--font-mono)' }}>
            <span style={{ color: 'var(--success)' }}>
              + {versions[hovered].changes.added} adicionados
            </span>
            <span style={{ color: 'var(--info)' }}>
              ~ {versions[hovered].changes.modified} alterados
            </span>
            <span style={{ color: 'var(--danger)' }}>
              − {versions[hovered].changes.removed} removidos
            </span>
          </div>
        </div>
        {!versions[hovered].current &&
        <button className="btn btn-sm">
            <I5 name="eye" size={12} />Ver esta versão
          </button>
        }
      </div>

      <style>{`
        @keyframes pinPulse5 {
          0% { transform: scale(1); opacity: 1; }
          100% { transform: scale(1.8); opacity: 0; }
        }
        @keyframes fadeUp5 {
          from { opacity: 0; transform: translateY(4px); }
          to { opacity: 1; transform: translateY(0); }
        }
      `}</style>
    </div>);

}

// ---------- Related docs grid ----------
function RelatedGrid() {
  return (
    <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: 10 }}>
      {D5.related.map((r) =>
      <a key={r.code} href="#" style={{
        background: 'var(--surface)',
        border: '1px solid var(--border)',
        borderRadius: 8,
        padding: 16,
        textDecoration: 'none', color: 'var(--text)',
        display: 'flex', flexDirection: 'column', gap: 10,
        transition: 'all 140ms'
      }} onMouseEnter={(e) => {
        e.currentTarget.style.borderColor = 'var(--brand)';
        e.currentTarget.style.boxShadow = 'var(--shadow-2)';
        e.currentTarget.style.transform = 'translateY(-2px)';
      }} onMouseLeave={(e) => {
        e.currentTarget.style.borderColor = 'var(--border)';
        e.currentTarget.style.boxShadow = 'none';
        e.currentTarget.style.transform = 'none';
      }}>
          <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
            <span style={{
            width: 28, height: 28, background: 'var(--brand-pale)',
            color: 'var(--brand)', borderRadius: 6,
            display: 'inline-flex', alignItems: 'center', justifyContent: 'center'
          }}>
              <I5 name="document" size={14} />
            </span>
            <span style={{
            fontSize: 9.5, fontFamily: 'var(--font-mono)', color: 'var(--text-muted)',
            letterSpacing: '0.06em', textTransform: 'uppercase'
          }}>{r.rel}</span>
          </div>
          <div style={{ fontSize: 13, color: 'var(--text)', fontWeight: 500, lineHeight: 1.35 }}>{r.title}</div>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginTop: 'auto' }}>
            <span className="mono" style={{ fontSize: 10.5, color: 'var(--text-muted)' }}>{r.code}</span>
            <span style={{ fontSize: 10.5, color: 'var(--text-muted)' }}>{r.type}</span>
          </div>
        </a>
      )}
    </div>);

}

// ---------- Coverage card (small, with link to fanout) ----------
function CoverageCard() {
  return (
    <div style={{
      background: 'var(--surface)', border: '1px solid var(--border)',
      borderRadius: 8, padding: 22,
      display: 'grid', gridTemplateColumns: '1fr auto', gap: 24, alignItems: 'center'
    }}>
      <div>
        <div style={{ display: 'flex', alignItems: 'baseline', gap: 8, marginBottom: 8 }}>
          <span style={{ fontSize: 28, fontWeight: 600, letterSpacing: '-0.02em',
            color: 'var(--text)', fontFamily: 'var(--font-display)' }}>84.5%</span>
          <span style={{ fontSize: 12, color: 'var(--text-muted)' }}>de 168 destinatários</span>
        </div>
        <div style={{ height: 6, background: 'var(--surface-2)', borderRadius: 999, overflow: 'hidden', marginBottom: 8 }}>
          <div style={{ width: '84.5%', height: '100%',
            background: 'linear-gradient(90deg, var(--brand) 0%, var(--brand-soft) 100%)',
            borderRadius: 999 }} />
        </div>
        <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: 11, color: 'var(--text-muted)' }}>
          <span>26 pendentes</span>
          <span>vence em 8 dias</span>
        </div>
      </div>
      <a href="#" className="btn btn-sm">
        Abrir Fanout<I5 name="chevron-right" size={12} />
      </a>
    </div>);

}

// ---------- Audit card ----------
function AuditCard() {
  return (
    <div style={{
      background: 'var(--surface)', border: '1px solid var(--border)',
      borderRadius: 8, padding: 22,
      display: 'flex', flexDirection: 'column', gap: 14,
      position: 'relative'
    }}>
      {/* Top row: title + hash chip on the right */}
      <div style={{ display: 'flex', alignItems: 'flex-start', justifyContent: 'space-between', gap: 12 }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 12, minWidth: 0 }}>
          <div style={{
            width: 36, height: 36, borderRadius: 8,
            background: 'var(--success-bg)', color: 'var(--success)',
            display: 'flex', alignItems: 'center', justifyContent: 'center', flexShrink: 0
          }}>
            <I5 name="shield" size={16} />
          </div>
          <div style={{ fontSize: 13.5, color: 'var(--text)', fontWeight: 500 }}>
            Cadeia de hash íntegra
          </div>
        </div>
        <ISOSeal5 hash={D5.hash} hashFull={D5.hashFull} />
      </div>

      <div style={{ fontSize: 12, color: 'var(--text-muted)', lineHeight: 1.5, marginTop: 'auto' }}>
        Documento congelado &middot; SHA-256 íntegro &middot; ISO 9001 cl. 7.5.3.
      </div>
    </div>);

}

// ---------- Comments thread ----------
function CommentsCard() {
  return (
    <div style={{
      background: 'var(--surface)', border: '1px solid var(--border)',
      borderRadius: 8, padding: '4px 22px 18px'
    }}>
      {D5.comments.map((c, i) =>
      <div key={i} style={{
        display: 'grid', gridTemplateColumns: '36px 1fr',
        gap: 14, padding: '18px 0',
        borderBottom: i < D5.comments.length - 1 ? '1px solid var(--border)' : 'none'
      }}>
          <A5 name={c.who} size={36} />
          <div>
            <div style={{ display: 'flex', alignItems: 'baseline', gap: 8, marginBottom: 4, flexWrap: 'wrap' }}>
              <span style={{ fontSize: 13, color: 'var(--text)', fontWeight: 500 }}>{c.who}</span>
              <span style={{ fontSize: 11, color: 'var(--text-muted)' }}>· {c.role}</span>
              <span style={{ fontSize: 11, color: 'var(--text-faint)', marginLeft: 'auto' }}>{c.when}</span>
            </div>
            <div style={{ fontSize: 13, color: 'var(--text-soft)', lineHeight: 1.55 }}>{c.text}</div>
          </div>
        </div>
      )}
      {/* Reply box */}
      <div style={{
        display: 'grid', gridTemplateColumns: '36px 1fr',
        gap: 14, padding: '14px 0 0',
        borderTop: '1px solid var(--border)'
      }}>
        <A5 name="Você" size={36} />
        <div style={{
          display: 'flex', alignItems: 'center', gap: 8,
          padding: '10px 12px',
          background: 'var(--surface-2)', border: '1px solid var(--border)',
          borderRadius: 6
        }}>
          <span style={{ flex: 1, fontSize: 13, color: 'var(--text-faint)' }}>
            Adicionar um comentário…
          </span>
          <button className="btn btn-sm btn-ghost"><I5 name="paperclip" size={12} /></button>
          <button className="btn btn-sm btn-primary">Comentar</button>
        </div>
      </div>
    </div>);

}

// =====================================================================
// MAIN PAGE
// =====================================================================
function PublicadoV5({ obsolete = false }) {
  return (
    <div className="palette-wine" style={{
      height: 980, display: 'flex', background: 'var(--bg)', position: 'relative',
      filter: obsolete ? 'grayscale(0.65)' : 'none',
      opacity: obsolete ? 0.85 : 1
    }}>
      <S5 variant="rail" active="documents" />
      <div style={{ flex: 1, overflow: 'hidden', display: 'flex', flexDirection: 'column', position: 'relative' }}>
        <div style={{ flex: 1, overflow: 'auto' }}>
          <HeroC5 obsolete={obsolete} />

          <div style={{ padding: '32px 56px 80px', maxWidth: 1180, margin: '0 auto',
            display: 'flex', flexDirection: 'column', gap: 36 }}>

            {/* KPI strip */}
            <KPIStrip />

            {/* About + Coverage row */}
            <section>
              <ST kicker="01 · Sobre" title="Identificação e responsabilidade" />
              <div style={{ display: 'grid', gridTemplateColumns: '1.4fr 1fr', gap: 20 }}>
                <AboutCard />
                <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
                  <CoverageCard />
                  <AuditCard />
                </div>
              </div>
            </section>

            {/* Sign-off pipeline */}
            <section>
              <ST kicker="02 · Cadeia de aprovação" title="Sign-offs desta versão" />
              <SignoffPipeline />
            </section>

            {/* Version timeline */}
            <section>
              <ST kicker="03 · Linhagem" title="Histórico de versões"
              aside={<span style={{ fontSize: 11, color: 'var(--text-muted)',
                fontFamily: 'var(--font-mono)', letterSpacing: '0.04em', textTransform: 'uppercase' }}>
                  passe o mouse para detalhar
                </span>} />
              <VersionTimeline />
            </section>

            {/* Related docs */}
            <section>
              <ST kicker="04 · Referências" title="Documentos relacionados"
              aside={<a href="#" style={{ fontSize: 12, color: 'var(--brand)', textDecoration: 'none' }}>
                  Ver todos os 7 vínculos →
                </a>} />
              <RelatedGrid />
            </section>

            {/* Comments */}
            <section>
              <ST kicker="05 · Discussão interna" title={`Comentários (${D5.comments.length})`} />
              <CommentsCard />
            </section>
          </div>
        </div>

        {obsolete &&
        <div style={{
          position: 'absolute', inset: 0, pointerEvents: 'none', zIndex: 30,
          display: 'flex', alignItems: 'flex-start', justifyContent: 'center',
          paddingTop: 80
        }}>
            <div style={{
            transform: 'rotate(-8deg)',
            border: '6px solid var(--danger)', color: 'var(--danger)',
            padding: '14px 36px',
            fontFamily: 'var(--font-mono)', fontSize: 36, fontWeight: 700,
            letterSpacing: '0.18em', textTransform: 'uppercase',
            background: 'rgba(255,255,255,0.65)', borderRadius: 4,
            boxShadow: '0 8px 32px rgba(200,54,74,0.20)'
          }}>OBSOLETO</div>
          </div>
        }
      </div>
    </div>);

}

window.MD_ONDA1_V5 = { PublicadoV5 };