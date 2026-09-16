const THEME_KEY = 'solvi-theme';
const DEFAULT_THEME = 'solvi';

function currentTheme() {
  return localStorage.getItem(THEME_KEY) || DEFAULT_THEME;
}

function applyTheme(theme) {
  document.documentElement.setAttribute('data-theme', theme);
  localStorage.setItem(THEME_KEY, theme);
}

document.addEventListener('alpine:init', () => {
  Alpine.store('ui', {
    sidebarOpen: true,
    theme: currentTheme(),
    toggleSidebar() {
      this.sidebarOpen = !this.sidebarOpen;
    },
    setTheme(theme) {
      this.theme = theme;
      applyTheme(theme);
    },
  });
});
