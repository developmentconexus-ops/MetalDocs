import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { ArtifactDecisionPanel } from "./ArtifactDecisionPanel";
import type { ArtifactDecisionModel } from "./types";

function makeModel(overrides: Partial<ArtifactDecisionModel> = {}): ArtifactDecisionModel {
  return {
    kicker: "Decisão requerida",
    heading: "Sua aprovação",
    description: "Registre sua decisão.",
    options: [
      {
        key: "approve",
        label: "Aprovar",
        description: "Aprova o documento.",
        tone: "approve",
        submitLabel: "Assinar e aprovar",
        requiresReason: false,
      },
      {
        key: "reject",
        label: "Devolver",
        description: "Devolve o documento.",
        tone: "reject",
        submitLabel: "Assinar e devolver",
        requiresReason: true,
      },
    ],
    reasonLabel: "Justificativa",
    reasonPlaceholder: "Comentário…",
    password: null,
    legal: null,
    signer: null,
    submit: vi.fn().mockResolvedValue(undefined),
    ...overrides,
  };
}

describe("ArtifactDecisionPanel", () => {
  it("preselects no option on mount", () => {
    render(<ArtifactDecisionPanel model={makeModel()} />);
    for (const radio of screen.getAllByRole("radio")) {
      expect(radio).toHaveAttribute("aria-checked", "false");
    }
  });

  it("disables the submit button until the user chooses an option", () => {
    render(<ArtifactDecisionPanel model={makeModel()} />);
    expect(screen.getByRole("button", { name: "Selecione uma decisão" })).toBeDisabled();
  });
});
