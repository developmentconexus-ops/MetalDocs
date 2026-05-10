/* global React, MD */
const { Icon, Status } = MD;
const { useState } = React;

// ============================================================
// TEMPLATE WIZARD — mesma DNA do doc wizard (selected-wizard.jsx)
// 5 etapas: Perfil · Identidade · Estrutura · Permissões · Confirmação
// ============================================================

const TPL_STEPS = ['Perfil', 'Identidade', 'Estrutura', 'Permissões', 'Confirmação'];

const TplStepper = ({ active, steps = TPL_STEPS }) => (
  <div style={{ display: 'flex', alignItems: 'center', gap: 0, marginBottom: 24, padding: '12px 18px', background: 'var(--surface)', border: '1px solid var(--border)', borderRadius: 'var(--r-3)' }}>
    {steps.map((s, i) => (
      <React.Fragment key={s}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
          <span style={{
            width: 22, height: 22, borderRadius: '50%',
            display: 'grid', placeItems: 'center',
            fontSize: 11, fontWeight: 600, fontFamily: 'var(--font-mono)',
            background: i < active ? 'var(--brand)' : i === active ? 'var(--surface)' : 'var(--surface-3)',
            color: i < active ? 'white' : i === active ? 'var(--brand)' : 'var(--text-muted)',
            border: i === active ? '1.5px solid var(--brand)' : i < active ? 'none' : '1px solid var(--border-strong)',
          }}>{i < active ? '✓' : i + 1}</span>
          <span style={{ fontSize: 13, fontWeight: i === active ? 600 : 400, color: i <= active ? 'var(--text)' : 'var(--text-muted)' }}>{s}</span>
        </div>
        {i < steps.length - 1 && <span style={{ flex: 1, height: 1, background: 'var(--border)', margin: '0 14px' }}/>}
      </React.Fragment>
    ))}
  </div>
);

const TplShell = ({ active, children, steps = TPL_STEPS }) => (
  <div style={{ flex: 1, overflow: 'auto', background: 'var(--bg)', padding: '32px 28px' }}>
    <div style={{ maxWidth: 880, margin: '0 auto' }}>
      <div className="kicker" style={{ marginBottom: 6 }}>Templates / Novo</div>
      <h1 className="display-1" style={{ marginBottom: 4 }}>Novo template reutilizável</h1>
      <p className="body" style={{ color: 'var(--text-muted)', marginBottom: 28 }}>
        Templates publicados ficam disponíveis para autores criarem novos documentos. Use placeholders <span className="mono">{`{{campo}}`}</span> para campos dinâmicos.
      </p>
      <TplStepper active={active} steps={steps}/>
      {children}
    </div>
  </div>
);

const TplFooter = ({ stepLabel, primary = 'Avançar →', back = true, primaryDisabled = false }) => (
  <>
    <div className="divider" style={{ margin: '0 -32px 20px' }}/>
    <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
      {back ? <button className="btn">← Voltar</button> : <span/>}
      <span className="spacer"/>
      <span className="caption">{stepLabel}</span>
      <button className="btn btn-primary" disabled={primaryDisabled}>{primary}</button>
    </div>
  </>
);

// ─── STEP 1 — Escopo (Perfil ou Documento específico) ───
const TplStep1 = ({ __initialScope = 'profile' } = {}) => {
  const [scope, setScope] = useState(__initialScope); // 'profile' | 'document'
  const [selected, setSelected] = useState('POP');
  const [docSel, setDocSel] = useState('POP-QUA-087');
  const documents = [
    { code: 'POP-QUA-087', title: 'Inspeção de Recebimento de Matéria-Prima', area: 'Qualidade', v: 'v3.2' },
    { code: 'POP-QUA-064', title: 'Calibração de Instrumentos de Medição', area: 'Qualidade', v: 'v2.1' },
    { code: 'IT-PROD-201', title: 'Setup de Linha de Produção — Estação 4', area: 'Produção', v: 'v1.4' },
    { code: 'POL-RH-008', title: 'Política de Trabalho Remoto', area: 'RH', v: 'v2.0' },
  ];
  const profiles = [
    { id: 'POP', name: 'Procedimento Operacional', desc: 'Sequência detalhada de execução de uma rotina operacional.', family: 'Qualidade', count: 12 },
    { id: 'IT', name: 'Instrução de Trabalho', desc: 'Passo-a-passo enxuto para tarefa específica em estação.', family: 'Qualidade', count: 18 },
    { id: 'POL', name: 'Política', desc: 'Diretrizes corporativas de alto nível com escopo organizacional.', family: 'Governança', count: 4 },
    { id: 'DC', name: 'Descrição de Cargo', desc: 'Identificação, responsabilidades e requisitos de uma posição.', family: 'RH', count: 9 },
    { id: 'CHK', name: 'Checklist Operacional', desc: 'Lista de verificação binária com observações por item.', family: 'Qualidade', count: 3, soon: true },
    { id: 'FOR', name: 'Formulário', desc: 'Captura estruturada de dados em campo ou em workflow.', family: 'Geral', count: 6 },
  ];
  return (
    <TplShell active={0}>
      <div className="card" style={{ padding: '28px 32px' }}>
        <div className="kicker" style={{ marginBottom: 8 }}>Etapa 1 de 5</div>
        <h2 className="h2" style={{ fontSize: 20, marginBottom: 6 }}>Escopo do template</h2>
        <p className="caption" style={{ marginBottom: 18 }}>Templates podem ser genéricos para um perfil (POP, IT, etc.) ou derivar de um documento específico já publicado.</p>

        {/* Tabs inside the card */}
        <div role="tablist" style={{ display: 'inline-flex', padding: 3, background: 'var(--surface-2)', border: '1px solid var(--border)', borderRadius: 'var(--r-2)', marginBottom: 22 }}>
          {[
            { id: 'profile', label: 'Para um perfil', sub: 'genérico (POP, IT, POL…)' },
            { id: 'document', label: 'A partir de um documento', sub: 'clona conteúdo de um doc publicado' },
          ].map(t => (
            <button key={t.id} role="tab" aria-selected={scope === t.id} onClick={() => setScope(t.id)} style={{
              padding: '8px 14px', border: 'none', cursor: 'pointer', borderRadius: 'var(--r-2)',
              background: scope === t.id ? 'var(--surface)' : 'transparent',
              boxShadow: scope === t.id ? '0 1px 2px rgba(0,0,0,0.06)' : 'none',
              color: scope === t.id ? 'var(--text)' : 'var(--text-muted)',
              fontSize: 12.5, fontWeight: scope === t.id ? 600 : 500,
              display: 'flex', flexDirection: 'column', alignItems: 'flex-start', gap: 1,
              transition: 'all 0.15s',
            }}>
              <span>{t.label}</span>
              <span className="tiny" style={{ color: 'var(--text-muted)', fontWeight: 400 }}>{t.sub}</span>
            </button>
          ))}
        </div>

        {scope === 'profile' ? (
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(2, 1fr)', gap: 10, marginBottom: 8 }}>
          {profiles.map(p => (
            <button key={p.id} onClick={() => !p.soon && setSelected(p.id)} disabled={p.soon} style={{
              textAlign: 'left', padding: 16,
              border: selected === p.id ? '2px solid var(--brand)' : '1px solid var(--border)',
              background: selected === p.id ? 'var(--brand-pale)' : 'var(--surface)',
              borderRadius: 'var(--r-3)',
              cursor: p.soon ? 'not-allowed' : 'pointer',
              opacity: p.soon ? 0.55 : 1,
              position: 'relative',
            }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 6 }}>
                <span className="mono" style={{ fontSize: 13, fontWeight: 600, color: selected === p.id ? 'var(--brand)' : 'var(--text)' }}>{p.id}</span>
                <span style={{ fontSize: 13, fontWeight: 500 }}>{p.name}</span>
                {p.soon && <span className="pill" style={{ fontSize: 9, marginLeft: 'auto' }}>em breve</span>}
                {selected === p.id && <span style={{ marginLeft: 'auto', color: 'var(--brand)' }}><Icon name="check" size={14}/></span>}
              </div>
              <p className="caption" style={{ marginBottom: 8, lineHeight: 1.4 }}>{p.desc}</p>
              <div style={{ display: 'flex', gap: 12, fontSize: 11, color: 'var(--text-muted)' }}>
                <span>Família: <span style={{ color: 'var(--text-soft)' }}>{p.family}</span></span>
                <span>·</span>
                <span><span className="mono" style={{ color: 'var(--text-soft)' }}>{p.count}</span> templates publicados</span>
              </div>
            </button>
          ))}
        </div>
        ) : (
          <div>
            <div style={{ position: 'relative', marginBottom: 12 }}>
              <input className="input" placeholder="Buscar por código ou título…" defaultValue="" style={{ height: 38, fontSize: 13, paddingLeft: 36 }}/>
              <span style={{ position: 'absolute', left: 12, top: 11, color: 'var(--text-muted)' }}><Icon name="search" size={16}/></span>
            </div>
            <div style={{ display: 'flex', flexDirection: 'column', gap: 8, marginBottom: 8 }}>
              {documents.map(d => (
                <button key={d.code} onClick={() => setDocSel(d.code)} style={{
                  textAlign: 'left', padding: '12px 14px',
                  border: docSel === d.code ? '2px solid var(--brand)' : '1px solid var(--border)',
                  background: docSel === d.code ? 'var(--brand-pale)' : 'var(--surface)',
                  borderRadius: 'var(--r-3)', cursor: 'pointer',
                  display: 'flex', alignItems: 'center', gap: 12,
                }}>
                  <span className="code-chip mono" style={{ fontSize: 11.5 }}>{d.code}</span>
                  <div style={{ flex: 1, minWidth: 0 }}>
                    <div style={{ fontSize: 13, fontWeight: 500, marginBottom: 2 }}>{d.title}</div>
                    <div className="tiny" style={{ color: 'var(--text-muted)' }}>{d.area} · publicado · <span className="mono">{d.v}</span></div>
                  </div>
                  {docSel === d.code && <span style={{ color: 'var(--brand)' }}><Icon name="check" size={16}/></span>}
                </button>
              ))}
            </div>
            <div style={{ padding: 12, background: 'var(--info-bg)', border: '1px dashed var(--info)', borderRadius: 'var(--r-2)', fontSize: 11.5, color: 'var(--text-soft)', lineHeight: 1.5 }}>
              <strong style={{ color: 'var(--info)' }}>ℹ️</strong> O conteúdo do documento será clonado como ponto de partida. Você pode adicionar placeholders <span className="mono">{`{{campo}}`}</span> no editor para tornar o template reutilizável.
            </div>
          </div>
        )}
        <TplFooter stepLabel={scope === 'profile' ? 'Etapa 1 de 5 · Selecione um perfil' : 'Etapa 1 de 5 · Selecione um documento'}/>
      </div>
    </TplShell>
  );
};

// ─── STEP 2 — Identidade ───
const TplStep2 = () => (
  <TplShell active={1}>
    <div className="card" style={{ padding: '28px 32px' }}>
      <div className="kicker" style={{ marginBottom: 8 }}>Etapa 2 de 5 · Perfil POP — Procedimento Operacional</div>
      <h2 className="h2" style={{ fontSize: 20, marginBottom: 20 }}>Identidade do template</h2>

      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 16, marginBottom: 18 }}>
        <div>
          <label className="kicker" style={{ display: 'block', marginBottom: 6 }}>Perfil selecionado</label>
          <div style={{ padding: '10px 12px', background: 'var(--surface-2)', border: '1px solid var(--border)', borderRadius: 'var(--r-2)', display: 'flex', alignItems: 'center', gap: 10 }}>
            <span className="code-chip mono">POP</span>
            <div style={{ flex: 1, fontSize: 13 }}>
              <div style={{ fontWeight: 500 }}>Procedimento Operacional</div>
              <div className="caption">Família: Qualidade</div>
            </div>
            <button className="btn btn-sm btn-ghost">Trocar</button>
          </div>
        </div>
        <div>
          <label className="kicker" style={{ display: 'block', marginBottom: 6 }}>Versão inicial</label>
          <div style={{ padding: '10px 12px', background: 'var(--surface-2)', border: '1px solid var(--border)', borderRadius: 'var(--r-2)', display: 'flex', alignItems: 'center', gap: 10 }}>
            <span className="mono" style={{ fontSize: 14, fontWeight: 600, color: 'var(--brand)' }}>v1.0</span>
            <span className="caption" style={{ flex: 1 }}>Atribuída automaticamente · incrementa em cada publicação</span>
          </div>
        </div>
      </div>

      <div style={{ marginBottom: 18 }}>
        <label className="kicker" style={{ display: 'block', marginBottom: 6 }}>Nome do template *</label>
        <input className="input" defaultValue="Inspeção de Recebimento de Matéria-Prima" style={{ height: 38, fontSize: 14 }}/>
        <div className="caption" style={{ marginTop: 4 }}>Aparece para os autores que forem cloná-lo no wizard de novo documento.</div>
      </div>

      <div style={{ marginBottom: 18 }}>
        <label className="kicker" style={{ display: 'block', marginBottom: 6 }}>Descrição</label>
        <textarea className="input" rows={3} defaultValue="Template para inspeção de matéria-prima recebida, com checklist de conformidade conforme NBR 6649." style={{ fontSize: 13, padding: 10, resize: 'vertical', fontFamily: 'inherit' }}/>
      </div>

      {/* Code preview — same DNA as doc wizard step 2 */}
      <div style={{ padding: 16, background: 'var(--brand-pale)', border: '1px dashed var(--brand-soft)', borderRadius: 'var(--r-3)', marginBottom: 18 }}>
        <div className="kicker" style={{ marginBottom: 6, color: 'var(--brand)' }}>Código sugerido · próximo template POP</div>
        <div className="mono" style={{ fontSize: 26, fontWeight: 600, color: 'var(--brand)', letterSpacing: '0.02em' }}>TPL-POP-009</div>
        <div className="caption" style={{ marginTop: 4 }}>Sequência ativa: 008 / Próxima: 009 · Códigos não são reutilizados.</div>
      </div>

      <div style={{ marginBottom: 24 }}>
        <label className="kicker" style={{ display: 'block', marginBottom: 6 }}>Tags</label>
        <div style={{ display: 'flex', flexWrap: 'wrap', gap: 6, padding: '8px 10px', border: '1px solid var(--border)', borderRadius: 'var(--r-2)', background: 'var(--surface)', minHeight: 40 }}>
          {['inspeção', 'recebimento', 'matéria-prima'].map((t, i) => (
            <span key={i} style={{ display: 'inline-flex', alignItems: 'center', gap: 4, padding: '3px 9px', background: 'var(--brand-pale)', borderRadius: 'var(--r-pill)', fontSize: 11.5, color: 'var(--brand-d)', fontWeight: 500 }}>
              {t}<button style={{ background: 'none', border: 'none', cursor: 'pointer', color: 'var(--brand)', padding: 0, fontSize: 13, lineHeight: 1 }}>×</button>
            </span>
          ))}
          <input type="text" placeholder="adicionar tag…" style={{ border: 'none', outline: 'none', flex: 1, minWidth: 100, fontSize: 12.5, background: 'transparent' }}/>
        </div>
      </div>

      <TplFooter stepLabel="Etapa 2 de 5 · Avance para subir a estrutura"/>
    </div>
  </TplShell>
);

// ─── STEP 3 — Estrutura ───
const TplStep3 = () => {
  const [phase] = useState('uploaded'); // 'empty' | 'uploaded'
  const placeholders = [
    { token: '{{cliente.nome}}', kind: 'texto', auto: false },
    { token: '{{material.codigo}}', kind: 'texto', auto: false },
    { token: '{{recebimento.data}}', kind: 'data', auto: true },
    { token: '{{lote.numero}}', kind: 'texto', auto: false },
    { token: '{{inspetor.nome}}', kind: 'auto', auto: true },
    { token: '{{conformidade.dimensional}}', kind: 'seleção', auto: false },
    { token: '{{observacoes}}', kind: 'texto longo', auto: false },
  ];
  return (
    <TplShell active={2}>
      <div className="card" style={{ padding: '28px 32px' }}>
        <div className="kicker" style={{ marginBottom: 8 }}>Etapa 3 de 5</div>
        <h2 className="h2" style={{ fontSize: 20, marginBottom: 6 }}>Estrutura do template</h2>
        <p className="caption" style={{ marginBottom: 24 }}>Suba um <span className="mono">.docx</span> base ou comece em branco. Placeholders no formato <span className="mono">{`{{campo}}`}</span> serão detectados automaticamente.</p>

        {/* Two starting points */}
        <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 10, marginBottom: 22 }}>
          <button style={{
            textAlign: 'left', padding: 16,
            border: '2px solid var(--brand)',
            background: 'var(--brand-pale)',
            borderRadius: 'var(--r-3)', cursor: 'pointer',
            display: 'flex', alignItems: 'center', gap: 14,
          }}>
            <div style={{ width: 60, height: 76, background: 'white', boxShadow: '0 2px 6px rgba(0,0,0,0.06)', borderRadius: 2, padding: '6px 5px', flexShrink: 0, position: 'relative' }}>
              <div style={{ height: 3, background: 'var(--brand)', marginBottom: 4, width: '60%' }}/>
              {[...Array(7)].map((_, k) => {
                const ph = k % 3 === 1;
                return <div key={k} style={{ height: 1.5, background: ph ? 'var(--brand)' : 'var(--border-strong)', marginBottom: 2.5, width: `${55 + (k * 9) % 35}%`, opacity: ph ? 0.7 : 1 }}/>;
              })}
            </div>
            <div style={{ flex: 1 }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 4 }}>
                <span style={{ fontSize: 13.5, fontWeight: 500 }}>A partir de um .docx</span>
                <span style={{ marginLeft: 'auto', color: 'var(--brand)' }}><Icon name="check" size={14}/></span>
              </div>
              <p className="caption" style={{ lineHeight: 1.4 }}>Importa formatação e detecta placeholders. Recomendado quando você já tem um documento de referência.</p>
            </div>
          </button>
          <button style={{
            textAlign: 'left', padding: 16,
            border: '1px solid var(--border)',
            background: 'var(--surface)',
            borderRadius: 'var(--r-3)', cursor: 'pointer',
            display: 'flex', alignItems: 'center', gap: 14,
          }}>
            <div style={{ width: 60, height: 76, background: 'white', boxShadow: '0 2px 6px rgba(0,0,0,0.06)', borderRadius: 2, padding: '6px 5px', flexShrink: 0, display: 'grid', placeItems: 'center' }}>
              <span style={{ color: 'var(--text-muted)', fontSize: 18 }}>+</span>
            </div>
            <div style={{ flex: 1 }}>
              <div style={{ fontSize: 13.5, fontWeight: 500, marginBottom: 4 }}>Em branco</div>
              <p className="caption" style={{ lineHeight: 1.4 }}>Cria estrutura no editor visual a partir do zero. Para templates novos sem documento de referência.</p>
            </div>
          </button>
        </div>

        {/* Uploaded file state */}
        {phase === 'uploaded' && (
          <>
            <div style={{ padding: 14, background: 'var(--surface-2)', border: '1px solid var(--border)', borderRadius: 'var(--r-2)', marginBottom: 14, display: 'flex', alignItems: 'center', gap: 14 }}>
              <div style={{ width: 36, height: 44, background: 'white', boxShadow: '0 1px 3px rgba(0,0,0,0.08)', borderRadius: 2, display: 'grid', placeItems: 'center', fontSize: 8.5, fontFamily: 'var(--font-mono)', color: 'var(--brand)', fontWeight: 700 }}>DOCX</div>
              <div style={{ flex: 1 }}>
                <div style={{ fontSize: 13.5, fontWeight: 500 }}>inspecao-mp-base.docx</div>
                <div className="tiny mono" style={{ color: 'var(--text-muted)' }}>147 KB · processado em 1.2s · 7 placeholders detectados</div>
              </div>
              <button className="btn btn-sm btn-ghost">Substituir</button>
            </div>

            {/* Placeholders block — feels like the code preview */}
            <div style={{ padding: 16, background: 'var(--brand-pale)', border: '1px dashed var(--brand-soft)', borderRadius: 'var(--r-3)', marginBottom: 22 }}>
              <div style={{ display: 'flex', alignItems: 'center', marginBottom: 12 }}>
                <div className="kicker" style={{ color: 'var(--brand)' }}>Placeholders detectados · 7 campos</div>
                <span className="spacer"/>
                <span className="tiny mono" style={{ color: 'var(--text-soft)' }}>2 auto-fill · 5 manuais</span>
              </div>
              <div style={{ display: 'flex', flexWrap: 'wrap', gap: 6 }}>
                {placeholders.map((p, i) => (
                  <span key={i} style={{
                    display: 'inline-flex', alignItems: 'center', gap: 5,
                    padding: '5px 10px',
                    background: p.auto ? 'rgba(56,178,134,0.15)' : 'var(--surface)',
                    border: '1px solid ' + (p.auto ? 'rgba(56,178,134,0.4)' : 'var(--border)'),
                    borderRadius: 'var(--r-pill)',
                    fontSize: 11, fontFamily: 'var(--font-mono)',
                    color: p.auto ? 'var(--success)' : 'var(--text-soft)',
                  }}>
                    {p.auto && <span style={{ fontSize: 9 }}>⚡</span>}
                    {p.token}
                    <span style={{ color: 'var(--text-muted)', fontFamily: 'var(--font-sans)', fontWeight: 400 }}>·</span>
                    <span style={{ fontFamily: 'var(--font-sans)', fontSize: 10.5, color: 'var(--text-muted)' }}>{p.kind}</span>
                  </span>
                ))}
              </div>
            </div>

            <div style={{ padding: 12, background: 'var(--info-bg)', border: '1px dashed var(--info)', borderRadius: 'var(--r-2)', fontSize: 12, color: 'var(--text-soft)', marginBottom: 24, lineHeight: 1.5 }}>
              <strong style={{ color: 'var(--info)' }}>ℹ️ Próximos passos:</strong> rótulos, tipos, validações e fontes de auto-fill são refinados no editor visual após criar o template.
            </div>
          </>
        )}

        <TplFooter stepLabel="Etapa 3 de 5 · Avance para definir permissões"/>
      </div>
    </TplShell>
  );
};

// ─── STEP 4 — Permissões ───
const TplStep4 = ({ initialMode = 'roles' } = {}) => {
  const [mode, setMode] = useState(initialMode); // 'roles' | 'areas' | 'all'
  const [areas, setAreas] = useState(['QUA']);
  const [perfis, setPerfis] = useState(['QUA-INSP', 'QUA-ANA', 'PROD-LIDER']);
  const all = [
    { id: 'QUA-INSP', name: 'Inspetor de Qualidade', area: 'Qualidade', count: 14 },
    { id: 'QUA-ANA', name: 'Analista de Qualidade', area: 'Qualidade', count: 8 },
    { id: 'QUA-COORD', name: 'Coordenador de Qualidade', area: 'Qualidade', count: 2 },
    { id: 'PROD-OP', name: 'Operador de Produção', area: 'Produção', count: 47 },
    { id: 'PROD-LIDER', name: 'Líder de Produção', area: 'Produção', count: 6 },
    { id: 'MAN-TEC', name: 'Técnico de Manutenção', area: 'Manutenção', count: 11 },
  ];
  const toggle = id => setPerfis(perfis.includes(id) ? perfis.filter(p => p !== id) : [...perfis, id]);
  const totalUsr = all.filter(p => perfis.includes(p.id)).reduce((s, p) => s + p.count, 0);
  return (
    <TplShell active={3}>
      <div className="card" style={{ padding: '28px 32px' }}>
        <div className="kicker" style={{ marginBottom: 8 }}>Etapa 4 de 5</div>
        <h2 className="h2" style={{ fontSize: 20, marginBottom: 6 }}>Quem pode usar este template?</h2>
        <p className="caption" style={{ marginBottom: 18 }}>Defina o público que verá este template ao criar um novo documento. Você pode ajustar depois.</p>

        {/* Mode segmented control */}
        <div role="tablist" style={{ display: 'inline-flex', padding: 3, background: 'var(--surface-2)', border: '1px solid var(--border)', borderRadius: 'var(--r-2)', marginBottom: 22 }}>
          {[
            { id: 'roles', label: 'Por funções', desc: 'Perfis específicos' },
            { id: 'areas', label: 'Por área', desc: 'Departamentos' },
            { id: 'all', label: 'Todos', desc: 'Empresa inteira' },
          ].map(t => (
            <button key={t.id} role="tab" aria-selected={mode === t.id} onClick={() => setMode(t.id)} style={{
              padding: '8px 14px', border: 'none', cursor: 'pointer', borderRadius: 'var(--r-2)',
              background: mode === t.id ? 'var(--surface)' : 'transparent',
              boxShadow: mode === t.id ? '0 1px 2px rgba(0,0,0,0.06)' : 'none',
              color: mode === t.id ? 'var(--text)' : 'var(--text-muted)',
              fontSize: 12.5, fontWeight: mode === t.id ? 600 : 500,
              display: 'flex', flexDirection: 'column', alignItems: 'flex-start', gap: 1,
              transition: 'all 0.15s',
            }}>
              <span>{t.label}</span>
              <span className="tiny" style={{ color: 'var(--text-muted)', fontWeight: 400 }}>{t.desc}</span>
            </button>
          ))}
        </div>

        {mode === 'all' && (
          <div style={{ padding: 18, background: 'var(--brand-pale)', border: '1px dashed var(--brand-soft)', borderRadius: 'var(--r-3)', marginBottom: 18, display: 'flex', alignItems: 'center', gap: 14 }}>
            <span style={{ width: 42, height: 42, background: 'var(--brand)', color: 'white', borderRadius: 8, display: 'grid', placeItems: 'center', flexShrink: 0 }}>
              <Icon name="home" size={20}/>
            </span>
            <div style={{ flex: 1 }}>
              <div style={{ fontSize: 14, fontWeight: 600, marginBottom: 2 }}>Toda a empresa</div>
              <p className="caption" style={{ lineHeight: 1.5 }}>Todos os colaboradores autenticados verão este template no wizard de novos documentos. Cobertura: <span className="mono" style={{ color: 'var(--brand)' }}>~340</span> usuários ativos em 6 áreas.</p>
            </div>
          </div>
        )}

        {mode === 'areas' && (
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: 10, marginBottom: 18 }}>
            {[
              { id: 'QUA', name: 'Qualidade', count: 28 },
              { id: 'PROD', name: 'Produção', count: 89 },
              { id: 'MAN', name: 'Manutenção', count: 34 },
              { id: 'RH', name: 'RH', count: 12 },
              { id: 'TI', name: 'TI & Segurança', count: 18 },
              { id: 'FIN', name: 'Financeiro', count: 9 },
            ].map(a => {
              const checked = areas.includes(a.id);
              return (
                <button key={a.id} onClick={() => setAreas(checked ? areas.filter(x => x !== a.id) : [...areas, a.id])} style={{
                  textAlign: 'left', padding: 14,
                  border: checked ? '2px solid var(--brand)' : '1px solid var(--border)',
                  background: checked ? 'var(--brand-pale)' : 'var(--surface)',
                  borderRadius: 'var(--r-3)', cursor: 'pointer',
                  display: 'flex', alignItems: 'center', gap: 12,
                }}>
                  <span style={{ width: 32, height: 32, background: checked ? 'var(--brand)' : 'var(--surface-3)', color: checked ? 'white' : 'var(--text-muted)', borderRadius: 6, display: 'grid', placeItems: 'center', flexShrink: 0 }}>
                    <Icon name="taxonomy" size={16}/>
                  </span>
                  <div style={{ flex: 1, minWidth: 0 }}>
                    <div className="mono" style={{ fontSize: 11.5, fontWeight: 600, color: checked ? 'var(--brand)' : 'var(--text-soft)', marginBottom: 2 }}>{a.id}</div>
                    <div style={{ fontSize: 13, fontWeight: 500, marginBottom: 2 }}>{a.name}</div>
                    <div className="tiny" style={{ color: 'var(--text-muted)' }}><span className="mono">{a.count}</span> usuários</div>
                  </div>
                  {checked && <span style={{ color: 'var(--brand)' }}><Icon name="check" size={16}/></span>}
                </button>
              );
            })}
          </div>
        )}

        {mode === 'roles' && (
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(2, 1fr)', gap: 10, marginBottom: 18 }}>
          {all.map(p => {
            const checked = perfis.includes(p.id);
            return (
              <button key={p.id} onClick={() => toggle(p.id)} style={{
                textAlign: 'left', padding: 14,
                border: checked ? '2px solid var(--brand)' : '1px solid var(--border)',
                background: checked ? 'var(--brand-pale)' : 'var(--surface)',
                borderRadius: 'var(--r-3)', cursor: 'pointer',
                display: 'flex', alignItems: 'center', gap: 12,
              }}>
                <span style={{
                  width: 32, height: 32,
                  background: checked ? 'var(--brand)' : 'var(--surface-3)',
                  color: checked ? 'white' : 'var(--text-muted)',
                  borderRadius: 6, display: 'grid', placeItems: 'center', flexShrink: 0,
                }}>
                  <Icon name="users" size={16}/>
                </span>
                <div style={{ flex: 1, minWidth: 0 }}>
                  <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 3 }}>
                    <span className="mono" style={{ fontSize: 11.5, fontWeight: 600, color: checked ? 'var(--brand)' : 'var(--text-soft)' }}>{p.id}</span>
                    <span className="tiny" style={{ color: 'var(--text-muted)' }}>{p.area}</span>
                  </div>
                  <div style={{ fontSize: 13, fontWeight: 500, marginBottom: 2 }}>{p.name}</div>
                  <div className="tiny" style={{ color: 'var(--text-muted)' }}><span className="mono">{p.count}</span> usuários ativos</div>
                </div>
                {checked && <span style={{ color: 'var(--brand)' }}><Icon name="check" size={16}/></span>}
              </button>
            );
          })}
        </div>
        )}

        {mode !== 'all' && (
          <div style={{ padding: 14, background: 'var(--brand-pale)', border: '1px dashed var(--brand-soft)', borderRadius: 'var(--r-3)', marginBottom: 24 }}>
            <div className="kicker" style={{ marginBottom: 4, color: 'var(--brand)' }}>Cobertura estimada</div>
            <div style={{ display: 'flex', alignItems: 'baseline', gap: 8 }}>
              <span className="mono" style={{ fontSize: 22, fontWeight: 600, color: 'var(--brand)' }}>{mode === 'roles' ? perfis.length : areas.length}</span>
              <span className="caption">
                {mode === 'roles'
                  ? <>{perfis.length === 1 ? 'perfil selecionado' : 'perfis selecionados'} · cobre <span className="mono" style={{ color: 'var(--brand)' }}>~{totalUsr}</span> usuários ativos</>
                  : <>{areas.length === 1 ? 'área selecionada' : 'áreas selecionadas'} · cobre todos os usuários das áreas marcadas</>}
              </span>
            </div>
          </div>
        )}

        <TplFooter stepLabel="Etapa 4 de 5 · Avance para revisar e confirmar" primaryDisabled={mode === 'roles' && perfis.length === 0}/>
      </div>
    </TplShell>
  );
};

// ─── STEP 5 — Confirmação ───
const TplStep5 = () => (
  <TplShell active={4}>
    <div className="card" style={{ padding: '28px 32px' }}>
      <div className="kicker" style={{ marginBottom: 8 }}>Etapa 5 de 5</div>
      <h2 className="h2" style={{ fontSize: 20, marginBottom: 6 }}>Revise e confirme a criação</h2>
      <p className="caption" style={{ marginBottom: 24 }}>O template será criado como rascunho. Você pode editar a estrutura no editor visual antes de submeter para revisão.</p>

      {/* Template preview card */}
      <div style={{ display: 'grid', gridTemplateColumns: '120px 1fr', gap: 18, padding: 18, background: 'var(--surface-2)', border: '1px solid var(--border)', borderRadius: 'var(--r-3)', marginBottom: 18 }}>
        <div style={{ width: 120, height: 152, background: 'white', boxShadow: '0 2px 8px rgba(0,0,0,0.06)', borderRadius: 2, padding: '10px 9px' }}>
          <div style={{ height: 4, background: 'var(--brand)', marginBottom: 6, width: '70%' }}/>
          <div style={{ fontSize: 7, fontFamily: 'var(--font-mono)', color: '#888', marginBottom: 4 }}>TPL-POP-009 v1.0</div>
          {[...Array(11)].map((_, k) => {
            const ph = k % 3 === 1;
            return <div key={k} style={{ height: 1.5, background: ph ? 'var(--brand-soft)' : '#e0d0d0', marginBottom: 3, width: `${55 + (k * 11) % 38}%`, opacity: ph ? 0.7 : 1 }}/>;
          })}
        </div>
        <div>
          <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 8 }}>
            <span className="code-chip mono" style={{ fontSize: 13, padding: '3px 9px' }}>TPL-POP-009</span>
            <Status s="draft"/>
            <span className="pill mono" style={{ background: 'transparent' }}>v1.0</span>
          </div>
          <div style={{ fontSize: 16, fontWeight: 500, marginBottom: 14 }}>Inspeção de Recebimento de Matéria-Prima</div>
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 8, fontSize: 12 }}>
            {[
              ['Perfil', 'POP — Procedimento Op.'],
              ['Família', 'Qualidade'],
              ['Origem', '.docx · 7 placeholders'],
              ['Auto-fill', '2 campos automáticos'],
              ['Permissões', '3 perfis · ~28 usuários'],
              ['Autor', 'Marina Silveira'],
            ].map(([k, v], i) => (
              <div key={i} style={{ display: 'flex', justifyContent: 'space-between', padding: '4px 0', borderBottom: '1px dashed var(--border)' }}>
                <span style={{ color: 'var(--text-muted)' }}>{k}</span>
                <span style={{ textAlign: 'right' }}>{v}</span>
              </div>
            ))}
          </div>
        </div>
      </div>

      {/* What happens next */}
      <div style={{ padding: 14, background: 'var(--brand-pale)', border: '1px dashed var(--brand-soft)', borderRadius: 'var(--r-2)', marginBottom: 22 }}>
        <div className="kicker" style={{ marginBottom: 8, color: 'var(--brand)' }}>Ao confirmar</div>
        <ol style={{ margin: 0, paddingLeft: 18, fontSize: 12.5, lineHeight: 1.7, color: 'var(--text-soft)' }}>
          <li>O slot <span className="mono" style={{ color: 'var(--brand)' }}>TPL-POP-009</span> será reservado permanentemente.</li>
          <li>O .docx será clonado como rascunho com os 7 placeholders detectados.</li>
          <li>Você será direcionado para o editor visual para refinar rótulos, tipos e validações.</li>
          <li>Submeta para revisão por <span className="mono" style={{ color: 'var(--brand)' }}>QUA-COORD</span>. Após publicado, o template aparece no wizard de novos documentos.</li>
        </ol>
      </div>

      <label style={{ display: 'flex', alignItems: 'flex-start', gap: 10, fontSize: 12.5, color: 'var(--text-soft)', marginBottom: 8 }}>
        <input type="checkbox" defaultChecked style={{ marginTop: 2 }}/>
        Confirmo que entendi que o código <span className="mono">TPL-POP-009</span> é definitivo e não pode ser reutilizado.
      </label>

      <TplFooter stepLabel="Etapa 5 de 5 · Tudo pronto para criar" primary="Criar e abrir editor →"/>
    </div>
  </TplShell>
);

// Variants for canvas — preselect tab/mode
const TplStep1Doc = () => <TplStep1 __initialScope="document"/>;
const TplStep4Areas = () => <TplStep4 initialMode="areas"/>;
const TplStep4All = () => <TplStep4 initialMode="all"/>;

window.MD_TPLWIZ = { TplStep1, TplStep1Doc, TplStep2, TplStep3, TplStep4, TplStep4Areas, TplStep4All, TplStep5 };
