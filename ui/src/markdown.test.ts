import { describe, expect, it } from "vitest";
import { roadmapUrl } from "./markdown";

describe("roadmapUrl", () => {
  it("resolves relative paths against /roadmap/", () => {
    expect(roadmapUrl("images/pic.png")).toBe("/roadmap/images/pic.png");
  });

  it("strips a leading ./", () => {
    expect(roadmapUrl("./images/pic.png")).toBe("/roadmap/images/pic.png");
  });

  it("leaves absolute urls alone", () => {
    expect(roadmapUrl("https://example.com/pic.png")).toBe("https://example.com/pic.png");
  });

  it("leaves root-relative paths and fragments alone", () => {
    expect(roadmapUrl("/already/rooted.png")).toBe("/already/rooted.png");
    expect(roadmapUrl("#section")).toBe("#section");
  });

  it("keeps the default dangerous-protocol sanitization", () => {
    expect(roadmapUrl("javascript:alert(1)")).toBe("");
  });
});
