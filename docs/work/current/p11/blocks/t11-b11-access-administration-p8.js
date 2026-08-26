(() => {
  "use strict";

  const $ = (selector) => document.querySelector(selector);
  const $$ = (selector) => [...document.querySelectorAll(selector)];

  const PAGE_SIZE = 3;
  const USER_PAGE_SIZE = 4;

  const USERS = [
    { id: "usr-001", name: "João Almeida", eligibility: "ENABLED" },
    { id: "usr-002", name: "Beatriz Silva", eligibility: "ENABLED" },
    { id: "usr-003", name: "Rafael Siqueira", eligibility: "ENABLED" },
    { id: "usr-004", name: "Ana Torres", eligibility: "ENABLED" },
    { id: "usr-005", name: "Bruno Vieira", eligibility: "ENABLED" },
    { id: "usr-006", name: "Carla Nunes", eligibility: "ENABLED" },
    { id: "usr-007", name: "Paulo Mendes", eligibility: "DISABLED" },
    { id: "usr-008", name: "Sofia Barros", eligibility: "ENABLED" },
    { id: "usr-009", name: "Luciana Prado", eligibility: "ENABLED" },
    { id: "usr-010", name: "Diego Ramos", eligibility: "ENABLED" },
    { id: "usr-011", name: "Mariana Costa", eligibility: "ENABLED" },
    { id: "usr-012", name: "Felipe Moraes", eligibility: "ENABLED" }
  ];

  const AREAS = [
    { id: "area-001", code: "ADM", name: "Administração" },
    { id: "area-002", code: "COM", name: "Comercial" },
    { id: "area-003", code: "FIN", name: "Financeiro" },
    { id: "area-004", code: "LOG", name: "Logística" },
    { id: "area-005", code: "QLD", name: "Qualidade" },
    { id: "area-006", code: "RH", name: "Recursos Humanos" },
    { id: "area-007", code: "SEG", name: "Segurança" },
    { id: "area-008", code: "TI", name: "Tecnologia" }
  ];

  const GROUPS = [
    { id: "grp-001", name: "Aprovadores Financeiro" },
    { id: "grp-002", name: "Equipe Comercial" },
    { id: "grp-003", name: "Diretoria" },
    { id: "grp-004", name: "Gestores da Qualidade" },
    { id: "grp-005", name: "Recursos Humanos" },
    { id: "grp-006", name: "Tecnologia" },
    { id: "grp-007", name: "Logística" },
    { id: "grp-008", name: "Segurança do Trabalho" }
  ];

  const ROLES = [
    {
      code: "governance_admin",
      label: "Administrador de governança",
      scopes: ["company"],
      permissions: ["organization.manage", "access.manage", "document_type.manage", "template_use.manage"]
    },
    {
      code: "area_manager",
      label: "Gestor de Área",
      scopes: ["area"],
      permissions: ["document.read_effective", "document.read_history", "document.read_working", "document.create", "document.edit", "document.submit", "document.cancel_revision", "document.obsolete", "document.owner.manage", "governance.act"]
    },
    {
      code: "author",
      label: "Autor",
      scopes: ["company", "area"],
      permissions: ["document.read_effective", "document.read_history", "document.read_working", "document.create", "document.edit", "document.submit"]
    },
    {
      code: "approver",
      label: "Aprovador",
      scopes: ["company", "area"],
      permissions: ["document.read_effective", "governance.act"]
    },
    {
      code: "viewer",
      label: "Visualizador",
      scopes: ["company", "area"],
      permissions: ["document.read_effective"]
    },
    {
      code: "governance_viewer",
      label: "Visualizador de governança",
      scopes: ["company", "area"],
      permissions: ["document.read_effective", "document.read_history", "audit.read"]
    }
  ];

  const BASE_ASSIGNMENTS = [
    { id: "asg-001", subjectKind: "group", subjectId: "grp-001", role: "approver", scopeKind: "area", scopeId: "area-003" },
    { id: "asg-002", subjectKind: "group", subjectId: "grp-001", role: "viewer", scopeKind: "area", scopeId: "area-002" },
    { id: "asg-003", subjectKind: "group", subjectId: "grp-001", role: "viewer", scopeKind: "company", scopeId: null },
    { id: "asg-004", subjectKind: "user", subjectId: "usr-011", role: "area_manager", scopeKind: "area", scopeId: "area-002" },
    { id: "asg-005", subjectKind: "group", subjectId: "grp-002", role: "author", scopeKind: "area", scopeId: "area-002" },
    { id: "asg-006", subjectKind: "user", subjectId: "usr-003", role: "viewer", scopeKind: "area", scopeId: "area-002" },
    { id: "asg-007", subjectKind: "group", subjectId: "grp-003", role: "governance_admin", scopeKind: "company", scopeId: null },
    { id: "asg-008", subjectKind: "user", subjectId: "usr-001", role: "author", scopeKind: "company", scopeId: null },
    { id: "asg-009", subjectKind: "group", subjectId: "grp-004", role: "governance_viewer", scopeKind: "company", scopeId: null },
    { id: "asg-010", subjectKind: "user", subjectId: "usr-002", role: "approver", scopeKind: "area", scopeId: "area-005" },
    { id: "asg-011", subjectKind: "group", subjectId: "grp-006", role: "viewer", scopeKind: "area", scopeId: "area-008" },
    { id: "asg-012", subjectKind: "user", subjectId: "usr-009", role: "author", scopeKind: "area", scopeId: "area-004" },
    { id: "asg-013", subjectKind: "group", subjectId: "grp-001", role: "governance_viewer", scopeKind: "area", scopeId: "area-005" },
    { id: "asg-014", subjectKind: "group", subjectId: "grp-001", role: "author", scopeKind: "area", scopeId: "area-004" }
  ];

  const BASE_MEMBERS = {
    "grp-001": ["usr-001", "usr-002", "usr-003", "usr-004", "usr-005", "usr-006", "usr-008"],
    "grp-002": ["usr-011", "usr-003"],
    "grp-003": ["usr-001"],
    "grp-004": ["usr-002"],
    "grp-005": ["usr-011"],
    "grp-006": ["usr-003"],
    "grp-007": ["usr-009"],
    "grp-008": ["usr-010"]
  };

  let assignments = [];
  let members = {};
  let disclosedGroups = [];
  let idempotencyStore = new Map();
  let nextAssignmentNumber = 100;
  let nextKeyNumber = 1;
  let semanticMutations = { membership: 0, grant: 0 };

  const flags = {
    pageFail: false,
    grantConflict: false,
    grantAmbiguous: false,
    memberConflict: false
  };

  const state = {
    lens: "area",
    areaId: "area-002",
    groupId: "grp-001",
    roleCode: "approver",
    revokeId: null,
    loadedMemberIds: new Set(),
    pages: {
      areas: 0,
      areaSpecific: 0,
      companyAssignments: 0,
      groups: 0,
      groupAccess: 0,
      members: 0,
      roleAssignments: 0,
      memberUsers: 0,
      grantSubjects: 0,
      grantAreas: 0
    },
    member: {
      userId: null,
      terminal: false
    },
    grant: {
      context: "general",
      contextSubjectId: null,
      contextAreaId: null,
      subjectKind: "user",
      subjectId: null,
      roleCode: "viewer",
      scopeKind: "company",
      areaId: null,
      key: null,
      fingerprint: null,
      pendingAmbiguous: false,
      terminal: false
    }
  };

  function user(id) { return USERS.find((item) => item.id === id); }
  function area(id) { return AREAS.find((item) => item.id === id); }
  function group(id) { return disclosedGroups.find((item) => item.id === id); }
  function role(code) { return ROLES.find((item) => item.code === code); }

  function cloneMembers() {
    return Object.fromEntries(Object.entries(BASE_MEMBERS).map(([groupId, userIds]) => [groupId, [...userIds]]));
  }

  function pageOf(items, pageIndex, size = PAGE_SIZE) {
    const start = pageIndex * size;
    return {
      items: items.slice(start, start + size),
      hasMore: start + size < items.length,
      pageNumber: pageIndex + 1
    };
  }

  function pageText(operation, page) {
    return `${operation} · Página ${page.pageNumber} · ${page.hasMore ? "há mais" : "fim da lista"}`;
  }

  function setPager(prevSelector, nextSelector, labelSelector, operation, pageIndex, page) {
    $(prevSelector).disabled = pageIndex === 0;
    $(nextSelector).disabled = !page.hasMore;
    $(labelSelector).textContent = pageText(operation, page);
  }

  function setState(selector, message = "", kind = "") {
    const element = $(selector);
    element.className = `state-line${kind ? ` ${kind}` : ""}`;
    element.textContent = message;
  }

  function setFlag(buttonSelector, flagName, value) {
    flags[flagName] = value;
    $(buttonSelector).setAttribute("aria-pressed", String(value));
  }

  function updateFixtureLog(message) {
    $("#fixture-log").textContent = `${message} Mutations: membership ${semanticMutations.membership} · grant ${semanticMutations.grant}.`;
  }

  function consumePageFailure(operation, stateSelector) {
    if (!flags.pageFail) return false;
    setFlag("#fx-page-fail", "pageFail", false);
    setState(stateSelector, `Falha na continuação de ${operation}. A página carregada permanece visível e não é apresentada como coleção completa.`, "error");
    updateFixtureLog(`Continuação de ${operation} falhou sem apagar a página atual.`);
    return true;
  }

  function nextPage(pageKey, operation, stateSelector, render) {
    if (consumePageFailure(operation, stateSelector)) return;
    state.pages[pageKey] += 1;
    setState(stateSelector);
    render();
  }

  function previousPage(pageKey, stateSelector, render) {
    if (state.pages[pageKey] === 0) return;
    state.pages[pageKey] -= 1;
    setState(stateSelector);
    render();
  }

  function subjectLabel(assignment) {
    return assignment.subjectKind === "group"
      ? group(assignment.subjectId)?.name || "Grupo indisponível"
      : user(assignment.subjectId)?.name || "User indisponível";
  }

  function scopeLabel(assignment) {
    if (assignment.scopeKind === "company") return "Toda a empresa";
    const selectedArea = area(assignment.scopeId);
    return selectedArea ? `${selectedArea.code} · ${selectedArea.name}` : "Área indisponível";
  }

  function assignmentRows(containerSelector, assignmentItems) {
    const container = $(containerSelector);
    if (!assignmentItems.length) {
      container.innerHTML = '<div class="empty">Nenhum RoleAssignment nesta página.</div>';
      return;
    }

    container.innerHTML = assignmentItems.map((assignment) => {
      const selectedRole = role(assignment.role);
      return `
        <div class="grant-row">
          <div><b>${subjectLabel(assignment)}</b><small>${assignment.subjectKind === "group" ? "Group" : "User"} · ${assignment.subjectId}</small></div>
          <div><b>${selectedRole.label}</b><small>${selectedRole.code}</small></div>
          <div><b>${scopeLabel(assignment)}</b><small>${assignment.scopeKind === "company" ? "Company scope" : "Area scope"}</small></div>
          <button class="row-action" type="button" data-revoke-id="${assignment.id}">Revogar</button>
        </div>`;
    }).join("");

    $$(`${containerSelector} [data-revoke-id]`).forEach((button) => {
      button.addEventListener("click", () => openRevoke(button.dataset.revokeId));
    });
  }

  function fixtureServerAssignmentPage(filter, pageIndex) {
    // P8 fixture server: op31 filters execute before pagination. The Product
    // consumer receives only this returned page; it never crawls all pages.
    const filtered = [...assignments].filter(filter).sort((a, b) => a.id.localeCompare(b.id));
    return pageOf(filtered, pageIndex);
  }

  function pagedAssignments(config) {
    const page = fixtureServerAssignmentPage(config.filter, state.pages[config.pageKey]);
    assignmentRows(config.container, page.items);
    setPager(config.prev, config.next, config.label, config.operation, state.pages[config.pageKey], page);
  }

  function renderAreaList() {
    const companyActive = state.areaId === "company";
    $("#company-scope-choice").innerHTML = `
      <button class="list-button${companyActive ? " active" : ""}" type="button" data-area-id="company" aria-pressed="${companyActive}">
        <span><strong>Toda a empresa</strong><span class="meta">Company scope · sem area_id</span></span>
        <span class="pill">Company</span>
      </button>`;

    const page = pageOf(AREAS, state.pages.areas);
    $("#area-list").innerHTML = page.items.map((item) => `
      <button class="list-button${state.areaId === item.id ? " active" : ""}" type="button" data-area-id="${item.id}" aria-pressed="${state.areaId === item.id}">
        <span><strong>${item.name}</strong><span class="meta">${item.code} · ${item.id}</span></span>
        <span class="pill">Area</span>
      </button>`).join("");

    $$('[data-area-id]').forEach((button) => {
      button.addEventListener("click", () => selectArea(button.dataset.areaId));
    });

    setPager("#area-prev", "#area-next", "#area-page", "op16 AreaPage", state.pages.areas, page);
  }

  function selectArea(areaId) {
    state.areaId = areaId;
    state.pages.areaSpecific = 0;
    state.pages.companyAssignments = 0;
    setState("#area-specific-state");
    setState("#company-assignment-state");
    renderArea();
  }

  function renderArea() {
    renderAreaList();
    const companySelected = state.areaId === "company";
    const selectedArea = area(state.areaId);
    $("#selected-area-title").textContent = companySelected ? "Toda a empresa" : `${selectedArea.code} · ${selectedArea.name}`;
    $("#selected-area-kind").textContent = companySelected ? "Company" : "Area";
    $("#area-specific-section").hidden = companySelected;
    $("#company-scope-copy").textContent = companySelected
      ? "RoleAssignments cujo scope real é Company."
      : "Também se aplicam nesta Área, mas continuam sendo grants Company-wide.";

    if (!companySelected) {
      pagedAssignments({
        filter: (assignment) => assignment.scopeKind === "area" && assignment.scopeId === state.areaId,
        pageKey: "areaSpecific",
        container: "#area-specific-list",
        prev: "#area-specific-prev",
        next: "#area-specific-next",
        label: "#area-specific-page",
        operation: "op31 scope_kind=area"
      });
    }

    pagedAssignments({
      filter: (assignment) => assignment.scopeKind === "company",
      pageKey: "companyAssignments",
      container: "#company-assignment-list",
      prev: "#company-assignment-prev",
      next: "#company-assignment-next",
      label: "#company-assignment-page",
      operation: "op31 scope_kind=company"
    });
  }

  function renderGroupList() {
    const page = pageOf(disclosedGroups, state.pages.groups);
    $("#group-list").innerHTML = page.items.map((item) => `
      <button class="list-button${state.groupId === item.id ? " active" : ""}" type="button" data-group-id="${item.id}" aria-pressed="${state.groupId === item.id}">
        <span><strong>${item.name}</strong><span class="meta">${item.id}</span></span>
        <span class="pill">Group</span>
      </button>`).join("");

    $$('#group-list [data-group-id]').forEach((button) => {
      button.addEventListener("click", () => selectGroup(button.dataset.groupId));
    });

    setPager("#group-prev", "#group-next", "#group-page", "op22 GroupPage", state.pages.groups, page);
  }

  function selectGroup(groupId) {
    state.groupId = groupId;
    state.pages.groupAccess = 0;
    state.pages.members = 0;
    state.loadedMemberIds = new Set();
    setState("#group-access-state");
    setState("#member-state");
    renderGroups();
  }

  function renderGroupMembers() {
    const memberIds = members[state.groupId] || [];
    const page = pageOf(memberIds, state.pages.members);
    page.items.forEach((userId) => state.loadedMemberIds.add(userId));

    if (!page.items.length) {
      $("#member-list").innerHTML = '<div class="empty">Nenhum membro nesta página.</div>';
    } else {
      $("#member-list").innerHTML = page.items.map((userId) => {
        const selectedUser = user(userId);
        return `
          <div class="member-row">
            <div><strong>${selectedUser.name}</strong><span class="meta">${selectedUser.id} · UserReference</span></div>
            <button class="row-action" type="button" data-remove-user-id="${selectedUser.id}">Remover</button>
          </div>`;
      }).join("");

      $$('#member-list [data-remove-user-id]').forEach((button) => {
        button.addEventListener("click", () => openRemoveMember(button.dataset.removeUserId));
      });
    }

    setPager("#member-prev", "#member-next", "#member-page", "op27 GroupMemberPage", state.pages.members, page);
    if (!$("#member-state").textContent) {
      setState("#member-state", page.hasMore ? "Mais memberships existem no servidor; use Próxima." : "Fim desta travessia op27; nenhuma contagem total foi inferida.");
    }
  }

  function renderGroups() {
    renderGroupList();
    const selectedGroup = group(state.groupId);
    $("#group-grant-context").disabled = !selectedGroup;
    $("#add-member").disabled = !selectedGroup;

    if (!selectedGroup) {
      $("#selected-group-title").textContent = "Grupo indisponível";
      $("#group-access-list").innerHTML = '<div class="empty">Nenhum dado protegido é mantido para o Group não divulgável.</div>';
      $("#member-list").innerHTML = '<div class="empty">Memberships não são apresentadas após a reconciliação 404.</div>';
      $("#group-access-page").textContent = "op31 group_id · contexto reconciliado";
      $("#member-page").textContent = "op27 GroupMemberPage · contexto reconciliado";
      $("#group-access-prev").disabled = true;
      $("#group-access-next").disabled = true;
      $("#member-prev").disabled = true;
      $("#member-next").disabled = true;
      if (!$("#group-access-state").textContent) {
        setState("#group-access-state", "404 notfound.resource — Group ausente ou não divulgável; selecione outra identidade retornada por op22.", "error");
      }
      if (!$("#member-state").textContent) {
        setState("#member-state", "O detalhe anterior foi removido sem inferir a existência protegida do Group.", "error");
      }
      return;
    }

    $("#selected-group-title").textContent = selectedGroup.name;

    pagedAssignments({
      filter: (assignment) => assignment.subjectKind === "group" && assignment.subjectId === state.groupId,
      pageKey: "groupAccess",
      container: "#group-access-list",
      prev: "#group-access-prev",
      next: "#group-access-next",
      label: "#group-access-page",
      operation: "op31 group_id"
    });

    renderGroupMembers();
  }

  function renderRoleList() {
    $("#role-list").innerHTML = ROLES.map((item) => `
      <button class="list-button${state.roleCode === item.code ? " active" : ""}" type="button" data-role-code="${item.code}" aria-pressed="${state.roleCode === item.code}">
        <span><strong>${item.label}</strong><span class="meta">${item.code}</span></span>
        <span class="pill">${item.scopes.map((scope) => scope === "company" ? "Empresa" : "Área").join(" + ")}</span>
      </button>`).join("");

    $$('#role-list [data-role-code]').forEach((button) => {
      button.addEventListener("click", () => selectRole(button.dataset.roleCode));
    });
  }

  function selectRole(roleCode) {
    state.roleCode = roleCode;
    state.pages.roleAssignments = 0;
    setState("#role-assignment-state");
    renderRoles();
  }

  function renderRoles() {
    renderRoleList();
    const selectedRole = role(state.roleCode);
    $("#role-title").textContent = selectedRole.label;
    $("#role-detail").innerHTML = `
      <div class="read-only"><strong>${selectedRole.code}</strong><br>Escopos admitidos: ${selectedRole.scopes.map((scope) => scope === "company" ? "Empresa" : "Área").join(" e ")}.</div>
      <h3>Permissões retornadas pelo servidor</h3>
      <ul class="role-permissions">${selectedRole.permissions.map((permission) => `<li><code>${permission}</code></li>`).join("")}</ul>`;

    pagedAssignments({
      filter: (assignment) => assignment.role === state.roleCode,
      pageKey: "roleAssignments",
      container: "#role-assignment-list",
      prev: "#role-assignment-prev",
      next: "#role-assignment-next",
      label: "#role-assignment-page",
      operation: "op31 role"
    });
  }

  function setLens(lens) {
    state.lens = lens;
    $$("[data-lens]").forEach((button) => {
      button.setAttribute("aria-selected", String(button.dataset.lens === lens));
      button.tabIndex = button.dataset.lens === lens ? 0 : -1;
    });
    $$("[data-lens-panel]").forEach((panel) => { panel.hidden = panel.dataset.lensPanel !== lens; });
    if (lens === "area") renderArea();
    if (lens === "groups") renderGroups();
    if (lens === "roles") renderRoles();
  }

  function renderMemberUsers() {
    const page = pageOf(USERS, state.pages.memberUsers, USER_PAGE_SIZE);
    $("#member-user-list").innerHTML = page.items.map((item) => {
      const disabled = item.eligibility === "DISABLED" || state.member.terminal;
      const locallyKnown = state.loadedMemberIds.has(item.id);
      const statePill = item.eligibility === "DISABLED"
        ? "DISABLED · indisponível"
        : locallyKnown
          ? "Relação vista em op27"
          : "ENABLED";
      return `
        <button class="list-button${state.member.userId === item.id ? " active" : ""}" type="button" data-member-user-id="${item.id}" aria-pressed="${state.member.userId === item.id}" ${disabled ? "disabled" : ""}>
          <span><strong>${item.name}</strong><span class="meta">${item.id}</span></span>
          <span class="pill">${statePill}</span>
        </button>`;
    }).join("");

    $$('#member-user-list [data-member-user-id]').forEach((button) => {
      button.addEventListener("click", () => selectMemberUser(button.dataset.memberUserId));
    });

    setPager("#member-user-prev", "#member-user-next", "#member-user-page", "op6 UserPage bruta", state.pages.memberUsers, page);
    $("#member-user-prev").disabled = state.member.terminal || state.pages.memberUsers === 0;
    $("#member-user-next").disabled = state.member.terminal || !page.hasMore;
  }

  function openAddMember() {
    state.pages.memberUsers = 0;
    state.member.userId = null;
    state.member.terminal = false;
    $("#member-review").hidden = true;
    $("#member-result").hidden = true;
    $("#member-confirm").hidden = true;
    setState("#member-user-state", "Limites de página op6 preservados; nenhuma membership foi pré-filtrada.");
    renderMemberUsers();
    $("#member-dialog").showModal();
  }

  function selectMemberUser(userId) {
    const selectedUser = user(userId);
    if (!selectedUser || selectedUser.eligibility !== "ENABLED" || state.member.terminal) return;
    state.member.userId = userId;
    const locallyKnown = state.loadedMemberIds.has(userId);
    $("#member-review-copy").innerHTML = `<strong>${selectedUser.name}</strong> → <strong>${group(state.groupId).name}</strong><br>${locallyKnown
      ? "A relação aparece em uma página op27 já carregada; o PUT ainda reconcilia a verdade atual."
      : "As páginas op27 carregadas não provam se a relação já existe; o PUT reconciliará 201 versus 204."}`;
    $("#member-review").hidden = false;
    $("#member-confirm").hidden = false;
    renderMemberUsers();
  }

  function confirmMember() {
    if (!state.member.userId || state.member.terminal) return;
    if (flags.memberConflict) {
      setFlag("#fx-member-conflict", "memberConflict", false);
      setState("#member-user-state", "409 state.conflict — offboarding venceu a serialização; nenhuma membership foi criada.", "error");
      updateFixtureLog("op28 terminou em conflito e preservou a seleção para revisão.");
      return;
    }

    const groupMembers = members[state.groupId] || (members[state.groupId] = []);
    const existed = groupMembers.includes(state.member.userId);
    if (!existed) {
      groupMembers.push(state.member.userId);
      groupMembers.sort();
      semanticMutations.membership += 1;
      state.loadedMemberIds.add(state.member.userId);
    }

    state.member.terminal = true;
    $("#member-confirm").hidden = true;
    $("#member-review").hidden = true;
    $("#member-result").hidden = false;
    $("#member-result").innerHTML = existed
      ? `<strong>204 · a relação já existia</strong>Nenhuma segunda membership foi criada. O resultado foi reconciliado sem conhecimento completo prévio.`
      : `<strong>201 · membership criada</strong>${user(state.member.userId).name} agora participa de ${group(state.groupId).name}.`;
    setState("#member-user-state", existed ? "op28 confirmou relação existente." : "op28 criou a primeira relação.", "success");
    renderMemberUsers();
    renderGroups();
    updateFixtureLog(existed ? "op28 retornou 204 e zero nova mutação." : "op28 retornou 201 e criou uma membership.");
  }

  function openRemoveMember(userId) {
    state.member.userId = userId;
    $("#remove-member-copy").innerHTML = `A membership de <strong>${user(userId).name}</strong> será removida de <strong>${group(state.groupId).name}</strong>.<br>Os acessos derivados deste Grupo deixarão de se aplicar. Grants diretos ou por outros Grupos podem continuar válidos.`;
    $("#remove-member-dialog").showModal();
  }

  function confirmRemoveMember() {
    const groupMembers = members[state.groupId] || [];
    const before = groupMembers.length;
    members[state.groupId] = groupMembers.filter((userId) => userId !== state.member.userId);
    if (members[state.groupId].length !== before) semanticMutations.membership += 1;
    const lastPage = Math.max(0, Math.ceil(members[state.groupId].length / PAGE_SIZE) - 1);
    state.pages.members = Math.min(state.pages.members, lastPage);
    $("#remove-member-dialog").close();
    renderGroups();
    updateFixtureLog("op29 removeu a relação exata; nenhum acesso alternativo foi inferido.");
  }

  function newGrantKey() {
    const suffix = String(nextKeyNumber++).padStart(12, "0");
    return `00000000-0000-4000-8000-${suffix}`;
  }

  function grantSubjectSource() {
    return state.grant.subjectKind === "group" ? disclosedGroups : USERS;
  }

  function invalidateGrantReview() {
    if (state.grant.terminal) return;
    state.grant.key = null;
    state.grant.fingerprint = null;
    $("#grant-review-box").hidden = true;
    $("#grant-confirm").hidden = true;
    $("#grant-ambiguous").hidden = true;
    $("#grant-result").hidden = true;
    setState("#grant-command-state");
  }

  function setGrantTerminal(terminal) {
    state.grant.terminal = terminal;
    $("#grant-subject-kind").disabled = terminal;
    $("#grant-role").disabled = terminal;
    $("#grant-scope-kind").disabled = terminal;
    $("#grant-review").disabled = terminal;
    renderGrantSubjectPicker();
    renderGrantAreaPicker();
  }

  function renderGrantSubjectPicker() {
    const source = grantSubjectSource();
    const size = state.grant.subjectKind === "user" ? USER_PAGE_SIZE : PAGE_SIZE;
    const operation = state.grant.subjectKind === "user" ? "op6 UserPage bruta" : "op22 GroupPage";
    const page = pageOf(source, state.pages.grantSubjects, size);
    const contextual = state.grant.contextSubjectId ? (state.grant.subjectKind === "group" ? group(state.grant.contextSubjectId) : user(state.grant.contextSubjectId)) : null;

    $("#grant-subject-context").hidden = !contextual;
    if (contextual) {
      $("#grant-subject-context").innerHTML = `<span><strong>Preseleção da lente:</strong> ${contextual.name}</span><span class="pill">identidade já conhecida</span>`;
    }

    $("#grant-subject-list").innerHTML = page.items.map((item) => {
      const disabledUser = state.grant.subjectKind === "user" && item.eligibility === "DISABLED";
      const disabled = disabledUser || state.grant.terminal;
      return `
        <button class="list-button${state.grant.subjectId === item.id ? " active" : ""}" type="button" data-grant-subject-id="${item.id}" aria-pressed="${state.grant.subjectId === item.id}" ${disabled ? "disabled" : ""}>
          <span><strong>${item.name}</strong><span class="meta">${item.id}</span></span>
          <span class="pill">${disabledUser ? "DISABLED · indisponível" : state.grant.subjectKind === "group" ? "Group" : "ENABLED"}</span>
        </button>`;
    }).join("");

    $$('#grant-subject-list [data-grant-subject-id]').forEach((button) => {
      button.addEventListener("click", () => {
        if (state.grant.terminal) return;
        state.grant.subjectId = button.dataset.grantSubjectId;
        state.grant.contextSubjectId = null;
        invalidateGrantReview();
        renderGrantSubjectPicker();
      });
    });

    setPager("#grant-subject-prev", "#grant-subject-next", "#grant-subject-page", operation, state.pages.grantSubjects, page);
    $("#grant-subject-prev").disabled = state.grant.terminal || state.pages.grantSubjects === 0;
    $("#grant-subject-next").disabled = state.grant.terminal || !page.hasMore;
  }

  function renderGrantAreaPicker() {
    const isArea = state.grant.scopeKind === "area";
    $("#grant-area-picker").hidden = !isArea;
    if (!isArea) return;

    const page = pageOf(AREAS, state.pages.grantAreas);
    const contextual = state.grant.contextAreaId ? area(state.grant.contextAreaId) : null;
    $("#grant-area-context").hidden = !contextual;
    if (contextual) {
      $("#grant-area-context").innerHTML = `<span><strong>Preseleção da lente:</strong> ${contextual.code} · ${contextual.name}</span><span class="pill">identidade já conhecida</span>`;
    }

    $("#grant-area-list").innerHTML = page.items.map((item) => `
      <button class="list-button${state.grant.areaId === item.id ? " active" : ""}" type="button" data-grant-area-id="${item.id}" aria-pressed="${state.grant.areaId === item.id}" ${state.grant.terminal ? "disabled" : ""}>
        <span><strong>${item.name}</strong><span class="meta">${item.code} · ${item.id}</span></span>
        <span class="pill">Area</span>
      </button>`).join("");

    $$('#grant-area-list [data-grant-area-id]').forEach((button) => {
      button.addEventListener("click", () => {
        if (state.grant.terminal) return;
        state.grant.areaId = button.dataset.grantAreaId;
        state.grant.contextAreaId = null;
        invalidateGrantReview();
        renderGrantAreaPicker();
      });
    });

    setPager("#grant-area-prev", "#grant-area-next", "#grant-area-page", "op16 AreaPage", state.pages.grantAreas, page);
    $("#grant-area-prev").disabled = state.grant.terminal || state.pages.grantAreas === 0;
    $("#grant-area-next").disabled = state.grant.terminal || !page.hasMore;
  }

  function renderGrantRoleMeaning() {
    const selectedRole = role(state.grant.roleCode);
    $("#grant-role-meaning").innerHTML = `<strong>${selectedRole.label}</strong><br>${selectedRole.permissions.map((permission) => `<code>${permission}</code>`).join(" · ")}`;
    [...$("#grant-scope-kind").options].forEach((option) => {
      option.disabled = !selectedRole.scopes.includes(option.value);
    });
    if (!selectedRole.scopes.includes(state.grant.scopeKind)) {
      state.grant.scopeKind = selectedRole.scopes[0];
      $("#grant-scope-kind").value = state.grant.scopeKind;
    }
    renderGrantAreaPicker();
  }

  function resetGrantState(context) {
    state.pages.grantSubjects = 0;
    state.pages.grantAreas = 0;
    state.grant = {
      context: context.kind,
      contextSubjectId: context.kind === "group" ? state.groupId : null,
      contextAreaId: context.kind === "area" && state.areaId !== "company" ? state.areaId : null,
      subjectKind: context.kind === "group" ? "group" : "user",
      subjectId: context.kind === "group" ? state.groupId : null,
      roleCode: context.kind === "area" ? (state.areaId === "company" ? "governance_admin" : "area_manager") : "viewer",
      scopeKind: context.kind === "area" ? (state.areaId === "company" ? "company" : "area") : "company",
      areaId: context.kind === "area" && state.areaId !== "company" ? state.areaId : null,
      key: null,
      fingerprint: null,
      pendingAmbiguous: false,
      terminal: false
    };
  }

  function openGrant(context = { kind: "general" }) {
    resetGrantState(context);
    $("#grant-subject-kind").value = state.grant.subjectKind;
    $("#grant-role").innerHTML = ROLES.map((item) => `<option value="${item.code}">${item.label}</option>`).join("");
    $("#grant-role").value = state.grant.roleCode;
    $("#grant-scope-kind").value = state.grant.scopeKind;
    $("#grant-review-box").hidden = true;
    $("#grant-confirm").hidden = true;
    $("#grant-ambiguous").hidden = true;
    $("#grant-result").hidden = true;
    $("#grant-close").disabled = false;
    setState("#grant-subject-state", state.grant.subjectKind === "user" ? "UserPage bruta: DISABLED permanece visível e indisponível." : "GroupPage com travessia visível.");
    setState("#grant-area-state");
    setState("#grant-command-state");
    setGrantTerminal(false);
    renderGrantRoleMeaning();
    $("#grant-dialog").showModal();
  }

  function selectedGrantSubject() {
    return state.grant.subjectKind === "group" ? group(state.grant.subjectId) : user(state.grant.subjectId);
  }

  function grantReview() {
    if (state.grant.terminal || state.grant.pendingAmbiguous) {
      setState("#grant-command-state", "O resultado anterior ainda está ambíguo. Reenvie somente o mesmo comando com a mesma chave.", "error");
      return;
    }
    const subject = selectedGrantSubject();
    const selectedRole = role(state.grant.roleCode);
    const selectedArea = state.grant.scopeKind === "area" ? area(state.grant.areaId) : null;

    if (!subject) {
      setState("#grant-command-state", "Selecione um Subject visível antes de revisar.", "error");
      return;
    }
    if (state.grant.subjectKind === "user" && subject.eligibility !== "ENABLED") {
      setState("#grant-command-state", "User DISABLED não pode formar o comando.", "error");
      return;
    }
    if (!selectedRole.scopes.includes(state.grant.scopeKind) || (state.grant.scopeKind === "area" && !selectedArea)) {
      setState("#grant-command-state", "Selecione um Scope admitido pela Função.", "error");
      return;
    }

    const command = {
      subject: { kind: state.grant.subjectKind, id: subject.id },
      role: selectedRole.code,
      scope: state.grant.scopeKind === "company" ? { kind: "company" } : { kind: "area", area_id: selectedArea.id }
    };
    state.grant.key = newGrantKey();
    state.grant.fingerprint = JSON.stringify(command);

    $("#grant-review-box").innerHTML = `
      <dl>
        <dt>Quem</dt><dd>${subject.name} · ${state.grant.subjectKind === "group" ? "Group" : "User"}</dd>
        <dt>Função</dt><dd>${selectedRole.label} · ${selectedRole.code}</dd>
        <dt>Onde</dt><dd>${state.grant.scopeKind === "company" ? "Toda a empresa" : `${selectedArea.code} · ${selectedArea.name}`}</dd>
        <dt>Permite</dt><dd>${selectedRole.permissions.join(" · ")}</dd>
        <dt>Chave</dt><dd><code>${state.grant.key}</code></dd>
      </dl>
      <p>Esta concessão é aditiva. Grants existentes não são alterados.</p>`;
    $("#grant-review-box").hidden = false;
    $("#grant-confirm").hidden = false;
    setState("#grant-command-state", "Comando lógico pronto. A chave permanece fixa em qualquer retry deste review.");
  }

  function createAssignmentFromGrant() {
    const id = `asg-${String(nextAssignmentNumber++).padStart(3, "0")}`;
    assignments.push({
      id,
      subjectKind: state.grant.subjectKind,
      subjectId: state.grant.subjectId,
      role: state.grant.roleCode,
      scopeKind: state.grant.scopeKind,
      scopeId: state.grant.scopeKind === "area" ? state.grant.areaId : null
    });
    assignments.sort((a, b) => a.id.localeCompare(b.id));
    semanticMutations.grant += 1;
    return { assignmentId: id, status: 201 };
  }

  function executeGrantCommand() {
    const existing = idempotencyStore.get(state.grant.key);
    if (existing) {
      if (existing.fingerprint !== state.grant.fingerprint) {
        return { kind: "error", message: "422 validation.idempotency_key_reused — mesma chave com fingerprint diferente." };
      }
      return { kind: "success", replay: true, record: existing };
    }

    if (flags.grantConflict) {
      setFlag("#fx-grant-conflict", "grantConflict", false);
      return { kind: "error", message: "409 state.conflict — comando rejeitado; review e chave permanecem para decisão segura." };
    }

    const created = createAssignmentFromGrant();
    const record = { ...created, fingerprint: state.grant.fingerprint };
    idempotencyStore.set(state.grant.key, record);

    if (flags.grantAmbiguous) {
      setFlag("#fx-grant-ambiguous", "grantAmbiguous", false);
      return { kind: "ambiguous", record };
    }

    return { kind: "success", replay: false, record };
  }

  function renderCurrentLens() {
    if (state.lens === "area") renderArea();
    if (state.lens === "groups") renderGroups();
    if (state.lens === "roles") renderRoles();
  }

  function showGrantSuccess(record, replay) {
    state.grant.pendingAmbiguous = false;
    $("#grant-close").disabled = false;
    $("#grant-ambiguous").hidden = true;
    $("#grant-result").hidden = false;
    $("#grant-result-title").textContent = replay ? "Replay do sucesso armazenado · zero segunda mutação" : "201 · RoleAssignment criado";
    $("#grant-result-id").textContent = record.assignmentId;
    $("#grant-result-key").textContent = state.grant.key;
    $("#grant-result-count").textContent = `Mutações semânticas de grant nesta fixture: ${semanticMutations.grant}.`;
    $("#grant-confirm").hidden = true;
    setState("#grant-command-state", replay ? "Mesmo status/identity reproduzido; AuthZ atual seria revalidada pelo servidor." : "Comando concluído; confirmação terminal neste diálogo.", "success");
    setGrantTerminal(true);
    renderCurrentLens();
    updateFixtureLog(replay ? "Replay reconhecido: assignment_id estável, zero nova mutação." : "Grant criado uma única vez.");
  }

  function confirmGrant() {
    if (!state.grant.key || state.grant.terminal) return;
    const result = executeGrantCommand();
    if (result.kind === "error") {
      setState("#grant-command-state", result.message, "error");
      updateFixtureLog("Grant não produziu mutação semântica.");
      return;
    }
    if (result.kind === "ambiguous") {
      state.grant.pendingAmbiguous = true;
      $("#grant-confirm").hidden = true;
      $("#grant-ambiguous").hidden = false;
      $("#grant-close").disabled = true;
      setGrantTerminal(true);
      setState("#grant-command-state", "A resposta se perdeu depois do commit; a mesma chave deve ser reutilizada.", "error");
      updateFixtureLog("Servidor concluiu uma mutação, mas a resposta ficou ambígua.");
      renderCurrentLens();
      return;
    }
    showGrantSuccess(result.record, result.replay);
  }

  function retryGrant() {
    if (!state.grant.key) return;
    const result = executeGrantCommand();
    if (result.kind === "success") showGrantSuccess(result.record, result.replay);
  }

  function replayGrantProof() {
    const before = semanticMutations.grant;
    const result = executeGrantCommand();
    const after = semanticMutations.grant;
    if (result.kind === "success" && result.replay && before === after) {
      showGrantSuccess(result.record, true);
      $("#grant-result-count").textContent = `Prova: contador permaneceu ${before} → ${after}; nenhuma segunda linha foi anexada.`;
    }
  }

  function openRevoke(assignmentId) {
    const assignment = assignments.find((item) => item.id === assignmentId);
    if (!assignment) return;
    state.revokeId = assignmentId;
    $("#revoke-copy").innerHTML = `
      <dl>
        <dt>assignment_id</dt><dd>${assignment.id}</dd>
        <dt>Quem</dt><dd>${subjectLabel(assignment)}</dd>
        <dt>Função</dt><dd>${role(assignment.role).label}</dd>
        <dt>Onde</dt><dd>${scopeLabel(assignment)}</dd>
      </dl>`;
    $("#revoke-dialog").showModal();
  }

  function confirmRevoke() {
    assignments = assignments.filter((assignment) => assignment.id !== state.revokeId);
    $("#revoke-dialog").close();
    renderCurrentLens();
    updateFixtureLog(`op33 revogou exatamente ${state.revokeId}; nenhum outro grant foi interpretado.`);
  }

  function closeAllDialogs() {
    ["#member-dialog", "#remove-member-dialog", "#grant-dialog", "#revoke-dialog"].forEach((selector) => {
      if ($(selector).open) $(selector).close();
    });
  }

  function setGlobalNav(open, { returnFocus = true } = {}) {
    const toggle = $("#global-nav-toggle");
    const nav = $("#global-nav");
    document.body.classList.toggle("global-nav-open", open);
    toggle.setAttribute("aria-expanded", String(open));
    toggle.setAttribute("aria-label", open ? "Fechar navegação global" : "Abrir navegação global");
    $(".main").inert = open;

    if (open) {
      (nav.querySelector("button") || nav).focus();
    } else if (returnFocus) {
      toggle.focus();
    }
  }

  function showAccessDenied() {
    closeAllDialogs();
    setGlobalNav(false, { returnFocus: false });
    $("#access-content").classList.add("is-denied");
    $("#access-denied").hidden = false;
    $("#access-denied-reload").focus();
    updateFixtureLog("403 permission.denied operado: superfície, lentes e mutações foram retiradas; denial não virou coleção vazia.");
  }

  function reconcileMissingGroup() {
    setLens("groups");
    const missingId = state.groupId;
    disclosedGroups = disclosedGroups.filter((item) => item.id !== missingId);
    const lastPage = Math.max(0, Math.ceil(disclosedGroups.length / PAGE_SIZE) - 1);
    state.pages.groups = Math.min(state.pages.groups, lastPage);
    state.pages.groupAccess = 0;
    state.pages.members = 0;
    state.groupId = null;
    state.loadedMemberIds = new Set();
    setState("#group-access-state", `404 notfound.resource para ${missingId} — ausente ou não divulgável; op22 foi reconciliada sem essa identidade.`, "error");
    setState("#member-state", "Detalhe e ações do Group anterior foram removidos; nenhuma existência protegida foi inferida.", "error");
    renderGroups();
    updateFixtureLog(`404 operado para ${missingId}: seleção, detalhes e ações foram reconciliados.`);
  }

  function reset() {
    assignments = BASE_ASSIGNMENTS.map((assignment) => ({ ...assignment }));
    members = cloneMembers();
    disclosedGroups = GROUPS.map((item) => ({ ...item }));
    idempotencyStore = new Map();
    nextAssignmentNumber = 100;
    nextKeyNumber = 1;
    semanticMutations = { membership: 0, grant: 0 };
    Object.keys(flags).forEach((flagName) => { flags[flagName] = false; });
    $$('.fixture button[aria-pressed]').forEach((button) => button.setAttribute("aria-pressed", "false"));
    Object.keys(state.pages).forEach((pageKey) => { state.pages[pageKey] = 0; });
    state.lens = "area";
    state.areaId = "area-002";
    state.groupId = "grp-001";
    state.roleCode = "approver";
    state.revokeId = null;
    state.loadedMemberIds = new Set();
    state.member = { userId: null, terminal: false };
    $$('.state-line').forEach((element) => {
      element.className = "state-line";
      element.textContent = "";
    });
    closeAllDialogs();
    setGlobalNav(false, { returnFocus: false });
    $("#access-content").classList.remove("is-denied");
    $("#access-denied").hidden = true;
    setLens("area");
    updateFixtureLog("Fixture base restaurada.");
  }

  $$(".tab").forEach((button, index, buttons) => {
    button.addEventListener("click", () => setLens(button.dataset.lens));
    button.addEventListener("keydown", (event) => {
      if (!['ArrowLeft', 'ArrowRight'].includes(event.key)) return;
      event.preventDefault();
      const direction = event.key === 'ArrowRight' ? 1 : -1;
      const next = (index + direction + buttons.length) % buttons.length;
      buttons[next].focus();
      setLens(buttons[next].dataset.lens);
    });
  });

  $("#global-nav-toggle").addEventListener("click", () => {
    const open = !document.body.classList.contains("global-nav-open");
    setGlobalNav(open);
  });

  $("#global-nav").addEventListener("keydown", (event) => {
    if (event.key !== "Tab") return;
    const focusable = [...$("#global-nav").querySelectorAll("button:not([disabled])")];
    if (!focusable.length) return;
    const first = focusable[0];
    const last = focusable[focusable.length - 1];
    if (event.shiftKey && document.activeElement === first) {
      event.preventDefault();
      last.focus();
    } else if (!event.shiftKey && document.activeElement === last) {
      event.preventDefault();
      first.focus();
    }
  });

  document.addEventListener("keydown", (event) => {
    if (event.key === "Escape" && document.body.classList.contains("global-nav-open")) {
      event.preventDefault();
      setGlobalNav(false);
    }
  });

  $("#bell").addEventListener("click", () => {
    const inbox = $("#quick-inbox");
    inbox.hidden = !inbox.hidden;
    $("#bell").setAttribute("aria-expanded", String(!inbox.hidden));
  });

  $("#area-prev").addEventListener("click", () => previousPage("areas", "#area-list-state", renderAreaList));
  $("#area-next").addEventListener("click", () => nextPage("areas", "op16 AreaPage", "#area-list-state", renderAreaList));
  $("#area-specific-prev").addEventListener("click", () => previousPage("areaSpecific", "#area-specific-state", renderArea));
  $("#area-specific-next").addEventListener("click", () => nextPage("areaSpecific", "op31 scope_kind=area", "#area-specific-state", renderArea));
  $("#company-assignment-prev").addEventListener("click", () => previousPage("companyAssignments", "#company-assignment-state", renderArea));
  $("#company-assignment-next").addEventListener("click", () => nextPage("companyAssignments", "op31 scope_kind=company", "#company-assignment-state", renderArea));

  $("#group-prev").addEventListener("click", () => previousPage("groups", "#group-list-state", renderGroupList));
  $("#group-next").addEventListener("click", () => nextPage("groups", "op22 GroupPage", "#group-list-state", renderGroupList));
  $("#group-access-prev").addEventListener("click", () => previousPage("groupAccess", "#group-access-state", renderGroups));
  $("#group-access-next").addEventListener("click", () => nextPage("groupAccess", "op31 group_id", "#group-access-state", renderGroups));
  $("#member-prev").addEventListener("click", () => previousPage("members", "#member-state", renderGroupMembers));
  $("#member-next").addEventListener("click", () => nextPage("members", "op27 GroupMemberPage", "#member-state", renderGroupMembers));

  $("#role-assignment-prev").addEventListener("click", () => previousPage("roleAssignments", "#role-assignment-state", renderRoles));
  $("#role-assignment-next").addEventListener("click", () => nextPage("roleAssignments", "op31 role", "#role-assignment-state", renderRoles));

  $("#add-member").addEventListener("click", openAddMember);
  $("#member-cancel").addEventListener("click", () => $("#member-dialog").close());
  $("#member-user-prev").addEventListener("click", () => previousPage("memberUsers", "#member-user-state", renderMemberUsers));
  $("#member-user-next").addEventListener("click", () => nextPage("memberUsers", "op6 UserPage", "#member-user-state", renderMemberUsers));
  $("#member-confirm").addEventListener("click", confirmMember);
  $("#remove-member-cancel").addEventListener("click", () => $("#remove-member-dialog").close());
  $("#remove-member-confirm").addEventListener("click", confirmRemoveMember);

  $("#grant-open").addEventListener("click", () => openGrant({ kind: "general" }));
  $("#area-grant-context").addEventListener("click", () => openGrant({ kind: "area" }));
  $("#group-grant-context").addEventListener("click", () => openGrant({ kind: "group" }));
  $("#grant-close").addEventListener("click", () => {
    if (state.grant.pendingAmbiguous) return;
    $("#grant-dialog").close();
  });
  $("#grant-dialog").addEventListener("cancel", (event) => {
    if (!state.grant.pendingAmbiguous) return;
    event.preventDefault();
    setState("#grant-command-state", "Resolva o resultado ambíguo reutilizando a mesma chave antes de fechar.", "error");
  });

  $("#grant-subject-kind").addEventListener("change", (event) => {
    state.grant.subjectKind = event.target.value;
    state.grant.subjectId = null;
    state.grant.contextSubjectId = null;
    state.pages.grantSubjects = 0;
    invalidateGrantReview();
    setState("#grant-subject-state", state.grant.subjectKind === "user" ? "UserPage bruta: DISABLED permanece visível e indisponível." : "GroupPage com travessia visível.");
    renderGrantSubjectPicker();
  });

  $("#grant-subject-prev").addEventListener("click", () => previousPage("grantSubjects", "#grant-subject-state", renderGrantSubjectPicker));
  $("#grant-subject-next").addEventListener("click", () => nextPage("grantSubjects", state.grant.subjectKind === "user" ? "op6 UserPage" : "op22 GroupPage", "#grant-subject-state", renderGrantSubjectPicker));

  $("#grant-role").addEventListener("change", (event) => {
    state.grant.roleCode = event.target.value;
    invalidateGrantReview();
    renderGrantRoleMeaning();
  });

  $("#grant-scope-kind").addEventListener("change", (event) => {
    state.grant.scopeKind = event.target.value;
    state.grant.contextAreaId = null;
    state.pages.grantAreas = 0;
    invalidateGrantReview();
    renderGrantAreaPicker();
  });

  $("#grant-area-prev").addEventListener("click", () => previousPage("grantAreas", "#grant-area-state", renderGrantAreaPicker));
  $("#grant-area-next").addEventListener("click", () => nextPage("grantAreas", "op16 AreaPage", "#grant-area-state", renderGrantAreaPicker));
  $("#grant-review").addEventListener("click", grantReview);
  $("#grant-confirm").addEventListener("click", confirmGrant);
  $("#grant-retry").addEventListener("click", retryGrant);
  $("#grant-replay-proof").addEventListener("click", replayGrantProof);

  $("#revoke-cancel").addEventListener("click", () => $("#revoke-dialog").close());
  $("#revoke-confirm").addEventListener("click", confirmRevoke);

  $("#fx-reset").addEventListener("click", reset);
  $("#access-denied-reload").addEventListener("click", reset);
  $("#fx-page-fail").addEventListener("click", () => {
    setFlag("#fx-page-fail", "pageFail", !flags.pageFail);
    updateFixtureLog("A próxima continuação op6/op16/op22/op27/op31 usará falha one-shot.");
  });
  $("#fx-grant-conflict").addEventListener("click", () => {
    setFlag("#fx-grant-conflict", "grantConflict", !flags.grantConflict);
    updateFixtureLog("O próximo grant simulará 409 antes de qualquer mutação.");
  });
  $("#fx-grant-ambiguous").addEventListener("click", () => {
    setFlag("#fx-grant-ambiguous", "grantAmbiguous", !flags.grantAmbiguous);
    updateFixtureLog("O próximo grant será commitado, mas sua resposta ficará ambígua.");
  });
  $("#fx-member-conflict").addEventListener("click", () => {
    setFlag("#fx-member-conflict", "memberConflict", !flags.memberConflict);
    updateFixtureLog("A próxima op28 simulará conflito com offboarding.");
  });
  $("#fx-403").addEventListener("click", () => {
    showAccessDenied();
  });
  $("#fx-404").addEventListener("click", () => {
    reconcileMissingGroup();
  });

  reset();
})();
