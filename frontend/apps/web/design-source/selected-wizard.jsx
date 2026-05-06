/* global React, MD */
const { Icon, Status, Avatar } = MD;
const { useState } = React;

// Stepper component
const Stepper = ({ active }) => {
  const steps = ['Perfil', 'Área & Código', 'Template', 'Confirmação'];
  return (
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
};

const WizardShell = ({ active, children }) => (
  <div style={{ flex: 1, overflow: 'auto', background: 'var(--bg)', padding: '32px 28px' }}>
    <div style={{ maxWidth: 880, margin: '0 auto' }}>
      <div className="kicker" style={{ marginBottom: 6 }}>Documentos / Novo</div>
      <h1 className="display-1" style={{ marginBottom: 4 }}>Novo documento controlado</h1>
      <p className="body" style={{ color: 'var(--text-muted)', marginBottom: 28 }}>
        Cada documento recebe um código único <span className="mono">{`{perfil}-{área}-{seq}`}</span> e clona o template publicado para o perfil.
      </p>
      <Stepper active={active}/>
      {children}
    </div>
  </div>
);

const Footer = ({ stepLabel, primary = 'Avançar →', back = true, primaryDisabled = false }) => (
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

// ─── STEP 1 — Perfil ───
const WizardStep1 = () => {
  const [selected, setSelected] = useState('POP');
  const profiles = [
    { id: 'POP', name: 'Procedimento Operacional', desc: 'Sequência detalhada de execução de uma rotina operacional.', family: 'Qualidade', count: 187 },
    { id: 'IT', name: 'Instrução de Trabalho', desc: 'Passo-a-passo enxuto para tarefa específica em estação.', family: 'Qualidade', count: 312 },
    { id: 'POL', name: 'Política', desc: 'Diretrizes corporativas de alto nível com escopo organizacional.', family: 'Governança', count: 24 },
    { id: 'DC', name: 'Descrição de Cargo', desc: 'Identificação, responsabilidades e requisitos de uma posição.', family: 'RH', count: 156 },
    { id: 'CHK', name: 'Checklist Operacional', desc: 'Lista de verificação binária com observações por item.', family: 'Qualidade', count: 41, soon: true },
    { id: 'FOR', name: 'Formulário', desc: 'Captura estruturada de dados em campo ou em workflow.', family: 'Geral', count: 78 },
  ];
  return (
    <WizardShell active={0}>
      <div className="card" style={{ padding: '28px 32px' }}>
        <div className="kicker" style={{ marginBottom: 8 }}>Etapa 1 de 4</div>
        <h2 className="h2" style={{ fontSize: 20, marginBottom: 6 }}>Que tipo de documento você está criando?</h2>
        <p className="caption" style={{ marginBottom: 24 }}>O perfil determina o template e a sequência de codificação. Esta escolha não pode ser alterada depois.</p>

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
                <span><span className="mono" style={{ color: 'var(--text-soft)' }}>{p.count}</span> existentes</span>
              </div>
            </button>
          ))}
        </div>
        <Footer stepLabel="Etapa 1 de 4 · Selecione um perfil para continuar"/>
      </div>
    </WizardShell>
  );
};

// ─── STEP 3 — Template ───
const WizardStep3 = () => {
  const [selected, setSelected] = useState('latest');
  const templates = [
    { id: 'latest', v: 'v3.2', d: '12 dias atrás', author: 'Juliana Prado', label: 'Versão publicada (recomendada)', notes: 'Inclui as 7 seções obrigatórias + bloco LGPD atualizado.', current: true },
    { id: 'v31', v: 'v3.1', d: '3 meses atrás', author: 'Juliana Prado', label: 'Versão anterior', notes: 'Sem bloco LGPD. Use apenas se requerido por contrato vigente.' },
    { id: 'blank', v: '—', d: '', author: '', label: 'Em branco', notes: 'Inicia sem template. Estrutura precisa ser definida manualmente. Não recomendado para POPs.', warning: true },
  ];
  return (
    <WizardShell active={2}>
      <div className="card" style={{ padding: '28px 32px' }}>
        <div className="kicker" style={{ marginBottom: 8 }}>Etapa 3 de 4 · Perfil POP — Procedimento Operacional</div>
        <h2 className="h2" style={{ fontSize: 20, marginBottom: 6 }}>Selecione o template a ser clonado</h2>
        <p className="caption" style={{ marginBottom: 24 }}>O conteúdo do template será copiado para o documento. Tokens fixos serão resolvidos no freeze.</p>

        <div style={{ display: 'flex', flexDirection: 'column', gap: 10, marginBottom: 8 }}>
          {templates.map(t => (
            <button key={t.id} onClick={() => setSelected(t.id)} style={{
              textAlign: 'left', padding: 16,
              border: selected === t.id ? '2px solid var(--brand)' : '1px solid var(--border)',
              background: selected === t.id ? 'var(--brand-pale)' : 'var(--surface)',
              borderRadius: 'var(--r-3)', cursor: 'pointer',
              display: 'flex', alignItems: 'center', gap: 16,
            }}>
              {/* mini preview */}
              <div style={{ width: 60, height: 76, background: 'white', boxShadow: '0 2px 6px rgba(0,0,0,0.06)', borderRadius: 2, padding: '6px 5px', flexShrink: 0 }}>
                <div style={{ height: 3, background: t.id === 'blank' ? 'var(--border-strong)' : 'var(--brand)', marginBottom: 4, width: '60%' }}/>
                {t.id !== 'blank' && [...Array(7)].map((_, k) => <div key={k} style={{ height: 1.5, background: 'var(--border-strong)', marginBottom: 2.5, width: `${55 + (k * 9) % 35}%` }}/>)}
              </div>
              <div style={{ flex: 1, minWidth: 0 }}>
                <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 4 }}>
                  <span style={{ fontSize: 14, fontWeight: 500 }}>{t.label}</span>
                  {t.v !== '—' && <span className="pill mono" style={{ fontSize: 10 }}>{t.v}</span>}
                  {t.current && <span className="pill pill-frozen"><span className="dot"/>publicada</span>}
                  {t.warning && <span className="pill" style={{ fontSize: 10, color: 'var(--warning)' }}>cuidado</span>}
                </div>
                <p className="caption" style={{ marginBottom: 6, lineHeight: 1.4 }}>{t.notes}</p>
                {t.author && <div className="tiny" style={{ color: 'var(--text-muted)' }}>por {t.author} · atualizado {t.d}</div>}
              </div>
              {selected === t.id && <span style={{ color: 'var(--brand)' }}><Icon name="check" size={18}/></span>}
            </button>
          ))}
        </div>
        <Footer stepLabel="Etapa 3 de 4 · Avance para revisar e confirmar"/>
      </div>
    </WizardShell>
  );
};

// ─── STEP 4 — Confirmação ───
const WizardStep4 = () => (
  <WizardShell active={3}>
    <div className="card" style={{ padding: '28px 32px' }}>
      <div className="kicker" style={{ marginBottom: 8 }}>Etapa 4 de 4</div>
      <h2 className="h2" style={{ fontSize: 20, marginBottom: 6 }}>Confirme a criação do documento</h2>
      <p className="caption" style={{ marginBottom: 24 }}>Os campos abaixo serão registrados no slot. O código é definitivo e não pode ser reutilizado.</p>

      {/* Doc preview card */}
      <div style={{ display: 'grid', gridTemplateColumns: '120px 1fr', gap: 18, padding: 18, background: 'var(--surface-2)', border: '1px solid var(--border)', borderRadius: 'var(--r-3)', marginBottom: 18 }}>
        <div style={{ width: 120, height: 152, background: 'white', boxShadow: '0 2px 8px rgba(0,0,0,0.06)', borderRadius: 2, padding: '10px 9px' }}>
          <div style={{ height: 4, background: 'var(--brand)', marginBottom: 6, width: '70%' }}/>
          <div style={{ fontSize: 7, fontFamily: 'var(--font-mono)', color: '#888', marginBottom: 4 }}>POP-QUA-120 v1</div>
          {[...Array(11)].map((_, k) => <div key={k} style={{ height: 1.5, background: '#e0d0d0', marginBottom: 3, width: `${55 + (k * 11) % 38}%` }}/>)}
        </div>
        <div>
          <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 8 }}>
            <span className="code-chip mono" style={{ fontSize: 13, padding: '3px 9px' }}>POP-QUA-120</span>
            <Status s="draft"/>
            <span className="pill mono" style={{ background: 'transparent' }}>v1</span>
          </div>
          <div style={{ fontSize: 16, fontWeight: 500, marginBottom: 14 }}>Inspeção de Recebimento de Matéria-Prima</div>
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 8, fontSize: 12 }}>
            {[
              ['Perfil', 'POP — Procedimento Op.'],
              ['Família', 'Qualidade'],
              ['Área', 'QUA — Qualidade'],
              ['Template', 'POP v3.2 (publicada)'],
              ['Autor', 'Marina Silveira'],
              ['Criado em', '03/05/2026 · 14:48'],
            ].map(([k, v], i) => (
              <div key={i} style={{ display: 'flex', justifyContent: 'space-between', padding: '4px 0', borderBottom: '1px dashed var(--border)' }}>
                <span style={{ color: 'var(--text-muted)' }}>{k}</span>
                <span>{v}</span>
              </div>
            ))}
          </div>
        </div>
      </div>

      {/* What happens next */}
      <div style={{ padding: 14, background: 'var(--brand-pale)', border: '1px dashed var(--brand-soft)', borderRadius: 'var(--r-2)', marginBottom: 22 }}>
        <div className="kicker" style={{ marginBottom: 8, color: 'var(--brand)' }}>Ao confirmar</div>
        <ol style={{ margin: 0, paddingLeft: 18, fontSize: 12.5, lineHeight: 1.7, color: 'var(--text-soft)' }}>
          <li>O slot <span className="mono" style={{ color: 'var(--brand)' }}>POP-QUA-120</span> será reservado permanentemente — códigos não são reutilizados mesmo após arquivamento.</li>
          <li>Uma cópia do template <span className="mono" style={{ color: 'var(--brand)' }}>POP v3.2</span> será clonada como rascunho.</li>
          <li>Você será direcionado para o editor para preencher o conteúdo.</li>
          <li>Tokens fixos (código, autor, vigência, aprovadores) serão resolvidos no momento do freeze.</li>
        </ol>
      </div>

      <label style={{ display: 'flex', alignItems: 'flex-start', gap: 10, fontSize: 12.5, color: 'var(--text-soft)', marginBottom: 8 }}>
        <input type="checkbox" defaultChecked style={{ marginTop: 2 }}/>
        Confirmo que entendi que o código <span className="mono">POP-QUA-120</span> é definitivo e não pode ser reutilizado.
      </label>

      <Footer stepLabel="Etapa 4 de 4 · Tudo pronto para criar" primary="Criar documento"/>
    </div>
  </WizardShell>
);

window.MD_WIZARD = { WizardStep1, WizardStep3, WizardStep4 };
