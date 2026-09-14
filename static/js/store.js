document.addEventListener('alpine:init', () => {
  Alpine.store('ui', {
    sidebarOpen: true,
    toggleSidebar() {
      this.sidebarOpen = !this.sidebarOpen;
    },
  });
});
