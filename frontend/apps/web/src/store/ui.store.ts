import { create } from "zustand";
import type { WorkspaceView } from "../components/DocumentWorkspaceShell";

interface UiStore {
  message: string;
  error: string;
  isCreateSubmitting: boolean;
  activeView: WorkspaceView;
  pendingViewNavigation: WorkspaceView | null;
  searchQuery: string;
  setMessage: (message: string) => void;
  setError: (error: string) => void;
  setIsCreateSubmitting: (isCreateSubmitting: boolean) => void;
  setActiveView: (activeView: WorkspaceView) => void;
  requestViewNavigation: (activeView: WorkspaceView) => void;
  clearPendingViewNavigation: () => void;
  setSearchQuery: (searchQuery: string) => void;
}

export const useUiStore = create<UiStore>((set) => ({
  message: "",
  error: "",
  isCreateSubmitting: false,
  activeView: "operations",
  pendingViewNavigation: null,
  searchQuery: "",
  setMessage: (message) => set({ message }),
  setError: (error) => set({ error }),
  setIsCreateSubmitting: (isCreateSubmitting) => set({ isCreateSubmitting }),
  setActiveView: (activeView) => set({ activeView }),
  requestViewNavigation: (activeView) => set({ activeView, pendingViewNavigation: activeView }),
  clearPendingViewNavigation: () => set({ pendingViewNavigation: null }),
  setSearchQuery: (searchQuery) => set({ searchQuery }),
}));
