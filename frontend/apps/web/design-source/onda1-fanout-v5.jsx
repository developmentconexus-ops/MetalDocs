/* global React, MD */
const { Icon: IF5, Avatar: AF5, Sidebar: SF5 } = MD;
const { useState: usf5, useEffect: uef5, useMemo: umf5, useRef: urf5 } = React;

// =====================================================================
// FANOUT v5 — Distribuição & cobertura
// Same DNA as PublicadoV5: wine palette, hero gradient, KPI strip, sections
// =====================================================================

const DF5 = {
  code: 'PR-EHS-014',
  title: 'Procedimento de Bloqueio e Etiquetagem (LOTO)',
  version: 'v3.2',
  area: 'Saúde, Segurança & Meio Ambiente',
  areaShort: 'SSMA',
  typeShort: 'Procedimento',
  hash: 'a7f3c9e1d8b4',

  // Distribution
  publishedAt: '12 mar 2026',
  publishedAtFull: '12 mar 2026 · 14:32',
  deadline: '22 mar 2026',
  deadlineFull: '22 mar 2026 · 23:59',
  daysLeft: 8,
  daysSince: 8,
  policy: 'Reconhecimento obrigatório (assinatura)',
  channel: 'E-mail + portal interno',
  remindersSent: 2,
  remindersScheduled: '20 mar · 08:00',

  // Counts
  totalTargets: 248,
  read: 182,
  acknowledged: 156,
  pending: 66,
  overdue: 12,

  // Daily cumulative ack since publish (8 days)
  dailyReads: [
    { day: '12/03', cumAck: 22 },
    { day: '13/03', cumAck: 58 },
    { day: '14/03', cumAck: 84 },
    { day: '15/03', cumAck: 102 },
    { day: '16/03', cumAck: 118 },
    { day: '17/03', cumAck: 132 },
    { day: '18/03', cumAck: 144 },
    { day: '19/03', cumAck: 156 }
  ],

  byArea: [
    { area: 'Manutenção',          total: 64, read: 58, ack: 52 },
    { area: 'Produção · Linha A',  total: 42, read: 38, ack: 34 },
    { area: 'SSMA',                total: 18, read: 18, ack: 18 },
    { area: 'Engenharia',          total: 24, read: 19, ack: 16 },
    { area: 'Produção · Linha B',  total: 38, read: 22, ack: 18 },
    { area: 'Logística',           total: 28, read: 14, ack: 10 },
    { area: 'Qualidade',           total: 16, read: 8,  ack: 5 },
    { area: 'Administrativo',      total: 18, read: 5,  ack: 3 }
  ],

  people: [
    { who: 'Roberto Carvalho',   area: 'Logística',         last: 'nunca abriu',           when: '—',           status: 'pending', overdue: false, role: 'Op. Empilhadeira' },
    { who: 'Ana Paula Mendes',   area: 'Administrativo',    last: 'nunca abriu',           when: '—',           status: 'pending', overdue: true,  role: 'Analista RH' },
    { who: 'Thiago Rocha',       area: 'Qualidade',         last: 'apenas leu',            when: '08 mar 2026', status: 'read',    overdue: false, role: 'Insp. Qualidade' },
    { who: 'Felipe Andrade',     area: 'Produção · B',      last: 'aberto há 6 dias',      when: '14 mar 2026', status: 'pending', overdue: false, role: 'Op. Multifuncional' },
    { who: 'Beatriz Moreira',    area: 'Administrativo',    last: 'nunca abriu',           when: '—',           status: 'pending', overdue: true,  role: 'Aux. Administrativo' },
    { who: 'Gabriel Costa',      area: 'Logística',         last: 'nunca abriu',           when: '—',           status: 'pending', overdue: true,  role: 'Conferente' },
    { who: 'Daniela Prado',      area: 'Manutenção',        last: 'reconheceu há 4 min',   when: '19 mar · 14:08', status: 'ack',  overdue: false, role: 'Téc. Mecânica' },
    { who: 'Eduardo Castro',     area: 'Engenharia',        last: 'reconheceu há 28 min',  when: '19 mar · 13:44', status: 'ack',  overdue: false, role: 'Eng. Processo' },
    { who: 'Marina Albuquerque', area: 'Manutenção',        last: 'apenas leu há 1h',      when: '19 mar · 13:08', status: 'read', overdue: false, role: 'Téc. Elétrica' },
    { who: 'João Vitor Lima',    area: 'Produção · Linha A',last: 'reconheceu ontem',      when: '18 mar · 17:22', status: 'ack',  overdue: false, role: 'Líder de Turno' },
    { who: 'Helena Tavares',     area: 'SSMA',              last: 'reconheceu há 2 dias',  when: '17 mar · 09:14', status: 'ack',  overdue: false, role: 'Téc. Segurança' }
  ]
};

const STATUS_TONE = {
  ack:     { label: 'Reconheceu',    bg: 'var(--success-bg)', fg: 'var(--success)' },
  read:    { label: 'Apenas leu',    bg: 'var(--brand-pale)', fg: 'var(--brand)'   },
  pending: { label: 'Pendente',      bg: 'var(--surface-3)',  fg: 'var(--text-muted)' },
  overdue: { label: 'Em atraso',     bg: 'var(--danger-bg)',  fg: 'var(--danger)'  }
};

// ---------- Animated counter ----------
function useCounterF5(target, duration = 1200) {
  const [v, setV] = usf5(0);
  uef5(() => {
    let raf, start;
    const step = (ts) => {
      if (!start) start = ts;
      const p = Math.min(1, (ts - start) / duration);
      const eased = 1 - Math.pow(1 - p, 3);
      setV(target * eased);
      if (p < 1) raf = requestAnimationFrame(step);
    };
    raf = requestAnimationFrame(step);
    return () => cancelAnimationFrame(raf);
  }, [target, duration]);
  return v;
}

// ---------- Doc reference card (compact) ----------
function DocRefCardF5() {
  return (
    <div style={{
      width: 168, height: 224,
      background: '#fff',
      border: '1px solid var(--border)',
      borderRadius: 4,
      boxShadow: '20px 20px 48px rgba(74,33,33,0.16), 4px 4px 14px rgba(74,33,33,0.10)',
      transform: 'perspective(1200px) rotateY(-12deg) rotateX(4deg)',
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
      }}>{DF5.areaShort}</div>
      <div style={{ flex: 1, padding: 14, display: 'flex', flexDirection: 'column' }}>
        <div style={{
          fontFamily: 'var(--font-mono)', fontSize: 18,
          color: 'var(--brand-deep)', fontWeight: 600, letterSpacing: '-0.01em',
          marginBottom: 4
        }}>{DF5.code}</div>
        <div style={{
          fontSize: 9.5, color: 'var(--text-soft)',
          letterSpacing: '0.06em', textTransform: 'uppercase', fontWeight: 500,
          fontFamily: 'var(--font-mono)'
        }}>{DF5.typeShort}</div>
        <div style={{ flex: 1 }} />
        <div style={{ height: 1, background: 'var(--border)', marginBottom: 8 }} />
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <span style={{ fontSize: 9.5, color: 'var(--text-muted)',
            fontFamily: 'var(--font-mono)', letterSpacing: '0.04em' }}>{DF5.version}</span>
          <span style={{ width: 5, height: 5, borderRadius: '50%', background: 'var(--success)' }} />
        </div>
      </div>
    </div>
  );
}

// ---------- Hero ----------
function HeroF5() {
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
        <IF5 name="chevron-right" size={10} />
        <a href="#" style={{ color: 'inherit', textDecoration: 'none' }}>{DF5.areaShort}</a>
        <IF5 name="chevron-right" size={10} />
        <a href="#" style={{ color: 'inherit', textDecoration: 'none' }}>{DF5.code}</a>
        <IF5 name="chevron-right" size={10} />
        <span style={{ color: 'var(--text)' }}>Distribuição</span>
      </div>

      <div style={{
        position: 'relative', display: 'grid',
        gridTemplateColumns: '210px 1fr',
        gap: 40, alignItems: 'center'
      }}>
        <div style={{ display: 'flex', justifyContent: 'center' }}>
          <DocRefCardF5 />
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
            }}>{DF5.code} · {DF5.version}</span>
            <span style={{
              display: 'inline-flex', alignItems: 'center', gap: 5,
              height: 22, padding: '0 10px', fontSize: 11, fontWeight: 500,
              background: 'var(--warning-bg, color-mix(in oklab, var(--warning) 14%, transparent))',
              color: 'var(--warning)',
              borderRadius: 'var(--r-pill)', fontFamily: 'var(--font-mono)',
              letterSpacing: '0.04em', textTransform: 'uppercase'
            }}>
              <span style={{ width: 5, height: 5, background: 'currentColor', borderRadius: '50%' }} />
              Vence em {DF5.daysLeft} dias
            </span>
            <span style={{ fontSize: 11, color: 'var(--text-muted)',
              fontFamily: 'var(--font-mono)', letterSpacing: '0.06em', textTransform: 'uppercase' }}>
              §03.05 · Fanout
            </span>
          </div>

          <h1 style={{
            fontSize: 36, lineHeight: 1.1, letterSpacing: '-0.025em',
            margin: '0 0 12px', fontWeight: 600, color: 'var(--text)',
            maxWidth: 640, textWrap: 'balance'
          }}>Distribuição & cobertura de leitura</h1>

          <p style={{
            fontSize: 14.5, lineHeight: 1.55, color: 'var(--text-soft)',
            margin: '0 0 22px', maxWidth: 580
          }}>
            {DF5.title} · publicado em {DF5.publishedAtFull}.
            Reconhecimento obrigatório por assinatura para todas as 248 pessoas alcançadas.
          </p>

          <div style={{ display: 'flex', alignItems: 'center', gap: 8, flexWrap: 'wrap' }}>
            <button className="btn btn-primary btn-lg"
              style={{ height: 42, padding: '0 20px', fontSize: 14, fontWeight: 600 }}>
              <IF5 name="mail" size={15} />Lembrete em massa
            </button>
            <button className="btn">
              <IF5 name="download" size={13} />Exportar relatório
            </button>
            <button className="btn">
              <IF5 name="users" size={13} />Adicionar destinatários
            </button>
            <button className="btn btn-ghost">
              <IF5 name="cog" size={13} />Política de fanout
            </button>
          </div>
        </div>
      </div>
    </header>
  );
}

// ---------- KPI strip ----------
function KPIStripF5() {
  const ackPct = (DF5.acknowledged / DF5.totalTargets) * 100;
  const readPct = (DF5.read / DF5.totalTargets) * 100;
  const animatedAck = useCounterF5(ackPct, 1400);

  const kpis = [
    { label: 'Alvos totais', value: DF5.totalTargets, mono: true, sub: '8 áreas · 1 doc' },
    { label: 'Reconheceram', value: DF5.acknowledged, mono: true,
      sub: `${ackPct.toFixed(1)}% da base`, bar: animatedAck, tone: 'success' },
    { label: 'Apenas leram', value: DF5.read - DF5.acknowledged, mono: true,
      sub: `${(readPct - ackPct).toFixed(1)}% leram sem assinar` },
    { label: 'Pendentes', value: DF5.pending, mono: true, sub: '6 áreas com pendência' },
    { label: 'Em atraso', value: DF5.overdue, mono: true, sub: '+2 vs. ontem', tone: 'danger' }
  ];

  return (
    <div style={{
      display: 'grid', gridTemplateColumns: 'repeat(5, 1fr)',
      background: 'var(--surface)',
      border: '1px solid var(--border)',
      borderRadius: 8,
      overflow: 'hidden'
    }}>
      {kpis.map((k, i) => (
        <div key={k.label} style={{
          padding: '18px 22px',
          borderRight: i < kpis.length - 1 ? '1px solid var(--border)' : 'none'
        }}>
          <div className="kicker" style={{ fontSize: 9.5, marginBottom: 8 }}>{k.label}</div>
          <div style={{
            fontSize: 22, fontWeight: 600, letterSpacing: '-0.02em',
            color: k.tone === 'danger' ? 'var(--danger)' : 'var(--text)',
            lineHeight: 1.1,
            fontFamily: k.mono ? 'var(--font-mono)' : 'var(--font-display)',
            marginBottom: k.bar !== undefined ? 8 : 4
          }}>{k.value}</div>
          {k.bar !== undefined && (
            <div style={{ height: 4, background: 'var(--surface-3)', borderRadius: 999, overflow: 'hidden', marginBottom: 6 }}>
              <div style={{ width: `${k.bar}%`, height: '100%',
                background: k.tone === 'success' ? 'var(--success)' : 'var(--brand)',
                transition: 'width .3s ease-out' }} />
            </div>
          )}
          {k.sub && <div style={{ fontSize: 11, color: 'var(--text-muted)' }}>{k.sub}</div>}
        </div>
      ))}
    </div>
  );
}

// ---------- Section title ----------
function STF5({ kicker, title, aside }) {
  return (
    <div style={{ display: 'flex', alignItems: 'flex-end', justifyContent: 'space-between',
      marginBottom: 18 }}>
      <div>
        <div className="kicker" style={{ fontSize: 9.5, marginBottom: 4 }}>{kicker}</div>
        <h2 style={{ fontSize: 17, lineHeight: 1.2, margin: 0,
          fontWeight: 600, letterSpacing: '-0.01em', color: 'var(--text)' }}>{title}</h2>
      </div>
      {aside}
    </div>
  );
}

// =====================================================================
// SECTION 1: Donut + Distribution facts
// =====================================================================
function DonutCardF5() {
  const [ready, setReady] = usf5(false);
  uef5(() => { const t = setTimeout(() => setReady(true), 220); return () => clearTimeout(t); }, []);
  const ackPct = (DF5.acknowledged / DF5.totalTargets) * 100;
  const readPct = (DF5.read / DF5.totalTargets) * 100;
  const animated = useCounterF5(ready ? Math.round(ackPct) : 0, 1600);
  const C = 2 * Math.PI * 42;

  return (
    <div style={{
      background: 'var(--surface)', border: '1px solid var(--border)',
      borderRadius: 8, padding: 24,
      display: 'flex', flexDirection: 'column', gap: 20
    }}>
      <div className="kicker" style={{ fontSize: 9.5 }}>Cobertura geral</div>
      <div style={{ display: 'flex', alignItems: 'center', gap: 24 }}>
        <div style={{ position: 'relative', width: 140, height: 140, flexShrink: 0 }}>
          <svg viewBox="0 0 100 100" style={{ width: '100%', height: '100%', transform: 'rotate(-90deg)' }}>
            <circle cx="50" cy="50" r="42" fill="none" stroke="var(--surface-3)" strokeWidth="9" />
            <circle cx="50" cy="50" r="42" fill="none" stroke="var(--brand)" strokeWidth="9" strokeLinecap="round" opacity="0.32"
              strokeDasharray={C}
              strokeDashoffset={C * (1 - (ready ? readPct / 100 : 0))}
              style={{ transition: 'stroke-dashoffset 1.6s cubic-bezier(.2,.8,.2,1)' }} />
            <circle cx="50" cy="50" r="42" fill="none" stroke="var(--success)" strokeWidth="9" strokeLinecap="round"
              strokeDasharray={C}
              strokeDashoffset={C * (1 - (ready ? ackPct / 100 : 0))}
              style={{ transition: 'stroke-dashoffset 1.6s cubic-bezier(.2,.8,.2,1) .15s' }} />
          </svg>
          <div style={{ position: 'absolute', inset: 0, display: 'flex',
            flexDirection: 'column', alignItems: 'center', justifyContent: 'center' }}>
            <div style={{
              fontSize: 32, lineHeight: 1, fontWeight: 600,
              fontFamily: 'var(--font-display)', letterSpacing: '-0.02em',
              color: 'var(--success)'
            }}>
              {Math.round(animated)}<span style={{ fontSize: 16, color: 'var(--text-muted)' }}>%</span>
            </div>
            <div className="kicker" style={{ marginTop: 4, fontSize: 9 }}>reconheceu</div>
          </div>
        </div>
        <div style={{ flex: 1, display: 'flex', flexDirection: 'column', gap: 12 }}>
          <LegendRowF5 color="var(--success)" label="Reconheceu (assinou)"
            value={DF5.acknowledged} total={DF5.totalTargets} />
          <LegendRowF5 color="var(--brand)" faded label="Apenas leu"
            value={DF5.read - DF5.acknowledged} total={DF5.totalTargets} />
          <LegendRowF5 color="var(--surface-3)" label="Pendente"
            value={DF5.pending} total={DF5.totalTargets} />
        </div>
      </div>
      <div style={{
        marginTop: 4, paddingTop: 16, borderTop: '1px solid var(--border)',
        display: 'flex', justifyContent: 'space-between', alignItems: 'center',
        fontSize: 11, color: 'var(--text-muted)'
      }}>
        <span>
          Meta de cobertura: <strong style={{ color: 'var(--text)', fontFamily: 'var(--font-mono)' }}>92%</strong>
          {' '}até {DF5.deadline}
        </span>
        <span style={{ display: 'inline-flex', alignItems: 'center', gap: 6 }}>
          <span style={{ width: 6, height: 6, borderRadius: '50%', background: 'var(--warning)' }} />
          {DF5.totalTargets - DF5.acknowledged - 32} pessoas distantes da meta
        </span>
      </div>
    </div>
  );
}

function LegendRowF5({ color, label, value, total, faded }) {
  return (
    <div style={{ display: 'flex', alignItems: 'center', gap: 10, fontSize: 12.5 }}>
      <span style={{ width: 10, height: 10, background: color, borderRadius: 2,
        opacity: faded ? 0.4 : 1, flexShrink: 0 }} />
      <span style={{ flex: 1, color: 'var(--text-soft)' }}>{label}</span>
      <span style={{ fontFamily: 'var(--font-mono)', fontWeight: 500, color: 'var(--text)' }}>
        {value}
      </span>
      <span style={{ minWidth: 40, textAlign: 'right',
        fontFamily: 'var(--font-mono)', fontSize: 11, color: 'var(--text-muted)' }}>
        {Math.round(value / total * 100)}%
      </span>
    </div>
  );
}

function DistributionFactsF5() {
  const facts = [
    { icon: 'calendar',  label: 'Distribuído em',    value: DF5.publishedAtFull },
    { icon: 'clock',     label: 'Prazo',             value: DF5.deadlineFull,
      hint: `${DF5.daysLeft} dias restantes`, tone: 'warning' },
    { icon: 'shield',    label: 'Política',          value: DF5.policy },
    { icon: 'mail',      label: 'Canal',             value: DF5.channel },
    { icon: 'bell',      label: 'Lembretes enviados', value: `${DF5.remindersSent}`, mono: true,
      hint: `próximo: ${DF5.remindersScheduled}` },
    { icon: 'users',     label: 'Grupos',            value: '8 áreas · 11 cargos' }
  ];
  return (
    <div style={{
      background: 'var(--surface)', border: '1px solid var(--border)',
      borderRadius: 8, overflow: 'hidden'
    }}>
      <div style={{
        padding: '14px 22px', borderBottom: '1px solid var(--border)',
        background: 'var(--surface-2)',
        display: 'flex', alignItems: 'center', justifyContent: 'space-between'
      }}>
        <div className="kicker" style={{ fontSize: 9.5 }}>Detalhes da distribuição</div>
        <button className="btn btn-sm btn-ghost">
          <IF5 name="edit" size={11} />Editar política
        </button>
      </div>
      <div style={{ padding: '8px 22px' }}>
        {facts.map((f, i) => (
          <div key={i} style={{
            display: 'grid', gridTemplateColumns: '32px 1fr auto', gap: 14,
            alignItems: 'center',
            padding: '14px 0',
            borderBottom: i < facts.length - 1 ? '1px solid var(--border)' : 'none'
          }}>
            <div style={{
              width: 32, height: 32, borderRadius: 6,
              background: f.tone === 'warning' ? 'color-mix(in oklab, var(--warning) 14%, transparent)' : 'var(--surface-2)',
              color: f.tone === 'warning' ? 'var(--warning)' : 'var(--text-soft)',
              display: 'flex', alignItems: 'center', justifyContent: 'center'
            }}>
              <IF5 name={f.icon} size={14} />
            </div>
            <div style={{ minWidth: 0 }}>
              <div className="kicker" style={{ fontSize: 9, marginBottom: 2 }}>{f.label}</div>
              <div style={{ fontSize: 12.5, color: 'var(--text)', fontWeight: 500,
                fontFamily: f.mono ? 'var(--font-mono)' : 'inherit' }}>
                {f.value}
              </div>
            </div>
            {f.hint && (
              <div style={{ fontSize: 11, color: 'var(--text-muted)',
                fontFamily: 'var(--font-mono)' }}>
                {f.hint}
              </div>
            )}
          </div>
        ))}
      </div>
    </div>
  );
}

// =====================================================================
// SECTION 2: Coverage by area (bullet bars)
// =====================================================================
function AreaRowF5({ a, ready, max }) {
  const pctAck = (a.ack / a.total) * 100;
  const pctRead = (a.read / a.total) * 100;
  const animatedAck = useCounterF5(ready ? pctAck : 0, 1400);
  const animatedRead = useCounterF5(ready ? pctRead : 0, 1400);
  const tone = pctAck > 90 ? 'var(--success)' : pctAck > 60 ? 'var(--brand)' : pctAck > 35 ? 'var(--warning)' : 'var(--danger)';

  return (
    <div style={{ display: 'grid', gridTemplateColumns: '180px 1fr 100px', gap: 16, alignItems: 'center' }}>
      <div style={{ minWidth: 0 }}>
        <div style={{ fontSize: 13, color: 'var(--text)', fontWeight: 500, marginBottom: 2,
          whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }}>{a.area}</div>
        <div style={{ fontSize: 10.5, color: 'var(--text-muted)',
          fontFamily: 'var(--font-mono)', letterSpacing: '0.04em' }}>
          {a.ack}/{a.total} reconheceram
        </div>
      </div>
      <div style={{ position: 'relative', height: 22 }}>
        {/* track */}
        <div style={{
          position: 'absolute', inset: 0, background: 'var(--surface-3)', borderRadius: 4,
          overflow: 'hidden'
        }}>
          {/* read bar (lighter) */}
          <div style={{
            position: 'absolute', top: 0, left: 0, bottom: 0,
            width: `${animatedRead}%`,
            background: 'var(--brand)', opacity: 0.22,
            transition: 'width 1.4s cubic-bezier(.2,.8,.2,1)'
          }} />
          {/* ack bar (solid) */}
          <div style={{
            position: 'absolute', top: 0, left: 0, bottom: 0,
            width: `${animatedAck}%`,
            background: tone,
            transition: 'width 1.4s cubic-bezier(.2,.8,.2,1) .1s'
          }} />
        </div>
        {/* goal marker (92%) */}
        <div style={{
          position: 'absolute', top: -3, bottom: -3, left: '92%',
          borderLeft: '1.5px dashed var(--text-faint)'
        }} />
      </div>
      <div style={{ display: 'flex', alignItems: 'baseline', gap: 6, justifyContent: 'flex-end' }}>
        <span style={{ fontFamily: 'var(--font-mono)', fontSize: 14, fontWeight: 600,
          color: tone, letterSpacing: '-0.01em' }}>
          {Math.round(pctAck)}%
        </span>
        <span style={{ fontSize: 10.5, color: 'var(--text-muted)',
          fontFamily: 'var(--font-mono)' }}>
          ack
        </span>
      </div>
    </div>
  );
}

function CoverageByAreaF5() {
  const [ready, setReady] = usf5(false);
  uef5(() => { const t = setTimeout(() => setReady(true), 350); return () => clearTimeout(t); }, []);
  const sorted = umf5(() => [...DF5.byArea].sort((a, b) => (a.ack / a.total) - (b.ack / b.total)), []);

  return (
    <div style={{
      background: 'var(--surface)', border: '1px solid var(--border)',
      borderRadius: 8, overflow: 'hidden'
    }}>
      <div style={{
        padding: '14px 22px', borderBottom: '1px solid var(--border)',
        background: 'var(--surface-2)',
        display: 'flex', alignItems: 'center', justifyContent: 'space-between'
      }}>
        <span style={{ fontSize: 11, color: 'var(--text-muted)' }}>
          Áreas ordenadas por menor cobertura primeiro · linha tracejada = meta de 92%
        </span>
        <div style={{ display: 'flex', gap: 14, fontSize: 10.5, color: 'var(--text-muted)' }}>
          <span style={{ display: 'inline-flex', alignItems: 'center', gap: 5 }}>
            <span style={{ width: 10, height: 10, background: 'var(--success)', borderRadius: 2 }} />Reconheceu
          </span>
          <span style={{ display: 'inline-flex', alignItems: 'center', gap: 5 }}>
            <span style={{ width: 10, height: 10, background: 'var(--brand)', opacity: 0.22, borderRadius: 2 }} />Apenas leu
          </span>
          <span style={{ display: 'inline-flex', alignItems: 'center', gap: 5 }}>
            <span style={{ width: 10, height: 10, border: '1px dashed var(--text-faint)' }} />Meta 92%
          </span>
        </div>
      </div>
      <div style={{ padding: '20px 22px', display: 'flex', flexDirection: 'column', gap: 14 }}>
        {sorted.map((a, i) => <AreaRowF5 key={i} a={a} ready={ready} />)}
      </div>
    </div>
  );
}

// =====================================================================
// SECTION 3: Cumulative ack timeline
// =====================================================================
function TimelineCardF5() {
  const [ready, setReady] = usf5(false);
  uef5(() => { const t = setTimeout(() => setReady(true), 350); return () => clearTimeout(t); }, []);

  const data = DF5.dailyReads;
  const goal = Math.round(DF5.totalTargets * 0.92); // 228
  const W = 1100, H = 180, padL = 50, padR = 30, padT = 20, padB = 30;
  const innerW = W - padL - padR, innerH = H - padT - padB;
  const xs = (i) => padL + (i / (data.length - 1)) * innerW;
  const ys = (v) => padT + innerH - (v / DF5.totalTargets) * innerH;

  const linePath = data.map((d, i) => `${i === 0 ? 'M' : 'L'} ${xs(i)} ${ys(d.cumAck)}`).join(' ');
  const areaPath = `${linePath} L ${xs(data.length - 1)} ${padT + innerH} L ${xs(0)} ${padT + innerH} Z`;

  return (
    <div style={{
      background: 'var(--surface)', border: '1px solid var(--border)',
      borderRadius: 8, padding: '20px 22px'
    }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'baseline', marginBottom: 8 }}>
        <div style={{ fontSize: 11, color: 'var(--text-muted)' }}>
          Reconhecimentos acumulados desde a publicação
        </div>
        <div style={{ display: 'flex', gap: 16, fontSize: 10.5, color: 'var(--text-muted)' }}>
          <span style={{ display: 'inline-flex', alignItems: 'center', gap: 5 }}>
            <span style={{ width: 16, height: 2, background: 'var(--brand)' }} />Reconhecimentos
          </span>
          <span style={{ display: 'inline-flex', alignItems: 'center', gap: 5 }}>
            <span style={{ width: 16, height: 0, borderTop: '2px dashed var(--text-faint)' }} />Meta · 228 (92%)
          </span>
        </div>
      </div>
      <svg viewBox={`0 0 ${W} ${H}`} style={{ width: '100%', height: H, display: 'block' }}>
        <defs>
          <linearGradient id="trendGradF5" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stopColor="var(--brand)" stopOpacity="0.18" />
            <stop offset="100%" stopColor="var(--brand)" stopOpacity="0" />
          </linearGradient>
          <clipPath id="revealF5">
            <rect x={padL} y="0" width={ready ? innerW : 0} height={H}
              style={{ transition: 'width 1.6s cubic-bezier(.2,.8,.2,1)' }} />
          </clipPath>
        </defs>

        {/* Y grid lines */}
        {[0, 0.25, 0.5, 0.75, 1].map((p, i) => (
          <line key={i} x1={padL} x2={W - padR}
            y1={padT + innerH * (1 - p)} y2={padT + innerH * (1 - p)}
            stroke="var(--border)" strokeWidth="1" strokeDasharray={i === 0 ? '' : '2 4'} />
        ))}
        {[0, 0.25, 0.5, 0.75, 1].map((p, i) => (
          <text key={`yl-${i}`} x={padL - 8} y={padT + innerH * (1 - p) + 3}
            fontSize="9.5" fill="var(--text-muted)" fontFamily="var(--font-mono)" textAnchor="end">
            {Math.round(DF5.totalTargets * p)}
          </text>
        ))}

        {/* Goal line */}
        <line x1={padL} x2={W - padR} y1={ys(goal)} y2={ys(goal)}
          stroke="var(--text-faint)" strokeWidth="1.5" strokeDasharray="4 4" />
        <text x={W - padR - 4} y={ys(goal) - 4}
          fontSize="9.5" fill="var(--text-muted)" fontFamily="var(--font-mono)" textAnchor="end">
          meta · 228
        </text>

        <g clipPath="url(#revealF5)">
          {/* area */}
          <path d={areaPath} fill="url(#trendGradF5)" />
          {/* line */}
          <path d={linePath} fill="none" stroke="var(--brand)" strokeWidth="2" />
          {/* points */}
          {data.map((d, i) => (
            <circle key={i} cx={xs(i)} cy={ys(d.cumAck)} r={i === data.length - 1 ? 4.5 : 2.8}
              fill="var(--brand)" stroke="var(--surface)" strokeWidth={i === data.length - 1 ? 3 : 1.5} />
          ))}
        </g>

        {/* X labels */}
        {data.map((d, i) => (
          <text key={`x-${i}`} x={xs(i)} y={H - 10}
            fontSize="9.5" fill="var(--text-muted)" fontFamily="var(--font-mono)" textAnchor="middle">
            {d.day}
          </text>
        ))}

        {/* Last value label */}
        <g transform={`translate(${xs(data.length - 1) - 60}, ${ys(data[data.length - 1].cumAck) - 28})`}>
          <rect width="56" height="20" rx="4" fill="var(--brand)" />
          <text x="28" y="13" textAnchor="middle"
            fontSize="11" fontWeight="600" fill="#fff" fontFamily="var(--font-mono)">
            +{data[data.length - 1].cumAck - data[data.length - 2].cumAck} hoje
          </text>
        </g>
      </svg>
    </div>
  );
}

// =====================================================================
// SECTION 4: Recipient list
// =====================================================================
function StatusPillF5({ p }) {
  const key = p.overdue ? 'overdue' : p.status;
  const t = STATUS_TONE[key];
  return (
    <span style={{
      display: 'inline-flex', alignItems: 'center', gap: 5,
      height: 20, padding: '0 8px',
      fontSize: 10.5, fontWeight: 500,
      background: t.bg, color: t.fg,
      borderRadius: 'var(--r-pill)',
      fontFamily: 'var(--font-mono)', letterSpacing: '0.04em', textTransform: 'uppercase'
    }}>
      <span style={{ width: 5, height: 5, background: 'currentColor', borderRadius: '50%' }} />
      {t.label}
    </span>
  );
}

function RecipientRowF5({ p, last, selected, onToggle }) {
  return (
    <div style={{
      display: 'grid', gridTemplateColumns: '32px 1.6fr 1.1fr 1fr 130px 130px 90px',
      gap: 14, alignItems: 'center',
      padding: '12px 22px',
      borderBottom: last ? 'none' : '1px solid var(--border)',
      background: selected ? 'var(--brand-pale)' : 'transparent',
      transition: 'background .12s'
    }}>
      <input type="checkbox" checked={selected} onChange={onToggle}
        style={{ accentColor: 'var(--brand)' }} />
      <div style={{ display: 'flex', alignItems: 'center', gap: 10, minWidth: 0 }}>
        <AF5 name={p.who} size={28} />
        <div style={{ minWidth: 0 }}>
          <div style={{ fontSize: 13, color: 'var(--text)', fontWeight: 500,
            whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }}>{p.who}</div>
          <div style={{ fontSize: 10.5, color: 'var(--text-muted)',
            fontFamily: 'var(--font-mono)', letterSpacing: '0.04em' }}>
            {p.role}
          </div>
        </div>
      </div>
      <span style={{ fontSize: 12, color: 'var(--text-soft)' }}>{p.area}</span>
      <StatusPillF5 p={p} />
      <span style={{ fontSize: 12, color: p.last.includes('nunca') ? 'var(--danger)' : 'var(--text-soft)' }}>
        {p.last}
      </span>
      <span style={{ fontSize: 11, color: 'var(--text-muted)',
        fontFamily: 'var(--font-mono)' }}>{p.when}</span>
      <div style={{ display: 'flex', gap: 4, justifyContent: 'flex-end' }}>
        <button className="btn btn-sm btn-ghost" title="Enviar lembrete">
          <IF5 name="mail" size={11} />
        </button>
        <button className="btn btn-sm btn-ghost" title="Ver perfil">
          <IF5 name="more" size={11} />
        </button>
      </div>
    </div>
  );
}

function RecipientsCardF5() {
  const [tab, setTab] = usf5('pending');
  const [selected, setSelected] = usf5(new Set());
  const [search, setSearch] = usf5('');

  const tabCounts = umf5(() => ({
    pending: DF5.people.filter(p => p.status === 'pending' && !p.overdue).length,
    overdue: DF5.people.filter(p => p.overdue).length,
    read:    DF5.people.filter(p => p.status === 'read').length,
    ack:     DF5.people.filter(p => p.status === 'ack').length,
    all:     DF5.people.length
  }), []);

  const filtered = umf5(() => {
    let list = DF5.people;
    if (tab === 'pending') list = list.filter(p => p.status === 'pending' && !p.overdue);
    else if (tab === 'overdue') list = list.filter(p => p.overdue);
    else if (tab === 'read') list = list.filter(p => p.status === 'read');
    else if (tab === 'ack') list = list.filter(p => p.status === 'ack');
    if (search) {
      const q = search.toLowerCase();
      list = list.filter(p => p.who.toLowerCase().includes(q) || p.area.toLowerCase().includes(q) || p.role.toLowerCase().includes(q));
    }
    return list;
  }, [tab, search]);

  const toggleAll = () => {
    if (selected.size === filtered.length) setSelected(new Set());
    else setSelected(new Set(filtered.map(p => p.who)));
  };
  const togglePerson = (who) => {
    const next = new Set(selected);
    if (next.has(who)) next.delete(who);
    else next.add(who);
    setSelected(next);
  };

  const tabs = [
    { id: 'pending', label: 'Pendentes',     n: tabCounts.pending },
    { id: 'overdue', label: 'Em atraso',     n: tabCounts.overdue, tone: 'var(--danger)' },
    { id: 'read',    label: 'Apenas leram',  n: tabCounts.read },
    { id: 'ack',     label: 'Reconheceram',  n: tabCounts.ack,     tone: 'var(--success)' },
    { id: 'all',     label: 'Todos',         n: tabCounts.all }
  ];

  const hasSelection = selected.size > 0;

  return (
    <div style={{
      background: 'var(--surface)', border: '1px solid var(--border)',
      borderRadius: 8, overflow: 'hidden'
    }}>
      {/* Filter bar */}
      <div style={{
        padding: '14px 22px', borderBottom: '1px solid var(--border)',
        background: 'var(--surface-2)',
        display: 'flex', alignItems: 'center', gap: 12, flexWrap: 'wrap'
      }}>
        <div style={{
          display: 'inline-flex', background: 'var(--surface)',
          border: '1px solid var(--border)', borderRadius: 6, padding: 3
        }}>
          {tabs.map(t => (
            <button key={t.id} onClick={() => setTab(t.id)} style={{
              height: 28, padding: '0 12px', border: 'none', cursor: 'pointer',
              background: tab === t.id ? 'var(--surface-2)' : 'transparent',
              boxShadow: tab === t.id ? 'var(--shadow-1)' : 'none',
              borderRadius: 4, fontSize: 12, fontWeight: tab === t.id ? 500 : 400,
              color: tab === t.id ? 'var(--text)' : 'var(--text-muted)',
              display: 'inline-flex', alignItems: 'center', gap: 6
            }}>
              {t.label}
              <span style={{
                fontFamily: 'var(--font-mono)', fontSize: 10,
                padding: '1px 6px', borderRadius: 3,
                background: tab === t.id ? (t.tone || 'var(--brand-pale)') : 'var(--surface-3)',
                color: tab === t.id ? (t.tone ? '#fff' : 'var(--brand)') : 'var(--text-muted)'
              }}>{t.n}</span>
            </button>
          ))}
        </div>
        <div style={{
          display: 'inline-flex', alignItems: 'center', gap: 8,
          height: 32, padding: '0 12px',
          background: 'var(--surface)', border: '1px solid var(--border)',
          borderRadius: 6, color: 'var(--text-muted)',
          width: 240, marginLeft: 'auto'
        }}>
          <IF5 name="search" size={13} />
          <input value={search} onChange={(e) => setSearch(e.target.value)}
            placeholder="Filtrar por nome, área, cargo…"
            style={{ flex: 1, border: 'none', background: 'transparent', outline: 'none',
              fontSize: 12, color: 'var(--text)' }} />
        </div>
        <button className="btn btn-sm">
          <IF5 name="filter" size={11} />Filtros avançados
        </button>
      </div>

      {/* Bulk action bar (when selection exists) */}
      {hasSelection && (
        <div style={{
          padding: '10px 22px',
          background: 'var(--brand-pale)',
          borderBottom: '1px solid var(--border)',
          display: 'flex', alignItems: 'center', gap: 10
        }}>
          <span style={{ fontSize: 12, color: 'var(--brand-deep)', fontWeight: 500 }}>
            {selected.size} selecionado{selected.size > 1 ? 's' : ''}
          </span>
          <button className="btn btn-sm btn-primary">
            <IF5 name="mail" size={11} />Enviar lembrete
          </button>
          <button className="btn btn-sm">
            <IF5 name="users" size={11} />Reatribuir
          </button>
          <button className="btn btn-sm">
            <IF5 name="download" size={11} />Exportar seleção
          </button>
          <button className="btn btn-sm btn-ghost" onClick={() => setSelected(new Set())}
            style={{ marginLeft: 'auto' }}>
            <IF5 name="x" size={11} />Limpar seleção
          </button>
        </div>
      )}

      {/* Header row */}
      <div style={{
        display: 'grid', gridTemplateColumns: '32px 1.6fr 1.1fr 1fr 130px 130px 90px',
        gap: 14, padding: '10px 22px',
        background: 'var(--surface-2)', borderBottom: '1px solid var(--border)',
        fontSize: 9.5, color: 'var(--text-muted)',
        textTransform: 'uppercase', letterSpacing: '0.06em', fontWeight: 500,
        fontFamily: 'var(--font-mono)'
      }}>
        <span>
          <input type="checkbox"
            checked={selected.size === filtered.length && filtered.length > 0}
            onChange={toggleAll}
            style={{ accentColor: 'var(--brand)' }} />
        </span>
        <span>Destinatário</span>
        <span>Área</span>
        <span>Status</span>
        <span>Última atividade</span>
        <span>Quando</span>
        <span></span>
      </div>

      {/* Rows */}
      {filtered.length === 0 ? (
        <div style={{ padding: 48, textAlign: 'center', color: 'var(--text-muted)', fontSize: 13 }}>
          Nenhum destinatário encontrado nesta visualização.
        </div>
      ) : filtered.map((p, i) => (
        <RecipientRowF5 key={p.who} p={p} last={i === filtered.length - 1}
          selected={selected.has(p.who)} onToggle={() => togglePerson(p.who)} />
      ))}

      {/* Footer */}
      <div style={{
        padding: '12px 22px', background: 'var(--surface-2)',
        borderTop: '1px solid var(--border)',
        display: 'flex', justifyContent: 'space-between', alignItems: 'center',
        fontSize: 11, color: 'var(--text-muted)'
      }}>
        <span>
          Mostrando <strong style={{ color: 'var(--text)' }}>{filtered.length}</strong> de
          {' '}<strong style={{ color: 'var(--text)' }}>{DF5.totalTargets}</strong> destinatários
        </span>
        <span style={{ display: 'inline-flex', gap: 4 }}>
          <button className="btn btn-sm btn-ghost"><IF5 name="chevron-left" size={12} /></button>
          <span style={{ display: 'inline-flex', alignItems: 'center', padding: '0 10px',
            fontSize: 11, fontFamily: 'var(--font-mono)' }}>1 / 23</span>
          <button className="btn btn-sm btn-ghost"><IF5 name="chevron-right" size={12} /></button>
        </span>
      </div>
    </div>
  );
}

// =====================================================================
// MAIN PAGE
// =====================================================================
function FanoutV5() {
  return (
    <div className="palette-wine" style={{
      height: 1200, display: 'flex', background: 'var(--bg)', position: 'relative'
    }}>
      <SF5 variant="rail" active="documents" />
      <div style={{ flex: 1, overflow: 'hidden', display: 'flex', flexDirection: 'column' }}>
        <div style={{ flex: 1, overflow: 'auto' }}>
          <HeroF5 />

          <div style={{ padding: '32px 56px 80px', maxWidth: 1180, margin: '0 auto',
            display: 'flex', flexDirection: 'column', gap: 36 }}>

            <KPIStripF5 />

            {/* §01 — Cobertura geral + detalhes */}
            <section>
              <STF5 kicker="01 · Status" title="Cobertura geral e detalhes da distribuição" />
              <div style={{ display: 'grid', gridTemplateColumns: '1.4fr 1fr', gap: 20 }}>
                <DonutCardF5 />
                <DistributionFactsF5 />
              </div>
            </section>

            {/* §02 — Por área */}
            <section>
              <STF5 kicker="02 · Por área"
                title="Onde está a pendência"
                aside={<a href="#" style={{ fontSize: 12, color: 'var(--brand)',
                  textDecoration: 'none' }}>Lembrar todas as áreas críticas →</a>} />
              <CoverageByAreaF5 />
            </section>

            {/* §03 — Linha do tempo */}
            <section>
              <STF5 kicker="03 · Linha do tempo"
                title="Curva de adoção desde a publicação"
                aside={<span style={{ fontSize: 11, color: 'var(--text-muted)',
                  fontFamily: 'var(--font-mono)', letterSpacing: '0.04em', textTransform: 'uppercase' }}>
                  últimos 8 dias
                </span>} />
              <TimelineCardF5 />
            </section>

            {/* §04 — Lista detalhada */}
            <section>
              <STF5 kicker="04 · Destinatários"
                title="Lista detalhada"
                aside={<span style={{ fontSize: 11, color: 'var(--text-muted)' }}>
                  selecione para ações em massa
                </span>} />
              <RecipientsCardF5 />
            </section>
          </div>
        </div>
      </div>
    </div>
  );
}

window.MD_ONDA1_FANOUT_V5 = { FanoutV5 };
