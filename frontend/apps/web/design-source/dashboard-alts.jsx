/* global React, MD */
const { Icon, Avatar } = window.MD;
const { useState, useEffect, useRef } = React;

/* ============================================================
   Shared CSS animations (injected once)
   ============================================================ */
const ALT_CSS = `
@keyframes md-pulse-soft { 0%,100% { opacity: 0.7; } 50% { opacity: 1; } }
@keyframes md-shimmer { 0% { transform: translateX(-100%); } 100% { transform: translateX(200%); } }
@keyframes md-fade-up { from { opacity: 0; transform: translateY(8px); } to { opacity: 1; transform: translateY(0); } }
@keyframes md-slide-in { from { transform: translateX(-12px); opacity: 0; } to { transform: translateX(0); opacity: 1; } }
@keyframes md-tick { 0% { width: 0; } 100% { width: var(--w); } }
@keyframes md-bg-pan { 0% { background-position: 0 0; } 100% { background-position: 40px 40px; } }
@keyframes md-blink { 0%,100% { opacity: 1; } 50% { opacity: 0.2; } }
@keyframes md-rise { from { transform: translateY(20px); opacity: 0; } to { transform: translateY(0); opacity: 1; } }
@keyframes md-count { from { opacity: 0; transform: scale(0.95); } to { opacity: 1; transform: scale(1); } }

.md-pulse { animation: md-pulse-soft 2s ease-in-out infinite; }
.md-fade-up { animation: md-fade-up 0.5s ease-out backwards; }
.md-rise { animation: md-rise 0.6s cubic-bezier(0.16, 1, 0.3, 1) backwards; }
.md-blink-dot { animation: md-blink 1.4s ease-in-out infinite; }

.md-shimmer-track { position: relative; overflow: hidden; }
.md-shimmer-track::after {
  content: ''; position: absolute; inset: 0;
  background: linear-gradient(90deg, transparent, rgba(255,255,255,0.18), transparent);
  animation: md-shimmer 2.5s ease-in-out infinite;
}

.md-counter { transition: color 0.3s ease; }
.md-row-hover { transition: background 0.18s ease, transform 0.18s ease; }
.md-row-hover:hover { background: var(--surface-2); }

.md-card-lift { transition: transform 0.25s cubic-bezier(0.16, 1, 0.3, 1), box-shadow 0.25s; }
.md-card-lift:hover { transform: translateY(-2px); box-shadow: 0 12px 32px rgba(0,0,0,0.08); }

.md-grid-bg {
  background-image:
    linear-gradient(var(--grid-line) 1px, transparent 1px),
    linear-gradient(90deg, var(--grid-line) 1px, transparent 1px);
  background-size: 24px 24px;
}

.md-noise {
  background-image: url("data:image/svg+xml,%3Csvg viewBox='0 0 200 200' xmlns='http://www.w3.org/2000/svg'%3E%3Cfilter id='n'%3E%3CfeTurbulence type='fractalNoise' baseFrequency='0.9' numOctaves='3' /%3E%3C/filter%3E%3Crect width='100%25' height='100%25' filter='url(%23n)' opacity='0.5'/%3E%3C/svg%3E");
}
`;

const StyleInject = () => {
  React.useEffect(() => {
    if (document.getElementById('md-alt-css')) return;
    const s = document.createElement('style');
    s.id = 'md-alt-css';
    s.textContent = ALT_CSS;
    document.head.appendChild(s);
  }, []);
  return null;
};

/* Animated counter — counts from 0 to value */
const AnimatedNum = ({ value, duration = 800, fontSize = 32, color }) => {
  const [n, setN] = useState(0);
  useEffect(() => {
    let raf, start;
    const step = (t) => {
      if (!start) start = t;
      const p = Math.min(1, (t - start) / duration);
      // ease-out cubic
      const eased = 1 - Math.pow(1 - p, 3);
      setN(Math.round(value * eased));
      if (p < 1) raf = requestAnimationFrame(step);
    };
    raf = requestAnimationFrame(step);
    return () => cancelAnimationFrame(raf);
  }, [value, duration]);
  return <span className="num" style={{ fontSize, color, fontVariantNumeric: 'tabular-nums', fontWeight: 600, letterSpacing: '-0.02em' }}>{n}</span>;
};

/* Live time updating */
const LiveTime = () => {
  const [t, setT] = useState(new Date());
  useEffect(() => { const id = setInterval(() => setT(new Date()), 1000); return () => clearInterval(id); }, []);
  return <span className="mono" style={{ fontSize: 11, color: 'var(--text-muted)' }}>{t.toLocaleTimeString('pt-BR')}</span>;
};

/* Sparkline mini-chart */
const Spark = ({ data, color = 'var(--brand)', height = 28, width = 100 }) => {
  const max = Math.max(...data);
  const min = Math.min(...data);
  const pts = data.map((v, i) => {
    const x = (i / (data.length - 1)) * width;
    const y = height - ((v - min) / (max - min || 1)) * (height - 4) - 2;
    return `${x},${y}`;
  }).join(' ');
  return (
    <svg width={width} height={height} style={{ display: 'block' }}>
      <polyline points={pts} fill="none" stroke={color} strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" style={{ filter: 'drop-shadow(0 1px 2px rgba(0,0,0,0.08))' }}/>
      <circle cx={width} cy={height - ((data[data.length-1] - min) / (max - min || 1)) * (height - 4) - 2} r="2.5" fill={color}/>
    </svg>
  );
};

/* ============================================================
   DASHBOARD · ALT 1 — EDITORIAL / MAGAZINE
   Big headline, asymmetric grid, deep brand sidebar, less chrome
   ============================================================ */
const DashboardEditorial = () => {
  return (
    <div style={{ flex: 1, overflow: 'auto', background: 'var(--bg)', position: 'relative' }}>
      <StyleInject/>

      {/* Hero — full-bleed dark editorial band */}
      <div style={{
        background: 'linear-gradient(135deg, var(--brand-deep) 0%, var(--brand) 65%, var(--brand-soft) 100%)',
        color: 'white', position: 'relative', overflow: 'hidden',
        padding: '40px 48px 56px',
      }}>
        {/* Decorative grid */}
        <div style={{ position: 'absolute', inset: 0, opacity: 0.08,
          backgroundImage: 'linear-gradient(rgba(255,255,255,0.4) 1px, transparent 1px), linear-gradient(90deg, rgba(255,255,255,0.4) 1px, transparent 1px)',
          backgroundSize: '32px 32px', animation: 'md-bg-pan 30s linear infinite'
        }}/>
        {/* Decorative number — huge */}
        <div style={{ position: 'absolute', right: -20, top: -30, fontSize: 320, fontWeight: 700,
          color: 'rgba(255,255,255,0.04)', lineHeight: 1, fontFamily: 'var(--font-display)', letterSpacing: '-0.05em' }}>
          04
        </div>

        <div style={{ position: 'relative', maxWidth: 980 }}>
          <div className="md-fade-up" style={{ display: 'flex', alignItems: 'center', gap: 12, marginBottom: 18, animationDelay: '0ms' }}>
            <span style={{ fontFamily: 'var(--font-mono)', fontSize: 11, letterSpacing: '0.16em', color: 'rgba(255,255,255,0.7)' }}>
              ED. SEX 04 · MAI · 2026 · VOL.III
            </span>
            <span style={{ width: 6, height: 6, borderRadius: '50%', background: 'var(--accent)' }} className="md-blink-dot"/>
            <span style={{ fontFamily: 'var(--font-mono)', fontSize: 11, color: 'rgba(255,255,255,0.6)' }}>
              LIVE · <LiveTime/>
            </span>
          </div>

          <h1 className="md-fade-up" style={{
            fontSize: 64, lineHeight: 1.02, letterSpacing: '-0.035em', fontWeight: 600,
            color: 'white', marginBottom: 20, animationDelay: '80ms', textWrap: 'pretty'
          }}>
            Bom dia,<br/>
            <span style={{ fontStyle: 'italic', fontWeight: 400, color: 'rgba(255,255,255,0.85)' }}>Marina.</span>
            <span style={{ display: 'block', fontSize: 28, fontWeight: 400, marginTop: 14, color: 'rgba(255,255,255,0.75)', letterSpacing: '-0.02em' }}>
              Hoje você tem <span style={{ color: 'var(--accent)', fontWeight: 600, position: 'relative' }}>
                quatro decisões
                <span style={{ position: 'absolute', left: 0, right: 0, bottom: -2, height: 2, background: 'var(--accent)', opacity: 0.6 }}/>
              </span> que param na sua mesa.
            </span>
          </h1>

          <div className="md-fade-up" style={{ display: 'flex', gap: 28, marginTop: 32, animationDelay: '160ms', flexWrap: 'wrap' }}>
            {[
              { label: 'Vence em', val: '3h 28min', sub: 'POP-QUA-0148', accent: true },
              { label: 'Aprovados (sua semana)', val: '12', sub: '↑ 33% vs. semana passada' },
              { label: 'Cobertura ISO', val: '94.2%', sub: 'cl. 7.5 · documentos vivos' },
            ].map((s, i) => (
              <div key={i} style={{ borderLeft: '1px solid rgba(255,255,255,0.2)', paddingLeft: 16 }}>
                <div style={{ fontFamily: 'var(--font-mono)', fontSize: 10, letterSpacing: '0.1em', color: 'rgba(255,255,255,0.55)', marginBottom: 4 }}>{s.label.toUpperCase()}</div>
                <div style={{ fontSize: 22, fontWeight: 600, color: s.accent ? 'var(--accent)' : 'white', marginBottom: 2, fontFamily: 'var(--font-mono)', letterSpacing: '-0.02em' }}>{s.val}</div>
                <div style={{ fontSize: 11, color: 'rgba(255,255,255,0.6)', fontFamily: 'var(--font-mono)' }}>{s.sub}</div>
              </div>
            ))}
          </div>
        </div>
      </div>

      {/* Body — asymmetric editorial grid */}
      <div style={{ padding: '40px 48px 48px', display: 'grid', gridTemplateColumns: '1fr 1fr 1fr', gap: 28 }}>

        {/* === Lead story (2 cols) === */}
        <article style={{ gridColumn: 'span 2' }} className="md-rise" >
          <div style={{ display: 'flex', alignItems: 'baseline', gap: 12, marginBottom: 16, paddingBottom: 12, borderBottom: '2px solid var(--text)' }}>
            <span style={{ fontSize: 11, fontFamily: 'var(--font-mono)', fontWeight: 700, letterSpacing: '0.14em', color: 'var(--text)' }}>§ AGUARDANDO SUA ASSINATURA</span>
            <span style={{ flex: 1 }}/>
            <span className="caption mono" style={{ color: 'var(--text-muted)' }}>4 de 4 visíveis</span>
          </div>

          <div style={{ display: 'flex', flexDirection: 'column', gap: 0 }}>
            {[
              { code: 'POP-QUA-0148', title: 'Inspeção de recebimento — solda MIG/MAG', author: 'Carla Pinheiro', area: 'Qualidade', deadline: '3h 28min', urgent: true, kind: 'POP', summary: 'Adiciona §3.2 sobre inspeção visual ampliada NBR ISO 5817. Etapa final.' },
              { code: 'IT-PROD-0072', title: 'Setup da prensa hidráulica P-440', author: 'João Aguiar', area: 'Produção', deadline: '9h', urgent: true, kind: 'IT', summary: 'Procedimento revisado após incidente de 12/abr. Aprovação técnica.' },
              { code: 'POL-RH-0011', title: 'Política de afastamento parcial', author: 'Fábio Ribas', area: 'Recursos Humanos', deadline: '1 dia', urgent: false, kind: 'POL', summary: 'Atualização para incluir modalidade híbrida pós-licença médica.' },
              { code: 'DC-TI-0203', title: 'Diagrama de rede — DMZ planta 2', author: 'Ana Iuri', area: 'TI & Segurança', deadline: '2 dias', urgent: false, kind: 'DC', summary: 'Reflete a segregação para ICS/SCADA aprovada no Q1.' },
            ].map((r, i) => (
              <div key={i} className="md-row-hover md-fade-up" style={{
                display: 'grid', gridTemplateColumns: '60px 1fr auto', gap: 18,
                padding: '18px 4px', borderBottom: '1px solid var(--border)',
                alignItems: 'flex-start', cursor: 'pointer', position: 'relative',
                animationDelay: `${i * 80 + 200}ms`,
              }}>
                <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', gap: 4, paddingTop: 4 }}>
                  <span style={{ fontFamily: 'var(--font-mono)', fontSize: 28, fontWeight: 700, color: 'var(--text)', lineHeight: 1, letterSpacing: '-0.02em' }}>0{i+1}</span>
                  <span style={{ fontFamily: 'var(--font-mono)', fontSize: 9, color: 'var(--text-muted)', letterSpacing: '0.1em' }}>{r.kind}</span>
                </div>
                <div>
                  <div style={{ fontSize: 11, fontFamily: 'var(--font-mono)', fontWeight: 600, color: 'var(--brand)', marginBottom: 4 }}>{r.code}</div>
                  <h3 style={{ fontSize: 19, lineHeight: 1.25, letterSpacing: '-0.018em', fontWeight: 600, marginBottom: 6, color: 'var(--text)', textWrap: 'pretty' }}>{r.title}</h3>
                  <p style={{ fontSize: 12.5, lineHeight: 1.55, color: 'var(--text-muted)', marginBottom: 10, maxWidth: 540 }}>{r.summary}</p>
                  <div style={{ display: 'flex', alignItems: 'center', gap: 10, fontSize: 11, color: 'var(--text-muted)' }}>
                    <Avatar name={r.author} size="sm"/>
                    <span style={{ color: 'var(--text-soft)', fontWeight: 500 }}>{r.author}</span>
                    <span style={{ color: 'var(--text-faint)' }}>—</span>
                    <span style={{ fontStyle: 'italic' }}>{r.area}</span>
                  </div>
                </div>
                <div style={{ textAlign: 'right', paddingTop: 6 }}>
                  <div style={{
                    fontFamily: 'var(--font-mono)', fontSize: 13, fontWeight: 600, marginBottom: 8,
                    color: r.urgent ? 'var(--accent)' : 'var(--text-soft)',
                  }}>
                    {r.urgent && <span className="md-blink-dot" style={{ display: 'inline-block', width: 6, height: 6, borderRadius: '50%', background: 'var(--accent)', marginRight: 6, verticalAlign: 'middle' }}/>}
                    vence em {r.deadline}
                  </div>
                  <button className="btn btn-sm" style={{ background: 'var(--text)', color: 'var(--surface)', border: 'none' }}>
                    Revisar →
                  </button>
                </div>
              </div>
            ))}
          </div>
        </article>

        {/* === Side column === */}
        <aside style={{ display: 'flex', flexDirection: 'column', gap: 28 }}>

          {/* Pulse — your week */}
          <div className="md-rise" style={{ animationDelay: '300ms' }}>
            <div style={{ display: 'flex', alignItems: 'baseline', gap: 8, marginBottom: 14, paddingBottom: 10, borderBottom: '2px solid var(--text)' }}>
              <span style={{ fontSize: 11, fontFamily: 'var(--font-mono)', fontWeight: 700, letterSpacing: '0.14em' }}>§ SEU PULSO</span>
              <span style={{ flex: 1 }}/>
              <span className="caption mono" style={{ color: 'var(--text-faint)' }}>5 dias úteis</span>
            </div>
            <div style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>
              {[
                { label: 'Aprovados', val: 12, data: [4, 6, 5, 8, 7, 12], color: 'var(--success)' },
                { label: 'Devolvidos', val: 3, data: [1, 2, 0, 1, 2, 3], color: 'var(--warning)' },
                { label: 'Tempo médio decisão', val: '1h 24min', data: [3, 2.5, 2, 1.8, 1.4, 1.4], color: 'var(--brand)' },
              ].map((m, i) => (
                <div key={i} style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
                  <div style={{ flex: 1 }}>
                    <div style={{ fontSize: 11, fontFamily: 'var(--font-mono)', color: 'var(--text-muted)', letterSpacing: '0.06em', marginBottom: 2 }}>{m.label.toUpperCase()}</div>
                    <div style={{ fontSize: 22, fontWeight: 600, fontFamily: 'var(--font-mono)', letterSpacing: '-0.02em' }}>{m.val}</div>
                  </div>
                  <Spark data={m.data} color={m.color} width={80}/>
                </div>
              ))}
            </div>
          </div>

          {/* Activity — vertical timeline */}
          <div className="md-rise" style={{ animationDelay: '380ms' }}>
            <div style={{ display: 'flex', alignItems: 'baseline', gap: 8, marginBottom: 14, paddingBottom: 10, borderBottom: '2px solid var(--text)' }}>
              <span style={{ fontSize: 11, fontFamily: 'var(--font-mono)', fontWeight: 700, letterSpacing: '0.14em' }}>§ MURMÚRIOS</span>
              <span style={{ flex: 1 }}/>
              <span className="caption mono" style={{ color: 'var(--text-faint)' }}><span className="md-blink-dot" style={{ display: 'inline-block', width: 5, height: 5, borderRadius: '50%', background: 'var(--success)', marginRight: 4 }}/>ao vivo</span>
            </div>
            <div style={{ position: 'relative', paddingLeft: 18 }}>
              <span style={{ position: 'absolute', left: 4, top: 4, bottom: 4, width: 1, background: 'var(--border)' }}/>
              {[
                { who: 'Carla Pinheiro', what: 'enviou', code: 'POP-QUA-0148', t: 'há 32min' },
                { who: 'Bruno Said', what: 'aprovou', code: 'IT-QUA-0091', t: 'há 3h' },
                { who: 'Sistema', what: 'fanout PDF de', code: 'POL-QUA-0007', t: 'há 4h' },
                { who: 'Carla Pinheiro', what: 'devolveu', code: 'IT-QUA-0088', t: 'ontem' },
              ].map((a, i) => (
                <div key={i} className="md-fade-up" style={{ position: 'relative', marginBottom: 14, fontSize: 12, lineHeight: 1.55, animationDelay: `${i*60+400}ms` }}>
                  <span style={{ position: 'absolute', left: -18, top: 4, width: 9, height: 9, borderRadius: '50%', background: 'var(--surface)', border: '2px solid var(--text)' }}/>
                  <div style={{ color: 'var(--text-muted)' }}>
                    <span style={{ color: 'var(--text)', fontWeight: 600 }}>{a.who}</span>
                    {' '}{a.what}{' '}
                    <span className="mono" style={{ color: 'var(--brand)', fontWeight: 600 }}>{a.code}</span>
                  </div>
                  <div className="caption mono" style={{ color: 'var(--text-faint)', fontSize: 10 }}>{a.t}</div>
                </div>
              ))}
            </div>
          </div>

        </aside>
      </div>
    </div>
  );
};

/* ============================================================
   DASHBOARD · ALT 2 — BENTO / DEPTH
   Layered cards, generous depth, stacked tiles, soft shadows
   ============================================================ */
const DashboardBento = () => {
  return (
    <div style={{ flex: 1, overflow: 'auto', background: 'var(--bg)', position: 'relative' }}>
      <StyleInject/>
      <div className="md-grid-bg" style={{ position: 'absolute', inset: 0, opacity: 0.4, pointerEvents: 'none' }}/>

      <div style={{ padding: '32px 36px', display: 'flex', flexDirection: 'column', gap: 24, position: 'relative' }}>

        {/* Greeting strip */}
        <div className="md-fade-up" style={{ display: 'flex', alignItems: 'flex-end', gap: 24, paddingBottom: 6 }}>
          <div style={{ flex: 1 }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: 10, marginBottom: 8 }}>
              <span style={{ width: 6, height: 6, borderRadius: '50%', background: 'var(--success)' }} className="md-blink-dot"/>
              <span className="kicker">Sex · 04 mai · 14:32 · São Paulo</span>
            </div>
            <h1 style={{ fontSize: 38, fontWeight: 600, letterSpacing: '-0.028em', lineHeight: 1.05, marginBottom: 6, textWrap: 'balance' }}>
              Bom dia, Marina.
            </h1>
            <p className="body" style={{ color: 'var(--text-muted)', maxWidth: 500 }}>
              Tudo organizado pra sua sexta. Comecemos pelas <strong style={{ color: 'var(--accent)' }}>4 pendências</strong>.
            </p>
          </div>
          <div style={{ display: 'flex', gap: 8 }}>
            <button className="btn"><Icon name="plus" size={14}/>Novo documento</button>
            <button className="btn btn-primary">Abrir caixa →</button>
          </div>
        </div>

        {/* === BENTO grid === */}
        <div style={{ display: 'grid', gap: 16,
          gridTemplateColumns: 'repeat(6, 1fr)',
          gridTemplateRows: 'auto auto auto',
        }}>

          {/* Hero tile — pendentes (huge) */}
          <div className="md-card-lift md-rise" style={{
            gridColumn: 'span 3', gridRow: 'span 2', animationDelay: '50ms',
            background: 'linear-gradient(155deg, var(--brand) 0%, var(--brand-deep) 100%)',
            color: 'white', borderRadius: 16, padding: 28, position: 'relative', overflow: 'hidden',
            border: '1px solid var(--brand-deep)', boxShadow: '0 12px 40px rgba(0,0,0,0.18)'
          }}>
            <div style={{ position: 'absolute', right: -40, bottom: -60, width: 240, height: 240,
              borderRadius: '50%', background: 'radial-gradient(circle, rgba(255,255,255,0.12), transparent 70%)' }}/>
            <div style={{ position: 'absolute', right: 24, top: 24, fontFamily: 'var(--font-mono)', fontSize: 10, color: 'rgba(255,255,255,0.5)', letterSpacing: '0.1em' }}>
              SUA CAIXA · ATUALIZADO 14:32
            </div>
            <div className="kicker" style={{ color: 'rgba(255,255,255,0.7)', marginBottom: 8 }}>§ AGUARDANDO VOCÊ</div>

            <div style={{ display: 'flex', alignItems: 'flex-end', gap: 14, marginBottom: 4 }}>
              <AnimatedNum value={4} fontSize={108} color="white"/>
              <div style={{ paddingBottom: 14 }}>
                <div style={{ fontSize: 16, fontWeight: 600, color: 'white' }}>pendências</div>
                <div style={{ fontSize: 12, color: 'rgba(255,255,255,0.7)' }}>2 vencem hoje</div>
              </div>
            </div>

            {/* Mini priority list */}
            <div style={{ marginTop: 24, display: 'flex', flexDirection: 'column', gap: 8 }}>
              {[
                { code: 'POP-QUA-0148', t: '3h 28min', urgent: true },
                { code: 'IT-PROD-0072', t: '9h', urgent: true },
                { code: 'POL-RH-0011', t: 'amanhã' },
                { code: 'DC-TI-0203', t: '2 dias' },
              ].map((p, i) => (
                <div key={i} className="md-fade-up" style={{
                  animationDelay: `${i*70+200}ms`,
                  display: 'flex', alignItems: 'center', gap: 10,
                  padding: '10px 12px', background: 'rgba(255,255,255,0.08)',
                  borderRadius: 8, fontSize: 12.5, backdropFilter: 'blur(8px)',
                  border: '1px solid rgba(255,255,255,0.12)'
                }}>
                  {p.urgent && <span className="md-blink-dot" style={{ width: 6, height: 6, borderRadius: '50%', background: 'var(--accent)', flexShrink: 0 }}/>}
                  {!p.urgent && <span style={{ width: 6, height: 6, borderRadius: '50%', background: 'rgba(255,255,255,0.4)', flexShrink: 0 }}/>}
                  <span className="mono" style={{ color: 'white', fontWeight: 600 }}>{p.code}</span>
                  <span style={{ flex: 1 }}/>
                  <span className="mono" style={{ color: p.urgent ? 'var(--accent)' : 'rgba(255,255,255,0.6)', fontSize: 11 }}>{p.t}</span>
                  <span style={{ color: 'rgba(255,255,255,0.6)' }}>→</span>
                </div>
              ))}
            </div>
          </div>

          {/* Stats row */}
          {[
            { label: 'Aprovados (sem.)', val: 12, sub: '↑ 33%', color: 'var(--success)', spark: [4,6,5,8,7,12] },
            { label: 'Devolvidos', val: 3, sub: 'aguardam', color: 'var(--warning)', spark: [1,2,0,1,2,3] },
            { label: 'Cobertura ISO', val: 94, suf: '%', sub: 'cl. 7.5', color: 'var(--info)', spark: [88,89,91,92,93,94] },
          ].map((s, i) => (
            <div key={i} className="md-card-lift md-rise card" style={{
              gridColumn: 'span 1', padding: 18, animationDelay: `${i*50+100}ms`,
              borderRadius: 14, position: 'relative', overflow: 'hidden'
            }}>
              <div style={{ position: 'absolute', top: 0, right: 0, width: 4, height: '100%', background: s.color, opacity: 0.7 }}/>
              <div className="kicker" style={{ marginBottom: 10 }}>{s.label}</div>
              <div style={{ display: 'flex', alignItems: 'baseline', gap: 4, marginBottom: 6 }}>
                <AnimatedNum value={s.val} fontSize={32} color={s.color}/>
                {s.suf && <span style={{ fontSize: 16, color: s.color, fontWeight: 500 }}>{s.suf}</span>}
              </div>
              <div className="caption" style={{ color: 'var(--text-muted)', marginBottom: 8 }}>{s.sub}</div>
              <Spark data={s.spark} color={s.color} width={120} height={24}/>
            </div>
          ))}

          {/* Activity tile */}
          <div className="md-card-lift md-rise card" style={{
            gridColumn: 'span 3', padding: 20, animationDelay: '300ms', borderRadius: 14
          }}>
            <div style={{ display: 'flex', alignItems: 'baseline', marginBottom: 14 }}>
              <div className="h3" style={{ fontSize: 14 }}>Trilha · sua área</div>
              <span style={{ flex: 1 }}/>
              <span style={{ display: 'flex', alignItems: 'center', gap: 6, fontSize: 10.5, fontFamily: 'var(--font-mono)', color: 'var(--success)', letterSpacing: '0.1em' }}>
                <span className="md-blink-dot" style={{ width: 6, height: 6, borderRadius: '50%', background: 'var(--success)' }}/>
                AO VIVO
              </span>
            </div>
            <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
              {[
                { who: 'Carla Pinheiro', what: 'enviou', code: 'POP-QUA-0148', t: '14:32', dot: 'review' },
                { who: 'Sistema', what: 'fanout PDF de', code: 'POL-QUA-0007', t: '13:50', dot: 'frozen' },
                { who: 'Bruno Said', what: 'aprovou', code: 'IT-QUA-0091', t: '11:08', dot: 'approved' },
                { who: 'Marina S', what: 'criou rascunho', code: 'POP-QUA-0149', t: '10:22', dot: 'draft' },
              ].map((a, i) => (
                <div key={i} className="md-fade-up" style={{ display: 'flex', alignItems: 'center', gap: 10, fontSize: 12, animationDelay: `${i*40+400}ms` }}>
                  <span style={{ width: 7, height: 7, borderRadius: '50%', flexShrink: 0,
                    background: a.dot === 'approved' ? 'var(--success)' : a.dot === 'review' ? 'var(--warning)' : a.dot === 'frozen' ? 'var(--info)' : 'var(--text-muted)' }}/>
                  <Avatar name={a.who} size="sm"/>
                  <span style={{ color: 'var(--text)', fontWeight: 500, fontSize: 12 }}>{a.who}</span>
                  <span style={{ color: 'var(--text-muted)' }}>{a.what}</span>
                  <span className="mono" style={{ color: 'var(--brand)', fontWeight: 600 }}>{a.code}</span>
                  <span style={{ flex: 1 }}/>
                  <span className="caption mono" style={{ color: 'var(--text-faint)' }}>{a.t}</span>
                </div>
              ))}
            </div>
          </div>

          {/* Drafts tile */}
          <div className="md-card-lift md-rise card" style={{
            gridColumn: 'span 3', padding: 20, animationDelay: '350ms', borderRadius: 14
          }}>
            <div style={{ display: 'flex', alignItems: 'baseline', marginBottom: 14 }}>
              <div className="h3" style={{ fontSize: 14 }}>Seus rascunhos</div>
              <span className="caption" style={{ marginLeft: 8 }}>· 7 ativos</span>
              <span style={{ flex: 1 }}/>
              <button className="btn btn-sm btn-ghost">Ver todos →</button>
            </div>
            <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
              {[
                { code: 'POP-QUA-0149', title: 'Calibração de paquímetro digital', edited: '22min', progress: 80 },
                { code: 'IT-QUA-0102', title: 'Inspeção visual de juntas soldadas', edited: 'ontem', progress: 45 },
                { code: 'DC-QUA-0034', title: 'Especificação SAE 1045 endurecido', edited: '3d', progress: 90 },
              ].map((d, i) => (
                <div key={i} className="md-fade-up md-row-hover" style={{
                  animationDelay: `${i*60+450}ms`,
                  padding: '10px 12px', borderRadius: 8, cursor: 'pointer',
                  display: 'flex', alignItems: 'center', gap: 12,
                  background: 'var(--surface-2)', border: '1px solid var(--border)'
                }}>
                  <div style={{ flex: 1, minWidth: 0 }}>
                    <div className="mono" style={{ fontSize: 11, fontWeight: 600, color: 'var(--text-soft)', marginBottom: 2 }}>{d.code}</div>
                    <div style={{ fontSize: 12.5, fontWeight: 500, color: 'var(--text)', textOverflow: 'ellipsis', overflow: 'hidden', whiteSpace: 'nowrap' }}>{d.title}</div>
                  </div>
                  <div style={{ width: 80, flexShrink: 0 }}>
                    <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: 10, color: 'var(--text-muted)', fontFamily: 'var(--font-mono)', marginBottom: 3 }}>
                      <span>{d.progress}%</span><span>{d.edited}</span>
                    </div>
                    <div className="md-shimmer-track" style={{ height: 4, background: 'var(--surface-3)', borderRadius: 2, overflow: 'hidden' }}>
                      <div style={{ width: `${d.progress}%`, height: '100%', background: 'var(--brand)', borderRadius: 2 }}/>
                    </div>
                  </div>
                  <Icon name="chevron" size={14}/>
                </div>
              ))}
            </div>
          </div>

        </div>
      </div>
    </div>
  );
};

/* ============================================================
   DASHBOARD · ALT 3 — TERMINAL / COMMAND CENTER
   Dense, mono-heavy, ledger feel, status-first
   ============================================================ */
const DashboardTerminal = () => {
  return (
    <div style={{ flex: 1, overflow: 'auto', background: 'var(--bg)', fontFamily: 'var(--font-sans)' }}>
      <StyleInject/>

      {/* Top status ribbon */}
      <div style={{
        background: 'var(--text)', color: 'var(--surface)',
        padding: '8px 24px', display: 'flex', alignItems: 'center', gap: 20,
        fontFamily: 'var(--font-mono)', fontSize: 11, letterSpacing: '0.04em',
      }}>
        <span style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
          <span className="md-blink-dot" style={{ width: 7, height: 7, borderRadius: '50%', background: 'var(--success)' }}/>
          SYS · ONLINE
        </span>
        <span style={{ opacity: 0.5 }}>│</span>
        <span>SESSION marina.s@metalcorp.br</span>
        <span style={{ opacity: 0.5 }}>│</span>
        <span>TENANT metalcorp_sp_01</span>
        <span style={{ opacity: 0.5 }}>│</span>
        <span><LiveTime/> · BRT</span>
        <span style={{ flex: 1 }}/>
        <span>FANOUT QUEUE: <span style={{ color: 'var(--success)' }}>0</span></span>
        <span style={{ opacity: 0.5 }}>│</span>
        <span>WORKERS: 4/4</span>
        <span style={{ opacity: 0.5 }}>│</span>
        <span>v3.4.2</span>
      </div>

      <div style={{ padding: '24px 32px 32px' }}>

        {/* Header */}
        <div className="md-fade-up" style={{ marginBottom: 24, display: 'flex', alignItems: 'flex-end', gap: 24 }}>
          <div style={{ flex: 1 }}>
            <div className="kicker" style={{ marginBottom: 6 }}>// HOME · PERSONAL DASHBOARD</div>
            <h1 style={{ fontSize: 30, fontWeight: 600, letterSpacing: '-0.025em', lineHeight: 1.1 }}>
              <span style={{ color: 'var(--text-muted)', fontWeight: 400 }}>$</span> Marina, você tem{' '}
              <span style={{ color: 'var(--accent)', fontFamily: 'var(--font-mono)' }}>4</span> decisões na fila.
            </h1>
            <p className="caption" style={{ color: 'var(--text-muted)', marginTop: 6, fontFamily: 'var(--font-mono)' }}>
              press <span className="kbd">G</span> + <span className="kbd">A</span> para ir à caixa de aprovação · <span className="kbd">⌘K</span> para tudo
            </p>
          </div>
        </div>

        {/* KPI ledger row */}
        <div style={{
          background: 'var(--surface)', border: '1px solid var(--border)',
          borderRadius: 8, marginBottom: 20, overflow: 'hidden',
          display: 'grid', gridTemplateColumns: 'repeat(6, 1fr)'
        }}>
          {[
            { label: 'PENDING_YOU', val: 4, color: 'var(--accent)', delta: '+1' },
            { label: 'OVERDUE', val: 0, color: 'var(--success)', delta: '0' },
            { label: 'RETURNED', val: 2, color: 'var(--warning)', delta: '+2' },
            { label: 'YOUR_DRAFTS', val: 7, color: 'var(--text-soft)', delta: '+0' },
            { label: 'APPROVED_WK', val: 12, color: 'var(--success)', delta: '+3' },
            { label: 'AVG_DECISION', val: '1h24', color: 'var(--brand)', delta: '-12m' },
          ].map((k, i) => (
            <div key={i} className="md-fade-up" style={{
              padding: '14px 16px', borderRight: i < 5 ? '1px solid var(--border)' : 'none',
              animationDelay: `${i*40}ms`,
              position: 'relative'
            }}>
              <div style={{ fontSize: 9.5, fontFamily: 'var(--font-mono)', color: 'var(--text-faint)', letterSpacing: '0.1em', marginBottom: 4 }}>
                {k.label}
              </div>
              <div style={{ display: 'flex', alignItems: 'baseline', gap: 6 }}>
                {typeof k.val === 'number'
                  ? <AnimatedNum value={k.val} fontSize={26} color={k.color}/>
                  : <span style={{ fontSize: 26, fontWeight: 600, color: k.color, fontFamily: 'var(--font-mono)', letterSpacing: '-0.02em' }}>{k.val}</span>
                }
                <span style={{ fontSize: 10, fontFamily: 'var(--font-mono)', color: 'var(--text-muted)', fontWeight: 500 }}>{k.delta}</span>
              </div>
            </div>
          ))}
        </div>

        {/* Main 2-col */}
        <div style={{ display: 'grid', gridTemplateColumns: '2fr 1fr', gap: 20 }}>

          {/* Approval queue table — terminal style */}
          <div style={{ background: 'var(--surface)', border: '1px solid var(--border)', borderRadius: 8, overflow: 'hidden' }}>
            <div style={{
              padding: '12px 16px', borderBottom: '1px solid var(--border)',
              background: 'var(--surface-2)', display: 'flex', alignItems: 'center', gap: 10,
              fontFamily: 'var(--font-mono)', fontSize: 11
            }}>
              <span style={{ color: 'var(--text-muted)' }}>$</span>
              <span style={{ color: 'var(--text)', fontWeight: 600 }}>md.approvals.pending</span>
              <span style={{ color: 'var(--text-muted)' }}>--user=marina --sort=deadline</span>
              <span style={{ flex: 1 }}/>
              <span style={{ color: 'var(--success)' }}>▸ 4 rows</span>
            </div>
            <table style={{ width: '100%', borderCollapse: 'collapse', fontFamily: 'var(--font-mono)', fontSize: 12 }}>
              <thead>
                <tr style={{ background: 'var(--surface-2)' }}>
                  <th style={{ padding: '8px 12px', textAlign: 'left', fontSize: 9.5, color: 'var(--text-faint)', letterSpacing: '0.1em', fontWeight: 600, borderBottom: '1px solid var(--border)' }}>#</th>
                  <th style={{ padding: '8px 12px', textAlign: 'left', fontSize: 9.5, color: 'var(--text-faint)', letterSpacing: '0.1em', fontWeight: 600, borderBottom: '1px solid var(--border)' }}>CODE</th>
                  <th style={{ padding: '8px 12px', textAlign: 'left', fontSize: 9.5, color: 'var(--text-faint)', letterSpacing: '0.1em', fontWeight: 600, borderBottom: '1px solid var(--border)' }}>TITLE</th>
                  <th style={{ padding: '8px 12px', textAlign: 'left', fontSize: 9.5, color: 'var(--text-faint)', letterSpacing: '0.1em', fontWeight: 600, borderBottom: '1px solid var(--border)' }}>AUTHOR</th>
                  <th style={{ padding: '8px 12px', textAlign: 'right', fontSize: 9.5, color: 'var(--text-faint)', letterSpacing: '0.1em', fontWeight: 600, borderBottom: '1px solid var(--border)' }}>ETA</th>
                  <th style={{ padding: '8px 12px', textAlign: 'right', fontSize: 9.5, color: 'var(--text-faint)', letterSpacing: '0.1em', fontWeight: 600, borderBottom: '1px solid var(--border)' }}></th>
                </tr>
              </thead>
              <tbody>
                {[
                  { code: 'POP-QUA-0148', title: 'Inspeção de recebimento — solda MIG/MAG', author: 'C. Pinheiro', eta: '03h28m', urgent: true },
                  { code: 'IT-PROD-0072', title: 'Setup da prensa hidráulica P-440', author: 'J. Aguiar', eta: '09h00m', urgent: true },
                  { code: 'POL-RH-0011', title: 'Política de afastamento parcial', author: 'F. Ribas', eta: '1d', urgent: false },
                  { code: 'DC-TI-0203', title: 'Diagrama de rede — DMZ planta 2', author: 'A. Iuri', eta: '2d', urgent: false },
                ].map((r, i) => (
                  <tr key={i} className="md-fade-up md-row-hover" style={{ animationDelay: `${i*50+100}ms`, cursor: 'pointer' }}>
                    <td style={{ padding: '10px 12px', borderBottom: '1px solid var(--border)', color: 'var(--text-faint)' }}>{String(i+1).padStart(2, '0')}</td>
                    <td style={{ padding: '10px 12px', borderBottom: '1px solid var(--border)', color: 'var(--brand)', fontWeight: 600 }}>{r.code}</td>
                    <td style={{ padding: '10px 12px', borderBottom: '1px solid var(--border)', color: 'var(--text)', fontFamily: 'var(--font-sans)', fontSize: 12.5 }}>{r.title}</td>
                    <td style={{ padding: '10px 12px', borderBottom: '1px solid var(--border)', color: 'var(--text-soft)' }}>{r.author}</td>
                    <td style={{ padding: '10px 12px', borderBottom: '1px solid var(--border)', textAlign: 'right', color: r.urgent ? 'var(--accent)' : 'var(--text-muted)', fontWeight: 600 }}>
                      {r.urgent && <span className="md-blink-dot" style={{ display: 'inline-block', width: 5, height: 5, borderRadius: '50%', background: 'var(--accent)', marginRight: 5, verticalAlign: 'middle' }}/>}
                      {r.eta}
                    </td>
                    <td style={{ padding: '10px 12px', borderBottom: '1px solid var(--border)', textAlign: 'right' }}>
                      <span style={{ color: 'var(--text-muted)', fontSize: 11 }}>↵</span>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          {/* Live activity feed */}
          <div style={{ background: 'var(--text)', color: 'var(--surface)', borderRadius: 8, padding: 16, fontFamily: 'var(--font-mono)', fontSize: 11.5, lineHeight: 1.7, position: 'relative', overflow: 'hidden' }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 12, fontSize: 10, color: 'rgba(255,255,255,0.5)', letterSpacing: '0.12em' }}>
              <span className="md-blink-dot" style={{ width: 6, height: 6, borderRadius: '50%', background: 'var(--success)' }}/>
              MD.STREAM · LIVE · TAIL -F
            </div>
            <div style={{ display: 'flex', flexDirection: 'column' }}>
              {[
                { t: '14:32:18', sev: 'INFO', msg: 'doc.submit POP-QUA-0148 by carla.p', color: 'var(--warning)' },
                { t: '14:30:52', sev: 'OK', msg: 'fanout.complete POL-QUA-0007 (38 cópias)', color: 'var(--success)' },
                { t: '14:29:11', sev: 'INFO', msg: 'doc.draft.save POP-QUA-0149 by marina.s', color: 'rgba(255,255,255,0.6)' },
                { t: '13:08:44', sev: 'OK', msg: 'doc.approve IT-QUA-0091 by bruno.s', color: 'var(--success)' },
                { t: '12:55:02', sev: 'WARN', msg: 'sla.breach.imminent POP-QUA-0148 (3h)', color: 'var(--accent)' },
                { t: '11:42:30', sev: 'INFO', msg: 'doc.return IT-QUA-0088 by carla.p', color: 'rgba(255,255,255,0.6)' },
                { t: '11:08:14', sev: 'OK', msg: 'doc.approve.l1 POP-QUA-0148 by bruno.s', color: 'var(--success)' },
                { t: '10:22:09', sev: 'INFO', msg: 'doc.create POP-QUA-0149 by marina.s', color: 'rgba(255,255,255,0.6)' },
              ].map((l, i) => (
                <div key={i} className="md-fade-up" style={{ display: 'flex', gap: 8, animationDelay: `${i*30+200}ms`, fontSize: 10.5 }}>
                  <span style={{ color: 'rgba(255,255,255,0.4)' }}>{l.t}</span>
                  <span style={{ color: l.color, fontWeight: 600, minWidth: 36 }}>{l.sev}</span>
                  <span style={{ color: 'rgba(255,255,255,0.85)', flex: 1, minWidth: 0, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{l.msg}</span>
                </div>
              ))}
            </div>
            <div style={{ marginTop: 14, padding: '8px 0 0', borderTop: '1px solid rgba(255,255,255,0.1)', display: 'flex', alignItems: 'center', gap: 6 }}>
              <span style={{ color: 'var(--success)' }}>▸</span>
              <span style={{ color: 'rgba(255,255,255,0.6)' }}>tail</span>
              <span className="md-blink-dot" style={{ width: 7, height: 14, background: 'var(--success)', display: 'inline-block' }}/>
            </div>
          </div>

        </div>

        {/* Bottom: drafts as cards */}
        <div style={{ marginTop: 20 }}>
          <div className="kicker" style={{ marginBottom: 10 }}>// YOUR_DRAFTS · 7 ACTIVE · LAST_EDIT 22m</div>
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: 12 }}>
            {[
              { code: 'POP-QUA-0149', title: 'Calibração de paquímetro digital', edited: '22m', progress: 80 },
              { code: 'IT-QUA-0102', title: 'Inspeção visual de juntas soldadas', edited: '1d', progress: 45 },
              { code: 'DC-QUA-0034', title: 'Especificação SAE 1045 endurecido', edited: '3d', progress: 90 },
            ].map((d, i) => (
              <div key={i} className="md-card-lift md-fade-up card" style={{ padding: 14, animationDelay: `${i*60+200}ms`, borderRadius: 8 }}>
                <div style={{ display: 'flex', alignItems: 'baseline', gap: 8, marginBottom: 8 }}>
                  <span className="mono" style={{ fontSize: 11, fontWeight: 600, color: 'var(--brand)' }}>{d.code}</span>
                  <span style={{ flex: 1 }}/>
                  <span className="mono" style={{ fontSize: 9.5, color: 'var(--text-faint)', letterSpacing: '0.08em' }}>EDITED {d.edited}</span>
                </div>
                <div style={{ fontSize: 13, fontWeight: 500, color: 'var(--text)', marginBottom: 12, lineHeight: 1.3 }}>{d.title}</div>
                <div className="md-shimmer-track" style={{ height: 3, background: 'var(--surface-3)', borderRadius: 2, overflow: 'hidden', position: 'relative' }}>
                  <div style={{ '--w': `${d.progress}%`, width: `${d.progress}%`, height: '100%', background: 'var(--brand)', animation: 'md-tick 1.2s cubic-bezier(0.16,1,0.3,1) backwards' }}/>
                </div>
                <div className="mono" style={{ fontSize: 10, color: 'var(--text-muted)', marginTop: 6, letterSpacing: '0.05em' }}>
                  PROGRESS {d.progress}% · DRAFT
                </div>
              </div>
            ))}
          </div>
        </div>

      </div>
    </div>
  );
};

window.MD_DASH_ALT = { DashboardEditorial, DashboardBento, DashboardTerminal };
