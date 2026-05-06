/* global React, MD */
const { Icon, Status, Avatar } = window.MD;

/* ============================================================
   § 02.01 · DASHBOARD PESSOAL
   Landing pós-login. Caixa de aprovação + atividade + atalhos.
   ============================================================ */
const Dashboard = () => {
  return (
    <div style={{ flex: 1, overflow: 'auto', background: 'var(--bg)' }}>
      {/* === Hero === */}
      <div style={{ padding: '28px 32px 20px', borderBottom: '1px solid var(--border)', background: 'var(--surface)' }}>
        <div style={{ display: 'flex', alignItems: 'baseline', gap: 14, marginBottom: 6 }}>
          <span className="kicker">§ Início · Sexta · 04 mai 2026</span>
          <span style={{ width: 4, height: 4, borderRadius: '50%', background: 'var(--text-faint)' }}/>
          <span className="caption mono" style={{ color: 'var(--text-muted)' }}>último login 14:21 · São Paulo</span>
        </div>
        <h1 className="display-1" style={{ fontSize: 32, letterSpacing: '-0.025em', marginBottom: 8 }}>
          Bom dia, Marina.
        </h1>
        <p className="body" style={{ color: 'var(--text-muted)', maxWidth: 560 }}>
          Você tem <strong style={{ color: 'var(--accent)' }}>4 documentos aguardando aprovação</strong> e 2 revisões devolvidas.
          A sua área (Qualidade) publicou 12 documentos esta semana.
        </p>
      </div>

      {/* === Stats row === */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(4, 1fr)', gap: 0, borderBottom: '1px solid var(--border)', background: 'var(--surface)' }}>
        {[
          { label: 'Pendentes p/ você', value: '4', sub: '2 vencem hoje', color: 'var(--accent)' },
          { label: 'Devolvidas', value: '2', sub: 'aguardam ajuste', color: 'var(--warning)' },
          { label: 'Em rascunho', value: '7', sub: 'autoria sua', color: 'var(--text-soft)' },
          { label: 'Lidas hoje', value: '23', sub: 'auditoria de leitura', color: 'var(--success)' },
        ].map((s, i) => (
          <div key={i} style={{ padding: '20px 28px', borderRight: i < 3 ? '1px solid var(--border)' : 'none' }}>
            <div className="kicker" style={{ marginBottom: 8 }}>{s.label}</div>
            <div style={{ display: 'flex', alignItems: 'baseline', gap: 10 }}>
              <span className="display-1 num" style={{ fontSize: 32, color: s.color }}>{s.value}</span>
              <span className="caption" style={{ color: 'var(--text-muted)' }}>{s.sub}</span>
            </div>
          </div>
        ))}
      </div>

      {/* === Main grid === */}
      <div style={{ display: 'grid', gridTemplateColumns: '1.6fr 1fr', gap: 20, padding: '24px 32px' }}>

        {/* ---- LEFT: Caixa de aprovação preview ---- */}
        <div className="card" style={{ padding: 0, overflow: 'hidden' }}>
          <div style={{ padding: '14px 20px', borderBottom: '1px solid var(--border)', display: 'flex', alignItems: 'center', gap: 12 }}>
            <div>
              <div className="h3" style={{ fontSize: 14, marginBottom: 2 }}>Sua caixa de aprovação</div>
              <div className="caption">4 pendentes · ISO segregation: você não pode aprovar seus próprios envios</div>
            </div>
            <span style={{ flex: 1 }}/>
            <button className="btn btn-sm btn-ghost">Ver todas →</button>
          </div>

          <table style={{ width: '100%', borderCollapse: 'collapse' }}>
            <tbody>
              {[
                { code: 'POP-QUA-0148', title: 'Inspeção de recebimento — solda MIG/MAG', author: 'Carla Pinheiro', area: 'Qualidade', deadline: 'Hoje · 18:00', urgent: true, version: 'v3.2' },
                { code: 'IT-PROD-0072', title: 'Setup da prensa hidráulica P-440', author: 'João Aguiar', area: 'Produção', deadline: 'Hoje · 23:59', urgent: true, version: 'v1.4' },
                { code: 'POL-RH-0011', title: 'Política de afastamento parcial', author: 'Fábio Ribas', area: 'RH', deadline: 'Amanhã', urgent: false, version: 'v2.0' },
                { code: 'DC-TI-0203', title: 'Diagrama de rede — DMZ planta 2', author: 'Ana Iuri', area: 'TI & Segurança', deadline: 'Em 2 dias', urgent: false, version: 'v1.0' },
              ].map((r, i) => (
                <tr key={i} style={{ borderBottom: i < 3 ? '1px solid var(--border)' : 'none' }}>
                  <td style={{ padding: '14px 20px', verticalAlign: 'top', width: 1, whiteSpace: 'nowrap' }}>
                    <div className="mono" style={{ fontSize: 11, color: 'var(--text-soft)', fontWeight: 600 }}>{r.code}</div>
                    <div className="caption mono" style={{ color: 'var(--text-faint)' }}>{r.version}</div>
                  </td>
                  <td style={{ padding: '14px 12px', verticalAlign: 'top' }}>
                    <div style={{ fontSize: 13.5, fontWeight: 500, color: 'var(--text)', marginBottom: 3 }}>{r.title}</div>
                    <div style={{ display: 'flex', alignItems: 'center', gap: 8, fontSize: 11, color: 'var(--text-muted)' }}>
                      <Avatar name={r.author} size="sm"/>
                      <span>{r.author}</span>
                      <span style={{ color: 'var(--text-faint)' }}>·</span>
                      <span>{r.area}</span>
                    </div>
                  </td>
                  <td style={{ padding: '14px 16px', verticalAlign: 'top', width: 1, whiteSpace: 'nowrap', textAlign: 'right' }}>
                    <div style={{ display: 'flex', alignItems: 'center', gap: 6, justifyContent: 'flex-end', fontSize: 11.5, color: r.urgent ? 'var(--accent)' : 'var(--text-muted)', fontFamily: 'var(--font-mono)', fontWeight: 600 }}>
                      {r.urgent && <span style={{ width: 5, height: 5, borderRadius: '50%', background: 'var(--accent)' }}/>}
                      {r.deadline}
                    </div>
                    <button className="btn btn-sm btn-primary" style={{ marginTop: 8 }}>Revisar →</button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>

        {/* ---- RIGHT: shortcuts + activity ---- */}
        <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>

          {/* Shortcuts */}
          <div className="card" style={{ padding: 18 }}>
            <div className="h3" style={{ fontSize: 14, marginBottom: 12 }}>Atalhos</div>
            <div style={{ display: 'grid', gridTemplateColumns: 'repeat(2, 1fr)', gap: 8 }}>
              {[
                { icon: 'plus', label: 'Novo documento', sub: 'wizard 4 passos' },
                { icon: 'docs', label: 'Meus rascunhos', sub: '7 documentos' },
                { icon: 'history', label: 'Recentes', sub: 'últimos 12' },
                { icon: 'audit', label: 'Trilha de auditoria', sub: 'em tempo real' },
              ].map((s, i) => (
                <button key={i} className="card" style={{ padding: 12, textAlign: 'left', display: 'flex', flexDirection: 'column', gap: 4, cursor: 'pointer', background: 'var(--surface-2)', border: '1px solid var(--border)' }}>
                  <span style={{ color: 'var(--brand)' }}><Icon name={s.icon} size={16}/></span>
                  <div style={{ fontSize: 12.5, fontWeight: 500, color: 'var(--text)', marginTop: 4 }}>{s.label}</div>
                  <div className="caption" style={{ color: 'var(--text-muted)' }}>{s.sub}</div>
                </button>
              ))}
            </div>
          </div>

          {/* Activity stream */}
          <div className="card" style={{ padding: 18 }}>
            <div style={{ display: 'flex', alignItems: 'baseline', marginBottom: 10 }}>
              <div className="h3" style={{ fontSize: 14 }}>Trilha · sua área</div>
              <span style={{ flex: 1 }}/>
              <span className="caption mono" style={{ color: 'var(--text-faint)' }}>tempo real</span>
            </div>
            <div style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>
              {[
                { who: 'Carla Pinheiro', what: 'enviou para revisão', code: 'POP-QUA-0148', t: '14:32', dot: 'review' },
                { who: 'Sistema', what: 'fanout PDF de', code: 'POL-QUA-0007', t: '13:50', dot: 'frozen' },
                { who: 'Bruno Said', what: 'aprovou', code: 'IT-QUA-0091', t: '11:08', dot: 'approved' },
                { who: 'Marina S', what: 'criou rascunho', code: 'POP-QUA-0149', t: '10:22', dot: 'draft' },
                { who: 'Carla Pinheiro', what: 'devolveu para ajuste', code: 'IT-QUA-0088', t: 'ontem', dot: 'rejected' },
              ].map((a, i) => (
                <div key={i} style={{ display: 'flex', alignItems: 'flex-start', gap: 10, fontSize: 12 }}>
                  <span style={{ width: 7, height: 7, borderRadius: '50%', marginTop: 5, flexShrink: 0,
                    background: a.dot === 'approved' ? 'var(--success)'
                              : a.dot === 'review' ? 'var(--warning)'
                              : a.dot === 'rejected' ? 'var(--danger)'
                              : a.dot === 'frozen' ? 'var(--info)'
                              : 'var(--text-muted)' }}/>
                  <div style={{ flex: 1, lineHeight: 1.5 }}>
                    <span style={{ color: 'var(--text)', fontWeight: 500 }}>{a.who}</span>
                    <span style={{ color: 'var(--text-muted)' }}> {a.what} </span>
                    <span className="mono" style={{ color: 'var(--brand)', fontWeight: 600 }}>{a.code}</span>
                  </div>
                  <span className="caption mono" style={{ color: 'var(--text-faint)' }}>{a.t}</span>
                </div>
              ))}
            </div>
          </div>
        </div>
      </div>

      {/* ---- Bottom row: my drafts ---- */}
      <div style={{ padding: '0 32px 32px' }}>
        <div className="card" style={{ padding: 0, overflow: 'hidden' }}>
          <div style={{ padding: '14px 20px', borderBottom: '1px solid var(--border)', display: 'flex', alignItems: 'center' }}>
            <div className="h3" style={{ fontSize: 14 }}>Seus rascunhos</div>
            <span className="caption" style={{ marginLeft: 10, color: 'var(--text-muted)' }}>· 7 documentos · última edição há 22min</span>
            <span style={{ flex: 1 }}/>
            <button className="btn btn-sm btn-ghost">Abrir todos →</button>
          </div>
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: 0 }}>
            {[
              { code: 'POP-QUA-0149', title: 'Calibração de paquímetro digital', edited: 'há 22min', progress: 80 },
              { code: 'IT-QUA-0102', title: 'Inspeção visual de juntas soldadas', edited: 'ontem', progress: 45 },
              { code: 'DC-QUA-0034', title: 'Especificação SAE 1045 endurecido', edited: '3 dias', progress: 90 },
            ].map((d, i) => (
              <div key={i} style={{ padding: 18, borderRight: i < 2 ? '1px solid var(--border)' : 'none', display: 'flex', flexDirection: 'column', gap: 10 }}>
                <div className="mono" style={{ fontSize: 11, fontWeight: 600, color: 'var(--text-soft)' }}>{d.code}</div>
                <div style={{ fontSize: 13.5, fontWeight: 500, color: 'var(--text)', lineHeight: 1.35 }}>{d.title}</div>
                <div style={{ flex: 1 }}/>
                <div>
                  <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: 10.5, color: 'var(--text-muted)', marginBottom: 4, fontFamily: 'var(--font-mono)' }}>
                    <span>RASCUNHO · {d.progress}%</span><span>{d.edited}</span>
                  </div>
                  <div style={{ height: 3, background: 'var(--surface-3)', borderRadius: 2, overflow: 'hidden' }}>
                    <div style={{ width: `${d.progress}%`, height: '100%', background: 'var(--brand)' }}/>
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>
    </div>
  );
};

/* ============================================================
   § 02.02 · CAIXA DE APROVAÇÃO (full)
   Lista única, focada, com filtros & multi-seleção.
   ============================================================ */
const ApprovalInbox = () => {
  const items = [
    { code: 'POP-QUA-0148', title: 'Inspeção de recebimento — solda MIG/MAG', author: 'Carla Pinheiro', area: 'Qualidade', deadline: 'Hoje · 18:00', urgency: 'high', kind: 'POP', stage: 'Revisão técnica', version: 'v3.2', sender_dept: 'Qualidade' },
    { code: 'IT-PROD-0072', title: 'Setup da prensa hidráulica P-440', author: 'João Aguiar', area: 'Produção', deadline: 'Hoje · 23:59', urgency: 'high', kind: 'IT', stage: 'Aprovação final', version: 'v1.4', sender_dept: 'Produção' },
    { code: 'POL-RH-0011', title: 'Política de afastamento parcial', author: 'Fábio Ribas', area: 'RH', deadline: 'Amanhã · 12:00', urgency: 'mid', kind: 'POL', stage: 'Revisão jurídica', version: 'v2.0', sender_dept: 'RH' },
    { code: 'DC-TI-0203', title: 'Diagrama de rede — DMZ planta 2', author: 'Ana Iuri', area: 'TI & Segurança', deadline: 'Em 2 dias', urgency: 'low', kind: 'DC', stage: 'Revisão técnica', version: 'v1.0', sender_dept: 'TI' },
    { code: 'POP-PROD-0312', title: 'Lubrificação rotativa — bomba B-12', author: 'Rafael Couto', area: 'Produção', deadline: 'Em 3 dias', urgency: 'low', kind: 'POP', stage: 'Aprovação final', version: 'v2.1', sender_dept: 'Produção' },
    { code: 'POL-FIN-0004', title: 'Aprovação de despesas até R$ 50k', author: 'Mauro Lacerda', area: 'Financeiro', deadline: 'Em 4 dias', urgency: 'low', kind: 'POL', stage: 'Revisão diretoria', version: 'v1.2', sender_dept: 'Financeiro' },
  ];
  const [selected, setSelected] = React.useState(new Set());
  const [filter, setFilter] = React.useState('all');

  const toggle = code => {
    const next = new Set(selected);
    next.has(code) ? next.delete(code) : next.add(code);
    setSelected(next);
  };

  return (
    <div style={{ flex: 1, display: 'flex', flexDirection: 'column', background: 'var(--bg)', overflow: 'hidden' }}>

      {/* Header */}
      <div style={{ padding: '24px 32px 16px', background: 'var(--surface)', borderBottom: '1px solid var(--border)' }}>
        <div className="kicker" style={{ marginBottom: 6 }}>§ 02 · Caixa de aprovação</div>
        <div style={{ display: 'flex', alignItems: 'flex-end', gap: 16 }}>
          <div>
            <h1 className="display-1" style={{ fontSize: 28, letterSpacing: '-0.02em', marginBottom: 4 }}>
              6 pendências · 2 vencem hoje
            </h1>
            <p className="caption" style={{ color: 'var(--text-muted)' }}>
              ISO 9001 cl. 7.5.3 — segregação de funções aplicada. Você não pode aprovar documentos que assinou como autor.
            </p>
          </div>
          <span style={{ flex: 1 }}/>
          <button className="btn btn-sm">Exportar relatório</button>
        </div>
      </div>

      {/* Filter strip */}
      <div style={{ padding: '12px 32px', background: 'var(--surface)', borderBottom: '1px solid var(--border)', display: 'flex', alignItems: 'center', gap: 8 }}>
        <span className="caption mono" style={{ color: 'var(--text-faint)', marginRight: 4 }}>Filtros:</span>
        {[
          { id: 'all', label: 'Todas', count: 6 },
          { id: 'urgent', label: 'Urgentes', count: 2 },
          { id: 'mid', label: 'Esta semana', count: 1 },
          { id: 'low', label: 'Próximas', count: 3 },
        ].map(f => (
          <button key={f.id} onClick={() => setFilter(f.id)} className="btn btn-sm" style={{
            background: filter === f.id ? 'var(--text)' : 'var(--surface-2)',
            color: filter === f.id ? 'var(--surface)' : 'var(--text-soft)',
            borderColor: filter === f.id ? 'var(--text)' : 'var(--border)'
          }}>
            {f.label} <span className="mono" style={{ marginLeft: 6, opacity: 0.7 }}>{f.count}</span>
          </button>
        ))}
        <span style={{ flex: 1 }}/>
        <button className="btn btn-sm btn-ghost"><Icon name="filter" size={14}/>Por área</button>
        <button className="btn btn-sm btn-ghost"><Icon name="filter" size={14}/>Por tipo</button>
      </div>

      {/* Bulk action bar (when selected) */}
      {selected.size > 0 && (
        <div style={{ padding: '10px 32px', background: 'var(--brand-pale)', borderBottom: '1px solid var(--border)', display: 'flex', alignItems: 'center', gap: 12, fontSize: 12.5 }}>
          <span style={{ fontWeight: 600, color: 'var(--brand)' }}>{selected.size} selecionado{selected.size > 1 ? 's' : ''}</span>
          <button className="btn btn-sm btn-primary"><Icon name="check" size={13}/>Aprovar em lote</button>
          <button className="btn btn-sm">Devolver com ajuste</button>
          <button className="btn btn-sm btn-ghost">Atribuir a outro</button>
          <span style={{ flex: 1 }}/>
          <button className="btn btn-sm btn-ghost" onClick={() => setSelected(new Set())}><Icon name="x" size={13}/>Limpar</button>
        </div>
      )}

      {/* Table */}
      <div style={{ flex: 1, overflow: 'auto', padding: '20px 32px 32px' }}>
        <div className="card" style={{ padding: 0, overflow: 'hidden' }}>
          <table style={{ width: '100%', borderCollapse: 'collapse' }}>
            <thead style={{ background: 'var(--surface-2)' }}>
              <tr>
                <th style={th(50)}><input type="checkbox" style={{ accentColor: 'var(--brand)' }}/></th>
                <th style={th(130)}>CÓDIGO</th>
                <th style={{ ...th(), textAlign: 'left' }}>DOCUMENTO</th>
                <th style={th(140)}>AUTOR</th>
                <th style={th(110)}>ETAPA</th>
                <th style={th(140)}>PRAZO</th>
                <th style={th(80)}></th>
              </tr>
            </thead>
            <tbody>
              {items.map((it, i) => {
                const checked = selected.has(it.code);
                return (
                  <tr key={i} style={{ borderTop: '1px solid var(--border)', background: checked ? 'var(--brand-pale)' : 'var(--surface)' }}>
                    <td style={td()}><input type="checkbox" checked={checked} onChange={() => toggle(it.code)} style={{ accentColor: 'var(--brand)' }}/></td>
                    <td style={td()}>
                      <div className="mono" style={{ fontSize: 11.5, fontWeight: 600, color: 'var(--text-soft)' }}>{it.code}</div>
                      <div className="caption mono" style={{ color: 'var(--text-faint)' }}>{it.version}</div>
                    </td>
                    <td style={{ ...td(), paddingLeft: 0 }}>
                      <div style={{ fontSize: 13, fontWeight: 500, color: 'var(--text)', marginBottom: 3 }}>{it.title}</div>
                      <div style={{ fontSize: 11, color: 'var(--text-muted)' }}>
                        <span className="mono" style={{ color: 'var(--text-soft)', fontWeight: 500 }}>{it.kind}</span>
                        <span style={{ margin: '0 6px', color: 'var(--text-faint)' }}>·</span>
                        <span>{it.area}</span>
                      </div>
                    </td>
                    <td style={td()}>
                      <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                        <Avatar name={it.author} size="sm"/>
                        <span style={{ fontSize: 12, color: 'var(--text-soft)' }}>{it.author}</span>
                      </div>
                    </td>
                    <td style={td()}>
                      <span className="caption" style={{ color: 'var(--text-soft)' }}>{it.stage}</span>
                    </td>
                    <td style={td()}>
                      <div style={{ display: 'flex', alignItems: 'center', gap: 6, fontSize: 11.5, fontFamily: 'var(--font-mono)', fontWeight: 600,
                        color: it.urgency === 'high' ? 'var(--accent)' : it.urgency === 'mid' ? 'var(--warning)' : 'var(--text-muted)' }}>
                        {it.urgency === 'high' && <span style={{ width: 5, height: 5, borderRadius: '50%', background: 'var(--accent)' }}/>}
                        {it.deadline}
                      </div>
                    </td>
                    <td style={{ ...td(), textAlign: 'right' }}>
                      <button className="btn btn-sm btn-primary">Abrir</button>
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>

        <div style={{ padding: '14px 4px 0', display: 'flex', alignItems: 'center', fontSize: 11, color: 'var(--text-muted)', fontFamily: 'var(--font-mono)' }}>
          <span>6 itens · ordenado por prazo asc.</span>
          <span style={{ flex: 1 }}/>
          <span>Atualizado há 12s · auto-refresh em 30s</span>
        </div>
      </div>
    </div>
  );
};
const th = (w) => ({ padding: '10px 14px', fontSize: 10, fontWeight: 600, color: 'var(--text-muted)', letterSpacing: '0.08em', fontFamily: 'var(--font-mono)', textAlign: 'left', borderBottom: '1px solid var(--border)', ...(w ? { width: w } : {}) });
const td = () => ({ padding: '14px', verticalAlign: 'middle' });

/* ============================================================
   § 06.02 · DETALHE + SIGNOFF
   Documento + diff + assinatura criptográfica.
   ============================================================ */
const SignoffDetail = () => {
  const [decision, setDecision] = React.useState(null); // 'approve' | 'reject' | null

  return (
    <div style={{ flex: 1, display: 'flex', background: 'var(--bg)', overflow: 'hidden' }}>

      {/* === MAIN: document + diff === */}
      <div style={{ flex: 1, display: 'flex', flexDirection: 'column', overflow: 'hidden' }}>

        {/* Doc header */}
        <div style={{ padding: '20px 32px 16px', background: 'var(--surface)', borderBottom: '1px solid var(--border)' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 10, marginBottom: 6 }}>
            <span className="kicker">§ Aprovação · etapa 2 de 2 · revisão final</span>
            <span style={{ flex: 1 }}/>
            <span className="caption mono" style={{ color: 'var(--text-faint)' }}>doc_pop_qua_0148_v3.2</span>
          </div>
          <div style={{ display: 'flex', alignItems: 'center', gap: 12, marginBottom: 8 }}>
            <span className="mono" style={{ fontSize: 13, fontWeight: 600, color: 'var(--brand)', padding: '3px 8px', background: 'var(--brand-pale)', borderRadius: 4, flexShrink: 0 }}>
              POP-QUA-0148
            </span>
            <h1 className="display-1" style={{ fontSize: 22, letterSpacing: '-0.015em', flex: 1, minWidth: 0, lineHeight: 1.25 }}>
              Inspeção de recebimento — solda MIG/MAG
            </h1>
          </div>
          <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
            <span className="pill pill-review"><span className="dot"/>Em revisão</span>
            <span className="caption mono" style={{ color: 'var(--text-muted)' }}>v3.2</span>
            <span style={{ color: 'var(--text-faint)' }}>·</span>
            <span className="caption mono" style={{ color: 'var(--accent)', fontWeight: 600 }}>vence hoje 18:00</span>
          </div>
          <div style={{ display: 'flex', alignItems: 'center', gap: 16, marginTop: 12, fontSize: 11.5, color: 'var(--text-muted)' }}>
            <span style={{ display: 'flex', alignItems: 'center', gap: 6 }}><Avatar name="Carla Pinheiro" size="sm"/>Carla Pinheiro · autora</span>
            <span style={{ color: 'var(--text-faint)' }}>·</span>
            <span>Qualidade</span>
            <span style={{ color: 'var(--text-faint)' }}>·</span>
            <span>Enviado para revisão há 3h</span>
            <span style={{ color: 'var(--text-faint)' }}>·</span>
            <span>1ª aprovação por <strong style={{ color: 'var(--text)' }}>Bruno Said</strong> · 11:08</span>
          </div>
        </div>

        {/* Tab strip */}
        <div style={{ padding: '0 32px', background: 'var(--surface)', borderBottom: '1px solid var(--border)', display: 'flex', gap: 4 }}>
          {[
            { id: 'doc', label: 'Documento', active: true },
            { id: 'diff', label: 'Mudanças vs v3.1', count: 14 },
            { id: 'comments', label: 'Comentários', count: 3 },
            { id: 'audit', label: 'Trilha' },
          ].map(t => (
            <button key={t.id} className="btn btn-sm" style={{
              background: 'transparent', border: 'none', borderBottom: t.active ? '2px solid var(--brand)' : '2px solid transparent',
              borderRadius: 0, color: t.active ? 'var(--text)' : 'var(--text-muted)',
              fontWeight: t.active ? 600 : 400, padding: '12px 14px', fontSize: 12.5
            }}>
              {t.label}
              {t.count !== undefined && <span className="mono" style={{ marginLeft: 6, fontSize: 10.5, color: 'var(--text-faint)' }}>{t.count}</span>}
            </button>
          ))}
        </div>

        {/* Document body — A4-ish */}
        <div style={{ flex: 1, overflow: 'auto', padding: '32px 0', background: 'var(--bg)' }}>
          <div style={{ maxWidth: 720, margin: '0 auto', background: 'var(--surface)', boxShadow: 'var(--shadow-2)', padding: '48px 64px', minHeight: 600, position: 'relative' }}>
            {/* Watermark */}
            <div style={{ position: 'absolute', top: 24, right: 28, fontSize: 9, fontFamily: 'var(--font-mono)', color: 'var(--text-faint)', letterSpacing: '0.1em', textAlign: 'right' }}>
              CÓPIA PARA REVISÃO<br/>NÃO CONTROLADA
            </div>

            <div className="kicker" style={{ marginBottom: 8 }}>POP · QUALIDADE · v3.2</div>
            <h2 style={{ fontSize: 22, fontWeight: 600, letterSpacing: '-0.015em', marginBottom: 24, color: 'var(--text)' }}>
              Inspeção de recebimento — solda MIG/MAG
            </h2>

            <div style={{ borderLeft: '3px solid var(--success)', background: 'var(--success-bg)', padding: '8px 12px', marginBottom: 24, fontSize: 11.5, color: 'var(--success)', fontFamily: 'var(--font-mono)' }}>
              + ADICIONADO §3.2 — Critérios para inspeção visual ampliada conforme NBR ISO 5817 · classe B
            </div>

            <h3 style={{ fontSize: 14, fontWeight: 600, marginBottom: 8, color: 'var(--text)' }}>1. Objetivo</h3>
            <p style={{ fontSize: 12.5, lineHeight: 1.7, color: 'var(--text-soft)', marginBottom: 18 }}>
              Estabelecer os critérios e procedimentos para inspeção de recebimento de peças metálicas com soldagem MIG/MAG,
              garantindo a conformidade dos lotes recebidos com as especificações técnicas e contratuais antes da liberação para produção.
            </p>

            <h3 style={{ fontSize: 14, fontWeight: 600, marginBottom: 8, color: 'var(--text)' }}>2. Aplicação</h3>
            <p style={{ fontSize: 12.5, lineHeight: 1.7, color: 'var(--text-soft)', marginBottom: 18 }}>
              Aplica-se a todos os lotes de peças soldadas com processos GMAW (MIG/MAG) recebidos pela área de Recebimento Fiscal,
              independente do fornecedor ou volume.
            </p>

            <h3 style={{ fontSize: 14, fontWeight: 600, marginBottom: 8, color: 'var(--text)' }}>3. Procedimento</h3>
            <ol style={{ fontSize: 12.5, lineHeight: 1.75, color: 'var(--text-soft)', paddingLeft: 22, marginBottom: 18 }}>
              <li><strong>3.1</strong> Verificar a documentação de envio (NF, certificado de qualidade, RPS).</li>
              <li style={{ background: 'var(--success-bg)', padding: '4px 6px', margin: '4px -6px', borderRadius: 3 }}>
                <strong>3.2</strong> <span style={{ background: 'var(--success-bg)', color: 'var(--success)', fontWeight: 600 }}>Realizar inspeção visual ampliada conforme NBR ISO 5817 classe B em 100% dos cordões.</span>
              </li>
              <li><strong>3.3</strong> Conferir dimensões com paquímetro digital calibrado (POP-QUA-0149).</li>
              <li><strong>3.4</strong> Em caso de não-conformidade, segregar e abrir RNC no sistema.</li>
            </ol>

            <h3 style={{ fontSize: 14, fontWeight: 600, marginBottom: 8, color: 'var(--text)' }}>4. Registros</h3>
            <p style={{ fontSize: 12.5, lineHeight: 1.7, color: 'var(--text-soft)', marginBottom: 18 }}>
              Os resultados serão registrados no formulário FQ-INS-001 e arquivados por 5 anos conforme política de retenção.
            </p>
          </div>
        </div>
      </div>

      {/* === SIDEBAR: signoff panel === */}
      <aside style={{ width: 360, background: 'var(--surface)', borderLeft: '1px solid var(--border)', display: 'flex', flexDirection: 'column', flexShrink: 0 }}>

        {/* Signoff header */}
        <div style={{ padding: '20px 22px 16px', borderBottom: '1px solid var(--border)' }}>
          <div className="kicker" style={{ marginBottom: 8, color: 'var(--brand)' }}>§ DECISÃO REQUERIDA</div>
          <h3 className="h3" style={{ fontSize: 16, marginBottom: 6 }}>Sua aprovação</h3>
          <p className="caption" style={{ color: 'var(--text-muted)', lineHeight: 1.5 }}>
            Aprovação final · etapa 2 de 2. Esta ação congelará a versão e disparará fanout PDF para todas as cópias controladas.
          </p>
        </div>

        {/* Approval flow viz */}
        <div style={{ padding: '18px 22px', borderBottom: '1px solid var(--border)' }}>
          <div className="kicker" style={{ marginBottom: 12 }}>Fluxo de aprovação</div>
          <div style={{ display: 'flex', flexDirection: 'column', gap: 0 }}>
            {[
              { who: 'Carla Pinheiro', role: 'Autora', status: 'sent', t: 'há 3h' },
              { who: 'Bruno Said', role: 'Revisão técnica', status: 'approved', t: '11:08' },
              { who: 'Marina S (você)', role: 'Aprovação final', status: 'pending', t: '—' },
            ].map((s, i, arr) => (
              <div key={i} style={{ display: 'flex', alignItems: 'flex-start', gap: 10, position: 'relative', paddingBottom: i < arr.length - 1 ? 14 : 0 }}>
                {i < arr.length - 1 && <span style={{ position: 'absolute', left: 9, top: 22, bottom: 0, width: 1, background: 'var(--border)' }}/>}
                <span style={{
                  width: 18, height: 18, borderRadius: '50%', flexShrink: 0,
                  background: s.status === 'approved' ? 'var(--success)' : s.status === 'sent' ? 'var(--text-muted)' : 'var(--surface)',
                  border: s.status === 'pending' ? '2px dashed var(--brand)' : 'none',
                  display: 'grid', placeItems: 'center', color: 'white', marginTop: 1
                }}>
                  {s.status === 'approved' && <Icon name="check" size={11}/>}
                  {s.status === 'sent' && <span style={{ width: 5, height: 5, borderRadius: '50%', background: 'white' }}/>}
                </span>
                <div style={{ flex: 1, fontSize: 12 }}>
                  <div style={{ fontWeight: 500, color: s.status === 'pending' ? 'var(--brand)' : 'var(--text)' }}>{s.who}</div>
                  <div className="caption" style={{ color: 'var(--text-muted)' }}>{s.role} · <span className="mono">{s.t}</span></div>
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Decision box */}
        <div style={{ padding: '20px 22px', flex: 1, overflow: 'auto' }}>
          <div className="kicker" style={{ marginBottom: 12 }}>Decisão</div>

          <div style={{ display: 'flex', flexDirection: 'column', gap: 8, marginBottom: 16 }}>
            <button onClick={() => setDecision('approve')} style={{
              padding: '14px 16px', borderRadius: 6,
              border: `1.5px solid ${decision === 'approve' ? 'var(--success)' : 'var(--border)'}`,
              background: decision === 'approve' ? 'var(--success-bg)' : 'var(--surface)',
              cursor: 'pointer', textAlign: 'left',
              display: 'flex', alignItems: 'flex-start', gap: 10
            }}>
              <span style={{ width: 16, height: 16, borderRadius: '50%', border: `1.5px solid ${decision === 'approve' ? 'var(--success)' : 'var(--border-strong)'}`, marginTop: 2, display: 'grid', placeItems: 'center', flexShrink: 0 }}>
                {decision === 'approve' && <span style={{ width: 8, height: 8, borderRadius: '50%', background: 'var(--success)' }}/>}
              </span>
              <div style={{ flex: 1 }}>
                <div style={{ fontSize: 13, fontWeight: 600, color: decision === 'approve' ? 'var(--success)' : 'var(--text)', marginBottom: 2 }}>Aprovar</div>
                <div className="caption" style={{ color: 'var(--text-muted)' }}>Congela a versão · dispara fanout PDF · publica</div>
              </div>
            </button>

            <button onClick={() => setDecision('reject')} style={{
              padding: '14px 16px', borderRadius: 6,
              border: `1.5px solid ${decision === 'reject' ? 'var(--danger)' : 'var(--border)'}`,
              background: decision === 'reject' ? 'var(--danger-bg)' : 'var(--surface)',
              cursor: 'pointer', textAlign: 'left',
              display: 'flex', alignItems: 'flex-start', gap: 10
            }}>
              <span style={{ width: 16, height: 16, borderRadius: '50%', border: `1.5px solid ${decision === 'reject' ? 'var(--danger)' : 'var(--border-strong)'}`, marginTop: 2, display: 'grid', placeItems: 'center', flexShrink: 0 }}>
                {decision === 'reject' && <span style={{ width: 8, height: 8, borderRadius: '50%', background: 'var(--danger)' }}/>}
              </span>
              <div style={{ flex: 1 }}>
                <div style={{ fontSize: 13, fontWeight: 600, color: decision === 'reject' ? 'var(--danger)' : 'var(--text)', marginBottom: 2 }}>Devolver com ajuste</div>
                <div className="caption" style={{ color: 'var(--text-muted)' }}>Volta para a autora · requer comentário</div>
              </div>
            </button>
          </div>

          <label style={{ display: 'block', marginBottom: 14 }}>
            <div className="kicker" style={{ marginBottom: 6 }}>Justificativa {decision === 'reject' && <span style={{ color: 'var(--danger)' }}>· obrigatória</span>}</div>
            <textarea className="input" rows={3} placeholder="Comentário visível na trilha de auditoria…" style={{ resize: 'none', fontSize: 12.5, padding: 10, fontFamily: 'var(--font-sans)' }}/>
          </label>

          <div style={{ background: 'var(--surface-2)', border: '1px solid var(--border)', borderRadius: 6, padding: 12, marginBottom: 14 }}>
            <div className="kicker" style={{ marginBottom: 6 }}>Assinatura</div>
            <div style={{ display: 'flex', alignItems: 'center', gap: 10, marginBottom: 8 }}>
              <Avatar name="Marina S"/>
              <div style={{ fontSize: 12, lineHeight: 1.4 }}>
                <div style={{ fontWeight: 500, color: 'var(--text)' }}>Marina Salgado</div>
                <div className="caption" style={{ color: 'var(--text-muted)' }}>Gerente de Qualidade · marina.s@metalcorp.br</div>
              </div>
            </div>
            <div className="caption" style={{ color: 'var(--text-muted)', lineHeight: 1.55, padding: '8px 10px', background: 'var(--surface)', borderRadius: 4, fontFamily: 'var(--font-mono)', fontSize: 10.5 }}>
              SHA-256 · timestamp ICP-Brasil · IP 200.142.84.12
            </div>
          </div>

          <label style={{ display: 'flex', alignItems: 'flex-start', gap: 8, fontSize: 11.5, color: 'var(--text-muted)', lineHeight: 1.5, marginBottom: 16, cursor: 'pointer' }}>
            <input type="checkbox" style={{ marginTop: 2, accentColor: 'var(--brand)' }}/>
            <span>Confirmo que revisei o conteúdo integralmente e que esta decisão tem efeito legal conforme MP 2.200-2/2001.</span>
          </label>
        </div>

        {/* Action footer */}
        <div style={{ padding: '14px 22px', borderTop: '1px solid var(--border)', background: 'var(--surface-2)', display: 'flex', gap: 8 }}>
          <button className="btn">Cancelar</button>
          <span style={{ flex: 1 }}/>
          <button className="btn btn-primary" disabled={!decision} style={{ opacity: decision ? 1 : 0.5 }}>
            <Icon name="lock" size={13}/>
            {decision === 'approve' ? 'Assinar e aprovar' : decision === 'reject' ? 'Assinar e devolver' : 'Selecione uma decisão'}
          </button>
        </div>
      </aside>
    </div>
  );
};

window.MD_PRIORITY = { Dashboard, ApprovalInbox, SignoffDetail };
