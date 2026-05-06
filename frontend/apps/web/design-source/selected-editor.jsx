/* global React, MD */
const { Icon, Status, Avatar } = MD;
const { useState } = React;

// === [C5] EDITOR — focused on document, right metadata sidebar, less chrome ===
const SelectedEditor = () => {
  const [tab, setTab] = useState('content');
  return (
    <div style={{ flex: 1, display: 'flex', overflow: 'hidden', background: 'var(--bg)' }}>
      {/* MAIN — paper-focused */}
      <div style={{ flex: 1, display: 'flex', flexDirection: 'column', minWidth: 0 }}>
        {/* Slim doc bar */}
        <div style={{ background: 'var(--surface)', borderBottom: '1px solid var(--border)', padding: '12px 28px', display: 'flex', alignItems: 'center', gap: 14, flexShrink: 0 }}>
          <span className="code-chip mono" style={{ fontSize: 12, padding: '3px 9px' }}>POP-RH-001</span>
          <div style={{ minWidth: 0 }}>
            <div style={{ fontSize: 14, fontWeight: 500, lineHeight: 1.2, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>Admissão e Onboarding de Colaboradores</div>
            <div className="caption" style={{ color: 'var(--text-muted)' }}>v4 · Vigente · editando rascunho v5</div>
          </div>
          <span className="spacer"/>
          <span className="caption" style={{ color: 'var(--text-muted)', display: 'flex', alignItems: 'center', gap: 6 }}>
            <span style={{ width: 6, height: 6, borderRadius: '50%', background: 'var(--success)' }}/>
            Salvo · há 12s
          </span>
          <button className="btn btn-sm btn-ghost"><Icon name="history" size={13}/></button>
          <button className="btn btn-sm">Salvar rascunho</button>
          <button className="btn btn-sm btn-primary">Submeter para revisão</button>
        </div>

        {/* Tabs */}
        <div style={{ display: 'flex', gap: 0, padding: '0 28px', background: 'var(--surface)', borderBottom: '1px solid var(--border)' }}>
          {[
            { id: 'content', l: 'Conteúdo' },
            { id: 'history', l: 'Revisões' },
            { id: 'approvals', l: 'Aprovações' },
            { id: 'audit', l: 'Auditoria' },
          ].map(t => (
            <button key={t.id} onClick={() => setTab(t.id)} style={{
              padding: '10px 14px', fontSize: 13, border: 'none', background: 'none',
              borderBottom: tab === t.id ? '2px solid var(--brand)' : '2px solid transparent',
              color: tab === t.id ? 'var(--text)' : 'var(--text-muted)',
              fontWeight: tab === t.id ? 500 : 400, cursor: 'pointer', marginBottom: -1,
            }}>{t.l}</button>
          ))}
        </div>

        {/* Paper canvas — generous whitespace, focus on document */}
        <div style={{ flex: 1, overflow: 'auto', padding: '40px 28px', display: 'flex', justifyContent: 'center' }}>
          <div style={{ width: 760, background: 'white', boxShadow: '0 4px 24px rgba(0,0,0,0.08)', padding: '64px 72px', minHeight: 700, fontFamily: 'Georgia, serif' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', borderBottom: '2px solid #1a0e0e', paddingBottom: 14, marginBottom: 32, fontSize: 11, fontFamily: 'var(--font-mono)', color: '#1a0e0e' }}>
              <span>POP-RH-001 · v5 (rascunho)</span>
              <span>EM EDIÇÃO</span>
              <span>METALDOCS</span>
            </div>
            <h1 style={{ fontSize: 24, fontWeight: 700, marginBottom: 22, fontFamily: 'Georgia, serif', borderBottom: '1px dashed #ccc', paddingBottom: 8, outline: 'none' }} contentEditable suppressContentEditableWarning>
              Admissão e Onboarding de Colaboradores
            </h1>
            <h3 style={{ fontSize: 14, fontWeight: 700, marginTop: 24, marginBottom: 8, fontFamily: 'Georgia, serif' }}>1. Objetivo</h3>
            <p style={{ fontSize: 13, lineHeight: 1.7, color: '#222', marginBottom: 14, outline: 'none' }} contentEditable suppressContentEditableWarning>
              Estabelecer o procedimento padrão para admissão de novos colaboradores na organização, garantindo a conformidade com os requisitos legais (CLT) e os padrões internos de qualidade definidos pelo Sistema de Gestão Integrado.
            </p>
            <h3 style={{ fontSize: 14, fontWeight: 700, marginTop: 24, marginBottom: 8, fontFamily: 'Georgia, serif' }}>2. Escopo</h3>
            <p style={{ fontSize: 13, lineHeight: 1.7, color: '#222', marginBottom: 14, outline: 'none' }} contentEditable suppressContentEditableWarning>
              Aplica-se a todos os processos de contratação realizados pela área de Recursos Humanos, incluindo cargos efetivos, temporários e estagiários, em todas as unidades operacionais.
            </p>
            <h3 style={{ fontSize: 14, fontWeight: 700, marginTop: 24, marginBottom: 8, fontFamily: 'Georgia, serif' }}>3. Responsabilidades</h3>
            <p style={{ fontSize: 13, lineHeight: 1.7, color: '#222', marginBottom: 14, outline: 'none' }} contentEditable suppressContentEditableWarning>
              O Coordenador de RH é responsável pela execução, condução das entrevistas técnicas e validação documental. O gestor da área requisitante valida competências técnicas e aprova a contratação final.
            </p>
            <h3 style={{ fontSize: 14, fontWeight: 700, marginTop: 24, marginBottom: 8, fontFamily: 'Georgia, serif' }}>4. Procedimento</h3>
            <ol style={{ fontSize: 13, lineHeight: 1.7, color: '#222', paddingLeft: 24, marginBottom: 14 }}>
              <li>Recebimento da requisição via sistema</li>
              <li>Triagem curricular e shortlisting</li>
              <li>Entrevistas técnicas e comportamentais</li>
              <li>Verificação de referências</li>
              <li>Proposta formal e negociação</li>
              <li>Coleta documental e exames admissionais</li>
              <li>Integração e treinamento inicial</li>
            </ol>
          </div>
        </div>
      </div>

      {/* RIGHT METADATA SIDEBAR */}
      <aside style={{ width: 320, borderLeft: '1px solid var(--border)', background: 'var(--surface-2)', overflow: 'auto', flexShrink: 0, padding: '20px 20px' }}>
        <div className="kicker" style={{ marginBottom: 10 }}>Metadados</div>
        <div style={{ display: 'flex', flexDirection: 'column', gap: 9, marginBottom: 22, fontSize: 12.5 }}>
          {[
            ['Código', <span className="mono">POP-RH-001</span>],
            ['Perfil', 'Procedimento Op.'],
            ['Área', 'Recursos Humanos'],
            ['Revisão atual', <span className="mono">v4 vigente</span>],
            ['Editando', <span className="mono">v5 rascunho</span>],
            ['Vigência', <span className="mono">15/04/2026</span>],
            ['Próx. revisão', <span className="mono">15/04/2027</span>],
          ].map(([k, v], i) => (
            <div key={i} style={{ display: 'flex', justifyContent: 'space-between', gap: 12 }}>
              <span style={{ color: 'var(--text-muted)' }}>{k}</span>
              <span style={{ color: 'var(--text)', textAlign: 'right' }}>{v}</span>
            </div>
          ))}
        </div>

        <div className="divider" style={{ margin: '0 -20px 16px' }}/>

        <div className="kicker" style={{ marginBottom: 10 }}>Revisões anteriores</div>
        <div style={{ position: 'relative', paddingLeft: 16, marginBottom: 22 }}>
          <span style={{ position: 'absolute', left: 5, top: 8, bottom: 8, width: 1, background: 'var(--border-strong)' }}/>
          {[
            { v: 'v5', d: 'em edição', n: 'Você · agora', cur: true, draft: true },
            { v: 'v4', d: '15/04/2026', n: 'Anexos LGPD' },
            { v: 'v3', d: '03/11/2025', n: 'Integração SAP' },
            { v: 'v2', d: '22/05/2024', n: 'Estágio' },
            { v: 'v1', d: '08/01/2024', n: 'Versão inicial' },
          ].map((r, i) => (
            <div key={i} style={{ position: 'relative', paddingBottom: 12 }}>
              <span style={{ position: 'absolute', left: -16, top: 4, width: 11, height: 11, borderRadius: '50%', background: r.draft ? 'var(--surface-2)' : r.cur ? 'var(--brand)' : 'var(--surface)', border: `2px ${r.draft ? 'dashed' : 'solid'} ${r.cur ? 'var(--brand)' : 'var(--border-strong)'}` }}/>
              <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: 12.5 }}>
                <span style={{ fontWeight: r.cur ? 500 : 400, fontFamily: 'var(--font-mono)' }}>{r.v}</span>
                <span className="tiny mono">{r.d}</span>
              </div>
              <div className="caption" style={{ color: 'var(--text-muted)' }}>{r.n}</div>
            </div>
          ))}
        </div>

        <div className="divider" style={{ margin: '0 -20px 16px' }}/>

        <div className="kicker" style={{ marginBottom: 10 }}>Rota de aprovação</div>
        <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
          {[
            { n: 'Marina Silveira', r: 'Autora', s: 'self' },
            { n: 'Rafael Castro', r: 'Líder Qualidade', s: 'next' },
            { n: 'Juliana Prado', r: 'Gerente Geral', s: 'wait' },
          ].map((a, i) => (
            <div key={i} style={{ display: 'flex', alignItems: 'center', gap: 8, fontSize: 12 }}>
              <Avatar name={a.n} size="sm"/>
              <div style={{ flex: 1, minWidth: 0 }}>
                <div style={{ fontWeight: 500 }}>{a.n}</div>
                <div className="tiny" style={{ color: 'var(--text-muted)' }}>{a.r}</div>
              </div>
              <span className="pill" style={{ fontSize: 10, background: a.s === 'self' ? 'var(--surface-3)' : a.s === 'next' ? 'var(--brand-pale)' : 'transparent', color: a.s === 'self' ? 'var(--text-muted)' : a.s === 'next' ? 'var(--brand)' : 'var(--text-faint)' }}>
                {a.s === 'self' ? 'você' : a.s === 'next' ? 'próximo' : 'aguardando'}
              </span>
            </div>
          ))}
        </div>
      </aside>
    </div>
  );
};

window.MD_EDITOR = { SelectedEditor };
