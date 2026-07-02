import { create } from "zustand";

interface UiStore {
  message: string;
  error: string;
  setMessage: (message: string) => void;
  setError: (error: string) => void;
}

export const useUiStore = create<UiStore>((set) => ({
  message: "",
  error: "",
  setMessage: (message) => set({ message }),
  setError: (error) => set({ error }),
}));
