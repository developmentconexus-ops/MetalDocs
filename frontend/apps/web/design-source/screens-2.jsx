/* global React, MD */
const { Icon, Status, Avatar, Logo } = MD;
const { useState } = React;

// === CREATE WIZARD ===
const CreateWizard = () => {
  const steps = ['Perfil', 'Área & Código', 'Template', 'Confirmação'];
  const active = 1;
  return (
    <div style={{ flex: 1, overflow: 'auto', background: 'var(--bg)', padding: '32px 28px' }}>
      <div style={{ maxWidth: 880, margin: '0 auto' }}>
        <div className="kicker" style={{ marginBottom: 6 }}>Documentos / Novo</div>
        <h1 className="display-1" style={{ marginBottom: 4 }}>Novo documento controlado</h1>
        <p className="body" style={{ color: 'var(--text-muted)', marginBottom: 28 }}>
          Cada documento recebe um código único <span className="mono">{`{perfil}-{área}-{seq}`}</span> e clona o template publicado para o perfil.
        </p>

        {/* Stepper */}
        <div style={{ display: 'flex', alignItems: 'center', gap: 0, marginBottom: 28, padding: '12px 18px', background: 'var(--surface)', border: '1px solid var(--border)', borderRadius: 'var(--r-3)' }}>
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

        <div className="card" style={{ padding: '28px 32px' }}>
          <div className="kicker" style={{ marginBottom: 8 }}>Etapa 2 de 4</div>
          <h2 className="h2" style={{ fontSize: 20, marginBottom: 20 }}>Selecione área e gere o código</h2>

          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 16, marginBottom: 20 }}>
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

          <div style={{ marginBottom: 20 }}>
            <label className="kicker" style={{ display: 'block', marginBottom: 6 }}>Título *</label>
            <input className="input" defaultValue="Inspeção de Recebimento de Matéria-Prima" style={{ height: 38, fontSize: 14 }}/>
          </div>

          {/* Code preview */}
          <div style={{ padding: 18, background: 'var(--brand-pale)', border: '1px dashed var(--brand-soft)', borderRadius: 'var(--r-3)', marginBottom: 20 }}>
            <div className="kicker" style={{ marginBottom: 8, color: 'var(--brand)' }}>Código gerado · próximo na sequência (POP, QUA)</div>
            <div className="mono" style={{ fontSize: 28, fontWeight: 600, color: 'var(--brand)', letterSpacing: '0.02em' }}>POP-QUA-120</div>
            <div className="caption" style={{ marginTop: 6 }}>Sequência ativa: 119 / Próxima: 120 · Esta é uma operação irreversível — códigos não são reutilizados mesmo após arquivamento.</div>
          </div>

          {/* Token preview */}
          <div className="kicker" style={{ marginBottom: 10 }}>Tokens fixos resolvidos no freeze</div>
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(2, 1fr)', gap: 8, marginBottom: 24, padding: 14, background: 'var(--surface-2)', borderRadius: 'var(--r-2)', border: '1px solid var(--border)' }}>
            {[
              ['{doc_code}', 'POP-QUA-120'],
              ['{doc_title}', 'Inspeção de Recebimento…'],
              ['{revision_number}', 'v1'],
              ['{author}', 'Marina Silveira'],
              ['{effective_date}', 'pendente'],
              ['{approvers}', 'pendente'],
              ['{controlled_by_area}', 'Qualidade'],
            ].map(([k, v], i) => (
              <div key={i} style={{ display: 'flex', justifyContent: 'space-between', padding: '4px 0', fontSize: 12, borderBottom: i < 6 ? '1px dashed var(--border)' : 'none' }}>
                <span className="mono" style={{ color: 'var(--brand)' }}>{k}</span>
                <span style={{ color: 'var(--text-soft)' }}>{v}</span>
              </div>
            ))}
          </div>

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

// === REGISTRY ===
const Registry = () => (
  <div style={{ flex: 1, overflow: 'auto', background: 'var(--bg)', padding: '24px 28px' }}>
    <div style={{ display: 'flex', alignItems: 'flex-end', gap: 16, marginBottom: 20 }}>
      <div>
        <div className="kicker" style={{ marginBottom: 6 }}>Registro · Documentos Controlados</div>
        <h1 className="display-1">Catálogo de slots</h1>
        <p className="body" style={{ color: 'var(--text-muted)', marginTop: 6 }}>Slots numerados por <span className="mono">(perfil, área)</span>. Cada slot abriga uma ou mais revisões versionadas.</p>
      </div>
      <span className="spacer"/>
      <button className="btn"><Icon name="download" size={13}/>Exportar CSV</button>
      <button className="btn btn-primary"><Icon name="plus" size={13}/>Novo slot</button>
    </div>

    {/* Sequence counters strip */}
    <div className="kicker" style={{ marginBottom: 10 }}>Contadores ativos por (perfil, área)</div>
    <div style={{ display: 'grid', gridTemplateColumns: 'repeat(6, 1fr)', gap: 8, marginBottom: 24 }}>
      {[
        ['POP', 'RH', 47], ['POP', 'QUA', 119], ['POP', 'PROD', 84], ['POP', 'TI', 22], ['POP', 'FIN', 19],
        ['DC', 'RH', 42], ['DC', 'PROD', 88], ['DC', 'QUA', 31], ['IT', 'PROD', 204], ['IT', 'QUA', 67],
        ['POL', 'TI', 7], ['POL', 'RH', 3],
      ].map(([p, a, n], i) => (
        <div key={i} className="card" style={{ padding: '10px 12px' }}>
          <div className="mono" style={{ fontSize: 11, color: 'var(--text-muted)' }}>{p}-{a}</div>
          <div style={{ display: 'flex', alignItems: 'baseline', gap: 6 }}>
            <span className="num" style={{ fontSize: 18, fontWeight: 600, fontFamily: 'var(--font-display)' }}>{String(n).padStart(3, '0')}</span>
            <span className="tiny" style={{ color: 'var(--text-faint)' }}>próx {String(n + 1).padStart(3, '0')}</span>
          </div>
        </div>
      ))}
    </div>

    <div className="card" style={{ padding: 0, overflow: 'hidden' }}>
      <div style={{ display: 'grid', gridTemplateColumns: '160px 1fr 110px 100px 90px 110px 90px 36px', gap: 12, padding: '10px 16px', background: 'var(--surface-2)', borderBottom: '1px solid var(--border)' }}>
        {['Código', 'Título', 'Área', 'Perfil', 'Revisões', 'Estado atual', 'Vigência', ''].map((h, i) => <div key={i} className="kicker" style={{ fontSize: 10 }}>{h}</div>)}
      </div>
      {[
        { c: 'POP-RH-001', t: 'Admissão e Onboarding de Colaboradores', a: 'RH', p: 'POP', r: 4, s: 'frozen', v: '15/04/2026' },
        { c: 'POP-RH-002', t: 'Demissão e Offboarding', a: 'RH', p: 'POP', r: 3, s: 'frozen', v: '03/02/2026' },
        { c: 'DC-RH-042', t: 'Descrição de Cargo — Analista Fiscal Sênior', a: 'RH', p: 'DC', r: 1, s: 'draft', v: '—' },
        { c: 'POP-QUA-118', t: 'Inspeção Dimensional em Componentes', a: 'Qualidade', p: 'POP', r: 2, s: 'review', v: '—' },
        { c: 'POP-QUA-119', t: 'Análise de Causa Raiz (RCA)', a: 'Qualidade', p: 'POP', r: 1, s: 'review', v: '—' },
        { c: 'IT-PROD-204', t: 'Calibração de Torquímetro', a: 'Produção', p: 'IT', r: 6, s: 'approved', v: '08/05/2026' },
        { c: 'POP-FIN-019', t: 'Conciliação Bancária Mensal', a: 'Financeiro', p: 'POP', r: 2, s: 'frozen', v: '20/02/2026' },
        { c: 'POL-TI-007', t: 'Política de Classificação da Informação', a: 'TI & Seg.', p: 'POL', r: 3, s: 'frozen', v: '01/03/2026' },
      ].map((d, i, a) => (
        <div key={i} className="row" style={{ display: 'grid', gridTemplateColumns: '160px 1fr 110px 100px 90px 110px 90px 36px', gap: 12, padding: '12px 16px', alignItems: 'center', borderBottom: i < a.length - 1 ? '1px solid var(--border)' : 'none', cursor: 'pointer' }}>
          <span className="code-chip mono">{d.c}</span>
          <span style={{ fontSize: 13.5, fontWeight: 500 }}>{d.t}</span>
          <span className="body-sm" style={{ color: 'var(--text-soft)' }}>{d.a}</span>
          <span className="mono" style={{ fontSize: 11, color: 'var(--text-muted)' }}>{d.p}</span>
          <span className="dotline">{[...Array(Math.max(d.r, 4))].map((_, k) => <span key={k} className={k < d.r ? 'on' : ''}/>)}<span className="num mono" style={{ marginLeft: 6, fontSize: 11, color: 'var(--text-muted)' }}>{d.r}</span></span>
          <Status s={d.s}/>
          <span className="caption mono">{d.v}</span>
          <button className="btn btn-icon btn-sm btn-ghost"><Icon name="more" size={13}/></button>
        </div>
      ))}
    </div>
  </div>
);

// === TEMPLATES LIST ===
const Templates = () => (
  <div style={{ flex: 1, overflow: 'auto', background: 'var(--bg)', padding: '24px 28px' }}>
    <div style={{ display: 'flex', alignItems: 'flex-end', gap: 16, marginBottom: 24 }}>
      <div>
        <div className="kicker" style={{ marginBottom: 6 }}>Templates</div>
        <h1 className="display-1">Layouts reutilizáveis</h1>
        <p className="body" style={{ color: 'var(--text-muted)', marginTop: 6 }}>Versionados, aprovados, publicados. Vinculados a perfis para clonagem em novos documentos.</p>
      </div>
      <span className="spacer"/>
      <button className="btn btn-primary"><Icon name="plus" size={13}/>Novo template</button>
    </div>

    <div style={{ display: 'flex', gap: 0, marginBottom: 16, borderBottom: '1px solid var(--border)' }}>
      {['Todos · 24', 'Publicados · 18', 'Em revisão · 4', 'Rascunhos · 2'].map((t, i) => (
        <button key={t} style={{ padding: '8px 16px', fontSize: 13, border: 'none', background: 'none', borderBottom: i === 0 ? '2px solid var(--brand)' : '2px solid transparent', color: i === 0 ? 'var(--text)' : 'var(--text-muted)', fontWeight: i === 0 ? 500 : 400, cursor: 'pointer', marginBottom: -1 }}>{t}</button>
      ))}
    </div>

    <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: 14 }}>
      {[
        { t: 'POP — Procedimento Operacional', d: 'Layout padrão para POPs com 7 seções obrigatórias (objetivo, escopo, responsabilidades, procedimento, registros, anexos, controle de revisões).', v: 'v3.2', s: 'frozen', bound: ['POP-RH', 'POP-QUA', 'POP-PROD', 'POP-FIN'], updated: '12 dias atrás', author: 'Juliana Prado' },
        { t: 'DC — Descrição de Cargo', d: 'Cabeçalho com identificação, missão do cargo, responsabilidades-chave, requisitos técnicos e comportamentais, condições de trabalho.', v: 'v2.0', s: 'frozen', bound: ['DC-RH', 'DC-PROD'], updated: '2 meses atrás', author: 'Marina Silveira' },
        { t: 'POL — Política', d: 'Layout corporativo para políticas com glossário, princípios, diretrizes, responsabilidades, exceções e revisão.', v: 'v1.4', s: 'frozen', bound: ['POL-TI', 'POL-RH'], updated: '5 dias atrás', author: 'André Lima' },
        { t: 'IT — Instrução de Trabalho', d: 'Modelo enxuto para passo-a-passo operacional: ferramenta, EPI, sequência, critério de aceitação, registros.', v: 'v4.1', s: 'frozen', bound: ['IT-PROD', 'IT-QUA'], updated: '1 mês atrás', author: 'Bruno Mendes' },
        { t: 'POP — Procedimento Operacional', d: 'Próxima versão em revisão. Inclui novo bloco LGPD e ajustes nos campos de aprovadores múltiplos.', v: 'v3.3', s: 'review', bound: [], updated: '4 horas atrás', author: 'Juliana Prado' },
        { t: 'CHK — Checklist Operacional', d: 'Layout de checklist com itens binários, observações e assinatura por linha. Em desenvolvimento.', v: 'v0.4', s: 'draft', bound: [], updated: 'ontem', author: 'Camila Tavares' },
      ].map((tpl, i) => (
        <div key={i} className="card" style={{ padding: 0, overflow: 'hidden', cursor: 'pointer' }}>
          {/* mini doc preview */}
          <div style={{ height: 110, background: 'var(--surface-2)', borderBottom: '1px solid var(--border)', padding: 14, position: 'relative', overflow: 'hidden' }}>
            <div style={{ width: 80, height: 100, background: 'white', boxShadow: '0 2px 8px rgba(0,0,0,0.06)', borderRadius: 2, padding: '8px 7px', position: 'absolute', left: 18, top: 14 }}>
              <div style={{ height: 4, background: 'var(--brand)', marginBottom: 5, width: '70%' }}/>
              {[...Array(8)].map((_, k) => <div key={k} style={{ height: 1.5, background: 'var(--border-strong)', marginBottom: 3, width: `${60 + (k * 7) % 35}%` }}/>)}
            </div>
            <div style={{ position: 'absolute', right: 14, top: 14, display: 'flex', flexDirection: 'column', gap: 5 }}>
              <Status s={tpl.s}/>
              <span className="pill mono" style={{ background: 'transparent' }}>{tpl.v}</span>
            </div>
          </div>
          <div style={{ padding: 14 }}>
            <div className="h3" style={{ fontSize: 14, marginBottom: 6, lineHeight: 1.3 }}>{tpl.t}</div>
            <p className="caption" style={{ marginBottom: 12, lineHeight: 1.45, height: 50, overflow: 'hidden' }}>{tpl.d}</p>
            <div style={{ display: 'flex', alignItems: 'center', gap: 6, marginBottom: 10, flexWrap: 'wrap' }}>
              {tpl.bound.length > 0 ? tpl.bound.slice(0, 3).map(b => <span key={b} className="pill mono" style={{ fontSize: 10 }}>{b}</span>) : <span className="caption" style={{ color: 'var(--text-faint)' }}>Não vinculado a perfis</span>}
              {tpl.bound.length > 3 && <span className="pill" style={{ fontSize: 10 }}>+{tpl.bound.length - 3}</span>}
            </div>
            <div className="divider" style={{ margin: '10px 0' }}/>
            <div style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
              <Avatar name={tpl.author} size="sm"/>
              <span className="caption" style={{ flex: 1 }}>{tpl.author.split(' ')[0]}</span>
              <span className="tiny mono">{tpl.updated}</span>
            </div>
          </div>
        </div>
      ))}
    </div>
  </div>
);

// === LOGIN ===
const Login = () => (
  <div style={{ flex: 1, display: 'flex', overflow: 'hidden', background: 'var(--bg)' }}>
    {/* Left visual */}
    <div style={{ flex: 1.1, background: 'var(--rail)', color: 'var(--rail-text)', padding: '48px 56px', display: 'flex', flexDirection: 'column', position: 'relative', overflow: 'hidden' }}>
      <div style={{ position: 'absolute', inset: 0, opacity: 0.06, backgroundImage: 'repeating-linear-gradient(90deg, transparent, transparent 39px, currentColor 39px, currentColor 40px), repeating-linear-gradient(0deg, transparent, transparent 39px, currentColor 39px, currentColor 40px)' }}/>
      <div style={{ display: 'flex', alignItems: 'center', gap: 10, position: 'relative', zIndex: 1 }}>
        <span style={{ width: 28, height: 28, borderRadius: 6, background: 'var(--brand)' }}/>
        <span style={{ fontSize: 16, fontWeight: 600, letterSpacing: '-0.02em' }}>MetalDocs</span>
      </div>
      <div style={{ flex: 1, display: 'flex', flexDirection: 'column', justifyContent: 'center', position: 'relative', zIndex: 1, maxWidth: 480 }}>
        <div className="kicker" style={{ color: 'var(--rail-text-muted)', marginBottom: 16 }}>§ Plataforma de Documentos Controlados</div>
        <h1 style={{ fontSize: 44, lineHeight: 1.05, letterSpacing: '-0.025em', fontWeight: 600, marginBottom: 18, fontFamily: 'var(--font-display)' }}>
          Auditoria<br/>
          <span style={{ color: 'var(--accent)', fontStyle: 'italic', fontWeight: 500 }}>imutável.</span><br/>
          Aprovação rastreável.
        </h1>
        <p style={{ fontSize: 14, lineHeight: 1.6, color: 'var(--rail-text-muted)', maxWidth: 380 }}>
          ISO 9001 · 14001 · 45001 · 27001. Versionamento criptográfico, segregação de funções e fanout automático para PDF.
        </p>
      </div>
      <div style={{ display: 'flex', gap: 28, fontSize: 11, color: 'var(--rail-text-muted)', fontFamily: 'var(--font-mono)', position: 'relative', zIndex: 1 }}>
        <span>v3.4.2</span>
        <span>BUILD a14d22</span>
        <span>SOC 2 TYPE II</span>
        <span style={{ display: 'flex', alignItems: 'center', gap: 6 }}><span style={{ width: 6, height: 6, borderRadius: '50%', background: 'var(--success)' }}/>OPERACIONAL</span>
      </div>
    </div>
    {/* Right form */}
    <div style={{ flex: 1, display: 'flex', alignItems: 'center', justifyContent: 'center', padding: 48 }}>
      <div style={{ width: 360 }}>
        <div className="kicker" style={{ marginBottom: 8 }}>Acesso · Tenant ACME</div>
        <h2 className="h2" style={{ fontSize: 26, marginBottom: 4, letterSpacing: '-0.02em' }}>Entrar na sua conta</h2>
        <p className="caption" style={{ marginBottom: 28 }}>Use suas credenciais corporativas ou SSO.</p>

        <button className="btn btn-lg" style={{ width: '100%', marginBottom: 12, justifyContent: 'center', height: 42 }}>
          <span style={{ width: 16, height: 16, background: 'var(--text)', display: 'inline-block', maskImage: 'radial-gradient(circle at 50% 50%, transparent 30%, black 30%)' }}/>
          Continuar com SSO corporativo
        </button>
        <div style={{ display: 'flex', alignItems: 'center', gap: 10, margin: '20px 0' }}>
          <div className="divider" style={{ flex: 1 }}/>
          <span className="tiny mono">OU</span>
          <div className="divider" style={{ flex: 1 }}/>
        </div>

        <label className="kicker" style={{ display: 'block', marginBottom: 6 }}>Identificador</label>
        <input className="input" defaultValue="marina.silveira" style={{ height: 40, fontSize: 14, marginBottom: 14 }}/>
        <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 6 }}>
          <span className="kicker">Senha</span>
          <a href="#" className="tiny" style={{ color: 'var(--brand)', textDecoration: 'none' }}>Esqueci minha senha</a>
        </div>
        <input className="input" type="password" defaultValue="••••••••••" style={{ height: 40, fontSize: 14, marginBottom: 16 }}/>
        <label style={{ display: 'flex', alignItems: 'center', gap: 8, fontSize: 12, color: 'var(--text-soft)', marginBottom: 18 }}>
          <input type="checkbox" defaultChecked/> Manter sessão por 8 horas
        </label>
        <button className="btn btn-primary btn-lg" style={{ width: '100%', justifyContent: 'center', height: 42 }}>Entrar</button>

        <p className="tiny" style={{ marginTop: 24, textAlign: 'center', color: 'var(--text-muted)', lineHeight: 1.5 }}>
          Suporte: <a href="#" style={{ color: 'var(--text-soft)' }}>helpdesk@metaldocs.acme</a><br/>
          Acessos administrados por <span className="mono">IAM</span> · ISO segregation enforced
        </p>
      </div>
    </div>
  </div>
);

// === SITE MAP ===
const SiteMap = () => {
  const sections = [
    { title: 'Auth', color: 'var(--info)', items: ['Login (SSO + senha)', 'Trocar senha (forçada)', 'Recuperação'] },
    { title: 'Início', color: 'var(--brand)', items: ['Dashboard', 'Caixa de aprovação', 'Trilha de auditoria'] },
    { title: 'Documentos', color: 'var(--brand)', items: ['Biblioteca (table)', 'Detalhe + revisões', 'Editor (eigenpal)', 'Wizard de criação', 'Histórico de versões'] },
    { title: 'Registro', color: 'var(--success)', items: ['Catálogo de slots', 'Detalhe do slot', 'Contadores por (perfil, área)'] },
    { title: 'Templates', color: 'var(--warning)', items: ['Lista de templates', 'Authoring (eigenpal)', 'Versões + lifecycle'] },
    { title: 'Aprovações', color: 'var(--accent)', items: ['Caixa de entrada', 'Detalhe + signoff', 'Rotas de aprovação (admin)'] },
    { title: 'Taxonomia', color: 'var(--info)', items: ['Famílias (global)', 'Áreas', 'Perfis + binding template'] },
    { title: 'Admin', color: 'var(--text-soft)', items: ['Usuários online', 'CRUD usuários', 'Capabilities', 'Áreas/perfis (membership)', 'Auditoria operacional'] },
    { title: 'Sistema', color: 'var(--text-muted)', items: ['Notificações', 'Configurações', 'Operações (workers, fanout)'] },
  ];
  return (
    <div style={{ flex: 1, overflow: 'auto', background: 'var(--bg)', padding: '32px 28px' }}>
      <div style={{ marginBottom: 24 }}>
        <div className="kicker" style={{ marginBottom: 6 }}>§ 00 · Mapa do produto</div>
        <h1 className="display-1">Arquitetura de telas — MetalDocs</h1>
        <p className="body" style={{ color: 'var(--text-muted)', marginTop: 8, maxWidth: 620 }}>
          9 áreas, ~40 telas. Núcleo do produto em Documentos / Registro / Aprovações. Authoring em Templates. Governança em Taxonomia / Admin. Esta versão entrega protótipos hi-fi das telas em destaque.
        </p>
      </div>
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: 14 }}>
        {sections.map((sec, i) => (
          <div key={i} className="card" style={{ padding: 18, position: 'relative', overflow: 'hidden' }}>
            <span style={{ position: 'absolute', left: 0, top: 0, bottom: 0, width: 3, background: sec.color }}/>
            <div className="kicker" style={{ marginBottom: 12, color: sec.color, fontSize: 11 }}>§ 0{i + 1} · {sec.title}</div>
            <ul style={{ margin: 0, padding: 0, listStyle: 'none', display: 'flex', flexDirection: 'column', gap: 8 }}>
              {sec.items.map((it, j) => (
                <li key={j} style={{ display: 'flex', alignItems: 'center', gap: 10, fontSize: 13 }}>
                  <span className="mono" style={{ fontSize: 10.5, color: 'var(--text-faint)', width: 22 }}>{String(j + 1).padStart(2, '0')}</span>
                  <span>{it}</span>
                </li>
              ))}
            </ul>
          </div>
        ))}
      </div>
    </div>
  );
};

window.MD_SCREENS_2 = { CreateWizard, Registry, Templates, Login, SiteMap };
