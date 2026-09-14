document.addEventListener('alpine:init', () => {
  Alpine.data('solviLogin', () => ({
    method: 'ldap',
    email: '',
    password: '',
    loading: false,

    init() {
      if (typeof lucide !== 'undefined') lucide.createIcons();
    },

    async doLogin() {
      this.loading = true;
      try {
        const res = await fetch('/api/v1/login', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            method: this.method,
            email: this.email,
            password: this.password,
          }),
        });
        if (!res.ok) {
          const msg = await res.json().catch(() => ({ error: 'ログインに失敗しました' }));
          window.notice.show({ message: msg.error || 'ログインに失敗しました', type: 'error' });
          return;
        }
        location.href = '/';
      } catch {
        window.notice.show({ message: 'サーバへ接続できませんでした', type: 'error' });
      } finally {
        this.loading = false;
      }
    },
  }));
});
