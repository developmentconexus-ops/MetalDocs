import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { MemoryRouter } from "react-router-dom";
import { ArtifactHero } from "./ArtifactHero";

describe("ArtifactHero", () => {
  it("renders the title as an h1", () => {
    render(
      <ArtifactHero
        breadcrumb={[{ label: "Home" }]}
        title="My Artifact Title"
      />,
    );
    expect(
      screen.getByRole("heading", { level: 1, name: "My Artifact Title" }),
    ).toBeTruthy();
  });

  it("renders each breadcrumb label", () => {
    render(
      <MemoryRouter>
        <ArtifactHero
          breadcrumb={[
            { label: "Biblioteca", href: "/documents" },
            { label: "RH", href: "/rh" },
            { label: "DC-RH-001" },
          ]}
          title="Title"
        />
      </MemoryRouter>,
    );
    expect(screen.getByText("Biblioteca")).toBeTruthy();
    expect(screen.getByText("RH")).toBeTruthy();
    expect(screen.getByText("DC-RH-001")).toBeTruthy();
  });

  it("renders non-last crumb with href as an anchor", () => {
    render(
      <MemoryRouter>
        <ArtifactHero
          breadcrumb={[
            { label: "Biblioteca", href: "/documents" },
            { label: "Last" },
          ]}
          title="Title"
        />
      </MemoryRouter>,
    );
    const link = screen.getByRole("link", { name: "Biblioteca" });
    expect(link.getAttribute("href")).toBe("/documents");
  });

  it("renders the last crumb as plain text even when it has an href", () => {
    render(
      <MemoryRouter>
        <ArtifactHero
          breadcrumb={[
            { label: "Biblioteca", href: "/documents" },
            { label: "Last Crumb", href: "/last" },
          ]}
          title="Title"
        />
      </MemoryRouter>,
    );
    // "Last Crumb" must appear but NOT as a link
    expect(screen.getByText("Last Crumb")).toBeTruthy();
    const links = screen.queryAllByRole("link");
    const lastCrumbLink = links.find(
      (el) => el.textContent === "Last Crumb",
    );
    expect(lastCrumbLink).toBeUndefined();
  });

  it("renders badges slot content when provided", () => {
    render(
      <ArtifactHero
        breadcrumb={[{ label: "Home" }]}
        title="Title"
        badges={<span data-testid="badge-slot">Badge Content</span>}
      />,
    );
    expect(screen.getByTestId("badge-slot")).toBeTruthy();
  });

  it("renders docCard slot content when provided", () => {
    render(
      <ArtifactHero
        breadcrumb={[{ label: "Home" }]}
        title="Title"
        docCard={<span data-testid="doccard-slot">Doc Card</span>}
      />,
    );
    expect(screen.getByTestId("doccard-slot")).toBeTruthy();
  });

  it("renders subtitle slot content when provided", () => {
    render(
      <ArtifactHero
        breadcrumb={[{ label: "Home" }]}
        title="Title"
        subtitle={<span data-testid="subtitle-slot">Subtitle Text</span>}
      />,
    );
    expect(screen.getByTestId("subtitle-slot")).toBeTruthy();
  });

  it("renders actions slot content when provided", () => {
    render(
      <ArtifactHero
        breadcrumb={[{ label: "Home" }]}
        title="Title"
        actions={<span data-testid="actions-slot">Action Button</span>}
      />,
    );
    expect(screen.getByTestId("actions-slot")).toBeTruthy();
  });

  it("does not crash when optional slots are omitted", () => {
    expect(() =>
      render(
        <ArtifactHero breadcrumb={[{ label: "Home" }]} title="Bare Title" />,
      ),
    ).not.toThrow();
    expect(screen.getByRole("heading", { level: 1, name: "Bare Title" })).toBeTruthy();
  });
});
