import { User } from '../model/user.js';

document.addEventListener('alpine:init', () => {
  Alpine.data('solviManagementUsers', () => ({
    users: [],

    init() {
      const el = document.getElementById('users-json');
      if (el) this.users = JSON.parse(el.textContent || '[]').map((dto) => User.fromJSON(dto));
      if (typeof lucide !== 'undefined') lucide.createIcons();
    },

    async toggleSupporter(user) {
      const next = !user.isSupporter;
      const res = await fetch(`/api/v1/management/users/${user.uuid}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ isSupporter: next }),
      });
      if (!res.ok) {
        window.notice.show({ message: '更新に失敗しました', type: 'error' });
        return;
      }
      user.isSupporter = next;
      window.notice.show({ message: '更新しました', type: 'success' });
    },
  }));
});
