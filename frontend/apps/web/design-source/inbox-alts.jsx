/* global React, MD */
const { Icon, Avatar } = window.MD;
const { useState, useEffect } = React;

/* ============================================================
   INBOX · ALT 1 — STACKED CARDS / TINDER FOR APPROVALS
   Single big card focused, swipe-style with stack behind
   ============================================================ */
const InboxStack = () => {
  const [idx, setIdx] = useState(0);
  const items = [
    { code: 'POP-QUA-0148', title: 'Inspeção de recebimento — solda MIG/MAG', author: 'Carla Pinheiro', area: 'Qualidade', deadline: '3h 28min', urgent: true, kind: 'POP', summary: 'Revisão da inspeção de soldas em juntas críticas. Adiciona §3.2 sobre inspeção visual ampliada conforme NBR ISO 5817 (qualidade B). Inclui ensaio por líquido penetrante em 5 pontos amostrais.', changes: 12, version: 'v2.3 → v2.4', stage: 'L2 · Você', timeline: [{ who: 'Carla Pinheiro', stage: 'Autora', t: '04/05 · 09:14', state: 'done' }, { who: 'Bruno Said', stage: 'L1 · Líder técnico', t: '04/05 · 11:08', state: 'done' }, { who: 'Marina Silveira', stage: 'L2 · Gestor da Qualidade', t: 'Aguardando', state: 'now' }, { who: 'Sistema', stage: 'Publicação + fanout', t: '—', state: 'next' }] },
    { code: 'IT-PROD-0072', title: 'Setup da prensa hidráulica P-440', author: 'João Aguiar', area: 'Produção', deadline: '9h', urgent: true, kind: 'IT', summary: 'Procedimento de setup revisado após incidente de 12/abr. Inclui dupla checagem de pressão (ponto crítico §4.1) e bloqueio mecânico antes do início da operação.', changes: 8, version: 'v1.6 → v1.7', stage: 'L2 · Você', timeline: [] },
    { code: 'POL-RH-0011', title: 'Política de afastamento parcial', author: 'Fábio Ribas', area: 'Recursos Humanos', deadline: '1 dia', urgent: false, kind: 'POL', summary: 'Atualização para incluir modalidade híbrida pós-licença médica.', changes: 4, version: 'v3.0 → v3.1', stage: 'L1 · Você', timeline: [] },
    { code: 'DC-TI-0203', title: 'Diagrama de rede — DMZ planta 2', author: 'Ana Iuri', area: 'TI & Segurança', deadline: '2 dias', urgent: false, kind: 'DC', summary: 'Reflete a segregação para ICS/SCADA aprovada no Q1.', changes: 6, version: 'v0.9 → v1.0', stage: 'L2 · Você', timeline: [] },
  ];
  const item = items[idx];

  const next = () => setIdx(i => (i + 1) % items.length);
  const prev = () => setIdx(i => (i - 1 + items.length) % items.length);

  return (
    <div style={{ flex: 1, overflow: 'auto', background: 'var(--bg)', position: 'relative' }}>
      <style>{`
        @keyframes alt1-card-in { from { opacity: 0; transform: translateY(20px) scale(0.97); } to { opacity: 1; transform: translateY(0) scale(1); } }
        @keyframes alt1-fade-up { from { opacity: 0; transform: translateY(8px); } to { opacity: 1; transform: translateY(0); } }
        @keyframes alt1-blink { 0%,100% { opacity: 1; } 50% { opacity: 0.3; } }
      `}</style>

      <div style={{ display: 'grid', gridTemplateColumns: '320px 1fr', gap: 0, height: '100%' }}>

        {/* Queue rail — vertical numbered list */}
        <aside style={{ borderRight: '1px solid var(--border)', background: 'var(--surface)', padding: '24px 0', overflow: 'auto' }}>
          <div style={{ padding: '0 24px 16px', borderBottom: '1px solid var(--border)', marginBottom: 8 }}>
            <div className="kicker" style={{ marginBottom: 4 }}>SUA FILA</div>
            <div style={{ display: 'flex', alignItems: 'baseline', gap: 6 }}>
              <span style={{ fontSize: 32, fontWeight: 600, fontFamily: 'var(--font-mono)', letterSpacing: '-0.02em' }}>{items.length}</span>
              <span className="caption">decisões pendentes</span>
            </div>
            <div className="caption" style={{ marginTop: 6, color: 'var(--text-muted)' }}>
              <span style={{ color: 'var(--accent)', fontWeight: 600 }}>2 vencem hoje</span>
            </div>
          </div>

          {items.map((it, i) => (
            <button key={it.code} onClick={() => setIdx(i)} style={{
              width: '100%', textAlign: 'left', padding: '14px 24px',
              border: 'none', borderLeft: i === idx ? '3px solid var(--brand)' : '3px solid transparent',
              background: i === idx ? 'var(--surface-2)' : 'transparent',
              cursor: 'pointer', display: 'flex', gap: 12, alignItems: 'flex-start',
              transition: 'all 0.18s ease', position: 'relative',
            }}>
              <div style={{ flexShrink: 0, width: 28, textAlign: 'center' }}>
                <div style={{ fontFamily: 'var(--font-mono)', fontSize: 22, fontWeight: 600, color: i === idx ? 'var(--text)' : 'var(--text-faint)', lineHeight: 1, letterSpacing: '-0.02em' }}>
                  {String(i+1).padStart(2,'0')}
                </div>
              </div>
              <div style={{ flex: 1, minWidth: 0 }}>
                <div style={{ display: 'flex', alignItems: 'center', gap: 6, marginBottom: 4 }}>
                  <span className="mono" style={{ fontSize: 10.5, fontWeight: 600, color: 'var(--brand)' }}>{it.code}</span>
                  {it.urgent && <span style={{ width: 5, height: 5, borderRadius: '50%', background: 'var(--accent)', animation: 'alt1-blink 1.4s ease-in-out infinite' }}/>}
                </div>
                <div style={{ fontSize: 12.5, fontWeight: 500, color: 'var(--text)', marginBottom: 4, lineHeight: 1.3, textOverflow: 'ellipsis', overflow: 'hidden', display: '-webkit-box', WebkitLineClamp: 2, WebkitBoxOrient: 'vertical' }}>{it.title}</div>
                <div style={{ display: 'flex', alignItems: 'center', gap: 6, fontSize: 11, color: 'var(--text-muted)' }}>
                  <span style={{ color: it.urgent ? 'var(--accent)' : 'var(--text-muted)', fontWeight: it.urgent ? 600 : 400, fontFamily: 'var(--font-mono)' }}>{it.deadline}</span>
                  <span style={{ color: 'var(--text-faint)' }}>·</span>
                  <span>{it.area}</span>
                </div>
              </div>
            </button>
          ))}
        </aside>

        {/* Card stack — main */}
        <main style={{ position: 'relative', overflow: 'hidden', display: 'flex', flexDirection: 'column' }}>
          <div style={{ padding: '20px 32px 0', display: 'flex', alignItems: 'center', gap: 12 }}>
            <span className="mono" style={{ fontSize: 11, fontWeight: 600, color: 'var(--text-faint)', letterSpacing: '0.08em' }}>
              {String(idx+1).padStart(2,'0')} / {String(items.length).padStart(2,'0')}
            </span>
            <span style={{ flex: 1 }}/>
            <button onClick={prev} className="btn btn-sm">← anterior</button>
            <button onClick={next} className="btn btn-sm">próximo →</button>
          </div>

          {/* Stacked cards */}
          <div style={{ flex: 1, position: 'relative', padding: 32, display: 'grid', placeItems: 'center' }}>
            {/* Behind card 2 */}
            {items[(idx+2) % items.length] && <div style={{ position: 'absolute', top: 60, left: '50%', transform: 'translateX(-50%) scale(0.92)', width: 'min(640px, 90%)', height: 480, background: 'var(--surface-2)', border: '1px solid var(--border)', borderRadius: 16, opacity: 0.4, zIndex: 1 }}/>}
            {/* Behind card 1 */}
            <div style={{ position: 'absolute', top: 50, left: '50%', transform: 'translateX(-50%) scale(0.96)', width: 'min(640px, 90%)', height: 500, background: 'var(--surface)', border: '1px solid var(--border)', borderRadius: 16, opacity: 0.7, zIndex: 2, boxShadow: 'var(--shadow-1)' }}/>

            {/* Active card */}
            <article key={item.code} style={{
              position: 'relative', zIndex: 3,
              width: 'min(640px, 90%)', background: 'var(--surface)',
              border: '1px solid var(--border)', borderRadius: 16,
              boxShadow: '0 24px 60px rgba(0,0,0,0.12), 0 4px 12px rgba(0,0,0,0.04)',
              overflow: 'hidden', animation: 'alt1-card-in 0.4s cubic-bezier(0.16, 1, 0.3, 1)'
            }}>
              {/* Card header — colored band by urgency */}
              <div style={{
                background: item.urgent ? 'linear-gradient(135deg, var(--brand) 0%, var(--brand-deep) 100%)' : 'var(--text)',
                color: 'white', padding: '20px 28px', position: 'relative', overflow: 'hidden'
              }}>
                <div style={{ display: 'flex', alignItems: 'flex-start', gap: 16, position: 'relative' }}>
                  <div style={{
                    flexShrink: 0, padding: '6px 10px', background: 'rgba(255,255,255,0.15)',
                    borderRadius: 6, fontFamily: 'var(--font-mono)', fontSize: 11, fontWeight: 700,
                    letterSpacing: '0.06em', backdropFilter: 'blur(8px)'
                  }}>{item.kind}</div>
                  <div style={{ flex: 1, minWidth: 0 }}>
                    <div className="mono" style={{ fontSize: 11, color: 'rgba(255,255,255,0.7)', marginBottom: 6, fontWeight: 500, letterSpacing: '0.04em' }}>
                      {item.code} · {item.version}
                    </div>
                    <h2 style={{ fontSize: 24, fontWeight: 600, color: 'white', lineHeight: 1.2, letterSpacing: '-0.02em', textWrap: 'balance' }}>{item.title}</h2>
                  </div>
                  {item.urgent && (
                    <div style={{ flexShrink: 0, textAlign: 'right' }}>
                      <div style={{ fontSize: 9.5, color: 'rgba(255,255,255,0.65)', fontFamily: 'var(--font-mono)', letterSpacing: '0.1em', marginBottom: 2 }}>VENCE EM</div>
                      <div style={{ fontSize: 18, fontWeight: 700, fontFamily: 'var(--font-mono)', color: 'var(--accent)', letterSpacing: '-0.02em' }}>{item.deadline}</div>
                    </div>
                  )}
                </div>
              </div>

              {/* Body */}
              <div style={{ padding: '24px 28px 20px' }}>
                <p style={{ fontSize: 14, lineHeight: 1.55, color: 'var(--text-soft)', marginBottom: 18, textWrap: 'pretty' }}>{item.summary}</p>

                <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr 1fr', gap: 16, padding: '16px 0', borderTop: '1px solid var(--border)', borderBottom: '1px solid var(--border)', marginBottom: 18 }}>
                  <div>
                    <div className="kicker" style={{ marginBottom: 6 }}>AUTOR</div>
                    <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                      <Avatar name={item.author} size="sm"/>
                      <div>
                        <div style={{ fontSize: 12.5, fontWeight: 500 }}>{item.author}</div>
                        <div className="caption" style={{ fontSize: 10.5 }}>{item.area}</div>
                      </div>
                    </div>
                  </div>
                  <div>
                    <div className="kicker" style={{ marginBottom: 6 }}>ALTERAÇÕES</div>
                    <div style={{ display: 'flex', alignItems: 'baseline', gap: 4 }}>
                      <span style={{ fontSize: 22, fontWeight: 600, fontFamily: 'var(--font-mono)', letterSpacing: '-0.02em' }}>{item.changes}</span>
                      <span className="caption">edições</span>
                    </div>
                  </div>
                  <div>
                    <div className="kicker" style={{ marginBottom: 6 }}>ESTÁGIO</div>
                    <div style={{ fontSize: 12.5, fontWeight: 500 }}>{item.stage}</div>
                    <div className="caption" style={{ fontSize: 10.5 }}>de 4 etapas</div>
                  </div>
                </div>

                {/* Action row */}
                <div style={{ display: 'flex', gap: 10 }}>
                  <button className="btn" style={{ flex: 1, justifyContent: 'center', height: 44, fontSize: 13 }}>
                    <Icon name="eye" size={14}/> Abrir documento
                  </button>
                  <button className="btn" style={{ height: 44, fontSize: 13, color: 'var(--warning)', borderColor: 'var(--warning)' }}>
                    Devolver
                  </button>
                  <button className="btn btn-primary" onClick={next} style={{ height: 44, fontSize: 13, padding: '0 24px' }}>
                    Aprovar e assinar →
                  </button>
                </div>
              </div>
            </article>
          </div>

          <div style={{ padding: '0 32px 20px', textAlign: 'center', fontSize: 11, color: 'var(--text-faint)', fontFamily: 'var(--font-mono)' }}>
            <span className="kbd">A</span> aprovar · <span className="kbd">D</span> devolver · <span className="kbd">←/→</span> navegar
          </div>
        </main>
      </div>
    </div>
  );
};

/* ============================================================
   INBOX · ALT 2 — KANBAN / SWIMLANES
   3 columns: aguardando você · sua revisão · aguardando outros
   Live, dragable feel with progress dots
   ============================================================ */
const InboxKanban = () => {
  const lanes = [
    {
      id: 'pending', title: 'Aguardando você', subtitle: 'decida hoje', color: 'var(--accent)', count: 4,
      items: [
        { code: 'POP-QUA-0148', title: 'Inspeção de recebimento — solda MIG/MAG', author: 'Carla Pinheiro', area: 'Qualidade', deadline: '3h 28min', urgent: true, kind: 'POP', stage: 2, total: 4, changes: 12 },
        { code: 'IT-PROD-0072', title: 'Setup da prensa hidráulica P-440', author: 'João Aguiar', area: 'Produção', deadline: '9h', urgent: true, kind: 'IT', stage: 2, total: 4, changes: 8 },
        { code: 'POL-RH-0011', title: 'Política de afastamento parcial', author: 'Fábio Ribas', area: 'RH', deadline: '1 dia', kind: 'POL', stage: 1, total: 3, changes: 4 },
        { code: 'DC-TI-0203', title: 'Diagrama de rede — DMZ planta 2', author: 'Ana Iuri', area: 'TI', deadline: '2 dias', kind: 'DC', stage: 2, total: 4, changes: 6 },
      ]
    },
    {
      id: 'mine', title: 'Sua revisão', subtitle: 'em andamento', color: 'var(--brand)', count: 2,
      items: [
        { code: 'POP-QUA-0145', title: 'Calibração de paquímetro digital', author: 'Carla Pinheiro', area: 'Qualidade', deadline: 'amanhã', kind: 'POP', stage: 2, total: 4, changes: 5, you: true },
        { code: 'IT-MAN-0058', title: 'Manutenção preventiva — torno CNC', author: 'Roberto N.', area: 'Manutenção', deadline: '3 dias', kind: 'IT', stage: 2, total: 3, changes: 9, you: true },
      ]
    },
    {
      id: 'others', title: 'Aguardando outros', subtitle: 'fora do seu controle', color: 'var(--text-muted)', count: 5,
      items: [
        { code: 'POP-QUA-0142', title: 'Auditoria interna · processo Q1', author: 'M. Silveira', area: 'Qualidade', deadline: '5 dias', kind: 'POP', stage: 3, total: 4, changes: 3, waiting: 'L3 · Diretoria' },
        { code: 'POL-DIR-0004', title: 'Código de conduta · revisão 2026', author: 'M. Silveira', area: 'Diretoria', deadline: '1 sem', kind: 'POL', stage: 4, total: 5, changes: 18, waiting: 'L4 · Jurídico' },
        { code: 'IT-COM-0021', title: 'Atendimento a auditoria externa', author: 'Você', area: 'Compliance', deadline: '1 sem', kind: 'IT', stage: 1, total: 3, changes: 2, waiting: 'L1 · L. Carvalho' },
      ]
    },
  ];

  return (
    <div style={{ flex: 1, overflow: 'hidden', background: 'var(--bg)', display: 'flex', flexDirection: 'column' }}>
      <style>{`
        @keyframes alt2-card-in { from { opacity: 0; transform: translateY(12px); } to { opacity: 1; transform: translateY(0); } }
        @keyframes alt2-pulse { 0%,100% { box-shadow: 0 0 0 0 rgba(200,74,42,0.4); } 50% { box-shadow: 0 0 0 6px rgba(200,74,42,0); } }
        @keyframes alt2-shimmer { 0% { background-position: -200% 0; } 100% { background-position: 200% 0; } }
        .alt2-card { transition: all 0.2s cubic-bezier(0.16, 1, 0.3, 1); }
        .alt2-card:hover { transform: translateY(-2px); box-shadow: 0 8px 24px rgba(0,0,0,0.08); }
        .alt2-urgent-pulse { animation: alt2-pulse 2s ease-in-out infinite; }
      `}</style>

      {/* Header */}
      <div style={{ padding: '20px 32px 16px', borderBottom: '1px solid var(--border)', background: 'var(--surface)', display: 'flex', alignItems: 'flex-end', gap: 24 }}>
        <div style={{ flex: 1 }}>
          <div className="kicker" style={{ marginBottom: 4 }}>APROVAÇÕES · CAIXA DE ENTRADA</div>
          <h1 style={{ fontSize: 24, fontWeight: 600, letterSpacing: '-0.02em' }}>11 documentos no seu radar</h1>
        </div>
        <div style={{ display: 'flex', gap: 6 }}>
          <button className="btn btn-sm">
            <Icon name="filter" size={13}/> Filtros
          </button>
          <button className="btn btn-sm">Agrupar por área</button>
          <span style={{ width: 1, background: 'var(--border)', margin: '0 4px' }}/>
          <span style={{ display: 'flex', alignItems: 'center', gap: 6, fontSize: 11, color: 'var(--success)', fontFamily: 'var(--font-mono)', letterSpacing: '0.08em' }}>
            <span style={{ width: 6, height: 6, borderRadius: '50%', background: 'var(--success)', animation: 'alt2-pulse 1.5s ease-in-out infinite' }}/>
            SINCRONIZADO
          </span>
        </div>
      </div>

      {/* Kanban grid */}
      <div style={{ flex: 1, overflow: 'auto', padding: 24, display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: 16 }}>
        {lanes.map((lane, li) => (
          <div key={lane.id} style={{ display: 'flex', flexDirection: 'column', minWidth: 0, background: 'var(--surface-2)', borderRadius: 12, padding: 12, border: '1px solid var(--border)' }}>
            {/* Lane header */}
            <div style={{ display: 'flex', alignItems: 'center', gap: 10, marginBottom: 12, padding: '4px 4px 12px', borderBottom: '1px solid var(--border)' }}>
              <span style={{ width: 8, height: 8, borderRadius: '50%', background: lane.color, flexShrink: 0 }}/>
              <div style={{ flex: 1, minWidth: 0 }}>
                <div style={{ fontSize: 13, fontWeight: 600, color: 'var(--text)', letterSpacing: '-0.005em' }}>{lane.title}</div>
                <div style={{ fontSize: 10.5, color: 'var(--text-muted)', fontFamily: 'var(--font-mono)', letterSpacing: '0.04em' }}>{lane.subtitle}</div>
              </div>
              <div style={{
                minWidth: 24, height: 24, padding: '0 8px', borderRadius: 12,
                background: lane.color, color: 'white', display: 'grid', placeItems: 'center',
                fontSize: 11.5, fontWeight: 700, fontFamily: 'var(--font-mono)'
              }}>{lane.items.length}</div>
            </div>

            {/* Cards */}
            <div style={{ display: 'flex', flexDirection: 'column', gap: 10, overflow: 'auto' }}>
              {lane.items.map((card, ci) => (
                <div key={card.code} className="alt2-card" style={{
                  background: 'var(--surface)', border: '1px solid var(--border)',
                  borderRadius: 8, padding: 14, cursor: 'pointer',
                  animation: `alt2-card-in 0.4s cubic-bezier(0.16, 1, 0.3, 1) ${li * 100 + ci * 60}ms backwards`,
                  position: 'relative', overflow: 'hidden',
                  ...(card.urgent ? { borderColor: 'var(--accent)', borderWidth: 1 } : {}),
                }}>
                  {/* urgency stripe */}
                  {card.urgent && <div style={{ position: 'absolute', top: 0, left: 0, right: 0, height: 2, background: 'var(--accent)' }}/>}

                  <div style={{ display: 'flex', alignItems: 'center', gap: 6, marginBottom: 8 }}>
                    <span style={{
                      padding: '2px 6px', background: 'var(--surface-3)', borderRadius: 4,
                      fontFamily: 'var(--font-mono)', fontSize: 9.5, fontWeight: 700, color: 'var(--text-soft)', letterSpacing: '0.06em'
                    }}>{card.kind}</span>
                    <span className="mono" style={{ fontSize: 10.5, color: 'var(--text-muted)', fontWeight: 500 }}>{card.code}</span>
                    <span style={{ flex: 1 }}/>
                    {card.urgent && <span className="alt2-urgent-pulse" style={{ width: 6, height: 6, borderRadius: '50%', background: 'var(--accent)' }}/>}
                  </div>

                  <div style={{ fontSize: 13, fontWeight: 500, color: 'var(--text)', lineHeight: 1.3, marginBottom: 10, textWrap: 'pretty' }}>{card.title}</div>

                  {/* Stage progress dots */}
                  <div style={{ display: 'flex', alignItems: 'center', gap: 4, marginBottom: 10 }}>
                    {Array.from({ length: card.total }).map((_, si) => (
                      <span key={si} style={{
                        flex: 1, height: 4, borderRadius: 2,
                        background: si < card.stage ? lane.color : 'var(--surface-3)',
                      }}/>
                    ))}
                    <span style={{ fontSize: 9.5, fontFamily: 'var(--font-mono)', color: 'var(--text-faint)', marginLeft: 4, letterSpacing: '0.04em' }}>
                      {card.stage}/{card.total}
                    </span>
                  </div>

                  {/* Bottom meta */}
                  <div style={{ display: 'flex', alignItems: 'center', gap: 8, fontSize: 11, color: 'var(--text-muted)' }}>
                    <Avatar name={card.author} size="xs"/>
                    <span style={{ flex: 1, minWidth: 0, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{card.author}</span>
                    <span className="mono" style={{ color: card.urgent ? 'var(--accent)' : 'var(--text-muted)', fontWeight: card.urgent ? 600 : 400, fontSize: 10.5 }}>{card.deadline}</span>
                  </div>

                  {card.waiting && (
                    <div style={{ marginTop: 8, padding: '6px 8px', background: 'var(--surface-2)', borderRadius: 5, fontSize: 10.5, color: 'var(--text-muted)', fontFamily: 'var(--font-mono)', letterSpacing: '0.02em' }}>
                      🕓 com <span style={{ color: 'var(--text)', fontWeight: 600 }}>{card.waiting}</span>
                    </div>
                  )}
                </div>
              ))}

              {lane.items.length === 0 && (
                <div style={{ padding: 24, textAlign: 'center', color: 'var(--text-faint)', fontSize: 12 }}>tudo limpo</div>
              )}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
};

/* ============================================================
   INBOX · ALT 3 — TIMELINE / DEADLINE-FIRST
   Horizontal timeline grouped by deadline tomorrow / week / month
   With heatmap of decisions made
   ============================================================ */
const InboxTimeline = () => {
  const buckets = [
    {
      label: 'Hoje', sub: 'até 18:00', count: 2, urgent: true,
      items: [
        { code: 'POP-QUA-0148', title: 'Inspeção de recebimento — solda MIG/MAG', author: 'Carla Pinheiro', area: 'Qualidade', time: '18:00', kind: 'POP', stage: 2, total: 4, deadline: '3h 28min', changes: 12 },
        { code: 'IT-PROD-0072', title: 'Setup da prensa hidráulica P-440', author: 'João Aguiar', area: 'Produção', time: '23:30', kind: 'IT', stage: 2, total: 4, deadline: '9h', changes: 8 },
      ]
    },
    {
      label: 'Amanhã', sub: 'sáb 05/05', count: 1,
      items: [
        { code: 'POL-RH-0011', title: 'Política de afastamento parcial', author: 'Fábio Ribas', area: 'RH', time: '09:00', kind: 'POL', stage: 1, total: 3, deadline: '18h', changes: 4 },
      ]
    },
    {
      label: 'Esta semana', sub: 'até 11/05', count: 1,
      items: [
        { code: 'DC-TI-0203', title: 'Diagrama de rede — DMZ planta 2', author: 'Ana Iuri', area: 'TI', time: 'seg 14:00', kind: 'DC', stage: 2, total: 4, deadline: '2 dias', changes: 6 },
      ]
    },
    {
      label: 'Próximo mês', sub: 'até 31/05', count: 0, items: []
    },
  ];

  // Heatmap: 14 days, count of decisions
  const heat = [3,5,2,7,4,6,8,5,3,9,4,7,5,2];

  return (
    <div style={{ flex: 1, overflow: 'auto', background: 'var(--bg)' }}>
      <style>{`
        @keyframes alt3-fade-in { from { opacity: 0; transform: translateX(-12px); } to { opacity: 1; transform: translateX(0); } }
        @keyframes alt3-pulse { 0%,100% { transform: scale(1); opacity: 1; } 50% { transform: scale(1.15); opacity: 0.7; } }
        @keyframes alt3-grow { from { transform: scaleY(0); } to { transform: scaleY(1); } }
        .alt3-row { transition: background 0.18s ease, transform 0.18s ease; }
        .alt3-row:hover { background: var(--surface-2); }
        .alt3-heat { transition: transform 0.2s ease; }
        .alt3-heat:hover { transform: scale(1.2); }
      `}</style>

      <div style={{ padding: '24px 32px' }}>

        {/* Header */}
        <div style={{ display: 'flex', alignItems: 'flex-start', gap: 24, marginBottom: 24, animation: 'alt3-fade-in 0.4s ease-out' }}>
          <div style={{ flex: 1 }}>
            <div className="kicker" style={{ marginBottom: 4 }}>APROVAÇÕES · LINHA DO TEMPO</div>
            <h1 style={{ fontSize: 28, fontWeight: 600, letterSpacing: '-0.022em', lineHeight: 1.15, marginBottom: 6 }}>
              Suas <span style={{ color: 'var(--brand)' }}>4 decisões</span> na ordem que importam
            </h1>
            <p className="caption" style={{ color: 'var(--text-muted)' }}>Organizadas por prazo. As de hoje pulsam em laranja.</p>
          </div>

          {/* Decisions heatmap */}
          <div style={{ background: 'var(--surface)', border: '1px solid var(--border)', borderRadius: 10, padding: '12px 16px' }}>
            <div className="kicker" style={{ marginBottom: 8 }}>SUAS DECISÕES · 14 DIAS</div>
            <div style={{ display: 'flex', gap: 3, alignItems: 'flex-end', height: 32 }}>
              {heat.map((v, i) => (
                <div key={i} className="alt3-heat" style={{
                  width: 8, height: `${(v/9)*100}%`, minHeight: 4,
                  background: `oklch(${50 + v*4}% 0.12 28)`,
                  borderRadius: 1, animation: `alt3-grow 0.6s cubic-bezier(0.16,1,0.3,1) ${i*30}ms backwards`,
                  transformOrigin: 'bottom'
                }} title={`${v} decisões`}/>
              ))}
            </div>
            <div style={{ display: 'flex', justifyContent: 'space-between', marginTop: 4 }}>
              <span style={{ fontSize: 9.5, color: 'var(--text-faint)', fontFamily: 'var(--font-mono)', letterSpacing: '0.04em' }}>22/abr</span>
              <span style={{ fontSize: 9.5, color: 'var(--text-faint)', fontFamily: 'var(--font-mono)', letterSpacing: '0.04em' }}>hoje</span>
            </div>
          </div>
        </div>

        {/* Timeline */}
        <div style={{ position: 'relative', paddingLeft: 32 }}>
          {/* Vertical rail */}
          <div style={{ position: 'absolute', left: 11, top: 8, bottom: 8, width: 2, background: 'linear-gradient(180deg, var(--accent) 0%, var(--brand) 25%, var(--text-muted) 60%, var(--text-faint) 100%)', borderRadius: 1 }}/>

          {buckets.map((bucket, bi) => (
            <section key={bucket.label} style={{ marginBottom: 32, position: 'relative', animation: `alt3-fade-in 0.5s cubic-bezier(0.16,1,0.3,1) ${bi*100}ms backwards` }}>
              {/* Bucket dot */}
              <div style={{
                position: 'absolute', left: -29, top: 0,
                width: 24, height: 24, borderRadius: '50%',
                background: bucket.urgent ? 'var(--accent)' : (bi === 1 ? 'var(--brand)' : 'var(--surface)'),
                border: bucket.urgent ? '4px solid var(--bg)' : `2px solid ${bi === 1 ? 'var(--brand)' : 'var(--border-strong)'}`,
                boxShadow: bucket.urgent ? '0 0 0 4px rgba(200,74,42,0.15)' : 'none',
                ...(bucket.urgent ? { animation: 'alt3-pulse 2s ease-in-out infinite' } : {}),
              }}/>

              {/* Bucket header */}
              <div style={{ display: 'flex', alignItems: 'baseline', gap: 12, marginBottom: 12, paddingTop: 0 }}>
                <h2 style={{ fontSize: 22, fontWeight: 600, letterSpacing: '-0.02em', color: bucket.urgent ? 'var(--accent)' : 'var(--text)', textTransform: bucket.urgent ? undefined : undefined }}>
                  {bucket.label}
                </h2>
                <span className="mono" style={{ fontSize: 12, color: 'var(--text-muted)', fontWeight: 500 }}>{bucket.sub}</span>
                <span style={{ flex: 1 }}/>
                <span style={{
                  padding: '3px 10px', borderRadius: 12, fontSize: 11, fontWeight: 600,
                  background: bucket.count > 0 ? (bucket.urgent ? 'var(--accent)' : 'var(--text)') : 'var(--surface-3)',
                  color: bucket.count > 0 ? 'white' : 'var(--text-muted)',
                  fontFamily: 'var(--font-mono)', letterSpacing: '0.04em'
                }}>{bucket.count} {bucket.count === 1 ? 'doc' : 'docs'}</span>
              </div>

              {/* Items list */}
              {bucket.items.length > 0 ? (
                <div style={{ display: 'flex', flexDirection: 'column', gap: 0, background: 'var(--surface)', border: '1px solid var(--border)', borderRadius: 10, overflow: 'hidden' }}>
                  {bucket.items.map((it, ii) => (
                    <div key={it.code} className="alt3-row" style={{
                      display: 'grid', gridTemplateColumns: '80px 1fr auto 200px auto',
                      gap: 16, alignItems: 'center', padding: '14px 18px',
                      borderBottom: ii < bucket.items.length - 1 ? '1px solid var(--border)' : 'none',
                      cursor: 'pointer'
                    }}>
                      <div style={{ textAlign: 'right' }}>
                        <div className="mono" style={{ fontSize: 16, fontWeight: 600, color: bucket.urgent ? 'var(--accent)' : 'var(--text)', letterSpacing: '-0.02em', lineHeight: 1 }}>{it.time}</div>
                        <div className="caption" style={{ fontSize: 10.5, marginTop: 2 }}>vence em {it.deadline}</div>
                      </div>

                      <div style={{ minWidth: 0 }}>
                        <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 4 }}>
                          <span style={{ padding: '2px 6px', background: 'var(--surface-3)', borderRadius: 4, fontFamily: 'var(--font-mono)', fontSize: 9.5, fontWeight: 700, color: 'var(--text-soft)', letterSpacing: '0.06em' }}>{it.kind}</span>
                          <span className="mono" style={{ fontSize: 11, color: 'var(--brand)', fontWeight: 600 }}>{it.code}</span>
                        </div>
                        <div style={{ fontSize: 13.5, fontWeight: 500, color: 'var(--text)', lineHeight: 1.3, textOverflow: 'ellipsis', overflow: 'hidden', whiteSpace: 'nowrap' }}>{it.title}</div>
                        <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginTop: 4, fontSize: 11, color: 'var(--text-muted)' }}>
                          <Avatar name={it.author} size="xs"/>
                          <span>{it.author}</span>
                          <span style={{ color: 'var(--text-faint)' }}>·</span>
                          <span>{it.area}</span>
                          <span style={{ color: 'var(--text-faint)' }}>·</span>
                          <span>{it.changes} alterações</span>
                        </div>
                      </div>

                      <div/>

                      {/* Stage progress */}
                      <div>
                        <div style={{ display: 'flex', alignItems: 'center', gap: 4, marginBottom: 4 }}>
                          {Array.from({ length: it.total }).map((_, si) => (
                            <span key={si} style={{
                              flex: 1, height: 4, borderRadius: 2,
                              background: si < it.stage ? (bucket.urgent ? 'var(--accent)' : 'var(--brand)') : 'var(--surface-3)',
                            }}/>
                          ))}
                        </div>
                        <div className="caption mono" style={{ fontSize: 10, letterSpacing: '0.04em' }}>
                          ETAPA {it.stage}/{it.total} · L2 — você
                        </div>
                      </div>

                      <button className="btn btn-primary btn-sm" style={{ height: 32 }}>
                        Revisar →
                      </button>
                    </div>
                  ))}
                </div>
              ) : (
                <div style={{ padding: '14px 18px', background: 'var(--surface)', border: '1px dashed var(--border)', borderRadius: 10, color: 'var(--text-faint)', fontSize: 12.5, fontStyle: 'italic' }}>
                  Nada ainda. Continue assim.
                </div>
              )}
            </section>
          ))}
        </div>
      </div>
    </div>
  );
};

window.MD_INBOX_ALT = { InboxStack, InboxKanban, InboxTimeline };
