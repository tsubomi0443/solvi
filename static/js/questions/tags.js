import { Tag } from '../model/tag.js';

document.addEventListener('alpine:init', () => {
  Alpine.data('solviTags', () => ({
    tags: [],
    filter: '',
    editing: '',
    editName: '',

    init() {
      const el = document.getElementById('tags-json');
      if (el) this.tags = JSON.parse(el.textContent || '[]').map((dto) => Tag.fromJSON(dto));
      if (typeof lucide !== 'undefined') lucide.createIcons();
    },

    filteredTags() {
      const q = this.filter.trim().toLowerCase();
      if (!q) return this.tags;
      return this.tags.filter((t) => t.name.toLowerCase().includes(q));
    },

    startEdit(name) {
      this.editing = name;
      this.editName = name;
    },

    async saveRename(from) {
      const res = await fetch('/api/v1/tags', {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ from, to: this.editName }),
      });
      if (!res.ok) {
        window.notice.show({ message: '更新に失敗しました', type: 'error' });
        return;
      }
      this.editing = '';
      await this.reload();
      window.notice.show({ message: 'タグ名を更新しました', type: 'success' });
    },

    async deleteTag(name) {
      if (!confirm(`「${name}」をすべての質問から削除しますか？`)) return;
      const res = await fetch('/api/v1/tags', {
        method: 'DELETE',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ name }),
      });
      if (!res.ok) {
        window.notice.show({ message: '削除に失敗しました', type: 'error' });
        return;
      }
      await this.reload();
      window.notice.show({ message: 'タグを削除しました', type: 'success' });
    },

    async reload() {
      const res = await fetch('/api/v1/tags');
      if (!res.ok) return;
      this.tags = (await res.json()).map((dto) => Tag.fromJSON(dto));
    },
  }));
});
