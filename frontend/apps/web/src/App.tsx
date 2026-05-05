import { Component } from "react";
import type { ReactNode } from "react";
import { RouterProvider } from "react-router-dom";
import { Toaster } from "sonner";
import { router } from "./app/AppRouter";
import { RootProviders } from "./app/RootProviders";

type AppErrorBoundaryState = {
  hasError: boolean;
  message: string;
};

class AppErrorBoundary extends Component<{ children: ReactNode }, AppErrorBoundaryState> {
  state: AppErrorBoundaryState = {
    hasError: false,
    message: "",
  };

  static getDerivedStateFromError(error: Error): AppErrorBoundaryState {
    return {
      hasError: true,
      message: error.message || "Falha inesperada ao renderizar a interface.",
    };
  }

  componentDidCatch(error: Error): void {
    console.error("MetalDocs UI render error", error);
  }

  render() {
    if (this.state.hasError) {
      return (
        <div className="app-shell">
          <section className="hero-panel stack">
            <strong>Falha ao montar a interface.</strong>
            <p className="hint">{this.state.message}</p>
            <p className="hint">A API respondeu, mas a interface encontrou um dado inesperado durante o render. Recarregue a pagina apos atualizar o frontend local.</p>
          </section>
        </div>
      );
    }
    return this.props.children;
  }
}

export default function App() {
  return (
    <AppErrorBoundary>
      <RootProviders>
        <RouterProvider router={router} />
      </RootProviders>
      <Toaster position="bottom-right" richColors closeButton />
    </AppErrorBoundary>
  );
}
