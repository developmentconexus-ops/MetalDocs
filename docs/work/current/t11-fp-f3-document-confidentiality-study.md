# T11 — FP2-F3 — Confidencialidade por documento — estudo de Global Maximum

> **Status:** UPSTREAM FINDING / ESTUDO — aguardando adjudicação do operador.
> **Trigger:** operador, 2026-08-26 — "documento que não deve ser visto por trabalhadores da produção, somente pelo gerente e diretoria".
> **Methods:** Engineering Method v1.0.0 + Frontend Method v2.3 §3.10A / §26 (categoria 7) / §12 (P6).
> **Implementation:** BLOCKED.

## 1. A necessidade humana (provada, não hipotética)

```text
ator          gerente de produção / diretoria
contexto      a Área "Produção" contém documentos operacionais lidos por todos os trabalhadores
              e documentos sensíveis (custo, margem, plano de reestruturação, laudo disciplinar)
necessidade   publicar o documento sensível DENTRO do contexto organizacional correto,
              legível apenas por um subconjunto
resultado     o documento vive na Área a que pertence, sem que sua sensibilidade vaze
```

A necessidade é ordinária em empresa real e independe do porte. Não é preferência de UI.

## 2. Limite exato da autoridade atual (provado)

```text
T3 equação canônica       leitura = document.read_effective em escopo Company | Area
viewer bundle             = { document.read_effective }
RoleAssignmentScopeKind   = company | area          (nenhuma outra dimensão existe)
predicados de documento   condicionam estado/relação/governança, NUNCA sensibilidade
T3 lista negativa         "materialized ACL" — sem autoridade semântica persistida
```

**Conclusão:** dois documentos na mesma Área são **indistinguíveis** para efeito de leitura. A autoridade atual **não consegue expressar** a necessidade §1. Isto é insuficiência real, não ignorância da UI.

## 3. P6 — Estudo de referências (a necessidade é padrão de mercado)

```text
SOURCE OBSERVATION  Veeva Vault — Dynamic Access Control: regras de compartilhamento dirigidas por
                    METADADO do documento atribuem grupos automaticamente; atribuição manual existe
                    como exceção, não como modelo
SOURCE OBSERVATION  Qualio — "private tags" associadas a User Groups; documento sem tag privada é
                    visível; o mapeamento tag→grupo é administrado centralmente
SOURCE OBSERVATION  M-Files — automatic permissions dirigidas por propriedade/classe/valor de
                    metadado; named ACLs existem como opção secundária

INFERENCE           as três plataformas maduras resolvem o mesmo problema pela MESMA forma:
                    uma CLASSIFICAÇÃO governada no documento, mapeada centralmente a grupos —
                    não um seletor livre de pessoas por documento

PRODUCT DECISION    se MetalDocs vier a expressar confidencialidade, a forma correta é
                    classificação + mapeamento central, e explicitamente NÃO o ACL por documento
                    do legado
```

O legado do operador fazia o seletor de pessoas por documento; as referências mostram que esse é o caminho que as plataformas profissionais deliberadamente **não** tomam como mecanismo primário.

## 4. Alternativas de Global Maximum

### A — Área dedicada (status quo; "crie a Área Produção-Restrita")

```text
custo arquitetural   zero
```

**Falsificada por Evidence.** Três consequências duras:

```text
1. IRREVERSIBILIDADE — `area_id` só existe em createDocument (op46). Não existe operação que
   mude a Área de um Documento. Reclassificar exige obsoletar e recriar, quebrando a
   continuidade de identidade/histórico que é a razão de ser do produto.
2. IDENTIDADE MENTE — com numeração document_type_area, o code embute a Área
   (PO-PRDR-014). O código do documento passaria a carregar seu nível de sigilo, e
   `committed Document.code = unique and never reused` o torna permanente.
3. ORGANIZAÇÃO POLUÍDA — Área é conceito ORGANIZACIONAL. Duplicar Áreas por sensibilidade
   fragmenta o organograma, a numeração e a navegação para resolver um problema de segurança.
```

Portanto A não é apenas subótima: ela cria dívida irreversível se recomendada como solução.

### B — Classificação de confidencialidade governada (convergente com as referências)

```text
forma      dimensão de classificação no Documento (ou herdada do DocumentType), fechada e
           product-owned, mapeada a concessões no B11
leitura    a equação T3 ganha uma dimensão: permissão + escopo + CLASSE
reclassif. operação governada e auditável, sem tocar code nem identidade
custo      reopen material do T3 (equação, predicados, read models), com impacto em
           B02 Biblioteca, B03 Oficial, B09 Auditoria, B11 Acesso e no censo de operações
```

### C — ACL por documento (o que o legado fazia)

```text
REJECTED — contradiz a lista negativa explícita do T3 (materialized ACL), destrói a
resposta auditável "quem podia ver o quê, e quando", produz deriva de permissão e
aproxima o produto do file drive genérico excluído pelo North Star §1.
As três referências o mantêm como exceção, nunca como modelo.
```

## 5. Disposição proposta ao operador

**UPSTREAM FINDING — REAL / MATERIAL, sem consumidor V1 nomeado.**

Proposta: **DEFERRED com forma nomeada**, não reopen agora.

```text
POR QUE NÃO REABRIR AGORA
  nenhum consumidor V1 nomeado (o próprio operador afirmou não ser escopo V1)
  reabrir o T3 hoje atinge a equação de autorização, 13 blocos LOCKED, o censo ratificado
  e a Biblioteca/Busca/Auditoria — custo alto contra necessidade ainda não datada

POR QUE REGISTRAR AGORA, E NÃO DEPOIS
  1. impede a armadilha: sem registro, alguém "resolve" com Área dedicada e cria
     documentos com sigilo no código, irreversíveis (§4-A)
  2. preserva a forma provada pelas referências, para que o reopen futuro não seja chute
  3. B13 fica honesto: a tela de criação não finge controle de visibilidade que não existe

CUSTO CONHECIDO DO ADIAMENTO
  quando a classificação entrar, a tela de criação ganha um campo e as projeções de leitura
  mudam — retrabalho bounded e barato, contra ripple global agora
```

Obrigação a registrar em `docs/decisions/forward-obligations.md`:

```text
SEC-XX — DEFERRED — Confidencialidade dentro da Área exige uma CLASSIFICAÇÃO governada no
Documento mapeada centralmente a concessões; nunca ACL por documento e nunca Área usada como
mecanismo de sigilo (Área é imutável por documento e o code a embute).
Gatilho de reabertura: um consumidor real nomeado precisar publicar documento sensível dentro
de uma Área compartilhada.
```

## 6. Efeito imediato no B13 (independe da adjudicação)

A escolha de Área **já é** a decisão de acesso e a tela não conta isso ao humano. Correção dentro da autoridade atual, sem reopen:

```text
declarar a REGRA de leitura resultante ("leitores com permissão nesta Área ou na Empresa
poderão ler quando o documento ficar efetivo") — regra, não enumeração de pessoas:
enumerar seria autorização calculada no frontend, proibida pelo T3 e por frontend.md
```

## 7. Fontes de referência (Evidence, nunca autoridade)

```text
Veeva Vault — About Dynamic Access Control for Documents
  https://platform.veevavault.help/en/lr/31824/
Veeva Vault — Using Document Sharing Settings
  https://platform.veevavault.help/en/lr/895/
Qualio — Create and Manage User Groups (private tags)
  https://docs.qualio.com/en/articles/6526547-create-and-manage-user-groups
Qualio — User Permissions
  https://docs.qualio.com/en/articles/6526420-user-permissions
M-Files — Automatic Permissions for Value List Items
  https://userguide.m-files.com/user-guide/latest/eng/Automatic_permissions.html
M-Files — Permissions and Automatic Permissions
  https://userguide.m-files.com/user-guide/latest/eng/Permissions_and_Default_permissions.html
```
