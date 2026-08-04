// Canonical PT-BR message for every backend Problem code (RFC 9457).
//
// Keyed by the closed code vocabulary of ADR 0089: ten SEMANTIC families, never
// module-named. Grouped by family here for the same reason the backend groups
// them — a reader looking for "why did this 409 happen" scans one section, not
// 141 lines.
//
// This file is hand-written on purpose: the codes are generated
// (`go run ./cmd/problem-codes-dump`), the Portuguese is not. Parity between the
// two is enforced in BOTH directions by
// src/lib/api/__tests__/errorMessages.coverage.test.ts:
//   - a backend code with no string here fails the frontend build;
//   - a string here for a code the backend no longer emits fails it too.
// So a new backend error cannot reach a user as a raw code, and a deleted one
// cannot leave a dead translation behind.
//
// Keep entries alphabetical within their family block.

export const errorMessages: Record<string, string> = {

  // ─── `request.` — 400 — the request itself is malformed: syntax, shape, headers, method.
  // ──────────────────────────────────────────────────────────────────────
  'request.body_too_large': 'O corpo da requisição excede o tamanho permitido.',
  'request.content_type_unsupported': 'Tipo de conteúdo inválido. Use application/json.',
  'request.cursor_expired': 'O cursor de paginação expirou. Recomece a listagem.',
  'request.cursor_invalid': 'O cursor informado é inválido ou expirou.',
  'request.empty_body': 'O corpo da requisição não pode estar vazio.',
  'request.header_required': 'Um cabeçalho obrigatório não foi enviado.',
  'request.idempotency_key_invalid': 'Identificador de operação inválido. Recarregue e tente novamente.',
  'request.idempotency_key_required': 'Identificador de operação ausente. Recarregue e tente novamente.',
  'request.if_match_malformed': 'A revisão informada é inválida.',
  'request.invalid': 'Há campos inválidos. Revise os dados informados.',
  'request.json_decode': 'Não foi possível ler o JSON enviado.',
  'request.json_type_error': 'O corpo da requisição deve ser um objeto JSON.',
  'request.method_not_allowed': 'Operação não suportada para este recurso.',
  'request.param_format': 'Um parâmetro foi enviado com formato inválido.',
  'request.param_required': 'Um parâmetro obrigatório não foi enviado.',
  'request.param_too_many_values': 'Um parâmetro foi enviado com valores demais.',
  'request.param_unmarshal': 'Um parâmetro não pôde ser interpretado.',

  // ─── `validation.` — 422 — the request is well-formed but its content is semantically wrong.
  // ──────────────────────────────────────────────────────────────────────
  'validation.approval_stage_required': 'Configure ao menos uma etapa de aprovação para esta rota.',
  'validation.area_code_immutable': 'O código da área não pode ser alterado após a criação.',
  'validation.area_parent_cycle': 'A área informada como superior criaria um ciclo na hierarquia.',
  'validation.delegation_window_invalid': 'A janela de vigência da delegação é inválida.',
  'validation.dictionary_token_missing': 'Um token de dicionário referenciado não existe.',
  'validation.doc_type_code_required': 'Informe o tipo de documento.',
  'validation.document_subject_key_mismatch': 'A rota informada não corresponde a este documento.',
  'validation.effective_date_required': 'Informe a data de vigência para concluir a publicação.',
  'validation.effective_to_not_after_effective_from': 'A data de fim de vigência deve ser posterior à de início.',
  'validation.empty_eligible_pool': 'Nenhum usuário elegível para uma das etapas da rota. Ajuste a rota ou os vínculos de área antes de submeter.',
  'validation.failed': 'A validação dos dados informados falhou.',
  'validation.family_unknown': 'Família de processos não encontrada.',
  'validation.field_immutable': 'Este campo não pode ser alterado.',
  'validation.manual_code_reason_required': 'Informe a justificativa para o código manual.',
  'validation.name_reserved': 'Este nome é reservado e não pode ser usado.',
  'validation.override_reason_required': 'Informe a justificativa da substituição.',
  'validation.placeholder_name_duplicate': 'Há campos com o mesmo nome no modelo. Os nomes devem ser únicos.',
  'validation.placeholder_name_invalid': 'O nome de um campo do modelo é inválido ou reservado.',
  'validation.placeholder_not_choice': 'Este campo não é uma lista de opções.',
  'validation.profile_code_immutable': 'O código do perfil não pode ser alterado.',
  'validation.profile_not_configured': 'Perfil de documento não configurado.',
  'validation.profile_unknown': 'O perfil informado não está cadastrado para este tenant. Verifique o código do perfil.',
  'validation.reason_category_invalid': 'A categoria do motivo informada é inválida.',
  'validation.reason_for_change_required': 'Informe o motivo da alteração para submeter.',
  'validation.reason_required': 'Informe o motivo para continuar.',
  'validation.review_due_before_effective': 'A data de revisão não pode ser anterior ao início da vigência.',
  'validation.revision_title_required': 'Informe o título da revisão para submeter.',
  'validation.role_unknown': 'Papel informado é desconhecido.',
  'validation.route_stage_required': 'Esta rota de aprovação exige ao menos uma etapa.',
  'validation.route_stages_not_permitted': 'Uma rota livre não pode declarar etapas de aprovação.',
  'validation.self_delegation': 'Não é possível delegar para si mesmo.',
  'validation.submit_choice_constraint_violated': 'Um dos aprovadores escolhidos não atende às restrições da etapa.',
  'validation.submit_choice_required': 'Escolha os aprovadores das etapas que exigem seleção no envio.',
  'validation.supersede_target_invalid': 'A versão alvo de substituição é inválida.',
  'validation.template_profile_mismatch': 'A versão do template não corresponde ao perfil do documento.',
  'validation.template_subject_key_mismatch': 'A chave do assunto deve ser igual ao código do perfil.',
  'validation.verdict_ready_on_approval_stage': 'Esta etapa é de aprovação, não de revisão — registre a assinatura em vez de um parecer.',
  'validation.visibility_scope_invalid': 'O escopo de visibilidade informado é inválido.',

  // ─── `auth.` — 401 — who you are is not established (or is established and refused).
  // ──────────────────────────────────────────────────────────────────────
  'auth.account_inactive': 'Esta conta está inativa.',
  'auth.account_locked': 'Conta temporariamente bloqueada por excesso de tentativas.',
  'auth.invalid_credentials': 'Usuário ou senha inválidos.',
  'auth.password_change_required': 'É necessário alterar a senha antes de continuar.',
  'auth.signature_invalid': 'A assinatura informada é inválida.',
  'auth.tenant_forbidden': 'Você não tem acesso ao tenant solicitado.',
  'auth.tenant_required': 'Selecione um tenant para continuar.',
  'auth.unauthenticated': 'Você precisa se autenticar para continuar.',

  // ─── `permission.` — 403 — who you are is established; you may not do this.
  // ──────────────────────────────────────────────────────────────────────
  'permission.capability_denied': 'Você não tem permissão para executar esta ação.',
  'permission.denied': 'Você não tem permissão para executar esta ação.',
  'permission.iso_segregation_violation': 'Você não pode atuar nesta etapa porque participou de uma etapa anterior do mesmo fluxo.',
  'permission.origin_forbidden': 'Origem da requisição não autorizada.',
  'permission.signoff_actor_not_eligible': 'Você não está elegível para assinar esta etapa.',
  'permission.sod_cross_stage_duplicate': 'Este usuário já assinou em outra etapa e não pode assinar novamente.',
  'permission.sod_submitter_cannot_sign': 'Você submeteu este documento e não pode aprová-lo. Outro usuário precisa assinar.',

  // ─── `notfound.` — 404 — the addressed resource does not exist for this tenant.
  // ──────────────────────────────────────────────────────────────────────
  'notfound.active_document_instance': 'Não há instância ativa para este documento controlado.',
  'notfound.approval_instance': 'Documento não encontrado.',
  'notfound.approval_instance_not_visible': 'Documento não encontrado.',
  'notfound.approval_route': 'Nenhuma rota de aprovação configurada para este perfil. Configure uma rota antes de continuar.',
  'notfound.controlled_document': 'Documento controlado não encontrado.',
  'notfound.delegation': 'Delegação não encontrada.',
  'notfound.document': 'Documento não encontrado.',
  'notfound.document_family': 'Família de processos não encontrada.',
  'notfound.document_profile': 'Perfil não encontrado.',
  'notfound.membership': 'Vínculo de usuário não encontrado.',
  'notfound.process_area': 'Área de processo não encontrada.',
  'notfound.resource': 'O recurso solicitado não foi encontrado.',
  'notfound.revision': 'Revisão do documento não encontrada.',

  // ─── `state.` — 409 — the resource exists but its current state forbids the operation.
  // ──────────────────────────────────────────────────────────────────────
  'state.active_revision_exists': 'Já existe uma revisão ativa para este documento.',
  'state.approval_blocked_unresolved_comments': 'Resolva os comentários pendentes antes de aprovar ou liberar este documento.',
  'state.approval_instance_completed': 'Este fluxo já foi concluído.',
  'state.approval_route_in_use': 'Esta rota está em uso e não pode ser alterada.',
  'state.approval_route_inactive': 'Esta rota está inativa e não pode ser usada.',
  'state.approval_route_missing': 'Perfil sem rota de aprovação ativa — configure a rota antes de submeter.',
  'state.approval_stage_not_active': 'Esta etapa de aprovação não está ativa no momento.',
  'state.controlled_document_not_active': 'O documento controlado não está ativo.',
  'state.default_template_obsolete': 'O template padrão do perfil está obsoleto.',
  'state.document_family_already_inactive': 'Esta família de documentos já está inativa.',
  'state.document_family_has_profiles': 'Não é possível remover: esta família ainda possui perfis vinculados.',
  'state.document_not_draft': 'O documento não está em rascunho.',
  'state.document_not_published': 'O documento não está publicado.',
  'state.document_profile_archived': 'O perfil está arquivado.',
  'state.fast_forward_not_eligible': 'Você não está elegível para aprovar a próxima etapa deste documento.',
  'state.fast_forward_stage_not_completed': 'A etapa atual ainda não foi concluída; não é possível aprovar a próxima etapa agora.',
  'state.override_template_deleted': 'O template de substituição foi excluído.',
  'state.override_template_not_published': 'O template de substituição não está publicado.',
  'state.placeholder_not_author_editable': 'Este campo não pode ser editado pelo autor.',
  'state.process_area_archived': 'A área de processo está arquivada.',
  'state.profile_class_route_conflict': 'A reclassificação conflita com uma rota de aprovação ativa deste perfil.',
  'state.profile_no_default_template': 'O perfil não possui template padrão definido.',
  'state.revision_not_draft': 'A revisão não está em rascunho e não pode ser alterada.',
  'state.system_template_immutable': 'Templates do sistema não podem ser alterados.',
  'state.template_artifact_missing': 'Não foi possível criar o documento porque o template base não está disponível no momento.',
  'state.template_not_published': 'A versão do template não está publicada.',
  'state.transition_invalid': 'A transição de estado solicitada não é permitida.',
  'state.upload_expired': 'O upload expirou. Tente enviar o arquivo novamente.',
  'state.upload_missing': 'O arquivo enviado não foi encontrado. Faça o upload novamente.',

  // ─── `precondition.` — 412 — a precondition the caller supplied does not hold.
  // ──────────────────────────────────────────────────────────────────────
  'precondition.content_hash_mismatch': 'O conteúdo enviado não corresponde à versão esperada.',
  'precondition.if_match_required': 'A revisão do documento é obrigatória para esta operação.',
  'precondition.lock_version_stale': 'Esta versão foi alterada por outra pessoa. Recarregue antes de salvar.',

  // ─── `conflict.` — 409 — the write collides with state that already exists.
  // ──────────────────────────────────────────────────────────────────────
  'conflict.already_exists': 'O recurso informado já existe.',
  'conflict.approval_route_duplicate_profile': 'Já existe uma rota para este perfil.',
  'conflict.concurrent_modification': 'O recurso foi alterado por outro usuário. Recarregue para continuar.',
  'conflict.controlled_document_code_archived': 'Não é possível reutilizar o código de um documento arquivado.',
  'conflict.controlled_document_code_taken': 'Este código de documento já está em uso.',
  'conflict.document_family_exists': 'Já existe uma família de documentos com este código.',
  'conflict.document_profile_exists': 'Já existe um perfil com este código.',
  'conflict.duplicate_submission': 'Esta submissão já foi registrada.',
  'conflict.generic': 'A operação conflita com o estado atual do recurso.',
  'conflict.idempotency_key_reused': 'A chave de idempotência já foi usada com outro conteúdo.',
  'conflict.mark_reviewed_stale_revision': 'A revisão mudou; recarregue antes de marcar como revisado.',
  'conflict.membership_exists': 'Já existe um vínculo ativo para este usuário nesta área.',
  'conflict.process_area_exists': 'Já existe uma área de processo com este código.',
  'conflict.signoff_duplicate': 'Esta aprovação já foi registrada.',
  'conflict.stale_base': 'A base da versão mudou. Recarregue e tente novamente.',
  'conflict.stale_revision': 'O documento foi alterado por outro usuário. Atualize e tente novamente.',

  // ─── `ratelimit.` — 429 — too many requests.
  // ──────────────────────────────────────────────────────────────────────
  'ratelimit.exceeded': 'Muitas requisições em sequência. Tente novamente em instantes.',

  // ─── `internal.` — 500 — the server failed.
  // ──────────────────────────────────────────────────────────────────────
  'internal.creation_context_unconfigured': 'O contexto de criação de documentos não está configurado. Contate o administrador.',
  'internal.db_privilege_missing': 'O servidor não tem privilégio suficiente para concluir esta operação.',
  'internal.db_unknown': 'Falha inesperada no banco de dados. Tente novamente.',
  'internal.not_implemented': 'Funcionalidade ainda não disponível.',
  'internal.signature_misconfigured': 'Falha de configuração na verificação de assinatura. Contate o administrador.',
  'internal.template_artifact_invariant_unconfigured': 'A configuração do artefato do template está incompleta. Contate o administrador.',
  'internal.unknown': 'Ocorreu um erro interno. Tente novamente.',
  'internal.upstream_timeout': 'A operação demorou mais que o esperado. Tente novamente.',
  'internal.verdict_wrong_stage_kind': 'Ocorreu um erro interno ao registrar o parecer. Tente novamente.',

  // ─── Frontend-synthetic ─────────────────────────────────────────────────
  // Dispatched by lib/api/client.ts when the session lapses; it is never a
  // backend Problem.code, so the coverage test allow-lists it explicitly.
  'authn.expired': 'Sua sessão expirou. Faça login novamente.',
};
