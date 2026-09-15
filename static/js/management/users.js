import { User } from '../model/user.js';

document.addEventListener('alpine:init', () => {
  Alpine.data('solviManagementUsers', () => ({
    users: [],
    currentUserUUID: window.solviCurrentUserUUID || '',

    init() {
      const el = document.getElementById('users-json');
      if (el) this.users = JSON.parse(el.textContent || '[]').map((dto) => User.fromJSON(dto));
      if (typeof lucide !== 'undefined') lucide.createIcons();
    },

    isSelf(user) {
      return Boolean(this.currentUserUUID && user.uuid === this.currentUserUUID);
    },

    async updateUser(user, payload) {
      const res = await fetch(`/api/v1/management/users/${user.uuid}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload),
      });
      if (!res.ok) {
        const msg = await res.json().catch(() => ({ error: '更新に失敗しました' }));
        window.notice.show({ message: msg.error || '更新に失敗しました', type: 'error' });
        return false;
      }
      window.notice.show({
        message: '更新しました。対象ユーザは再ログイン後に反映されます。',
        type: 'success',
      });
      return true;
    },

    async toggleSupporter(user) {
      const next = !user.isSupporter;
      const ok = await this.updateUser(user, { isSupporter: next });
      if (ok) user.isSupporter = next;
    },

    async toggleAdmin(user) {
      const next = !user.isAdmin;
      const ok = await this.updateUser(user, { isAdmin: next });
      if (ok) user.isAdmin = next;
    },
  }));
});
