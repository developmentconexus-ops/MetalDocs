(() => {
  "use strict";

  const $ = (selector) => document.querySelector(selector);

  const PAGE_SIZE = 4;
  const PICKER_PAGE_SIZE = 4;

  const USERS = [
    { id: "usr-001", name: "João Almeida", eligibility: "ENABLED" },
    { id: "usr-002", name: "Beatriz Silva", eligibility: "ENABLED" },
    { id: "usr-003", name: "Rafael Siqueira", eligibility: "ENABLED" },
    { id: "usr-004", name: "Ana Torres", eligibility: "ENABLED" },
    { id: "usr-005", name: "Bruno Vieira", eligibility: "ENABLED" },
    { id: "usr-006", name: "Carla Nunes", eligibility: "ENABLED" },
    { id: "usr-007", name: "Paulo Mendes", eligibility: "DISABLED" },
    { id: "usr-008", name: "Sofia Barros", eligibility: "ENABLED" },
    { id: "usr-009", name: "Luciana Prado", eligibility: "ENABLED" }
  ];

  const GROUPS = [
    { id: "grp-001", name: "Aprovadores Financeiro" },
    { id: "grp-002", name: "Equipe Comercial" },
    { id: "grp-003", name: "Diretoria" },
    { id: "grp-004", name: "Gestores da Qualidade" },
    { id: "grp-005", name: "Recursos Humanos" },
    { id: "grp-006", name: "Tecnologia" }
  ];

  const AREAS = [
    { id: "area-001", code: "ADM", name: "Administração" },
    { id: "area-002", code: "COM", name: "Comercial" },
    { id: "area-003", code: "FIN", name: "Financeiro" },
    { id: "area-005", code: "QLD", name: "Qualidade" },
    { id: "area-006", code: "RH", name: "Recursos Humanos" }
  ];

  const BASE_TYPES = [
    { id: "dt-001", code: "PO", name: "Procedimento Operacional", scope: "document_type_area", active: true,
      governance: { mode: "use_governance_route", steps: [
        { label: "Revisão da Qualidade", selectorKind: "group", selectorId: "grp-004", dueDays: 5 },
        { label: "Aprovação da Diretoria", selectorKind: "group", selectorId: "grp-003", dueDays: null }
      ] },
      representation: "require_official_rendition", nextSeq: 18 },
    { id: "dt-002", code: "IT", name: "Instrução de Trabalho", scope: "document_type_area", active: true,
      governance: { mode: "use_governance_route", steps: [
        { label: "Aprovação do gestor", selectorKind: "user", selectorId: "usr-004", dueDays: 3 }
      ] },
      representation: "require_official_rendition", nextSeq: 41 },
    { id: "dt-003", code: "FRM", name: "Formulário", scope: "document_type", active: true,
      governance: { mode: "no_human_approval", steps: [] },
      representation: "source_only", nextSeq: 7 },
    { id: "dt-004", code: "MAN", name: "Manual", scope: "document_type", active: true,
      governance: { mode: "use_governance_route", steps: [
        { label: "Revisão técnica", selectorKind: "user", selectorId: "usr-002", dueDays: 10 },
        { label: "Aprovação da Qualidade", selectorKind: "group", selectorId: "grp-004", dueDays: 5 },
        { label: "Aprovação final", selectorKind: "group", selectorId: "grp-003", dueDays: null }
      ] },
      representation: "require_official_rendition", nextSeq: 3 },
    { id: "dt-005", code: "POL", name: "Política", scope: "document_type", active: true,
      governance: { mode: "use_governance_route", steps: [
        { label: "Aprovação da Diretoria", selectorKind: "group", selectorId: "grp-003", dueDays: 7 }
      ] },
      representation: "require_official_rendition", nextSeq: 12 },
    { id: "dt-006", code: "REG", name: "Registro de Qualidade", scope: "document_type_area", active: false,
      governance: { mode: "no_human_approval", steps: [] },
      representation: "source_only", nextSeq: 96 }
  ];

  const BASE_TEMPLATES = [
    { docId: "doc-101", code: "FRM-001", isTemplate: true, hasEffective: true, title: "Modelo de ata de reunião", eligibleTypeIds: ["dt-003"] },
    { docId: "doc-102", code: "FRM-002", isTemplate: true, hasEffective: true, title: "Modelo de checklist de inspeção", eligibleTypeIds: ["dt-003", "dt-006"] },
    { docId: "doc-103", code: "PO-QLD-004", isTemplate: true, hasEffective: false, title: null, eligibleTypeIds: ["dt-001"] },
    { docId: "doc-104", code: "MAN-002", isTemplate: false, hasEffective: true, title: "Manual de integração", eligibleTypeIds: [] },
    { docId: "doc-105", code: "IT-COM-009", isTemplate: true, hasEffective: true, title: "Modelo de roteiro comercial", eligibleTypeIds: ["dt-002"] },
    { docId: "doc-106", code: "POL-003", isTemplate: false, hasEffective: true, title: "Política de segurança da informação", eligibleTypeIds: [] }
  ];

  let types = [];
  let templates = [];
  let etags = {};
  let idempotencyStore = new Map();
  let nextTypeNumber = 100;
  let nextKeyNumber = 1;
  let mutations = { create: 0, replace: 0 };

  const flags = {
    pageFail: false,
    createConflict: false,
    createAmbiguous: false,
    stale: false,
    inUse: false
  };

  const state = {
    lens: "types",
    typeId: "dt-001",
    typeMissing: false,
    pages: { types: 0, templates: 0, eligiblePicker: 0, selector: 0 },
    previewAreaId: "",
    create: {
      key: null,
      fingerprint: null,
      pendingAmbiguous: false,
      terminal: false,
      gov: null
    },
    base: { etag: null },
    governance: { etag: null, gov: null },
    eligible: { etag: null, chosen: new Set() },
    role: { docId: null, etag: null, nextValue: null },
    selector: { targetGov: null, stepIndex: null, kind: null }
  };

  function typeOf(id) { return types.find((item) => item.id === id); }
  function templateOf(docId) { return templates.find((item) => item.docId === docId); }
  function userOf(id) { return USERS.find((item) => item.id === id); }
  function groupOf(id) { return GROUPS.find((item) => item.id === id); }
  function areaOf(id) { return AREAS.find((item) => item.id === id); }

  function selectorLabel(step) {
    if (!step.selectorId) { return "— responsável não selecionado —"; }
    if (step.selectorKind === "user") {
      const person = userOf(step.selectorId);
      return "Usuário nomeado: " + (person ? person.name : step.selectorId);
    }
    const grp = groupOf(step.selectorId);
    return "Grupo: " + (grp ? grp.name : step.selectorId);
  }

  function cloneGov(gov) {
    return { mode: gov.mode, steps: gov.steps.map((step) => ({ ...step })) };
  }

  function pageOf(items, pageIndex, size) {
    const start = pageIndex * size;
    return { items: items.slice(start, start + size), hasMore: start + size < items.length };
  }

  function log(message) {
    $("#fixture-log").textContent = message + " Mutations: criar " + mutations.create + " · gravar " + mutations.replace + ".";
  }

  function setState(element, message, kind) {
    element.classList.remove("error", "success");
    if (kind) { element.classList.add(kind); }
    element.textContent = message;
  }

  function resetFixture() {
    types = BASE_TYPES.map((item) => ({ ...item, governance: cloneGov(item.governance) }));
    templates = BASE_TEMPLATES.map((item) => ({ ...item, eligibleTypeIds: [...item.eligibleTypeIds] }));
    etags = {};
    types.forEach((item) => { etags["base:" + item.id] = 1; etags["gov:" + item.id] = 1; etags["elig:" + item.id] = 1; });
    templates.forEach((item) => { etags["role:" + item.docId] = 1; });
    idempotencyStore = new Map();
    nextTypeNumber = 100;
    nextKeyNumber = 1;
    mutations = { create: 0, replace: 0 };
    Object.keys(flags).forEach((key) => { flags[key] = false; });
    ["#fx-page-fail", "#fx-create-conflict", "#fx-create-ambiguous", "#fx-stale", "#fx-in-use"].forEach((sel) => {
      $(sel).setAttribute("aria-pressed", "false");
    });
    state.lens = "types";
    state.typeId = "dt-001";
    state.typeMissing = false;
    state.pages = { types: 0, templates: 0, eligiblePicker: 0, selector: 0 };
    state.previewAreaId = "";
    $("#gov-denied").hidden = true;
    $("#gov-content").classList.remove("is-denied");
    log("Fixture base restaurada.");
    renderAll();
  }

  /* ---------------- lens ---------------- */

  function setLens(lens) {
    state.lens = lens;
    $("#tab-types").setAttribute("aria-selected", lens === "types" ? "true" : "false");
    $("#tab-templates").setAttribute("aria-selected", lens === "templates" ? "true" : "false");
    $("#lens-types").hidden = lens !== "types";
    $("#lens-templates").hidden = lens !== "templates";
  }

  /* ---------------- types list (op34) ---------------- */

  function renderTypesList() {
    const page = pageOf(types, state.pages.types, PAGE_SIZE);
    const list = $("#types-list");
    list.innerHTML = "";
    page.items.forEach((item) => {
      const button = document.createElement("button");
      button.type = "button";
      button.className = "list-button" + (item.id === state.typeId && !state.typeMissing ? " active" : "");
      button.setAttribute("role", "listitem");
      button.innerHTML = "<span><strong>" + item.code + " · " + item.name + "</strong>" +
        "<span class='meta'>numeração " + (item.scope === "document_type" ? "por tipo" : "por tipo + Área") + "</span></span>" +
        "<span class='pill'>" + (item.active ? "ATIVO" : "INATIVO") + "</span>";
      button.addEventListener("click", () => {
        state.typeId = item.id;
        state.typeMissing = false;
        state.previewAreaId = "";
        renderAll();
      });
      list.appendChild(button);
    });
    $("#types-page-label").textContent = "página " + (state.pages.types + 1) + (page.hasMore ? " · há continuação" : " · fim");
    $("#types-prev").disabled = state.pages.types === 0;
    $("#types-next").disabled = !page.hasMore;
  }

  function goTypesPage(delta) {
    if (delta > 0 && flags.pageFail) {
      flags.pageFail = false;
      $("#fx-page-fail").setAttribute("aria-pressed", "false");
      setState($("#types-state"), "Falha ao continuar a paginação. A página atual permanece a verdade canônica exibida; tente a continuação novamente.", "error");
      return;
    }
    state.pages.types += delta;
    setState($("#types-state"), "", null);
    renderTypesList();
  }

  /* ---------------- type detail ---------------- */

  function renderDetail() {
    const missing = state.typeMissing;
    $("#type-404").hidden = !missing;
    $("#type-detail").hidden = missing;
    if (missing) {
      $("#detail-title").textContent = "Detalhe do tipo";
      return;
    }
    const item = typeOf(state.typeId);
    $("#detail-title").textContent = item.code + " · " + item.name;

    $("#base-facts").innerHTML =
      "<dt>Code</dt><dd>" + item.code + "</dd>" +
      "<dt>Nome</dt><dd>" + item.name + "</dd>" +
      "<dt>Numeração</dt><dd>" + (item.scope === "document_type" ? "por tipo (ex.: " + item.code + "-001)" : "por tipo + Área (ex.: " + item.code + "-RH-001)") + "</dd>" +
      "<dt>Ativo</dt><dd>" + (item.active ? "sim" : "não — indisponível para novos documentos") + "</dd>" +
      "<dt>ETag</dt><dd><code>W/\"base-" + etags["base:" + item.id] + "\"</code></dd>";

    renderGovView(item);
    renderEligibleView(item);
    renderPreviewControls(item);
  }

  function renderGovView(item) {
    const view = $("#govpol-view");
    if (item.governance.mode === "no_human_approval") {
      view.innerHTML = "<div class='read-only'><strong>Sem aprovação humana.</strong> SUBMIT segue direto para o gate de representação; nenhum step existe nesta política.</div>" +
        representationLine(item);
      return;
    }
    const rows = item.governance.steps.map((step, index) =>
      "<li><strong>" + (index + 1) + ". " + step.label + "</strong><span class='meta'> — " + selectorLabel(step) + " · " +
      (step.dueDays ? "prazo: " + step.dueDays + " dia(s) corrido(s) após ativação" : "sem prazo configurado") + "</span></li>"
    ).join("");
    view.innerHTML = "<ol class='route-order'>" + rows + "</ol>" + representationLine(item) +
      "<div class='read-only'>ETag governança: <code>W/\"gov-" + etags["gov:" + item.id] + "\"</code>. Attempts existentes preservam o snapshot da rota vigente na criação.</div>";
  }

  function representationLine(item) {
    return "<div class='state-line'>Representação oficial: <strong>" +
      (item.representation === "source_only" ? "somente fonte (DOCX exato)" : "exige Official Rendition (PDF)") + "</strong></div>";
  }

  function renderEligibleView(item) {
    const chosen = templates.filter((tpl) => tpl.eligibleTypeIds.includes(item.id))
      .sort((a, b) => a.code.localeCompare(b.code));
    const view = $("#eligible-view");
    if (chosen.length === 0) {
      view.innerHTML = "<div class='empty'>Nenhum modelo elegível para este tipo. Conjunto vazio é um estado válido, não uma falha.</div>";
    } else {
      view.innerHTML = "<div class='chips'>" + chosen.map((tpl) =>
        "<span class='chip'>" + tpl.code + (tpl.hasEffective ? "" : " · sem revisão efetiva") + "</span>").join("") + "</div>" +
        "<div class='state-line'>ETag conjunto: <code>W/\"elig-" + etags["elig:" + item.id] + "\"</code></div>";
    }
  }

  function renderPreviewControls(item) {
    const select = $("#preview-area");
    const needsArea = item.scope === "document_type_area";
    $("#preview-area-label").hidden = !needsArea;
    select.hidden = !needsArea;
    if (needsArea) {
      select.innerHTML = "<option value=''>— selecione a Área —</option>" +
        AREAS.map((areaItem) => "<option value='" + areaItem.id + "'" + (state.previewAreaId === areaItem.id ? " selected" : "") + ">" + areaItem.code + " · " + areaItem.name + "</option>").join("");
    }
    setState($("#preview-state"), "", null);
  }

  function runPreview() {
    const item = typeOf(state.typeId);
    const needsArea = item.scope === "document_type_area";
    state.previewAreaId = needsArea ? $("#preview-area").value : "";
    if (needsArea && !state.previewAreaId) {
      setState($("#preview-state"), "422 · validation.failed — numeração deste tipo é por tipo + Área: informe area_id para gerar o preview.", "error");
      return;
    }
    const areaItem = needsArea ? areaOf(state.previewAreaId) : null;
    const seq = String(item.nextSeq).padStart(3, "0");
    const code = needsArea ? item.code + "-" + areaItem.code + "-" + seq : item.code + "-" + seq;
    setState($("#preview-state"), "Preview: " + code + " · reservation=false — nada foi reservado; o código final só existe na criação atômica do Documento.", "success");
  }

  /* ---------------- governance policy editor (shared) ---------------- */

  function buildGovEditor(container, gov, options) {
    container.innerHTML = "";
    const modeBox = document.createElement("div");
    modeBox.className = "mode-choice";
    modeBox.setAttribute("role", "radiogroup");
    modeBox.setAttribute("aria-label", "Modo de governança");
    modeBox.innerHTML =
      "<label><input type='radio' name='" + options.name + "-mode' value='no_human_approval'" + (gov.mode === "no_human_approval" ? " checked" : "") + ">" +
      "<span><strong>Sem aprovação humana</strong><small>steps são proibidos nesta política; SUBMIT segue direto para o gate de representação.</small></span></label>" +
      "<label><input type='radio' name='" + options.name + "-mode' value='use_governance_route'" + (gov.mode === "use_governance_route" ? " checked" : "") + ">" +
      "<span><strong>Rota de governança</strong><small>steps sequenciais; exige pelo menos 1 step. A ordem da lista é a ordem da rota.</small></span></label>";
    container.appendChild(modeBox);

    const stepsWrap = document.createElement("div");
    container.appendChild(stepsWrap);

    const repBox = document.createElement("div");
    repBox.className = "mode-choice";
    repBox.setAttribute("role", "radiogroup");
    repBox.setAttribute("aria-label", "Representação oficial");
    repBox.innerHTML =
      "<label><input type='radio' name='" + options.name + "-rep' value='source_only'" + (gov.representation === "source_only" ? " checked" : "") + ">" +
      "<span><strong>Somente fonte</strong><small>o efetivo é o DOCX exato submetido.</small></span></label>" +
      "<label><input type='radio' name='" + options.name + "-rep' value='require_official_rendition'" + (gov.representation === "require_official_rendition" ? " checked" : "") + ">" +
      "<span><strong>Exigir Official Rendition (PDF)</strong><small>a efetivação exige a renderização oficial em PDF.</small></span></label>";
    container.appendChild(repBox);

    modeBox.addEventListener("change", (event) => {
      gov.mode = event.target.value;
      if (gov.mode === "use_governance_route" && gov.steps.length === 0) {
        gov.steps.push({ label: "", selectorKind: "user", selectorId: null, dueDays: null });
      }
      renderSteps();
    });
    repBox.addEventListener("change", (event) => { gov.representation = event.target.value; });

    function renderSteps() {
      stepsWrap.innerHTML = "";
      if (gov.mode === "no_human_approval") {
        stepsWrap.innerHTML = "<div class='read-only'>Sem steps: <code>no_human_approval</code> proíbe o membro <code>steps</code>.</div>";
        return;
      }
      const head = document.createElement("div");
      head.className = "gov-steps-head";
      head.innerHTML = "<h4>Steps da rota (ordem = rota)</h4>";
      const addButton = document.createElement("button");
      addButton.type = "button";
      addButton.className = "row-action";
      addButton.textContent = "＋ Adicionar step";
      addButton.addEventListener("click", () => {
        gov.steps.push({ label: "", selectorKind: "user", selectorId: null, dueDays: null });
        renderSteps();
      });
      head.appendChild(addButton);
      stepsWrap.appendChild(head);

      gov.steps.forEach((step, index) => {
        const row = document.createElement("div");
        row.className = "step-row";
        const ordinal = document.createElement("div");
        ordinal.className = "step-ordinal";
        ordinal.textContent = String(index + 1);
        row.appendChild(ordinal);

        const fields = document.createElement("div");
        fields.className = "step-fields";
        const labelId = options.name + "-step-" + index;

        const labelBox = document.createElement("div");
        labelBox.innerHTML = "<label for='" + labelId + "-label'>Rótulo</label>";
        const labelInput = document.createElement("input");
        labelInput.type = "text";
        labelInput.id = labelId + "-label";
        labelInput.maxLength = 60;
        labelInput.value = step.label;
        labelInput.addEventListener("input", () => { step.label = labelInput.value; });
        labelBox.appendChild(labelInput);
        fields.appendChild(labelBox);

        const selectorBox = document.createElement("div");
        selectorBox.innerHTML = "<label>Responsável</label>";
        const kindWrap = document.createElement("div");
        ["user", "group"].forEach((kind) => {
          const radioLabel = document.createElement("label");
          radioLabel.className = "inline-check";
          radioLabel.style.paddingTop = "0";
          const radio = document.createElement("input");
          radio.type = "radio";
          radio.name = labelId + "-kind";
          radio.value = kind;
          radio.checked = step.selectorKind === kind;
          radio.addEventListener("change", () => {
            step.selectorKind = kind;
            step.selectorId = null;
            renderSteps();
          });
          radioLabel.appendChild(radio);
          radioLabel.appendChild(document.createTextNode(kind === "user" ? " Usuário nomeado" : " Grupo"));
          kindWrap.appendChild(radioLabel);
        });
        selectorBox.appendChild(kindWrap);
        const pickButton = document.createElement("button");
        pickButton.type = "button";
        pickButton.className = "row-action";
        pickButton.textContent = "Selecionar…";
        pickButton.addEventListener("click", () => openSelector(gov, index, renderSteps));
        selectorBox.appendChild(pickButton);
        const currentValue = document.createElement("div");
        currentValue.className = "step-selector-value";
        currentValue.textContent = selectorLabel(step);
        selectorBox.appendChild(currentValue);
        fields.appendChild(selectorBox);

        const dueBox = document.createElement("div");
        dueBox.innerHTML = "<label for='" + labelId + "-due'>Prazo (dias corridos)</label>";
        const dueInput = document.createElement("input");
        dueInput.type = "number";
        dueInput.id = labelId + "-due";
        dueInput.min = "1";
        dueInput.step = "1";
        dueInput.placeholder = "sem prazo";
        dueInput.value = step.dueDays === null || step.dueDays === undefined ? "" : String(step.dueDays);
        dueInput.addEventListener("input", () => {
          step.dueDays = dueInput.value === "" ? null : Number(dueInput.value);
        });
        dueBox.appendChild(dueInput);
        const dueHint = document.createElement("div");
        dueHint.className = "field-hint";
        dueHint.textContent = "Opcional. 1 dia = 24h corridas; vazio = sem prazo. Congela em due_at quando o step ativa.";
        dueBox.appendChild(dueHint);
        fields.appendChild(dueBox);

        row.appendChild(fields);

        const tools = document.createElement("div");
        tools.className = "step-tools";
        const upButton = document.createElement("button");
        upButton.type = "button";
        upButton.textContent = "↑ subir";
        upButton.disabled = index === 0;
        upButton.setAttribute("aria-label", "Mover step " + (index + 1) + " para cima");
        upButton.addEventListener("click", () => {
          [gov.steps[index - 1], gov.steps[index]] = [gov.steps[index], gov.steps[index - 1]];
          renderSteps();
        });
        const downButton = document.createElement("button");
        downButton.type = "button";
        downButton.textContent = "↓ descer";
        downButton.disabled = index === gov.steps.length - 1;
        downButton.setAttribute("aria-label", "Mover step " + (index + 1) + " para baixo");
        downButton.addEventListener("click", () => {
          [gov.steps[index + 1], gov.steps[index]] = [gov.steps[index], gov.steps[index + 1]];
          renderSteps();
        });
        const removeButton = document.createElement("button");
        removeButton.type = "button";
        removeButton.textContent = "Remover";
        removeButton.setAttribute("aria-label", "Remover step " + (index + 1));
        removeButton.addEventListener("click", () => {
          gov.steps.splice(index, 1);
          renderSteps();
        });
        tools.appendChild(upButton);
        tools.appendChild(downButton);
        tools.appendChild(removeButton);
        row.appendChild(tools);

        stepsWrap.appendChild(row);
      });
    }

    renderSteps();
  }

  function validateGov(gov) {
    if (gov.mode === "no_human_approval") { return null; }
    if (gov.steps.length === 0) { return "Rota de governança exige pelo menos 1 step."; }
    for (let index = 0; index < gov.steps.length; index += 1) {
      const step = gov.steps[index];
      if (!step.label.trim()) { return "Step " + (index + 1) + ": rótulo é obrigatório."; }
      if (!step.selectorId) { return "Step " + (index + 1) + ": selecione o responsável (usuário nomeado ou grupo)."; }
      if (step.dueDays !== null && (!Number.isInteger(step.dueDays) || step.dueDays < 1)) {
        return "Step " + (index + 1) + ": prazo deve ser um inteiro positivo de dias, ou vazio.";
      }
    }
    return null;
  }

  /* ---------------- selector picker (op6 / op22) ---------------- */

  function openSelector(gov, stepIndex, rerender) {
    state.selector = { targetGov: gov, stepIndex, kind: gov.steps[stepIndex].selectorKind, rerender };
    state.pages.selector = 0;
    $("#selector-help").textContent = state.selector.kind === "user"
      ? "Página bruta de usuários (op6). Usuários DISABLED aparecem, mas não são selecionáveis para novos steps."
      : "Página bruta de grupos (op22). O grupo resolve seus membros no momento da ativação do step.";
    renderSelectorPicker();
    $("#dlg-selector").showModal();
  }

  function renderSelectorPicker() {
    const source = state.selector.kind === "user" ? USERS : GROUPS;
    const page = pageOf(source, state.pages.selector, PICKER_PAGE_SIZE);
    const list = $("#selector-picker");
    list.innerHTML = "";
    page.items.forEach((item) => {
      const button = document.createElement("button");
      button.type = "button";
      button.className = "list-button";
      button.setAttribute("role", "listitem");
      const disabled = state.selector.kind === "user" && item.eligibility === "DISABLED";
      button.disabled = disabled;
      button.innerHTML = "<span><strong>" + item.name + "</strong>" +
        (disabled ? "<span class='meta'>DISABLED — visível, indisponível para novos steps</span>" : "") + "</span>";
      button.addEventListener("click", () => {
        const step = state.selector.targetGov.steps[state.selector.stepIndex];
        step.selectorId = item.id;
        $("#dlg-selector").close();
        state.selector.rerender();
      });
      list.appendChild(button);
    });
    $("#selector-label").textContent = "página " + (state.pages.selector + 1) + (page.hasMore ? " · há continuação" : " · fim");
    $("#selector-prev").disabled = state.pages.selector === 0;
    $("#selector-next").disabled = !page.hasMore;
  }

  function goSelectorPage(delta) {
    if (delta > 0 && flags.pageFail) {
      flags.pageFail = false;
      $("#fx-page-fail").setAttribute("aria-pressed", "false");
      setState($("#selector-state"), "Falha ao continuar. A página atual permanece; tente novamente.", "error");
      return;
    }
    state.pages.selector += delta;
    setState($("#selector-state"), "", null);
    renderSelectorPicker();
  }

  /* ---------------- create type (op35) ---------------- */

  function openCreate() {
    state.create = {
      key: "idem-b12-" + String(nextKeyNumber).padStart(3, "0"),
      fingerprint: null,
      pendingAmbiguous: false,
      terminal: false,
      gov: { mode: "use_governance_route", steps: [{ label: "", selectorKind: "user", selectorId: null, dueDays: null }], representation: "require_official_rendition" }
    };
    nextKeyNumber += 1;
    $("#create-code").value = "";
    $("#create-name").value = "";
    $("#create-scope").value = "document_type";
    $("#create-active").checked = true;
    $("#create-result").hidden = true;
    $("#create-ambiguous").hidden = true;
    $("#create-retry").hidden = true;
    $("#create-submit").hidden = false;
    setCreateFormDisabled(false);
    setState($("#create-state"), "Idempotency-Key desta intenção: " + state.create.key, null);
    buildGovEditor($("#create-gov-editor"), state.create.gov, { name: "create" });
    $("#dlg-create").showModal();
  }

  function setCreateFormDisabled(disabled) {
    ["#create-code", "#create-name", "#create-scope", "#create-active"].forEach((sel) => { $(sel).disabled = disabled; });
    $("#create-gov-editor").querySelectorAll("input, button, select").forEach((el) => { el.disabled = disabled; });
  }

  function normalizeCode(raw) {
    return raw.trim().toUpperCase();
  }

  function submitCreate() {
    if (state.create.terminal) { return; }
    const code = normalizeCode($("#create-code").value);
    const name = $("#create-name").value.trim();
    if (!code) { setState($("#create-state"), "Informe o code.", "error"); return; }
    if (!/^[A-Z0-9]+$/.test(code)) {
      setState($("#create-state"), "422 · validation.failed — code aceita apenas ASCII alfanumérico maiúsculo; “-” é proibido (o Produto é dono do separador).", "error");
      return;
    }
    if (!name) { setState($("#create-state"), "Informe o nome.", "error"); return; }
    const govError = validateGov(state.create.gov);
    if (govError) { setState($("#create-state"), "422 · validation.failed — " + govError, "error"); return; }

    const fingerprint = JSON.stringify({ code, name, scope: $("#create-scope").value, active: $("#create-active").checked, gov: state.create.gov });
    state.create.fingerprint = fingerprint;

    const replay = idempotencyStore.get(state.create.key);
    if (replay) {
      finishCreate(replay, true);
      return;
    }

    if (flags.createConflict) {
      flags.createConflict = false;
      $("#fx-create-conflict").setAttribute("aria-pressed", "false");
      setState($("#create-state"), "409 · conflict.code_taken — já existe um tipo com o code normalizado “" + code + "” nesta Company. Nenhum tipo foi criado; ajuste o code e envie novamente (nova intenção = nova chave).", "error");
      state.create.key = "idem-b12-" + String(nextKeyNumber).padStart(3, "0");
      nextKeyNumber += 1;
      return;
    }

    const created = {
      id: "dt-" + nextTypeNumber,
      code,
      name,
      scope: $("#create-scope").value,
      active: $("#create-active").checked,
      governance: cloneGov(state.create.gov),
      representation: state.create.gov.representation,
      nextSeq: 1
    };
    nextTypeNumber += 1;
    types.push(created);
    etags["base:" + created.id] = 1;
    etags["gov:" + created.id] = 1;
    etags["elig:" + created.id] = 1;
    mutations.create += 1;
    idempotencyStore.set(state.create.key, { typeId: created.id, code: created.code });

    if (flags.createAmbiguous) {
      flags.createAmbiguous = false;
      $("#fx-create-ambiguous").setAttribute("aria-pressed", "false");
      state.create.pendingAmbiguous = true;
      $("#create-ambiguous").hidden = false;
      $("#create-submit").hidden = true;
      $("#create-retry").hidden = false;
      setCreateFormDisabled(true);
      setState($("#create-state"), "A resposta do servidor se perdeu após o commit ser possível. Não altere a composição; use a mesma chave.", "error");
      log("op35 ambíguo após commit (tipo persistido no servidor, resposta perdida).");
      return;
    }

    finishCreate(idempotencyStore.get(state.create.key), false);
  }

  function retryCreate() {
    if (!state.create.pendingAmbiguous) { return; }
    const replay = idempotencyStore.get(state.create.key);
    finishCreate(replay, true);
  }

  function finishCreate(result, isReplay) {
    state.create.pendingAmbiguous = false;
    state.create.terminal = true;
    $("#create-ambiguous").hidden = true;
    $("#create-retry").hidden = true;
    $("#create-submit").hidden = true;
    setCreateFormDisabled(true);
    $("#create-result").hidden = false;
    $("#create-result").innerHTML = "<strong>" + (isReplay ? "200 · replay exato da mesma chave" : "201 · tipo criado") + "</strong>" +
      "document_type_id: <code>" + result.typeId + "</code> · code <code>" + result.code + "</code>" +
      "<div class='proof-box'>Prova de idempotência: mutations criar = " + mutations.create + " (uma intenção lógica → um tipo, mesmo após replay).</div>";
    setState($("#create-state"), "", null);
    state.typeId = result.typeId;
    state.typeMissing = false;
    log(isReplay ? "op35 replay recuperou o resultado exato armazenado." : "op35 criou " + result.code + ".");
    renderAll();
  }

  /* ---------------- edit base (op37) ---------------- */

  function openBase() {
    const item = typeOf(state.typeId);
    state.base.etag = etags["base:" + item.id];
    $("#base-code").value = item.code;
    $("#base-name").value = item.name;
    $("#base-scope").value = item.scope;
    $("#base-active").checked = item.active;
    $("#base-reload").hidden = true;
    $("#base-save").disabled = false;
    setState($("#dlg-base-state"), "If-Match: W/\"base-" + state.base.etag + "\"", null);
    $("#dlg-base").showModal();
  }

  function saveBase() {
    const item = typeOf(state.typeId);
    const code = normalizeCode($("#base-code").value);
    const name = $("#base-name").value.trim();
    if (!code || !/^[A-Z0-9]+$/.test(code)) {
      setState($("#dlg-base-state"), "422 · validation.failed — code aceita apenas ASCII alfanumérico maiúsculo, sem “-”.", "error");
      return;
    }
    if (!name) { setState($("#dlg-base-state"), "Informe o nome.", "error"); return; }

    if (flags.stale) {
      flags.stale = false;
      $("#fx-stale").setAttribute("aria-pressed", "false");
      etags["base:" + item.id] += 1;
      setState($("#dlg-base-state"), "412 · precondition.resource_changed — outra sessão gravou esta base primeiro. Zero mutação aplicada; releia a versão atual e reaplique sua intenção.", "error");
      $("#base-reload").hidden = false;
      $("#base-save").disabled = true;
      return;
    }

    const structuralChange = code !== item.code || $("#base-scope").value !== item.scope;
    if (flags.inUse && structuralChange) {
      flags.inUse = false;
      $("#fx-in-use").setAttribute("aria-pressed", "false");
      setState($("#dlg-base-state"), "409 · state.document_type_in_use — code e escopo de numeração tornaram-se imutáveis: um Documento comprometido já usa este tipo. Nome e ativo continuam editáveis. (B12-F1: este fato só é conhecido na falha da escrita.)", "error");
      return;
    }

    item.code = code;
    item.name = name;
    item.scope = $("#base-scope").value;
    item.active = $("#base-active").checked;
    etags["base:" + item.id] += 1;
    mutations.replace += 1;
    log("op37 gravou base de " + item.code + ".");
    $("#dlg-base").close();
    setState($("#base-state"), "Base gravada. Novo ETag: W/\"base-" + etags["base:" + item.id] + "\".", "success");
    renderAll();
  }

  function reloadBase() {
    openBaseAfterReload();
  }

  function openBaseAfterReload() {
    const item = typeOf(state.typeId);
    state.base.etag = etags["base:" + item.id];
    $("#base-code").value = item.code;
    $("#base-name").value = item.name;
    $("#base-scope").value = item.scope;
    $("#base-active").checked = item.active;
    $("#base-reload").hidden = true;
    $("#base-save").disabled = false;
    setState($("#dlg-base-state"), "Versão atual relida. If-Match: W/\"base-" + state.base.etag + "\". Reaplique sua intenção sobre o estado vigente.", null);
  }

  /* ---------------- edit governance (op39) ---------------- */

  function openGovernance() {
    const item = typeOf(state.typeId);
    state.governance.etag = etags["gov:" + item.id];
    state.governance.gov = cloneGov(item.governance);
    state.governance.gov.representation = item.representation;
    $("#governance-reload").hidden = true;
    $("#governance-save").disabled = false;
    setState($("#dlg-governance-state"), "If-Match: W/\"gov-" + state.governance.etag + "\"", null);
    buildGovEditor($("#gov-editor"), state.governance.gov, { name: "gov" });
    $("#dlg-governance").showModal();
  }

  function saveGovernance() {
    const item = typeOf(state.typeId);
    const govError = validateGov(state.governance.gov);
    if (govError) { setState($("#dlg-governance-state"), "422 · validation.failed — " + govError, "error"); return; }

    if (flags.stale) {
      flags.stale = false;
      $("#fx-stale").setAttribute("aria-pressed", "false");
      etags["gov:" + item.id] += 1;
      setState($("#dlg-governance-state"), "412 · precondition.resource_changed — a rota foi gravada por outra sessão. Zero mutação; releia e reaplique.", "error");
      $("#governance-reload").hidden = false;
      $("#governance-save").disabled = true;
      return;
    }

    item.governance = { mode: state.governance.gov.mode, steps: state.governance.gov.steps.map((step) => ({ ...step })) };
    item.representation = state.governance.gov.representation;
    etags["gov:" + item.id] += 1;
    mutations.replace += 1;
    log("op39 gravou governança de " + item.code + ".");
    $("#dlg-governance").close();
    setState($("#govpol-state"), "Governança gravada. Attempts existentes preservam o snapshot anterior; attempts futuros usam esta configuração.", "success");
    renderAll();
  }

  function reloadGovernance() {
    const item = typeOf(state.typeId);
    state.governance.etag = etags["gov:" + item.id];
    state.governance.gov = cloneGov(item.governance);
    state.governance.gov.representation = item.representation;
    $("#governance-reload").hidden = true;
    $("#governance-save").disabled = false;
    setState($("#dlg-governance-state"), "Versão atual relida. If-Match: W/\"gov-" + state.governance.etag + "\".", null);
    buildGovEditor($("#gov-editor"), state.governance.gov, { name: "gov" });
  }

  /* ---------------- eligible templates (op41) ---------------- */

  function openEligible() {
    const item = typeOf(state.typeId);
    state.eligible.etag = etags["elig:" + item.id];
    state.eligible.chosen = new Set(templates.filter((tpl) => tpl.eligibleTypeIds.includes(item.id)).map((tpl) => tpl.docId));
    state.pages.eligiblePicker = 0;
    $("#eligible-reload").hidden = true;
    $("#eligible-save").disabled = false;
    setState($("#dlg-eligible-state"), "If-Match: W/\"elig-" + state.eligible.etag + "\" · substituição integral do conjunto.", null);
    renderEligiblePicker();
    $("#dlg-eligible").showModal();
  }

  function renderEligiblePicker() {
    const ordered = [...templates].sort((a, b) => a.code.localeCompare(b.code));
    const page = pageOf(ordered, state.pages.eligiblePicker, PICKER_PAGE_SIZE);
    const list = $("#eligible-picker");
    list.innerHTML = "";
    page.items.forEach((tpl) => {
      const row = document.createElement("label");
      row.className = "check-row";
      const check = document.createElement("input");
      check.type = "checkbox";
      check.checked = state.eligible.chosen.has(tpl.docId);
      check.disabled = !tpl.isTemplate;
      check.addEventListener("change", () => {
        if (check.checked) { state.eligible.chosen.add(tpl.docId); } else { state.eligible.chosen.delete(tpl.docId); }
      });
      row.appendChild(check);
      const info = document.createElement("span");
      info.innerHTML = "<strong>" + tpl.code + "</strong><span class='meta'>" +
        (!tpl.isTemplate ? "sem papel de modelo — indisponível como elegível" :
          tpl.hasEffective ? (tpl.title || "") : "sem revisão efetiva — configurável, mas não aparecerá como opção de criação") + "</span>";
      row.appendChild(info);
      list.appendChild(row);
    });
    $("#eligible-picker-label").textContent = "página " + (state.pages.eligiblePicker + 1) + (page.hasMore ? " · há continuação" : " · fim");
    $("#eligible-picker-prev").disabled = state.pages.eligiblePicker === 0;
    $("#eligible-picker-next").disabled = !page.hasMore;
  }

  function goEligiblePage(delta) {
    if (delta > 0 && flags.pageFail) {
      flags.pageFail = false;
      $("#fx-page-fail").setAttribute("aria-pressed", "false");
      setState($("#eligible-picker-state"), "Falha ao continuar. Página atual preservada; tente novamente.", "error");
      return;
    }
    state.pages.eligiblePicker += delta;
    setState($("#eligible-picker-state"), "", null);
    renderEligiblePicker();
  }

  function saveEligible() {
    const item = typeOf(state.typeId);
    if (flags.stale) {
      flags.stale = false;
      $("#fx-stale").setAttribute("aria-pressed", "false");
      etags["elig:" + item.id] += 1;
      setState($("#dlg-eligible-state"), "412 · precondition.resource_changed — o conjunto mudou em outra sessão. Zero mutação; releia e reaplique.", "error");
      $("#eligible-reload").hidden = false;
      $("#eligible-save").disabled = true;
      return;
    }
    templates.forEach((tpl) => {
      const included = state.eligible.chosen.has(tpl.docId);
      const has = tpl.eligibleTypeIds.includes(item.id);
      if (included && !has) { tpl.eligibleTypeIds.push(item.id); }
      if (!included && has) { tpl.eligibleTypeIds = tpl.eligibleTypeIds.filter((id) => id !== item.id); }
    });
    etags["elig:" + item.id] += 1;
    mutations.replace += 1;
    log("op41 substituiu o conjunto elegível de " + item.code + ".");
    $("#dlg-eligible").close();
    setState($("#eligible-state"), "Conjunto gravado (substituição integral). Vazio é válido.", "success");
    renderAll();
  }

  function reloadEligible() {
    const item = typeOf(state.typeId);
    state.eligible.etag = etags["elig:" + item.id];
    state.eligible.chosen = new Set(templates.filter((tpl) => tpl.eligibleTypeIds.includes(item.id)).map((tpl) => tpl.docId));
    $("#eligible-reload").hidden = true;
    $("#eligible-save").disabled = false;
    setState($("#dlg-eligible-state"), "Versão atual relida. If-Match: W/\"elig-" + state.eligible.etag + "\".", null);
    renderEligiblePicker();
  }

  /* ---------------- templates lens (op43 / op51) ---------------- */

  function renderTemplatesList() {
    const ordered = [...templates].sort((a, b) => a.code.localeCompare(b.code));
    const page = pageOf(ordered, state.pages.templates, PAGE_SIZE);
    const list = $("#templates-list");
    list.innerHTML = "";
    page.items.forEach((tpl) => {
      const row = document.createElement("div");
      row.className = "template-row";
      row.setAttribute("role", "listitem");
      const eligibleChips = tpl.eligibleTypeIds.map((typeId) => {
        const target = typeOf(typeId);
        return target ? "<button type='button' class='chip' data-open-type='" + typeId + "'>" + target.code + " ↗</button>" : "";
      }).join("") || "<span class='meta'>nenhum tipo elegível</span>";
      row.innerHTML =
        "<div><b>" + tpl.code + "</b><small>" + (tpl.hasEffective ? (tpl.title || "") : "sem revisão efetiva") + "</small></div>" +
        "<div><span class='pill'>" + (tpl.isTemplate ? "PAPEL DE MODELO" : "SEM PAPEL DE MODELO") + "</span><small>" +
        (tpl.hasEffective ? "revisão efetiva presente" : "não aparecerá como opção de criação") + "</small></div>" +
        "<div><small>Elegível para (edite no tipo dono):</small><div class='chips'>" + eligibleChips + "</div></div>";
      const action = document.createElement("button");
      action.type = "button";
      action.className = "row-action";
      action.textContent = tpl.isTemplate ? "Remover papel de modelo" : "Atribuir papel de modelo";
      action.addEventListener("click", () => openRole(tpl.docId));
      row.appendChild(action);
      list.appendChild(row);
    });
    list.querySelectorAll("[data-open-type]").forEach((chip) => {
      chip.addEventListener("click", () => {
        state.typeId = chip.getAttribute("data-open-type");
        state.typeMissing = false;
        setLens("types");
        renderAll();
      });
    });
    $("#templates-page-label").textContent = "página " + (state.pages.templates + 1) + (page.hasMore ? " · há continuação" : " · fim");
    $("#templates-prev").disabled = state.pages.templates === 0;
    $("#templates-next").disabled = !page.hasMore;
  }

  function goTemplatesPage(delta) {
    if (delta > 0 && flags.pageFail) {
      flags.pageFail = false;
      $("#fx-page-fail").setAttribute("aria-pressed", "false");
      setState($("#templates-state"), "Falha ao continuar a paginação. Página atual preservada; tente novamente.", "error");
      return;
    }
    state.pages.templates += delta;
    setState($("#templates-state"), "", null);
    renderTemplatesList();
  }

  function openRole(docId) {
    const tpl = templateOf(docId);
    state.role = { docId, etag: etags["role:" + docId], nextValue: !tpl.isTemplate };
    $("#role-facts").innerHTML =
      "<dt>Documento</dt><dd>" + tpl.code + "</dd>" +
      "<dt>Papel atual</dt><dd>" + (tpl.isTemplate ? "modelo" : "sem papel de modelo") + "</dd>" +
      "<dt>Novo papel</dt><dd>" + (state.role.nextValue ? "modelo" : "sem papel de modelo") + "</dd>" +
      "<dt>If-Match</dt><dd><code>W/\"role-" + state.role.etag + "\"</code></dd>";
    $("#role-consequence").textContent = state.role.nextValue
      ? "O documento passará a aparecer na administração de modelos e poderá ser marcado elegível por tipo. Opções de criação continuam exigindo revisão EFETIVA e elegibilidade no momento da criação."
      : "O documento deixa de ser oferecido como modelo para novas criações. Documentos já derivados dele permanecem válidos e não são rebindados.";
    $("#role-reload").hidden = true;
    $("#role-confirm").disabled = false;
    setState($("#dlg-role-state"), "", null);
    if (!$("#dlg-role").open) { $("#dlg-role").showModal(); }
  }

  function confirmRole() {
    const tpl = templateOf(state.role.docId);
    if (flags.stale) {
      flags.stale = false;
      $("#fx-stale").setAttribute("aria-pressed", "false");
      etags["role:" + tpl.docId] += 1;
      setState($("#dlg-role-state"), "412 · precondition.resource_changed — o papel foi alterado em outra sessão. Zero mutação; releia e decida sobre o estado vigente.", "error");
      $("#role-reload").hidden = false;
      $("#role-confirm").disabled = true;
      return;
    }
    tpl.isTemplate = state.role.nextValue;
    if (!tpl.isTemplate) { tpl.eligibleTypeIds = []; }
    etags["role:" + tpl.docId] += 1;
    mutations.replace += 1;
    log("op51 gravou papel de modelo de " + tpl.code + ".");
    $("#dlg-role").close();
    setState($("#templates-state"), "Papel de modelo gravado para " + tpl.code + ".", "success");
    renderAll();
  }

  function reloadRole() {
    const tpl = templateOf(state.role.docId);
    state.role.etag = etags["role:" + tpl.docId];
    state.role.nextValue = !tpl.isTemplate;
    openRole(tpl.docId);
  }

  /* ---------------- denied / 404 ---------------- */

  function deny() {
    $("#gov-denied").hidden = false;
    $("#gov-content").classList.add("is-denied");
    log("403 em /admin/document-governance — superfície negada sem coleção vazia fingida.");
  }

  function undeny() {
    $("#gov-denied").hidden = true;
    $("#gov-content").classList.remove("is-denied");
    log("Contexto autorizado recarregado.");
    renderAll();
  }

  /* ---------------- render root ---------------- */

  function renderAll() {
    renderTypesList();
    renderDetail();
    renderTemplatesList();
  }

  /* ---------------- wiring ---------------- */

  $("#fx-reset").addEventListener("click", resetFixture);
  [["#fx-page-fail", "pageFail"], ["#fx-create-conflict", "createConflict"], ["#fx-create-ambiguous", "createAmbiguous"], ["#fx-stale", "stale"], ["#fx-in-use", "inUse"]].forEach(([sel, flag]) => {
    $(sel).addEventListener("click", () => {
      flags[flag] = !flags[flag];
      $(sel).setAttribute("aria-pressed", String(flags[flag]));
      log("Fixture “" + $(sel).textContent + "” = " + (flags[flag] ? "armado" : "desarmado") + ".");
    });
  });
  $("#fx-403").addEventListener("click", deny);
  $("#fx-404").addEventListener("click", () => {
    state.typeMissing = true;
    log("404 no tipo selecionado — ausência não prova existência oculta.");
    renderAll();
  });
  $("#gov-denied-reload").addEventListener("click", undeny);

  $("#tab-types").addEventListener("click", () => setLens("types"));
  $("#tab-templates").addEventListener("click", () => setLens("templates"));

  $("#types-prev").addEventListener("click", () => goTypesPage(-1));
  $("#types-next").addEventListener("click", () => goTypesPage(1));
  $("#templates-prev").addEventListener("click", () => goTemplatesPage(-1));
  $("#templates-next").addEventListener("click", () => goTemplatesPage(1));
  $("#eligible-picker-prev").addEventListener("click", () => goEligiblePage(-1));
  $("#eligible-picker-next").addEventListener("click", () => goEligiblePage(1));
  $("#selector-prev").addEventListener("click", () => goSelectorPage(-1));
  $("#selector-next").addEventListener("click", () => goSelectorPage(1));

  $("#open-create").addEventListener("click", openCreate);
  $("#create-submit").addEventListener("click", submitCreate);
  $("#create-retry").addEventListener("click", retryCreate);

  $("#edit-base").addEventListener("click", openBase);
  $("#base-save").addEventListener("click", saveBase);
  $("#base-reload").addEventListener("click", reloadBase);

  $("#edit-governance").addEventListener("click", openGovernance);
  $("#governance-save").addEventListener("click", saveGovernance);
  $("#governance-reload").addEventListener("click", reloadGovernance);

  $("#edit-eligible").addEventListener("click", openEligible);
  $("#eligible-save").addEventListener("click", saveEligible);
  $("#eligible-reload").addEventListener("click", reloadEligible);

  $("#role-confirm").addEventListener("click", confirmRole);
  $("#role-reload").addEventListener("click", reloadRole);

  $("#run-preview").addEventListener("click", runPreview);

  document.querySelectorAll("[data-close]").forEach((button) => {
    button.addEventListener("click", () => $("#" + button.getAttribute("data-close")).close());
  });

  $("#global-nav-toggle").addEventListener("click", () => {
    const open = document.body.classList.toggle("global-nav-open");
    $("#global-nav-toggle").setAttribute("aria-expanded", String(open));
    if (open) { $("#global-nav").focus(); }
  });

  resetFixture();
})();
