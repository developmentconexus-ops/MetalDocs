import React from "react";
import ReactDOM from "react-dom/client";
import App from "../App";
import { initFeatureFlags } from "../features/featureFlags";

// Fetch server-controlled feature flags before first render.
// .finally() ensures a network error still mounts the app (defaults apply).
initFeatureFlags().finally(() => {
  ReactDOM.createRoot(document.getElementById("root")!).render(
    <React.StrictMode>
      <App />
    </React.StrictMode>,
  );
});
