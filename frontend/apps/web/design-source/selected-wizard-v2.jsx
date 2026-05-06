/* global React, MD */
const { Icon } = MD;
const { useState } = React;

// Step 2 with visibility selector
const WizardStep2V2 = () => {
  const [vis, setVis] = useState('company');
  const visOptions = [
    { id: 'area', icon: 'taxonomy', label: 'Apenas a área', desc: 'Visível somente para membros da área (RH, Qualidade, etc).' },
    { id: 'people', icon: 'users', label: 'Pessoas selecionadas', desc: 'Lista nominal de usuários ou grupos. Acesso por convite.' },
    { id: 'company', icon: 'home', label: 'Toda a empresa', desc: 'Todos os colaboradores autenticados podem consultar.' },
    { id: 'external', icon: 'link', label: 'Externo / Cliente', desc: 'Compartilhável via link público autenticado. Watermark obrigatório.' },
  ];

  const Stepper = ({ active }) => {
    const steps = ['Perfil', 'Área & Código', 'Template', 'Confirmação'];
    return (
      <div style={{ display: 'flex', alignItems: 'center', gap: 0, marginBottom: 24, padding: '12px 18px', background: 'var(--surface)', border: '1px solid var(--border)', borderRadius: 'var(--r-3)' }}>
        {steps.map((s, i) => (
          <React.Fragment key={s}>
            <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
              <span style={{ width: 22, height: 22, borderRadius: '50%', display: 'grid', placeItems: 'center', fontSize: 11, fontWeight: 600, fontFamily: 'var(--font-mono)', background: i < active ? 'var(--brand)' : i === active ? 'var(--surface)' : 'var(--surface-3)', color: i < active ? 'white' : i === active ? 'var(--brand)' : 'var(--text-muted)', border: i === active ? '1.5px solid var(--brand)' : i < active ? 'none' : '1px solid var(--border-strong)' }}>{i < active ? '✓' : i + 1}</span>
              <span style={{ fontSize: 13, fontWeight: i === active ? 600 : 400, color: i <= active ? 'var(--text)' : 'var(--text-muted)' }}>{s}</span>
            </div>
            {i < steps.length - 1 && <span style={{ flex: 1, height: 1, background: 'var(--border)', margin: '0 14px' }}/>}
          </React.Fragment>
        ))}
      </div>
    );
  };

  return (
    <div style={{ flex: 1, overflow: 'auto', background: 'var(--bg)', padding: '32px 28px' }}>
      <div style={{ maxWidth: 880, margin: '0 auto' }}>
        <div className="kicker" style={{ marginBottom: 6 }}>Documentos / Novo</div>
        <h1 className="display-1" style={{ marginBottom: 4 }}>Novo documento controlado</h1>
        <p className="body" style={{ color: 'var(--text-muted)', marginBottom: 28 }}>
          Cada documento recebe um código único <span className="mono">{`{perfil}-{área}-{seq}`}</span>.
        </p>
        <Stepper active={1}/>

        <div className="card" style={{ padding: '28px 32px' }}>
          <div className="kicker" style={{ marginBottom: 8 }}>Etapa 2 de 4</div>
          <h2 className="h2" style={{ fontSize: 20, marginBottom: 20 }}>Área, código e visibilidade</h2>

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
              <label className="kicker" style={{ display: 'block', marginBottom: 6 }}>Área *</label>
              <select className="input" defaultValue="QUA" style={{ height: 38, padding: '0 12px', fontSize: 13 }}>
                <option value="QUA">QUA — Qualidade</option>
                <option value="RH">RH — Recursos Humanos</option>
                <option value="PROD">PROD — Produção</option>
                <option value="TI">TI — TI & Segurança</option>
                <option value="FIN">FIN — Financeiro</option>
              </select>
            </div>
          </div>

          <div style={{ marginBottom: 18 }}>
            <label className="kicker" style={{ display: 'block', marginBottom: 6 }}>Título *</label>
            <input className="input" defaultValue="Inspeção de Recebimento de Matéria-Prima" style={{ height: 38, fontSize: 14 }}/>
          </div>

          {/* Code preview */}
          <div style={{ padding: 16, background: 'var(--brand-pale)', border: '1px dashed var(--brand-soft)', borderRadius: 'var(--r-3)', marginBottom: 22 }}>
            <div className="kicker" style={{ marginBottom: 6, color: 'var(--brand)' }}>Código gerado · próximo em (POP, QUA)</div>
            <div className="mono" style={{ fontSize: 26, fontWeight: 600, color: 'var(--brand)', letterSpacing: '0.02em' }}>POP-QUA-120</div>
            <div className="caption" style={{ marginTop: 4 }}>Sequência ativa: 119 / Próxima: 120 · Códigos não são reutilizados.</div>
          </div>

          {/* Visibility selector — NEW */}
          <div className="kicker" style={{ marginBottom: 10 }}>Visibilidade *</div>
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(2, 1fr)', gap: 10, marginBottom: 24 }}>
            {visOptions.map(o => (
              <button key={o.id} onClick={() => setVis(o.id)} style={{
                textAlign: 'left', padding: 14,
                border: vis === o.id ? '2px solid var(--brand)' : '1px solid var(--border)',
                background: vis === o.id ? 'var(--brand-pale)' : 'var(--surface)',
                borderRadius: 'var(--r-3)', cursor: 'pointer',
                display: 'flex', alignItems: 'flex-start', gap: 10,
              }}>
                <span style={{ width: 32, height: 32, background: vis === o.id ? 'var(--brand)' : 'var(--surface-3)', color: vis === o.id ? 'white' : 'var(--text-muted)', borderRadius: 6, display: 'grid', placeItems: 'center', flexShrink: 0 }}>
                  <Icon name={o.icon} size={16}/>
                </span>
                <div style={{ flex: 1 }}>
                  <div style={{ fontSize: 13, fontWeight: 500, marginBottom: 3 }}>{o.label}</div>
                  <p className="caption" style={{ lineHeight: 1.4 }}>{o.desc}</p>
                </div>
                {vis === o.id && <span style={{ color: 'var(--brand)' }}><Icon name="check" size={14}/></span>}
              </button>
            ))}
          </div>

          {/* Conditional sub-controls */}
          {vis === 'people' && (
            <div style={{ padding: 12, background: 'var(--surface-2)', borderRadius: 'var(--r-2)', marginBottom: 22, border: '1px solid var(--border)' }}>
              <label className="kicker" style={{ display: 'block', marginBottom: 6 }}>Convidar pessoas ou grupos</label>
              <input className="input" placeholder="usuario@empresa.com, Grupo: RH-Liderança…" style={{ height: 34, fontSize: 13 }}/>
              <div style={{ display: 'flex', gap: 6, marginTop: 8, flexWrap: 'wrap' }}>
                {['Marina S.', 'Rafael C.', 'Grupo: RH-Liderança'].map(n => <span key={n} className="pill" style={{ fontSize: 11 }}>{n} ✕</span>)}
              </div>
            </div>
          )}
          {vis === 'external' && (
            <div style={{ padding: 12, background: 'var(--surface-2)', borderRadius: 'var(--r-2)', marginBottom: 22, border: '1px solid var(--border)' }}>
              <label style={{ display: 'flex', alignItems: 'center', gap: 8, fontSize: 12, marginBottom: 6 }}>
                <input type="checkbox" defaultChecked/> Exigir senha de acesso
              </label>
              <label style={{ display: 'flex', alignItems: 'center', gap: 8, fontSize: 12, marginBottom: 6 }}>
                <input type="checkbox" defaultChecked/> Aplicar watermark com email do visualizador
              </label>
              <label style={{ display: 'flex', alignItems: 'center', gap: 8, fontSize: 12 }}>
                <input type="checkbox"/> Expirar acesso após <input type="number" defaultValue="30" style={{ width: 50, marginLeft: 4, marginRight: 4, height: 22, padding: '0 4px' }}/> dias
              </label>
            </div>
          )}

          <div className="divider" style={{ margin: '0 -32px 20px' }}/>
          <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
            <button className="btn">← Voltar</button>
            <span className="spacer"/>
            <span className="caption">Etapa 2 de 4 · Avance para selecionar template</span>
            <button className="btn btn-primary">Avançar →</button>
          </div>
        </div>
      </div>
    </div>
  );
};

window.MD_WIZARD_V2 = { WizardStep2V2 };
