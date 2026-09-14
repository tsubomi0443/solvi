import { User } from '../model/user.js';

document.addEventListener('alpine:init', () => {
  Alpine.data('solviSetting', () => ({
    loading: false,
    form: { name: '', email: '', icon: '' },
    iconFile: null,

    init() {
      const el = document.getElementById('user-json');
      if (el) {
        const u = User.fromJSON(JSON.parse(el.textContent || '{}'));
        this.form = { name: u.name || '', email: u.email || '', icon: u.icon || '' };
      }
      if (typeof lucide !== 'undefined') lucide.createIcons();
    },

    onIconChange(e) {
      this.iconFile = e.target.files?.[0] || null;
    },

    async save() {
      this.loading = true;
      try {
        const res = await fetch('/api/v1/setting', {
          method: 'PUT',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ name: this.form.name, email: this.form.email }),
        });
        if (!res.ok) {
          window.notice.show({ message: '保存に失敗しました', type: 'error' });
          return;
        }
        if (this.iconFile) {
          const fd = new FormData();
          fd.append('icon', this.iconFile);
          const iconRes = await fetch('/api/v1/setting/icon', { method: 'POST', body: fd });
          if (!iconRes.ok) {
            window.notice.show({ message: 'アイコンの保存に失敗しました', type: 'warning' });
          }
        }
        window.notice.show({ message: '保存しました', type: 'success' });
      } catch {
        window.notice.show({ message: 'サーバへ接続できませんでした', type: 'error' });
      } finally {
        this.loading = false;
      }
    },

    async deleteIcon() {
      const res = await fetch('/api/v1/setting/icon', { method: 'DELETE' });
      if (!res.ok) {
        window.notice.show({ message: 'アイコン削除に失敗しました', type: 'error' });
        return;
      }
      this.form.icon = '';
      window.notice.show({ message: 'アイコンを削除しました', type: 'success' });
    },
  }));
});
