/* global React, MD */
const { Icon, Status, Avatar, Logo, Sidebar, Toolbar } = MD;

// === MOCK DATA ===
const DOCS = [
  { code: 'POP-RH-001', title: 'Admissão e Onboarding de Colaboradores', area: 'RH', profile: 'POP', rev: 'v4', status: 'frozen', author: 'Marina Silveira', updated: '2 dias atrás', approvers: ['MS', 'RC', 'JP'], hash: '7f3a91…b2', effective: '2026-04-15' },
  { code: 'POP-QUA-118', title: 'Inspeção Dimensional em Componentes Usinados', area: 'Qualidade', profile: 'POP', rev: 'v2', status: 'review', author: 'Rafael Castro', updated: '4 horas atrás', approvers: ['JP', 'AL'], hash: '—', effective: '—' },
  { code: 'DC-RH-042', title: 'Descrição de Cargo — Analista Fiscal Sênior', area: 'RH', profile: 'DC', rev: 'v1', status: 'draft', author: 'Camila Tavares', updated: '23 minutos atrás', approvers: [], hash: '—', effective: '—' },
  { code: 'POL-TI-007', title: 'Política de Classificação da Informação', area: 'TI & Segurança', profile: 'POL', rev: 'v3', status: 'frozen', author: 'André Lima', updated: 'ontem', approvers: ['AL', 'RC', 'BM', 'JP'], hash: '92cf41…a8', effective: '2026-03-01' },
  { code: 'IT-PROD-204', title: 'Instrução de Trabalho — Calibração de Torquímetro', area: 'Produção', profile: 'IT', rev: 'v6', status: 'approved', author: 'Bruno Mendes', updated: '6 horas atrás', approvers: ['BM', 'RC'], hash: 'a14d22…f0', effective: '2026-05-08' },
  { code: 'POP-FIN-019', title: 'Conciliação Bancária Mensal', area: 'Financeiro', profile: 'POP', rev: 'v2', status: 'frozen', author: 'Juliana Prado', updated: '3 dias atrás', approvers: ['JP', 'MS'], hash: '4e8b77…1c', effective: '2026-02-20' },
  { code: 'POP-QUA-119', title: 'Análise de Causa Raiz (RCA) para Não-Conformidades', area: 'Qualidade', profile: 'POP', rev: 'v1', status: 'review', author: 'Rafael Castro', updated: '1 hora atrás', approvers: ['MS', 'AL'], hash: '—', effective: '—' },
  { code: 'DC-PROD-088', title: 'Descrição de Cargo — Líder de Linha', area: 'Produção', profile: 'DC', rev: 'v2', status: 'frozen', author: 'Bruno Mendes', updated: '5 dias atrás', approvers: ['BM', 'RC'], hash: 'e9a3c0…d4', effective: '2026-04-02' },
  { code: 'POL-RH-003', title: 'Política de Equidade e Diversidade', area: 'RH', profile: 'POL', rev: 'v2', status: 'rejected', author: 'Camila Tavares', updated: '2 dias atrás', approvers: ['JP'], hash: '—', effective: '—' },
  { code: 'POP-TI-022', title: 'Resposta a Incidentes de Segurança', area: 'TI & Segurança', profile: 'POP', rev: 'v4', status: 'frozen', author: 'André Lima', updated: '1 semana atrás', approvers: ['AL', 'RC', 'JP'], hash: '6d12bb…f9', effective: '2026-01-18' },
];

// === DOCUMENTS LIBRARY — DIRECTION A: TABLE-FIRST DENSE ===
const LibraryA = () => (
  <div style={{ flex: 1, overflow: 'auto', background: 'var(--bg)' }}>
    <div style={{ padding: '24px 28px 0', maxWidth: 1400 }}>
      {/* Header */}
      <div style={{ display: 'flex', alignItems: 'flex-end', gap: 16, marginBottom: 4 }}>
        <div>
          <div className="kicker" style={{ marginBottom: 6 }}>Documentos · Biblioteca</div>
          <h1 className="display-1">Acervo controlado</h1>
        </div>
        <div className="spacer"/>
        <div style={{ display: 'flex', alignItems: 'center', gap: 8, color: 'var(--text-muted)', fontSize: 12 }}>
          <span className="mono">2.847 documentos</span>
          <span>·</span>
          <span>Última fanout <span className="mono" style={{ color: 'var(--text-soft)' }}>14:32</span></span>
        </div>
      </div>
      <p className="body" style={{ color: 'var(--text-muted)', maxWidth: 580, marginTop: 8 }}>
        Documentos controlados por código, perfil e área. Filtre por estado para revisões pendentes ou trilha de auditoria.
      </p>

      {/* Stat strip */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(4, 1fr)', gap: 12, margin: '24px 0 20px' }}>
        {[
          { l: 'Em revisão', v: '8', d: '+3 hoje', c: 'var(--warning)' },
          { l: 'Aprovação pendente', v: '3', d: '2 vencendo', c: 'var(--accent)' },
          { l: 'Frozen este mês', v: '47', d: '+12% vs mês anterior', c: 'var(--success)' },
          { l: 'Próx. revisão programada', v: '23', d: 'Maio · Jun', c: 'var(--info)' },
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

      {/* Filter bar */}
      <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 12, flexWrap: 'wrap' }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 0, background: 'var(--surface)', border: '1px solid var(--border-strong)', borderRadius: 'var(--r-2)', overflow: 'hidden' }}>
          {['Todos', 'Frozen', 'Em revisão', 'Rascunho', 'Arquivado'].map((t, i) => (
            <button key={t} style={{ height: 30, padding: '0 12px', fontSize: 12, border: 'none', borderLeft: i ? '1px solid var(--border)' : 'none', background: i === 0 ? 'var(--surface-3)' : 'transparent', color: i === 0 ? 'var(--text)' : 'var(--text-soft)', fontWeight: i === 0 ? 500 : 400, cursor: 'pointer' }}>
              {t} {i === 0 && <span className="num" style={{ color: 'var(--text-muted)', marginLeft: 4 }}>2.847</span>}
            </button>
          ))}
        </div>
        <button className="btn btn-sm"><Icon name="filter" size={13}/>Área <Icon name="chevdown" size={11}/></button>
        <button className="btn btn-sm"><Icon name="filter" size={13}/>Perfil <Icon name="chevdown" size={11}/></button>
        <button className="btn btn-sm"><Icon name="filter" size={13}/>Autor <Icon name="chevdown" size={11}/></button>
        <button className="btn btn-sm"><Icon name="filter" size={13}/>Período <Icon name="chevdown" size={11}/></button>
        <span className="spacer"/>
        <div style={{ display: 'flex', background: 'var(--surface)', border: '1px solid var(--border-strong)', borderRadius: 'var(--r-2)' }}>
          <button className="btn btn-sm btn-ghost" style={{ borderRadius: 0, background: 'var(--surface-3)' }}><Icon name="list" size={13}/></button>
          <button className="btn btn-sm btn-ghost" style={{ borderRadius: 0, borderLeft: '1px solid var(--border)' }}><Icon name="grid" size={13}/></button>
        </div>
        <button className="btn btn-sm"><Icon name="download" size={13}/>Exportar</button>
      </div>

      {/* Table */}
      <div className="card" style={{ overflow: 'hidden', padding: 0, marginBottom: 32 }}>
        <div style={{ display: 'grid', gridTemplateColumns: '160px 1fr 110px 100px 70px 110px 130px 90px 36px', alignItems: 'center', gap: 12, padding: '10px 16px', borderBottom: '1px solid var(--border)', background: 'var(--surface-2)' }}>
          {['Código', 'Documento', 'Área', 'Perfil', 'Rev.', 'Estado', 'Autor', 'Atualizado', ''].map((h, i) => (
            <div key={i} className="kicker" style={{ fontSize: 10 }}>{h}</div>
          ))}
        </div>
        {DOCS.map((d, i) => (
          <div key={i} className="row" style={{ display: 'grid', gridTemplateColumns: '160px 1fr 110px 100px 70px 110px 130px 90px 36px', alignItems: 'center', gap: 12, padding: '10px 16px', borderBottom: i < DOCS.length - 1 ? '1px solid var(--border)' : 'none', cursor: 'pointer' }}>
            <span className="code-chip mono">{d.code}</span>
            <div style={{ minWidth: 0 }}>
              <div style={{ fontSize: 13.5, fontWeight: 500, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{d.title}</div>
              {d.hash !== '—' && <div className="tiny mono" style={{ color: 'var(--text-faint)', marginTop: 2 }}>hash {d.hash} · vig. {d.effective}</div>}
              {d.hash === '—' && <div className="tiny" style={{ color: 'var(--text-faint)', marginTop: 2 }}>{d.approvers.length ? `${d.approvers.length} aprovador(es) pendente(s)` : 'Sem rota de aprovação'}</div>}
            </div>
            <span className="body-sm" style={{ color: 'var(--text-soft)' }}>{d.area}</span>
            <span className="mono" style={{ fontSize: 11, color: 'var(--text-muted)' }}>{d.profile}</span>
            <span className="mono num" style={{ fontSize: 12, color: 'var(--text-soft)' }}>{d.rev}</span>
            <Status s={d.status}/>
            <div style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
              <Avatar name={d.author} size="sm"/>
              <span className="caption" style={{ color: 'var(--text-soft)', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{d.author.split(' ')[0]}</span>
            </div>
            <span className="caption">{d.updated}</span>
            <button className="btn btn-icon btn-sm btn-ghost"><Icon name="more" size={14}/></button>
          </div>
        ))}
      </div>
    </div>
  </div>
);

// === DIRECTION B: CARD GRID + ACTIVITY SIDEBAR ===
const LibraryB = () => (
  <div style={{ flex: 1, display: 'flex', overflow: 'hidden', background: 'var(--bg)' }}>
    <div style={{ flex: 1, overflow: 'auto', padding: '24px 28px' }}>
      <div style={{ marginBottom: 20 }}>
        <div className="kicker" style={{ marginBottom: 6 }}>Acervo · 2.847 documentos</div>
        <h1 className="display-1">Boa tarde, Marina.</h1>
        <p className="body" style={{ color: 'var(--text-muted)', marginTop: 6 }}>3 documentos aguardam sua aprovação. <a href="#" style={{ color: 'var(--brand)', fontWeight: 500 }}>Abrir caixa de entrada →</a></p>
      </div>

      {/* Pinned section */}
      <div className="kicker" style={{ marginBottom: 10 }}>★ Fixados</div>
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: 12, marginBottom: 28 }}>
        {DOCS.slice(0, 3).map((d, i) => (
          <div key={i} className="card" style={{ padding: 16, cursor: 'pointer' }}>
            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 10 }}>
              <span className="code-chip mono">{d.code}</span>
              <Status s={d.status}/>
            </div>
            <div className="h3" style={{ fontSize: 14, marginBottom: 6, lineHeight: 1.3 }}>{d.title}</div>
            <div className="caption" style={{ marginBottom: 12 }}>{d.area} · {d.profile} · <span className="mono">{d.rev}</span></div>
            <div className="divider" style={{ marginBottom: 10 }}/>
            <div style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
              <Avatar name={d.author} size="sm"/>
              <span className="caption">{d.author}</span>
              <span className="spacer"/>
              <span className="tiny mono">{d.updated}</span>
            </div>
          </div>
        ))}
      </div>

      <div className="kicker" style={{ marginBottom: 10 }}>Recentes</div>
      <div className="card" style={{ padding: 0 }}>
        {DOCS.slice(3).map((d, i, a) => (
          <div key={i} className="row" style={{ display: 'flex', alignItems: 'center', gap: 14, padding: '12px 16px', borderBottom: i < a.length - 1 ? '1px solid var(--border)' : 'none', cursor: 'pointer' }}>
            <span className="code-chip mono" style={{ width: 110, textAlign: 'center' }}>{d.code}</span>
            <div style={{ flex: 1, minWidth: 0 }}>
              <div style={{ fontSize: 13.5, fontWeight: 500 }}>{d.title}</div>
              <div className="caption">{d.area} · {d.author} · <span className="mono">{d.rev}</span></div>
            </div>
            <Status s={d.status}/>
            <span className="caption" style={{ width: 90, textAlign: 'right' }}>{d.updated}</span>
          </div>
        ))}
      </div>
    </div>

    {/* Activity sidebar */}
    <aside style={{ width: 320, borderLeft: '1px solid var(--border)', background: 'var(--surface-2)', overflow: 'auto', padding: '24px 20px' }}>
      <div className="h3" style={{ marginBottom: 14 }}>Atividade da semana</div>
      <div style={{ display: 'flex', alignItems: 'flex-end', gap: 4, height: 60, marginBottom: 6 }}>
        {[12, 18, 8, 22, 31, 15, 9].map((h, i) => (
          <div key={i} style={{ flex: 1, height: `${h * 1.8}px`, background: i === 4 ? 'var(--brand)' : 'var(--brand-soft)', opacity: i === 4 ? 1 : 0.45, borderRadius: '2px 2px 0 0' }}/>
        ))}
      </div>
      <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: 10, color: 'var(--text-muted)', marginBottom: 24, fontFamily: 'var(--font-mono)' }}>
        {['SEG', 'TER', 'QUA', 'QUI', 'SEX', 'SÁB', 'DOM'].map(d => <span key={d}>{d}</span>)}
      </div>

      <div className="kicker" style={{ marginBottom: 10 }}>Trilha de auditoria</div>
      <div style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>
        {[
          { who: 'Juliana Prado', what: 'aprovou', code: 'POP-FIN-019', t: '14:32' },
          { who: 'Rafael Castro', what: 'submeteu para revisão', code: 'POP-QUA-119', t: '13:58' },
          { who: 'Sistema', what: 'gerou PDF de', code: 'IT-PROD-204', t: '13:51' },
          { who: 'André Lima', what: 'editou', code: 'POL-TI-007', t: '11:24' },
          { who: 'Marina S.', what: 'arquivou', code: 'POP-RH-099', t: '10:17' },
        ].map((a, i) => (
          <div key={i} style={{ display: 'flex', gap: 10 }}>
            <Avatar name={a.who} size="sm"/>
            <div style={{ flex: 1, fontSize: 12, lineHeight: 1.45 }}>
              <span style={{ fontWeight: 500 }}>{a.who}</span> <span style={{ color: 'var(--text-muted)' }}>{a.what}</span> <span className="code-chip mono" style={{ fontSize: 10.5, padding: '1px 5px' }}>{a.code}</span>
              <div className="tiny mono" style={{ color: 'var(--text-faint)', marginTop: 2 }}>{a.t}</div>
            </div>
          </div>
        ))}
      </div>
    </aside>
  </div>
);

// === DIRECTION C: EDITORIAL — full-bleed, typography-led ===
const LibraryC = () => (
  <div style={{ flex: 1, overflow: 'auto', background: 'var(--surface)' }}>
    <div style={{ padding: '36px 48px', borderBottom: '1px solid var(--border)' }}>
      <div style={{ display: 'flex', alignItems: 'flex-end', gap: 24 }}>
        <div style={{ flex: 1 }}>
          <div className="kicker" style={{ marginBottom: 10 }}>§ 04 · Acervo Controlado</div>
          <h1 style={{ fontSize: 52, lineHeight: 1, letterSpacing: '-0.035em', fontWeight: 600, fontFamily: 'var(--font-display)' }}>
            Documentos<br/>
            <span style={{ color: 'var(--brand)', fontStyle: 'italic', fontWeight: 500 }}>sob controle.</span>
          </h1>
        </div>
        <div style={{ display: 'flex', gap: 32, paddingBottom: 8 }}>
          {[['2.847', 'frozen'], ['8', 'em revisão'], ['23', 'rascunho'], ['v3.4', 'release']].map(([v, l], i) => (
            <div key={i} style={{ borderLeft: '1px solid var(--border)', paddingLeft: 16 }}>
              <div className="num" style={{ fontSize: 26, fontWeight: 500, fontFamily: 'var(--font-display)', letterSpacing: '-0.02em' }}>{v}</div>
              <div className="kicker" style={{ fontSize: 10 }}>{l}</div>
            </div>
          ))}
        </div>
      </div>
    </div>

    <div style={{ padding: '12px 48px', borderBottom: '1px solid var(--border)', display: 'flex', alignItems: 'center', gap: 24, fontSize: 12, color: 'var(--text-muted)' }}>
      <button style={{ border: 'none', background: 'none', padding: '4px 0', borderBottom: '2px solid var(--brand)', color: 'var(--text)', fontWeight: 500, cursor: 'pointer', fontSize: 12 }}>Todos</button>
      <button style={{ border: 'none', background: 'none', padding: '4px 0', cursor: 'pointer', fontSize: 12, color: 'var(--text-muted)' }}>Frozen</button>
      <button style={{ border: 'none', background: 'none', padding: '4px 0', cursor: 'pointer', fontSize: 12, color: 'var(--text-muted)' }}>Em revisão</button>
      <button style={{ border: 'none', background: 'none', padding: '4px 0', cursor: 'pointer', fontSize: 12, color: 'var(--text-muted)' }}>Rascunhos</button>
      <span className="spacer"/>
      <button className="btn btn-sm btn-ghost"><Icon name="filter" size={12}/>Filtros</button>
      <span className="mono">A–Z</span>
    </div>

    {/* Editorial list — large, generous */}
    <div>
      {DOCS.map((d, i) => (
        <a key={i} href="#" className="row" style={{
          display: 'grid',
          gridTemplateColumns: '60px 140px 1fr 120px 120px 110px 36px',
          alignItems: 'center',
          gap: 20,
          padding: '20px 48px',
          borderBottom: '1px solid var(--border)',
          textDecoration: 'none',
          color: 'inherit',
          cursor: 'pointer',
        }}>
          <span className="num mono" style={{ fontSize: 13, color: 'var(--text-faint)' }}>{String(i + 1).padStart(2, '0')}</span>
          <span className="mono" style={{ fontSize: 13, fontWeight: 500, color: 'var(--brand)' }}>{d.code}</span>
          <div>
            <div style={{ fontSize: 17, fontWeight: 500, letterSpacing: '-0.01em', lineHeight: 1.25, marginBottom: 4 }}>{d.title}</div>
            <div className="caption" style={{ color: 'var(--text-muted)' }}>
              {d.area} · {d.profile} · rev. <span className="mono">{d.rev}</span> {d.hash !== '—' && <>· hash <span className="mono">{d.hash}</span></>}
            </div>
          </div>
          <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
            <Avatar name={d.author} size="sm"/>
            <span className="caption" style={{ color: 'var(--text-soft)' }}>{d.author.split(' ')[0]}</span>
          </div>
          <Status s={d.status}/>
          <span className="caption" style={{ textAlign: 'right' }}>{d.updated}</span>
          <Icon name="chevron" size={14}/>
        </a>
      ))}
    </div>
  </div>
);

window.MD_LIBRARY = { LibraryA, LibraryB, LibraryC, DOCS };
