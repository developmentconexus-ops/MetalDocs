/* global React, Icons */

/* =========================================================
   BLOCKS LIBRARY PANEL
   ========================================================= */
function BlocksPanel({ onClose }) {
  const sections = [
    { title: "Editorial Blocks", count: 4, items: [
      { label: "Header Block", meta: "h1 + meta", icon: "heading" },
      { label: "Section Title", meta: "h2", icon: "heading" },
      { label: "Signature Line", meta: "auth · role", icon: "signature" },
      { label: "Quote / Note", meta: "callout", icon: "quote" },
    ]},
    { title: "Layout & Structure", count: 4, items: [
      { label: "Two-Column Grid", meta: "1fr 1fr", icon: "columns" },
      { label: "Three-Column Grid", meta: "1fr 1fr 1fr", icon: "gridThree" },
      { label: "Divider", meta: "horizontal rule", icon: "divider" },
      { label: "Variable Field", meta: "{{ slot }}", icon: "variable" },
    ]},
    { title: "Media & Interactivity", count: 3, items: [
      { label: "Media Drop Box", meta: "image · pdf", icon: "mediaDrop" },
      { label: "Image", meta: "inline", icon: "imageSm" },
      { label: "Approval Stamp", meta: "ISO 9001", icon: "stamp" },
    ]},
  ];

  return (
    <div className="rail-panel">
      <div className="rail-panel-head">
        <Icons.blocks size={14} />
        <h3>Blocks Library</h3>
        <button className="close" onClick={onClose} aria-label="Collapse">
          <Icons.close size={14} />
        </button>
      </div>
      <div className="rail-panel-body">
        <div className="blocks-search">
          <span className="ic"><Icons.search size={12} /></span>
          <input placeholder="Search blocks…" />
        </div>

        {sections.map((sec) => (
          <div className="block-section" key={sec.title}>
            <div className="block-section-head">
              <h4>{sec.title}</h4>
              <span className="count">{sec.count}</span>
            </div>
            <div className="block-grid">
              {sec.items.map((b) => {
                const Icn = I[b.icon];
                return (
                  <div className="block-card" key={b.label} draggable>
                    <div className="preview"><Icn size={18} /></div>
                    <div className="label">{b.label}</div>
                    <div className="meta">{b.meta}</div>
                  </div>
                );
              })}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

/* =========================================================
   INSPECTOR PANEL
   ========================================================= */
function InspectorPanel({ onClose }) {
  const [stroke, setStroke] = React.useState("solid");
  const [thickness, setThickness] = React.useState(1);
  const [surface, setSurface] = React.useState(1);

  const swatches = ["#ffffff", "#f8fafc", "#fef3c7", "#dcfce7", "#dbeafe", "#f3e6e8", "#1e293b"];

  return (
    <div className="rail-panel">
      <div className="rail-panel-head">
        <Icons.panelRight size={14} />
        <h3>Inspector</h3>
        <button className="close" onClick={onClose} aria-label="Collapse">
          <Icons.close size={14} />
        </button>
      </div>
      <div className="rail-panel-body">
        <div className="inspector-selection">
          <div className="chip"><Icons.heading size={12} /></div>
          <div className="name">Header Block</div>
          <div className="id">#hdr-01</div>
        </div>

        <div className="ins-section">
          <div className="ins-section-head">
            <h4>Dimensions</h4>
            <button className="reset">Reset</button>
          </div>
          <div className="ins-row">
            <div className="ins-field">
              <label>Width</label>
              <div className="ins-input-wrap">
                <span className="prefix">W</span>
                <input defaultValue="656" />
                <span className="unit">px</span>
              </div>
            </div>
            <div className="ins-field">
              <label>Height</label>
              <div className="ins-input-wrap">
                <span className="prefix">H</span>
                <input defaultValue="auto" />
                <span className="unit">—</span>
              </div>
            </div>
          </div>
        </div>

        <div className="ins-section">
          <div className="ins-section-head">
            <h4>Surface Color</h4>
          </div>
          <div className="swatch-row">
            {swatches.map((s, i) => (
              <div
                key={i}
                className={"swatch" + (i === surface ? " is-active" : "")}
                style={{ background: s }}
                onClick={() => setSurface(i)}
              />
            ))}
          </div>
        </div>

        <div className="ins-section">
          <div className="ins-section-head">
            <h4>Stroke & Border</h4>
          </div>
          <div className="stroke-pick">
            {["solid","dashed","double"].map((t) => (
              <button
                key={t}
                className={"stroke-opt" + (stroke === t ? " is-active" : "")}
                onClick={() => setStroke(t)}
              >
                <span className={"line-" + t} />
              </button>
            ))}
          </div>
          <div className="slider-row">
            <input
              type="range" min="0" max="4" step="0.5"
              value={thickness}
              onChange={(e) => setThickness(parseFloat(e.target.value))}
            />
            <span className="val">{thickness.toFixed(1)}px</span>
          </div>
        </div>

        <div className="ins-section">
          <div className="ins-section-head">
            <h4>Spacing</h4>
          </div>
          <div className="spacing-box">
            <span className="spacing-label" style={{ top: -7, left: 12 }}>MARGIN</span>
            <div style={{ position: "absolute", top: 4, left: "50%", transform: "translateX(-50%)" }}>
              <input className="spacing-input" defaultValue="16" />
            </div>
            <div style={{ position: "absolute", left: 4, top: "50%", transform: "translateY(-50%)" }}>
              <input className="spacing-input" defaultValue="0" />
            </div>
            <div style={{ position: "absolute", right: 4, top: "50%", transform: "translateY(-50%)" }}>
              <input className="spacing-input" defaultValue="0" />
            </div>
            <div style={{ position: "absolute", bottom: 4, left: "50%", transform: "translateX(-50%)" }}>
              <input className="spacing-input" defaultValue="24" />
            </div>
            <div className="spacing-inner">
              <span className="spacing-label" style={{ top: -7, left: 12, background: "var(--surface)" }}>PADDING</span>
              <div className="ins-row" style={{ marginBottom: 0 }}>
                <div className="ins-field">
                  <label>Top</label>
                  <div className="ins-input-wrap"><input defaultValue="12" /><span className="unit">px</span></div>
                </div>
                <div className="ins-field">
                  <label>Bottom</label>
                  <div className="ins-input-wrap"><input defaultValue="12" /><span className="unit">px</span></div>
                </div>
              </div>
              <div className="ins-row" style={{ marginBottom: 0, marginTop: 8 }}>
                <div className="ins-field">
                  <label>Left</label>
                  <div className="ins-input-wrap"><input defaultValue="20" /><span className="unit">px</span></div>
                </div>
                <div className="ins-field">
                  <label>Right</label>
                  <div className="ins-input-wrap"><input defaultValue="20" /><span className="unit">px</span></div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

Object.assign(window, { BlocksPanel, InspectorPanel });
