// Small glue for HTMX and Alpine. Keep this file short; prefer HTMX attributes
// and Alpine x-data in templates.

// Light/dark switch in the sidebar. The saved choice is applied early by an
// inline script in layouts/base; without one, the OS setting decides.
const THEME_LIGHT = "utility";
const THEME_DARK = "utility-dark";

function currentTheme() {
  const set = document.documentElement.getAttribute("data-theme");
  if (set) return set;
  return window.matchMedia("(prefers-color-scheme: dark)").matches ? THEME_DARK : THEME_LIGHT;
}

document.addEventListener("click", (event) => {
  if (!event.target.closest("[data-theme-toggle]")) return;
  const next = currentTheme() === THEME_DARK ? THEME_LIGHT : THEME_DARK;
  document.documentElement.setAttribute("data-theme", next);
  try {
    localStorage.setItem("theme", next);
  } catch (e) {
    // Storage blocked (private mode): the switch still works for this page.
  }
});

// Drag-and-drop file picker (templates/components/dropzone.html).
document.addEventListener("alpine:init", () => {
  Alpine.data("dropzone", () => ({
    name: "",
    size: "",
    over: false,

    pick() {
      const file = this.$refs.input.files[0];
      this.name = file ? file.name : "";
      this.size = file ? formatBytes(file.size) : "";
    },

    drop(event) {
      this.over = false;
      if (!event.dataTransfer.files.length) return;
      this.$refs.input.files = event.dataTransfer.files;
      this.pick();
    },
  }));
});

function formatBytes(n) {
  if (n < 1024) return `${n} B`;
  if (n < 1024 * 1024) return `${Math.round(n / 1024)} KB`;
  return `${(n / 1024 / 1024).toFixed(1)} MB`;
}

// Upload progress for forms that contain <progress data-upload-progress>.
document.addEventListener("htmx:xhr:progress", (event) => {
  const bar = event.target.querySelector?.("[data-upload-progress]");
  const { loaded, total } = event.detail;
  if (!bar || !total) return;
  bar.hidden = false;
  bar.value = Math.round((loaded / total) * 100);
});

document.addEventListener("htmx:afterRequest", (event) => {
  const bar = event.target.querySelector?.("[data-upload-progress]");
  if (!bar) return;
  bar.hidden = true;
  bar.value = 0;
});

// On small screens the result sits below the form; bring it into view.
document.addEventListener("htmx:afterSwap", (event) => {
  const target = event.detail.target;
  if (target.id !== "result") return;
  const rect = target.getBoundingClientRect();
  if (rect.top < 0 || rect.bottom > window.innerHeight) {
    const smooth = !window.matchMedia("(prefers-reduced-motion: reduce)").matches;
    target.scrollIntoView({ behavior: smooth ? "smooth" : "auto", block: "nearest" });
  }
});

// The server couldn't be reached at all (offline, proxy down, upload too big for nginx).
document.addEventListener("htmx:sendError", showConnectionError);
document.addEventListener("htmx:timeout", showConnectionError);

function showConnectionError(event) {
  const target = event.detail.target;
  if (!target) return;
  target.innerHTML = `
    <div role="alert" class="alert alert-error alert-soft items-start">
      <span>Couldn't reach the server. Check your connection and try again.</span>
    </div>`;
}
