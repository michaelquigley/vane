import { defaultUrlTransform } from "react-markdown";

// relative urls in item bodies resolve against the roadmap directory — the
// reading obsidian and github give them — via the server's read-only
// /roadmap/ route. absolute urls, root-relative paths, and fragments pass
// through untouched; react-markdown's default transform keeps its
// dangerous-protocol sanitization either way.
export function roadmapUrl(url: string): string {
  if (/^[a-z][a-z0-9+.-]*:/i.test(url) || url.startsWith("/") || url.startsWith("#")) {
    return defaultUrlTransform(url);
  }
  return defaultUrlTransform(`/roadmap/${url.replace(/^\.\//, "")}`);
}
