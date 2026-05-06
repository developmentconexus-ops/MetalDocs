/* global React, ReactDOM, Icons, BlocksPanel, InspectorPanel */
const { useState, useEffect } = React;
const I = Icons;

const TWEAK_DEFAULTS = /*EDITMODE-BEGIN*/{
  "variant": "hero",
  "docWidth": 800,
  "showTweaks": true,
  "accent": "wine"
}/*EDITMODE-END*/;

/* ---------- Topbar ---------- */
function Topbar({ onHamburger }) {
  return (
    <div className="topbar">
      <div className="topbar-left">
        <div className="brand-dot"><I.file size={12} /></div>
        <div className="brand-name">MetalDocs</div>
        <button className="icon-btn" onClick={onHamburger} aria-label="Menu"><I.menu size={16} /></button>
        <div className="divider-v" />
        <div className="crumbs">
          <span>Templates</span>
          <span className="sep">/</span>
          <span className="current">Purchase Order</span>
          <span className="version-badge">v1</span>
          <span className="status-pill draft">Draft</span>
        </div>
      </div>
      <div className="topbar-right">
        <span className="saved-indicator">
          <I.check size={14} className="check" /> Saved
        </span>
        <button className="icon-btn" aria-label="History"><I.history size={16} /></button>
        <button className="icon-btn" aria-label="Share"><I.share size={16} /></button>
        <button className="btn-primary"><I.send size={14} /> Submit for Review</button>
        <div className="avatar">LT</div>
      </div>
    </div>
  );
}

/* ---------- Left Rail ---------- */
function LeftRail({ active, onSelect, blocksOpen }) {
  const items = [
    { key: "blocks", icon: "blocks", tip: "Blocks", kbd: "B" },
    { key: "layout", icon: "layoutGrid", tip: "Layout" },
    { key: "image",  icon: "image",  tip: "Media" },
    { divider: true },
    { key: "outline", icon: "listTree", tip: "Outline" },
    { key: "search", icon: "search", tip: "Find", kbd: "⌘F" },
  ];
  return (
    <div className="rail rail-left" style={{ flexDirection: "row" }}>
      <div className="rail-icons">
        {items.map((it, i) => it.divider ? (
          <div key={i} className="rail-divider" />
        ) : (
          <button
            key={it.key}
            className={"rail-icon-btn" + (active === it.key ? " is-active" : "")}
            onClick={() => onSelect(it.key)}
          >
            {React.createElement(I[it.icon], { size: 18 })}
            <span className="tip">
              {it.tip}{it.kbd && <span className="kbd">{it.kbd}</span>}
            </span>
          </button>
        ))}
      </div>
      {blocksOpen && <BlocksPanel onClose={() => onSelect(null)} />}
    </div>
  );
}

/* ---------- Right Rail ---------- */
function RightRail({ active, onSelect, inspectorOpen }) {
  const items = [
    { key: "inspector", icon: "panelRight", tip: "Inspector" },
    { key: "variables", icon: "braces",     tip: "Variables" },
    { key: "comments",  icon: "message",    tip: "Comments" },
    { divider: true },
    { key: "versions",  icon: "gitBranch",  tip: "Versions" },
  ];
  return (
    <div className="rail rail-right" style={{ flexDirection: "row" }}>
      {inspectorOpen && <InspectorPanel onClose={() => onSelect(null)} />}
      <div className="rail-icons">
        {items.map((it, i) => it.divider ? (
          <div key={i} className="rail-divider" />
        ) : (
          <button
            key={it.key}
            className={"rail-icon-btn" + (active === it.key ? " is-active" : "")}
            onClick={() => onSelect(it.key)}
          >
            {React.createElement(I[it.icon], { size: 18 })}
            <span className="tip">{it.tip}</span>
          </button>
        ))}
      </div>
    </div>
  );
}

/* ---------- Toolbar ---------- */
function Toolbar({ formats, toggleFormat, align, setAlign }) {
  return (
    <div className="toolbar">
      <div className="toolbar-group">
        <button className="tb-btn" title="Undo"><I.undo size={15} /></button>
        <button className="tb-btn" title="Redo"><I.redo size={15} /></button>
      </div>
      <div className="toolbar-sep" />

      <div className="toolbar-group">
        <button className="tb-select" title="Paragraph style">
          Heading 1 <I.chevronDown size={12} className="chev" />
        </button>
        <button className="tb-select" title="Font">
          Source Serif <I.chevronDown size={12} className="chev" />
        </button>
        <button className="tb-select" title="Size">
          16 <I.chevronDown size={12} className="chev" />
        </button>
      </div>
      <div className="toolbar-sep" />

      <div className="toolbar-group">
        <button
          className={"tb-btn" + (formats.bold ? " is-active" : "")}
          onClick={() => toggleFormat("bold")}
          title="Bold (⌘B)"
        ><I.bold size={14} /></button>
        <button
          className={"tb-btn" + (formats.italic ? " is-active" : "")}
          onClick={() => toggleFormat("italic")}
          title="Italic (⌘I)"
        ><I.italic size={14} /></button>
        <button
          className={"tb-btn" + (formats.underline ? " is-active" : "")}
          onClick={() => toggleFormat("underline")}
          title="Underline (⌘U)"
        ><I.underline size={14} /></button>
        <button
          className={"tb-btn" + (formats.strike ? " is-active" : "")}
          onClick={() => toggleFormat("strike")}
          title="Strikethrough"
        ><I.strike size={14} /></button>
        <button className="tb-color" title="Text color">
          <I.textColor size={14} />
          <span className="swatch" style={{ background: "#1e293b" }} />
        </button>
        <button className="tb-color" title="Highlight">
          <I.highlight size={14} />
          <span className="swatch" style={{ background: "#fef3c7" }} />
        </button>
      </div>
      <div className="toolbar-sep" />

      <div className="toolbar-group">
        {[
          { k: "left",    icon: "alignLeft" },
          { k: "center",  icon: "alignCenter" },
          { k: "right",   icon: "alignRight" },
          { k: "justify", icon: "alignJustify" },
        ].map(a => (
          <button
            key={a.k}
            className={"tb-btn" + (align === a.k ? " is-active" : "")}
            onClick={() => setAlign(a.k)}
            title={"Align " + a.k}
          >{React.createElement(I[a.icon], { size: 14 })}</button>
        ))}
      </div>
      <div className="toolbar-sep" />

      <div className="toolbar-group">
        <button className="tb-btn" title="Bullet list"><I.listBullet size={14} /></button>
        <button className="tb-btn" title="Numbered list"><I.listNumber size={14} /></button>
        <button className="tb-btn" title="Outdent"><I.outdent size={14} /></button>
        <button className="tb-btn" title="Indent"><I.indent size={14} /></button>
      </div>
      <div className="toolbar-sep" />

      <div className="toolbar-group">
        <button className="tb-btn" title="Link (⌘K)"><I.link size={14} /></button>
        <button className="tb-btn" title="Table"><I.table size={14} /></button>
        <button className="tb-btn" title="Image"><I.imageSm size={14} /></button>
        <button className="tb-btn" title="Insert block"><I.plusSquare size={14} /></button>
      </div>
    </div>
  );
}

/* ---------- Document body variants ---------- */
function DocBody({ variant, align }) {
  if (variant === "empty") {
    return (
      <div className="empty-page">
        <div className="empty-icon"><I.plusSquare size={22} /></div>
        <h2>Start authoring your template</h2>
        <p>Drag reusable blocks from the library on the left — headers, two-column grids, media drops — or start typing to insert paragraphs inline.</p>
        <button className="empty-cta">
          <span className="arrow"><I.arrowLeft size={14} /></span>
          Open blocks library
        </button>
      </div>
    );
  }

  const selected = variant === "inspector";

  return (
    <div style={{ textAlign: align }}>
      <div className={selected ? "selected-block" : undefined} style={{ position: "relative", display: "inline-block", width: "100%" }}>
        {selected && <>
          <span className="handle tl" />
          <span className="handle tr" />
          <span className="handle bl" />
          <span className="handle br" />
        </>}
        <h1>Purchase Order</h1>
      </div>
      <p><span className="bold">Bold test.</span> Hello from in-app DOCX creation!</p>
      <p>
        This template generates a signed purchase order for the MetalDocs SGQ workflow.
        All bracketed fields below resolve against the requisition context at export time —
        vendor name, requester, cost center, and approval chain are pulled from the purchase
        requisition module. Reviewers see the rendered preview before signing.
      </p>
      <h2>Line Items</h2>
      <p>
        Each row of the items table binds to a requisition line. Quantities and unit prices
        roll up into the totals block beneath this section. Currency formatting follows
        the organization locale defined in settings.
      </p>
      <p className="muted">Start typing to continue the template…<span className="caret" /></p>
    </div>
  );
}

/* ---------- Tweaks panel ---------- */
function Tweaks({ tweaks, setTweaks, visible, onHide }) {
  if (!visible) return null;
  return (
    <div className="tweaks">
      <div className="tweaks-head">
        <h5>Tweaks</h5>
        <button onClick={onHide}>Hide</button>
      </div>
      <div className="tweaks-body">
        <div className="tweak-row">
          <label>Variant</label>
          <div className="seg">
            {["hero", "empty", "blocks", "inspector", "dark"].map(v => (
              <button
                key={v}
                className={tweaks.variant === v ? "is-active" : ""}
                onClick={() => setTweaks({ ...tweaks, variant: v })}
              >{v[0].toUpperCase()}</button>
            ))}
          </div>
        </div>
        <div className="tweak-row">
          <label>Doc width</label>
          <input
            type="range" min="640" max="960" step="20"
            value={tweaks.docWidth}
            onChange={(e) => setTweaks({ ...tweaks, docWidth: parseInt(e.target.value) })}
          />
        </div>
      </div>
    </div>
  );
}

/* =========================================================
   APP
   ========================================================= */
function App() {
  const [tweaks, setTweaks] = useState(TWEAK_DEFAULTS);
  const [editMode, setEditMode] = useState(false);

  // Editor state
  const [leftActive, setLeftActive] = useState("blocks");
  const [rightActive, setRightActive] = useState("inspector");
  const [formats, setFormats] = useState({ bold: false, italic: false, underline: false, strike: false });
  const [align, setAlign] = useState("left");

  const toggleFormat = (k) => setFormats(f => ({ ...f, [k]: !f[k] }));

  // Edit mode wiring
  useEffect(() => {
    const onMsg = (e) => {
      const d = e.data || {};
      if (d.type === "__activate_edit_mode") setEditMode(true);
      if (d.type === "__deactivate_edit_mode") setEditMode(false);
    };
    window.addEventListener("message", onMsg);
    window.parent.postMessage({ type: "__edit_mode_available" }, "*");
    return () => window.removeEventListener("message", onMsg);
  }, []);

  const persistTweaks = (next) => {
    setTweaks(next);
    window.parent.postMessage({ type: "__edit_mode_set_keys", edits: next }, "*");
  };

  const variant = tweaks.variant;

  const appClass = ["app"];
  if (variant === "dark") appClass.push("dark");
  const bodyClass = ["body"];
  if (variant === "blocks") bodyClass.push("blocks-open");
  if (variant === "inspector") bodyClass.push("inspector-open");

  const blocksOpen = variant === "blocks";
  const inspectorOpen = variant === "inspector";

  const variantLabels = {
    hero: "Hero · Composed template",
    empty: "Empty state · First open",
    blocks: "Blocks library · Expanded",
    inspector: "Inspector · Header block selected",
    dark: "Dark mode · Hero",
  };

  return (
    <div className="stage">
      <div className="stage-head">
        <div className="stage-title">
          <span className="dot" />
          MetalDocs <span className="sep">/</span> <span className="sub">Template Editor — Purchase Order v1 (Draft)</span>
        </div>
        <div className="stage-variants">
          {[
            { k: "hero", label: "1 · Hero" },
            { k: "empty", label: "2 · Empty" },
            { k: "blocks", label: "3 · Blocks" },
            { k: "inspector", label: "4 · Inspector" },
            { k: "dark", label: "5 · Dark" },
          ].map(v => (
            <button
              key={v.k}
              className={variant === v.k ? "is-active" : ""}
              onClick={() => persistTweaks({ ...tweaks, variant: v.k })}
            >{v.label}</button>
          ))}
        </div>
      </div>

      <div className="stage-meta">
        <span data-screen-label={"0" + (["hero","empty","blocks","inspector","dark"].indexOf(variant)+1) + " " + variantLabels[variant]}>
          1440 × 900 · {variantLabels[variant]}
        </span>
      </div>

      <div className={appClass.join(" ")} data-variant={variant}>
        <Topbar />
        <div className={bodyClass.join(" ")}>
          <LeftRail
            active={blocksOpen ? "blocks" : leftActive}
            onSelect={(k) => {
              if (k === null) persistTweaks({ ...tweaks, variant: "hero" });
              else setLeftActive(k);
            }}
            blocksOpen={blocksOpen}
          />
          <div className="canvas">
            <div className="dochead">
              <div className="left">
                <span>Purchase Order</span>
                <span className="sep">·</span>
                <span className="meta">Template v1 Draft</span>
              </div>
              <div className="right">
                <span>Last edited 2 min ago</span>
                <span className="dot" />
                <span>Leandro Theodoro</span>
              </div>
            </div>
            <Toolbar formats={formats} toggleFormat={toggleFormat} align={align} setAlign={setAlign} />
            <div className="canvas-scroll">
              <div className="page" style={{ width: tweaks.docWidth }}>
                <DocBody variant={variant} align={align} />
              </div>
            </div>
          </div>
          <RightRail
            active={inspectorOpen ? "inspector" : rightActive}
            onSelect={(k) => {
              if (k === null) persistTweaks({ ...tweaks, variant: "hero" });
              else setRightActive(k);
            }}
            inspectorOpen={inspectorOpen}
          />
        </div>
      </div>

      <Tweaks
        tweaks={tweaks}
        setTweaks={persistTweaks}
        visible={editMode}
        onHide={() => setEditMode(false)}
      />
    </div>
  );
}

ReactDOM.createRoot(document.getElementById("root")).render(<App />);
