/* global React, MD */
const { Icon, Status, Avatar, Logo, Sidebar, Toolbar } = MD;
const { useState } = React;

// === SHARED MOCK DATA ===
const DOCS = [
  { code: 'POP-RH-001', title: 'Admissão e Onboarding de Colaboradores', area: 'RH', profile: 'POP', rev: 'v4', status: 'frozen', author: 'Marina Silveira', updated: '2 dias atrás', approvers: ['MS', 'RC', 'JP'], effective: '15/04/2026' },
  { code: 'POP-QUA-118', title: 'Inspeção Dimensional em Componentes Usinados', area: 'Qualidade', profile: 'POP', rev: 'v2', status: 'review', author: 'Rafael Castro', updated: '4 horas atrás', approvers: ['JP', 'AL'], effective: '—' },
  { code: 'DC-RH-042', title: 'Descrição de Cargo — Analista Fiscal Sênior', area: 'RH', profile: 'DC', rev: 'v1', status: 'draft', author: 'Camila Tavares', updated: '23 minutos atrás', approvers: [], effective: '—' },
  { code: 'POL-TI-007', title: 'Política de Classificação da Informação', area: 'TI & Segurança', profile: 'POL', rev: 'v3', status: 'frozen', author: 'André Lima', updated: 'ontem', approvers: ['AL', 'RC', 'BM', 'JP'], effective: '01/03/2026' },
  { code: 'IT-PROD-204', title: 'Calibração de Torquímetro', area: 'Produção', profile: 'IT', rev: 'v6', status: 'approved', author: 'Bruno Mendes', updated: '6 horas atrás', approvers: ['BM', 'RC'], effective: '08/05/2026' },
  { code: 'POP-FIN-019', title: 'Conciliação Bancária Mensal', area: 'Financeiro', profile: 'POP', rev: 'v2', status: 'frozen', author: 'Juliana Prado', updated: '3 dias atrás', approvers: ['JP', 'MS'], effective: '20/02/2026' },
  { code: 'POP-QUA-119', title: 'Análise de Causa Raiz (RCA)', area: 'Qualidade', profile: 'POP', rev: 'v1', status: 'review', author: 'Rafael Castro', updated: '1 hora atrás', approvers: ['MS', 'AL'], effective: '—' },
  { code: 'DC-PROD-088', title: 'Descrição de Cargo — Líder de Linha', area: 'Produção', profile: 'DC', rev: 'v2', status: 'frozen', author: 'Bruno Mendes', updated: '5 dias atrás', approvers: ['BM', 'RC'], effective: '02/04/2026' },
  { code: 'POL-RH-003', title: 'Política de Equidade e Diversidade', area: 'RH', profile: 'POL', rev: 'v2', status: 'rejected', author: 'Camila Tavares', updated: '2 dias atrás', approvers: ['JP'], effective: '—' },
  { code: 'POP-TI-022', title: 'Resposta a Incidentes de Segurança', area: 'TI & Segurança', profile: 'POP', rev: 'v4', status: 'frozen', author: 'André Lima', updated: '1 semana atrás', approvers: ['AL', 'RC', 'JP'], effective: '18/01/2026' },
  { code: 'POP-PROD-088', title: 'Setup de Linha — Mudança de Ferramental', area: 'Produção', profile: 'POP', rev: 'v3', status: 'frozen', author: 'Bruno Mendes', updated: '1 mês atrás', approvers: ['BM', 'RC', 'JP'], effective: '12/04/2026' },
  { code: 'IT-QUA-067', title: 'Verificação de Calibração de Paquímetros', area: 'Qualidade', profile: 'IT', rev: 'v2', status: 'frozen', author: 'Rafael Castro', updated: '2 semanas atrás', approvers: ['RC', 'JP'], effective: '28/03/2026' },
];

// === [C6+C7] LIBRARY — Tabela densa Wine + collapsible activity sidebar ===
const SelectedLibrary = () => {
  const [activityOpen, setActivityOpen] = useState(true);
  const [filter, setFilter] = useState('todos');

  const filterTabs = [
    { id: 'todos', label: 'Todos', count: 2847 },
    { id: 'meus', label: 'Meus', count: 38 },
    { id: 'review', label: 'Em revisão', count: 8 },
    { id: 'pendente', label: 'Aprovação pendente', count: 3, urgent: true },
    { id: 'frozen', label: 'Frozen', count: 2612 },
    { id: 'rascunho', label: 'Rascunhos', count: 23 },
  ];

  return (
    <div style={{ flex: 1, display: 'flex', overflow: 'hidden' }}>
      {/* MAIN CONTENT */}
      <div style={{ flex: 1, overflow: 'auto', background: 'var(--bg)' }}>
        <div style={{ padding: '24px 28px 0' }}>
          <div style={{ display: 'flex', alignItems: 'flex-end', gap: 16, marginBottom: 4 }}>
            <div>
              <div className="kicker" style={{ marginBottom: 6 }}>Documentos · Biblioteca</div>
              <h1 className="display-1">Acervo controlado</h1>
            </div>
            <span className="spacer"/>
            <div style={{ display: 'flex', alignItems: 'center', gap: 8, color: 'var(--text-muted)', fontSize: 12 }}>
              <span className="mono">2.847 documentos</span>
              <span>·</span>
              <span>Última fanout <span className="mono" style={{ color: 'var(--text-soft)' }}>14:32</span></span>
            </div>
            <button className="btn btn-sm" onClick={() => setActivityOpen(!activityOpen)} title={activityOpen ? 'Recolher atividade' : 'Mostrar atividade'}>
              <Icon name={activityOpen ? 'chevron' : 'history'} size={13}/>
              {activityOpen ? 'Recolher atividade' : 'Atividade'}
            </button>
          </div>
          <p className="body" style={{ color: 'var(--text-muted)', maxWidth: 580, marginTop: 8 }}>
            Documentos controlados por código, perfil e área. Filtre por estado para revisões pendentes ou trilha de auditoria.
          </p>

          {/* Stat strip */}
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(4, 1fr)', gap: 12, margin: '24px 0 20px' }}>
            {[
              { l: 'Em revisão', v: '8', d: '+3 hoje', c: 'var(--warning)' },
              { l: 'Aprovação pendente', v: '3', d: '2 vencendo', c: 'var(--accent)' },
              { l: 'Frozen este mês', v: '47', d: '+12% vs anterior', c: 'var(--success)' },
              { l: 'Próx. revisão', v: '23', d: 'Maio · Jun', c: 'var(--info)' },
            ].map((s, i) => (
              <div key={i} className="card" style={{ padding: '14px 16px', position: 'relative', overflow: 'hidden' }}>
                <div className="kicker" style={{ marginBottom: 6 }}>{s.l}</div>
                <div style={{ display: 'flex', alignItems: 'baseline', gap: 8 }}>
                  <span className="display-1 num" style={{ fontSize: 28 }}>{s.v}</span>
                  <span style={{ fontSize: 11, color: s.c }}>{s.d}</span>
                </div>
                <span style={{ position: 'absolute', left: 0, top: 0, bottom: 0, width: 3, background: s.c, opacity: 0.7 }}/>
              </div>
            ))}
          </div>

          {/* Filter tabs */}
          <div style={{ display: 'flex', alignItems: 'center', gap: 4, marginBottom: 12, borderBottom: '1px solid var(--border)' }}>
            {filterTabs.map(t => (
              <button key={t.id} onClick={() => setFilter(t.id)} style={{
                padding: '8px 14px', fontSize: 13, border: 'none', background: 'none',
                borderBottom: filter === t.id ? '2px solid var(--brand)' : '2px solid transparent',
                color: filter === t.id ? 'var(--text)' : 'var(--text-muted)',
                fontWeight: filter === t.id ? 500 : 400, cursor: 'pointer', marginBottom: -1,
                display: 'flex', alignItems: 'center', gap: 6,
              }}>
                {t.label}
                <span className="mono" style={{ fontSize: 10.5, color: t.urgent ? 'var(--accent)' : 'var(--text-faint)' }}>{t.count}</span>
              </button>
            ))}
            <span className="spacer"/>
            <button className="btn btn-sm btn-ghost"><Icon name="filter" size={13}/>Filtros</button>
            <button className="btn btn-sm btn-ghost"><Icon name="download" size={13}/>Exportar</button>
          </div>
        </div>

        {/* Table */}
        <div style={{ padding: '0 28px 28px' }}>
          <div className="card" style={{ padding: 0, overflow: 'hidden' }}>
            <div style={{ display: 'grid', gridTemplateColumns: '160px 1fr 100px 90px 70px 110px 100px 110px 36px', gap: 12, padding: '10px 16px', background: 'var(--surface-2)', borderBottom: '1px solid var(--border)' }}>
              {['Código', 'Título', 'Área', 'Perfil', 'Rev.', 'Estado', 'Autor', 'Atualizado', ''].map((h, i) => <div key={i} className="kicker" style={{ fontSize: 10 }}>{h}</div>)}
            </div>
            {DOCS.map((d, i) => (
              <div key={i} className="row" style={{ display: 'grid', gridTemplateColumns: '160px 1fr 100px 90px 70px 110px 100px 110px 36px', gap: 12, padding: '11px 16px', alignItems: 'center', borderBottom: i < DOCS.length - 1 ? '1px solid var(--border)' : 'none', cursor: 'pointer' }}>
                <span className="code-chip mono">{d.code}</span>
                <span style={{ fontSize: 13.5, fontWeight: 500, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{d.title}</span>
                <span className="body-sm" style={{ color: 'var(--text-soft)' }}>{d.area}</span>
                <span className="mono" style={{ fontSize: 11, color: 'var(--text-muted)' }}>{d.profile}</span>
                <span className="mono" style={{ fontSize: 12 }}>{d.rev}</span>
                <Status s={d.status}/>
                <div style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
                  <Avatar name={d.author} size="sm"/>
                  <span style={{ fontSize: 12 }}>{d.author.split(' ')[0]}</span>
                </div>
                <span className="caption mono">{d.updated}</span>
                <button className="btn btn-icon btn-sm btn-ghost"><Icon name="more" size={13}/></button>
              </div>
            ))}
          </div>
        </div>
      </div>

      {/* COLLAPSIBLE ACTIVITY SIDEBAR (right) */}
      {activityOpen && (
        <aside style={{ width: 320, borderLeft: '1px solid var(--border)', background: 'var(--surface-2)', overflow: 'auto', flexShrink: 0 }}>
          <div style={{ padding: '18px 20px 12px', display: 'flex', alignItems: 'center' }}>
            <div>
              <div className="h3" style={{ fontSize: 14 }}>Atividade & auditoria</div>
              <div className="caption">Tempo real · últimas 8h</div>
            </div>
            <span className="spacer"/>
            <button className="btn btn-icon btn-sm btn-ghost" onClick={() => setActivityOpen(false)}><Icon name="x" size={13}/></button>
          </div>
          <div className="divider"/>

          {/* Pending approvals */}
          <div style={{ padding: '14px 20px' }}>
            <div className="kicker" style={{ marginBottom: 10 }}>Sua caixa · 3 pendentes</div>
            {[
              { code: 'POP-QUA-118', t: 'Vence em 6h', urgent: true },
              { code: 'POP-QUA-119', t: 'Vence em 2d' },
              { code: 'POL-RH-003', t: 'Sem prazo' },
            ].map((p, i) => (
              <div key={i} style={{ display: 'flex', alignItems: 'center', gap: 8, padding: '8px 0', borderBottom: i < 2 ? '1px dashed var(--border)' : 'none' }}>
                <span className="code-chip mono" style={{ fontSize: 10.5 }}>{p.code}</span>
                <span className="spacer"/>
                <span className="tiny" style={{ color: p.urgent ? 'var(--accent)' : 'var(--text-muted)', fontWeight: p.urgent ? 500 : 400 }}>{p.t}</span>
              </div>
            ))}
            <button className="btn btn-sm" style={{ width: '100%', justifyContent: 'center', marginTop: 10 }}>Abrir caixa de aprovação</button>
          </div>

          <div className="divider"/>

          {/* Activity stream */}
          <div style={{ padding: '14px 20px' }}>
            <div className="kicker" style={{ marginBottom: 12 }}>Trilha de auditoria</div>
            <div style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>
              {[
                { who: 'Juliana P.', what: 'aprovou', code: 'POP-FIN-019', t: '14:32' },
                { who: 'Sistema', what: 'gerou PDF', code: 'IT-PROD-204', t: '13:51' },
                { who: 'Rafael C.', what: 'submeteu', code: 'POP-QUA-119', t: '13:58' },
                { who: 'André L.', what: 'editou', code: 'POL-TI-007', t: '11:24' },
                { who: 'Marina S.', what: 'arquivou', code: 'POP-RH-099', t: '10:17' },
                { who: 'Bruno M.', what: 'congelou', code: 'IT-PROD-203', t: '09:44' },
                { who: 'Camila T.', what: 'criou rascunho', code: 'DC-RH-042', t: '09:12' },
              ].map((a, i) => (
                <div key={i} style={{ display: 'flex', gap: 10, alignItems: 'flex-start' }}>
                  <Avatar name={a.who} size="sm"/>
                  <div style={{ flex: 1, fontSize: 12, lineHeight: 1.45 }}>
                    <span style={{ fontWeight: 500 }}>{a.who}</span> <span style={{ color: 'var(--text-muted)' }}>{a.what}</span> <span className="code-chip mono" style={{ fontSize: 10.5, padding: '1px 5px' }}>{a.code}</span>
                  </div>
                  <span className="tiny mono" style={{ color: 'var(--text-faint)' }}>{a.t}</span>
                </div>
              ))}
            </div>
            <button className="btn btn-sm btn-ghost" style={{ width: '100%', marginTop: 14, justifyContent: 'center' }}>Ver tudo →</button>
          </div>
        </aside>
      )}

      {/* Floating reopen tab when collapsed */}
      {!activityOpen && (
        <button onClick={() => setActivityOpen(true)} style={{
          position: 'absolute', right: 0, top: 120,
          writingMode: 'vertical-rl', transform: 'rotate(180deg)',
          padding: '12px 8px', background: 'var(--surface)', border: '1px solid var(--border)', borderRight: 'none',
          borderRadius: 'var(--r-2) 0 0 var(--r-2)', cursor: 'pointer', fontSize: 11,
          color: 'var(--text-muted)', display: 'flex', alignItems: 'center', gap: 8,
        }}>
          <Icon name="history" size={12}/> Atividade
        </button>
      )}
    </div>
  );
};

window.MD_LIB_SELECTED = { SelectedLibrary, DOCS };
