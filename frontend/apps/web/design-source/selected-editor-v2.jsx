/* global React, MD */
const { Icon, Avatar } = MD;
const { useState } = React;

// === EDITOR v2 — minimal chrome, paper-focused, right metadata sidebar ===
const SelectedEditorV2 = () => {
  return (
    <div style={{ flex: 1, display: 'flex', overflow: 'hidden', background: 'var(--bg)' }}>
      {/* MAIN — paper only */}
      <div style={{ flex: 1, display: 'flex', flexDirection: 'column', minWidth: 0 }}>
        {/* Ultra slim doc bar */}
        <div style={{ background: 'var(--surface)', borderBottom: '1px solid var(--border)', padding: '10px 24px', display: 'flex', alignItems: 'center', gap: 12, flexShrink: 0, fontSize: 13 }}>
          <button className="btn btn-icon btn-sm btn-ghost" title="Voltar"><Icon name="chevron" size={14} style={{ transform: 'rotate(180deg)' }}/></button>
          <span className="code-chip mono" style={{ fontSize: 12, padding: '2px 8px' }}>POP-RH-001</span>
          <span style={{ color: 'var(--text-muted)' }}>Admissão e Onboarding de Colaboradores</span>
          <span className="caption mono" style={{ color: 'var(--text-faint)', marginLeft: 4 }}>v5 · rascunho</span>
          <span className="spacer"/>
          <span className="caption" style={{ color: 'var(--text-muted)', display: 'flex', alignItems: 'center', gap: 6 }}>
            <span style={{ width: 6, height: 6, borderRadius: '50%', background: 'var(--success)' }}/>
            Salvo · há 12s
          </span>
          <button className="btn btn-sm btn-ghost"><Icon name="history" size={13}/>Revisões</button>
          <button className="btn btn-sm">Salvar</button>
          <button className="btn btn-sm btn-primary">Submeter para revisão</button>
        </div>

        {/* Floating mini-toolbar (formatting) */}
        <div style={{ padding: '8px 24px', borderBottom: '1px solid var(--border)', background: 'var(--surface)', display: 'flex', alignItems: 'center', gap: 4, flexShrink: 0, fontSize: 12 }}>
          <select style={{ height: 26, fontSize: 12, padding: '0 6px', border: '1px solid var(--border)', background: 'var(--surface)', borderRadius: 4, color: 'var(--text-soft)' }}>
            <option>Parágrafo</option>
            <option>Título 1</option>
            <option>Título 2</option>
            <option>Título 3</option>
          </select>
          <span style={{ width: 1, height: 18, background: 'var(--border)', margin: '0 4px' }}/>
          {['B', 'I', 'U'].map(c => (
            <button key={c} style={{ width: 26, height: 26, border: 'none', background: 'transparent', cursor: 'pointer', borderRadius: 4, fontFamily: 'serif', fontWeight: c === 'B' ? 700 : 400, fontStyle: c === 'I' ? 'italic' : 'normal', textDecoration: c === 'U' ? 'underline' : 'none', fontSize: 13, color: 'var(--text-soft)' }}>{c}</button>
          ))}
          <span style={{ width: 1, height: 18, background: 'var(--border)', margin: '0 4px' }}/>
          {[
            <span key="ul" style={{ fontSize: 13 }}>≡</span>,
            <span key="ol" style={{ fontSize: 11, fontFamily: 'var(--font-mono)' }}>1.</span>,
            <span key="link" key2="link"><Icon name="link" size={13}/></span>,
          ].map((c, i) => (
            <button key={i} style={{ width: 26, height: 26, border: 'none', background: 'transparent', cursor: 'pointer', borderRadius: 4, color: 'var(--text-soft)', display: 'grid', placeItems: 'center' }}>{c}</button>
          ))}
          <span className="spacer"/>
          <span className="caption" style={{ color: 'var(--text-faint)' }}>1.247 palavras · 7 seções</span>
        </div>

        {/* Paper canvas — generous whitespace */}
        <div style={{ flex: 1, overflow: 'auto', padding: '36px 28px', display: 'flex', justifyContent: 'center' }}>
          <div style={{ width: 760, background: 'white', boxShadow: '0 4px 32px rgba(26, 14, 14, 0.10)', padding: '64px 76px', minHeight: 700, fontFamily: 'Georgia, serif' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', borderBottom: '2px solid #1a0e0e', paddingBottom: 14, marginBottom: 32, fontSize: 11, fontFamily: 'var(--font-mono)', color: '#1a0e0e' }}>
              <span>POP-RH-001 · v5 (rascunho)</span>
              <span>EM EDIÇÃO</span>
              <span>METALDOCS</span>
            </div>
            <h1 style={{ fontSize: 24, fontWeight: 700, marginBottom: 22, fontFamily: 'Georgia, serif', outline: 'none' }} contentEditable suppressContentEditableWarning>
              Admissão e Onboarding de Colaboradores
            </h1>
            <h3 style={{ fontSize: 14, fontWeight: 700, marginTop: 24, marginBottom: 8 }}>1. Objetivo</h3>
            <p style={{ fontSize: 13, lineHeight: 1.7, color: '#222', marginBottom: 14, outline: 'none' }} contentEditable suppressContentEditableWarning>
              Estabelecer o procedimento padrão para admissão de novos colaboradores na organização, garantindo a conformidade com os requisitos legais (CLT) e os padrões internos de qualidade definidos pelo Sistema de Gestão Integrado.
            </p>
            <h3 style={{ fontSize: 14, fontWeight: 700, marginTop: 24, marginBottom: 8 }}>2. Escopo</h3>
            <p style={{ fontSize: 13, lineHeight: 1.7, color: '#222', marginBottom: 14 }} contentEditable suppressContentEditableWarning>
              Aplica-se a todos os processos de contratação realizados pela área de Recursos Humanos, incluindo cargos efetivos, temporários e estagiários.
            </p>
            <h3 style={{ fontSize: 14, fontWeight: 700, marginTop: 24, marginBottom: 8 }}>3. Responsabilidades</h3>
            <p style={{ fontSize: 13, lineHeight: 1.7, color: '#222', marginBottom: 14 }} contentEditable suppressContentEditableWarning>
              O Coordenador de RH é responsável pela execução. O gestor da área requisitante valida competências técnicas e aprova a contratação final.
            </p>
            <h3 style={{ fontSize: 14, fontWeight: 700, marginTop: 24, marginBottom: 8 }}>4. Procedimento</h3>
            <ol style={{ fontSize: 13, lineHeight: 1.7, color: '#222', paddingLeft: 24, marginBottom: 14 }}>
              <li contentEditable suppressContentEditableWarning>Recebimento da requisição via sistema</li>
              <li contentEditable suppressContentEditableWarning>Triagem curricular e shortlisting</li>
              <li contentEditable suppressContentEditableWarning>Entrevistas técnicas e comportamentais</li>
              <li contentEditable suppressContentEditableWarning>Verificação de referências</li>
              <li contentEditable suppressContentEditableWarning>Proposta formal e negociação</li>
              <li contentEditable suppressContentEditableWarning>Coleta documental e exames admissionais</li>
              <li contentEditable suppressContentEditableWarning>Integração e treinamento inicial</li>
            </ol>
          </div>
        </div>
      </div>

      {/* RIGHT METADATA SIDEBAR — kept */}
      <aside style={{ width: 300, borderLeft: '1px solid var(--border)', background: 'var(--surface-2)', overflow: 'auto', flexShrink: 0, padding: '20px 18px' }}>
        <div className="kicker" style={{ marginBottom: 10 }}>Metadados</div>
        <div style={{ display: 'flex', flexDirection: 'column', gap: 9, marginBottom: 22, fontSize: 12.5 }}>
          {[
            ['Código', <span className="mono">POP-RH-001</span>],
            ['Perfil', 'Procedimento Op.'],
            ['Área', 'RH'],
            ['Vigência atual', <span className="mono">v4 · 15/04/2026</span>],
            ['Próx. revisão', <span className="mono">15/04/2027</span>],
            ['Visibilidade', 'Toda empresa'],
          ].map(([k, v], i) => (
            <div key={i} style={{ display: 'flex', justifyContent: 'space-between', gap: 12 }}>
              <span style={{ color: 'var(--text-muted)' }}>{k}</span>
              <span style={{ color: 'var(--text)', textAlign: 'right' }}>{v}</span>
            </div>
          ))}
        </div>

        <div className="divider" style={{ margin: '0 -18px 16px' }}/>

        <div className="kicker" style={{ marginBottom: 10 }}>Revisões</div>
        <div style={{ position: 'relative', paddingLeft: 16, marginBottom: 22 }}>
          <span style={{ position: 'absolute', left: 5, top: 8, bottom: 8, width: 1, background: 'var(--border-strong)' }}/>
          {[
            { v: 'v5', d: 'agora', n: 'Você · em edição', cur: true, draft: true },
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

        <div className="divider" style={{ margin: '0 -18px 16px' }}/>

        <div className="kicker" style={{ marginBottom: 10 }}>Próximos aprovadores</div>
        <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
          {[
            { n: 'Rafael Castro', r: 'Líder Qualidade', s: 'next' },
            { n: 'Juliana Prado', r: 'Gerente Geral', s: 'wait' },
          ].map((a, i) => (
            <div key={i} style={{ display: 'flex', alignItems: 'center', gap: 8, fontSize: 12 }}>
              <Avatar name={a.n} size="sm"/>
              <div style={{ flex: 1, minWidth: 0 }}>
                <div style={{ fontWeight: 500 }}>{a.n}</div>
                <div className="tiny" style={{ color: 'var(--text-muted)' }}>{a.r}</div>
              </div>
              <span className="pill" style={{ fontSize: 10, background: a.s === 'next' ? 'var(--brand-pale)' : 'transparent', color: a.s === 'next' ? 'var(--brand)' : 'var(--text-faint)' }}>
                {a.s === 'next' ? 'próximo' : 'aguarda'}
              </span>
            </div>
          ))}
        </div>
      </aside>
    </div>
  );
};

window.MD_EDITOR_V2 = { SelectedEditorV2 };
