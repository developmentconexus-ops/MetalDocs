import { describe, expect, it } from "vitest";
import { matchRoutes } from "react-router-dom";
import { templatesRoutes } from "../routes";

describe("templatesRoutes", () => {
  it("mounts the list route at /templates without a catch-all self-redirect", () => {
    expect(templatesRoutes.filter((route) => route.path === "templates")).toHaveLength(1);
    expect(templatesRoutes.some((route) => route.path === "templates/*")).toBe(false);

    const matches = matchRoutes(templatesRoutes, "/templates");
    expect(matches).not.toBeNull();
    expect(matches?.at(-1)?.route.path).toBe("templates");
  });

  it("keeps the explicit child routes for new and editor screens", () => {
    expect(matchRoutes(templatesRoutes, "/templates/new")?.at(-1)?.route.path).toBe("templates/new");
    expect(matchRoutes(templatesRoutes, "/templates/abc/versions/2")?.at(-1)?.route.path).toBe(
      "templates/:templateId/versions/:versionNum",
    );
  });
});
