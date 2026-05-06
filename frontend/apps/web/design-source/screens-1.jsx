/* global React, MD, MD_LIBRARY */
const { Icon, Status, Avatar, Logo } = MD;

// === DASHBOARD / HOME ===
const Dashboard = () => (
  <div style={{ flex: 1, overflow: 'auto', background: 'var(--bg)', padding: '24px 28px' }}>
    <div style={{ marginBottom: 24 }}>
      <div className="kicker" style={{ marginBottom: 6 }}>Início · 03 · Maio · 2026</div>
      <h1 className="display-1">Bom dia, Marina.</h1>
      <p className="body" style={{ color: 'var(--text-muted)', marginTop: 6, maxWidth: 560 }}>
        3 documentos aguardam sua aprovação · 2 templates em revisão · próxima auditoria ISO 9001 em <span className="mono" style={{ color: 'var(--text)' }}>34 dias</span>.
      </p>
    </div>

    <div style={{ display: 'grid', gridTemplateColumns: 'repeat(4, 1fr)', gap: 12, marginBottom: 24 }}>
      {[
        { l: 'Acervo total', v: '2.847', d: 'frozen + ativos', spark: [12, 18, 14, 22, 28, 31, 38] },
        { l: 'Aprovações pendentes', v: '11', d: '3 vencendo em 24h', spark: [4, 6, 5, 9, 11, 8, 11], c: 'var(--warning)' },
        { l: 'Em rascunho', v: '23', d: 'autores ativos: 8', spark: [18, 21, 19, 24, 22, 20, 23] },
        { l: 'Frozen este mês', v: '47', d: '+12% MoM', spark: [2, 8, 14, 22, 31, 39, 47], c: 'var(--success)' },
      ].map((s, i) => (
        <div key={i} className="card" style={{ padding: 16 }}>
          <div className="kicker" style={{ marginBottom: 8 }}>{s.l}</div>
          <div style={{ display: 'flex', alignItems: 'flex-end', justifyContent: 'space-between', marginBottom: 8 }}>
            <span className="num" style={{ fontSize: 30, fontWeight: 600, fontFamily: 'var(--font-display)', letterSpacing: '-0.02em' }}>{s.v}</span>
            <svg width="64" height="28" viewBox="0 0 64 28" style={{ flexShrink: 0 }}>
              <polyline fill="none" stroke={s.c || 'var(--brand)'} strokeWidth="1.5" points={s.spark.map((v, i) => `${i * 10 + 2},${28 - v / Math.max(...s.spark) * 22}`).join(' ')}/>
            </svg>
          </div>
          <div className="caption" style={{ color: s.c || 'var(--text-muted)' }}>{s.d}</div>
        </div>
      ))}
    </div>

    <div style={{ display: 'grid', gridTemplateColumns: '1.5fr 1fr', gap: 16 }}>
      {/* My approval queue */}
      <div className="card" style={{ padding: 0, overflow: 'hidden' }}>
        <div style={{ padding: '14px 18px', borderBottom: '1px solid var(--border)', display: 'flex', alignItems: 'center' }}>
          <div>
            <div className="h3" style={{ fontSize: 14 }}>Sua caixa de aprovação</div>
            <div className="caption">3 pendentes · ISO segregation: você não pode aprovar seus próprios envios</div>
          </div>
          <span className="spacer"/>
          <button className="btn btn-sm">Abrir caixa <Icon name="arrow" size={12}/></button>
        </div>
        {[
          { code: 'POP-QUA-118', title: 'Inspeção Dimensional em Componentes Usinados', author: 'Rafael Castro', stage: 'Etapa 2 de 3', due: 'Vence em 6h', urgent: true },
          { code: 'POP-QUA-119', title: 'Análise de Causa Raiz para Não-Conformidades', author: 'Rafael Castro', stage: 'Etapa 1 de 2', due: 'Vence em 2 dias' },
          { code: 'POL-RH-003', title: 'Política de Equidade e Diversidade', author: 'Camila Tavares', stage: 'Etapa 1 de 3', due: 'Sem prazo' },
        ].map((t, i, a) => (
          <div key={i} className="row" style={{ padding: '14px 18px', display: 'flex', alignItems: 'center', gap: 14, borderBottom: i < a.length - 1 ? '1px solid var(--border)' : 'none', cursor: 'pointer' }}>
            <span className="code-chip mono">{t.code}</span>
            <div style={{ flex: 1, minWidth: 0 }}>
              <div style={{ fontSize: 13.5, fontWeight: 500, marginBottom: 3 }}>{t.title}</div>
              <div className="caption">por {t.author} · {t.stage}</div>
            </div>
            <span className="caption" style={{ color: t.urgent ? 'var(--accent)' : 'var(--text-muted)', fontWeight: t.urgent ? 500 : 400 }}>{t.due}</span>
            <button className="btn btn-sm btn-primary">Revisar</button>
          </div>
        ))}
      </div>

      {/* Activity */}
      <div className="card" style={{ padding: 18 }}>
        <div className="h3" style={{ fontSize: 14, marginBottom: 4 }}>Trilha de auditoria</div>
        <div className="caption" style={{ marginBottom: 16 }}>Tempo real · últimas 8h</div>
        <div style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>
          {[
            { who: 'Juliana P.', what: 'aprovou', code: 'POP-FIN-019', t: '14:32' },
            { who: 'Sistema', what: 'gerou PDF de', code: 'IT-PROD-204', t: '13:51' },
            { who: 'Rafael C.', what: 'submeteu', code: 'POP-QUA-119', t: '13:58' },
            { who: 'André L.', what: 'editou', code: 'POL-TI-007', t: '11:24' },
            { who: 'Marina S.', what: 'arquivou', code: 'POP-RH-099', t: '10:17' },
            { who: 'Bruno M.', what: 'congelou', code: 'IT-PROD-203', t: '09:44' },
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
      </div>
    </div>
  </div>
);

// === DOCUMENT DETAIL ===
const DocumentDetail = () => {
  const d = MD_LIBRARY.DOCS[0];
  return (
    <div style={{ flex: 1, display: 'flex', overflow: 'hidden', background: 'var(--bg)' }}>
      <div style={{ flex: 1, overflow: 'auto' }}>
        {/* doc header */}
        <div style={{ background: 'var(--surface)', borderBottom: '1px solid var(--border)', padding: '20px 32px' }}>
          <div className="caption" style={{ marginBottom: 10 }}>
            <a href="#" style={{ color: 'var(--text-muted)' }}>Documentos</a> <span style={{ color: 'var(--text-faint)' }}>/</span> <a href="#" style={{ color: 'var(--text-muted)' }}>RH</a> <span style={{ color: 'var(--text-faint)' }}>/</span> <span style={{ color: 'var(--text)' }}>POP-RH-001</span>
          </div>
          <div style={{ display: 'flex', alignItems: 'flex-start', gap: 16 }}>
            <div style={{ flex: 1 }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: 10, marginBottom: 8 }}>
                <span className="code-chip mono" style={{ fontSize: 13, padding: '3px 9px' }}>POP-RH-001</span>
                <Status s="frozen"/>
                <span className="pill" style={{ background: 'transparent' }}><Icon name="lock" size={10}/>Imutável</span>
                <span className="pill mono" style={{ background: 'transparent' }}>v4 · revisão atual</span>
              </div>
              <h1 className="h2" style={{ fontSize: 24, letterSpacing: '-0.02em', marginBottom: 6 }}>{d.title}</h1>
              <div className="caption">Procedimento Operacional · Recursos Humanos · Vigente desde 15/04/2026</div>
            </div>
            <div style={{ display: 'flex', gap: 6 }}>
              <button className="btn btn-sm"><Icon name="download" size={13}/>PDF</button>
              <button className="btn btn-sm"><Icon name="download" size={13}/>DOCX</button>
              <button className="btn btn-sm"><Icon name="link" size={13}/></button>
              <button className="btn btn-sm"><Icon name="more" size={13}/></button>
              <button className="btn btn-sm btn-primary"><Icon name="plus" size={13}/>Nova revisão</button>
            </div>
          </div>

          <div style={{ display: 'flex', gap: 0, marginTop: 18, marginBottom: -1 }}>
            {['Conteúdo', 'Histórico de revisões', 'Aprovações', 'Auditoria', 'Distribuição'].map((t, i) => (
              <button key={t} style={{ padding: '8px 14px', fontSize: 13, border: 'none', background: 'none', borderBottom: i === 0 ? '2px solid var(--brand)' : '2px solid transparent', color: i === 0 ? 'var(--text)' : 'var(--text-muted)', fontWeight: i === 0 ? 500 : 400, cursor: 'pointer' }}>{t}</button>
            ))}
          </div>
        </div>

        {/* doc preview — paper */}
        <div style={{ padding: '28px', display: 'flex', justifyContent: 'center' }}>
          <div style={{ width: 720, background: 'white', boxShadow: '0 4px 24px rgba(0,0,0,0.08)', padding: '56px 64px', minHeight: 600, fontFamily: 'Georgia, serif' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', borderBottom: '2px solid #1a0e0e', paddingBottom: 12, marginBottom: 28, fontSize: 11, fontFamily: 'var(--font-mono)', color: '#1a0e0e' }}>
              <span>POP-RH-001 · v4</span>
              <span>VIGENTE 15/04/2026</span>
              <span>METALDOCS</span>
            </div>
            <h1 style={{ fontSize: 22, fontWeight: 700, marginBottom: 18, fontFamily: 'Georgia, serif' }}>Admissão e Onboarding de Colaboradores</h1>
            <p style={{ fontSize: 12.5, lineHeight: 1.65, color: '#333', marginBottom: 14 }}>
              <strong>1. Objetivo.</strong> Estabelecer o procedimento padrão para admissão de novos colaboradores na organização, garantindo a conformidade com os requisitos legais (CLT) e os padrões internos de qualidade definidos pelo Sistema de Gestão Integrado.
            </p>
            <p style={{ fontSize: 12.5, lineHeight: 1.65, color: '#333', marginBottom: 14 }}>
              <strong>2. Escopo.</strong> Aplica-se a todos os processos de contratação realizados pela área de Recursos Humanos, incluindo cargos efetivos, temporários e estagiários, em todas as unidades operacionais.
            </p>
            <p style={{ fontSize: 12.5, lineHeight: 1.65, color: '#333', marginBottom: 14 }}>
              <strong>3. Responsabilidades.</strong> O Coordenador de RH é responsável pela execução, condução das entrevistas técnicas e validação documental. O gestor da área requisitante valida competências técnicas e aprova a contratação final.
            </p>
            <p style={{ fontSize: 12.5, lineHeight: 1.65, color: '#333', marginBottom: 14 }}>
              <strong>4. Procedimento.</strong> 4.1 Recebimento da requisição via sistema. 4.2 Triagem curricular e shortlisting. 4.3 Entrevistas técnicas e comportamentais. 4.4 Verificação de referências. 4.5 Proposta formal e negociação. 4.6 Coleta documental e exames admissionais. 4.7 Integração e treinamento inicial.
            </p>
            <div style={{ marginTop: 36, paddingTop: 16, borderTop: '1px solid #ccc', fontSize: 10, color: '#666', fontFamily: 'var(--font-mono)', display: 'flex', justifyContent: 'space-between' }}>
              <span>Autor: Marina Silveira</span>
              <span>Aprovadores: M. Silveira · R. Castro · J. Prado</span>
              <span>hash 7f3a91…b2</span>
            </div>
          </div>
        </div>
      </div>

      {/* RIGHT PANEL: metadata + history */}
      <aside style={{ width: 320, borderLeft: '1px solid var(--border)', background: 'var(--surface-2)', overflow: 'auto', padding: '24px 20px' }}>
        <div className="kicker" style={{ marginBottom: 10 }}>Metadados</div>
        <div style={{ display: 'flex', flexDirection: 'column', gap: 10, marginBottom: 24, fontSize: 12.5 }}>
          {[
            ['Código', <span className="mono">POP-RH-001</span>],
            ['Perfil', 'Procedimento Operacional'],
            ['Área', 'Recursos Humanos'],
            ['Família', 'Qualidade'],
            ['Revisão', <span className="mono">v4</span>],
            ['Vigência', <span className="mono">15/04/2026</span>],
            ['Próx. revisão', <span className="mono">15/04/2027</span>],
            ['Content hash', <span className="mono" style={{ fontSize: 11 }}>7f3a91…b2</span>],
            ['Values hash', <span className="mono" style={{ fontSize: 11 }}>e2d0c8…14</span>],
          ].map(([k, v], i) => (
            <div key={i} style={{ display: 'flex', justifyContent: 'space-between', gap: 12 }}>
              <span style={{ color: 'var(--text-muted)' }}>{k}</span>
              <span style={{ color: 'var(--text)', textAlign: 'right' }}>{v}</span>
            </div>
          ))}
        </div>

        <div className="kicker" style={{ marginBottom: 10 }}>Linha de revisões</div>
        <div style={{ position: 'relative', paddingLeft: 16, marginBottom: 24 }}>
          <span style={{ position: 'absolute', left: 5, top: 8, bottom: 8, width: 1, background: 'var(--border-strong)' }}/>
          {[
            { v: 'v4', d: '15/04/2026', n: 'Atualização anexos LGPD', cur: true },
            { v: 'v3', d: '03/11/2025', n: 'Revisão integração SAP' },
            { v: 'v2', d: '22/05/2024', n: 'Ajuste procedimento estágio' },
            { v: 'v1', d: '08/01/2024', n: 'Versão inicial' },
          ].map((r, i) => (
            <div key={i} style={{ position: 'relative', paddingBottom: 14 }}>
              <span style={{ position: 'absolute', left: -16, top: 4, width: 11, height: 11, borderRadius: '50%', background: r.cur ? 'var(--brand)' : 'var(--surface)', border: `2px solid ${r.cur ? 'var(--brand)' : 'var(--border-strong)'}` }}/>
              <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: 12.5 }}>
                <span style={{ fontWeight: r.cur ? 500 : 400, fontFamily: 'var(--font-mono)' }}>{r.v}</span>
                <span className="tiny mono">{r.d}</span>
              </div>
              <div className="caption" style={{ color: 'var(--text-muted)' }}>{r.n}</div>
            </div>
          ))}
        </div>

        <div className="kicker" style={{ marginBottom: 10 }}>Aprovadores</div>
        <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
          {[
            { n: 'Marina Silveira', r: 'Coordenadora RH', s: 'signed', t: '14/04 · 17:22' },
            { n: 'Rafael Castro', r: 'Líder Qualidade', s: 'signed', t: '15/04 · 09:18' },
            { n: 'Juliana Prado', r: 'Gerente Geral', s: 'signed', t: '15/04 · 11:04' },
          ].map((a, i) => (
            <div key={i} style={{ display: 'flex', alignItems: 'center', gap: 8, fontSize: 12 }}>
              <Avatar name={a.n} size="sm"/>
              <div style={{ flex: 1, minWidth: 0 }}>
                <div style={{ fontWeight: 500 }}>{a.n}</div>
                <div className="tiny" style={{ color: 'var(--text-muted)' }}>{a.r}</div>
              </div>
              <span style={{ color: 'var(--success)' }}><Icon name="check" size={14}/></span>
            </div>
          ))}
        </div>
      </aside>
    </div>
  );
};

window.MD_OTHER = { Dashboard, DocumentDetail };
