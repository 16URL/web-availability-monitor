// Shapes mirror the Go structs, which are serialized without json tags.
export type Site = { ID: number; Alias: string; URL: string };
export type Check = {
  ID: number;
  HTTPStatus: number;
  ResponseTimeMs: number;
  CheckedAt: string;
};

export const API_URL: string =
  import.meta.env.PUBLIC_API_URL ?? "http://localhost:8082";

async function request(path: string, init?: RequestInit): Promise<Response> {
  const res = await fetch(API_URL + path, {
    ...init,
    signal: AbortSignal.timeout(15_000),
  });
  if (!res.ok) throw new Error((await res.text()).trim() || `HTTP ${res.status}`);
  return res;
}

const enc = encodeURIComponent;

export const getSites = async (): Promise<Site[]> =>
  (await (await request("/urls")).json()) ?? [];

export const getChecks = async (alias: string): Promise<Check[]> =>
  (await (await request(`/urls/${enc(alias)}/checks`)).json()) ?? [];

export const addSite = (alias: string, url: string) =>
  request("/urls", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ alias, url }),
  });

export const deleteSite = (alias: string) =>
  request(`/urls/${enc(alias)}`, { method: "DELETE" });

export const checkNow = (alias: string) => request(`/urls/${enc(alias)}/check`);

export function errorMessage(e: unknown): string {
  if (e instanceof DOMException && e.name === "TimeoutError") {
    return "the server did not answer in time.";
  }
  if (e instanceof TypeError) {
    return `cannot reach ${API_URL}. Check that the Go server is running and the page is opened at http://localhost:4321.`;
  }
  return e instanceof Error ? e.message : String(e);
}

export const isUp = (c: Check) => c.HTTPStatus >= 200 && c.HTTPStatus < 400;

export function describe(c?: Check): { cls: "up" | "down" | "none"; label: string } {
  if (!c) return { cls: "none", label: "No checks" };
  return isUp(c) ? { cls: "up", label: "Up" } : { cls: "down", label: "Down" };
}

const timeFormat = new Intl.DateTimeFormat(undefined, {
  dateStyle: "short",
  timeStyle: "medium",
});

export function formatTime(iso: string): string {
  const d = new Date(iso);
  return Number.isNaN(d.getTime()) ? iso : timeFormat.format(d);
}

export const byId = <T extends HTMLElement>(id: string) =>
  document.getElementById(id) as T;

// Builds nodes with textContent only, so API data is never parsed as HTML.
export function el<K extends keyof HTMLElementTagNameMap>(
  tag: K,
  attrs: Record<string, string> = {},
  ...children: (Node | string)[]
): HTMLElementTagNameMap[K] {
  const node = document.createElement(tag);
  for (const [k, v] of Object.entries(attrs)) node.setAttribute(k, v);
  node.append(...children);
  return node;
}

// Shows a message in a status/alert paragraph; empty text hides it.
export function setNote(node: HTMLElement, text: string, kind: "" | "error" | "ok" = "") {
  node.textContent = text;
  node.hidden = !text;
  node.className = node.className.replace(/\b(error|ok)\b/g, "").trim();
  if (kind) node.classList.add(kind);
  node.setAttribute("role", kind === "error" ? "alert" : "status");
}
