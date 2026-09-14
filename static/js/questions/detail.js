import { Question } from '../model/question.js';
import { User } from '../model/user.js';

function statusLabel(s) {
  return { pending: '未対応', supporting: '対応中', done: '完了' }[s] || s;
}

function statusBadge(s) {
  return { pending: 'badge-ghost', supporting: 'badge-warning', done: 'badge-success' }[s] || 'badge-ghost';
}

document.addEventListener('alpine:init', () => {
  Alpine.data('solviDetail', () => ({
    question: {},
    currentUser: {},
    isSupporter: Boolean(window.solviIsSupporter),
    composerText: '',

    init() {
      const qEl = document.getElementById('question-json');
      const uEl = document.getElementById('user-json');
      if (qEl) this.question = Question.fromJSON(JSON.parse(qEl.textContent || '{}'));
      if (uEl) this.currentUser = User.fromJSON(JSON.parse(uEl.textContent || '{}'));
      if (typeof lucide !== 'undefined') lucide.createIcons();
      ['create-content', 'create-answer', 'create-memo', 'create-refer', 'update-question'].forEach((ev) => {
        document.addEventListener(ev, (e) => {
          if (e.detail?.uuid === this.question.uuid) this.fetchQuestion();
        });
      });
    },

    initial(name) {
      return (name || '?').slice(0, 1);
    },

    firstContent() {
      const contents = this.question.contents || [];
      return contents.length ? contents[0] : null;
    },

    chatTimelineItems() {
      const items = [];
      const contents = this.question.contents || [];
      contents.slice(1).forEach((c) => {
        items.push({ ...c, kind: 'content' });
      });
      (this.question.answers || []).forEach((a) => {
        items.push({ ...a, kind: 'answer', refers: this.refersForAnswer(a) });
      });
      if (this.isSupporter) {
        (this.question.memos || []).forEach((m) => {
          items.push({ ...m, kind: 'memo' });
        });
      }
      return items.sort((a, b) => new Date(a.createdAt) - new Date(b.createdAt));
    },

    refersForAnswer(answer) {
      return (this.question.refers || []).filter(() => false);
    },

    isSelfItem(item) {
      if (item.kind === 'content') {
        return this.currentUser.uuid && this.question.questionUserUuid === this.currentUser.uuid;
      }
      return item.userName === this.currentUser.name;
    },

    chatBubbleClass(kind) {
      if (kind === 'memo') return 'chat-bubble-accent';
      if (kind === 'content') return 'chat-bubble-secondary';
      return '';
    },

    formatDate(v) {
      if (!v) return '';
      try { return new Date(v).toLocaleString('ja-JP'); } catch { return v; }
    },

    canComplete() {
      if (this.question.supportStatus === 'done') return false;
      if (this.isSupporter) return true;
      return !this.question.isRequireHumanSupport;
    },

    async fetchQuestion() {
      const res = await fetch(`/api/v1/questions/${this.question.uuid}`);
      if (!res.ok) return;
      this.question = Question.fromJSON(await res.json());
      this.$nextTick(() => {
        if (typeof lucide !== 'undefined') lucide.createIcons();
      });
    },

    async appendContent() {
      const res = await fetch(`/api/v1/questions/${this.question.uuid}/contents`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ content: this.composerText }),
      });
      if (!res.ok) {
        window.notice.show({ message: '追記に失敗しました', type: 'error' });
        return;
      }
      this.composerText = '';
      await this.fetchQuestion();
    },

    async addAnswer() {
      const res = await fetch(`/api/v1/questions/${this.question.uuid}/answers`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ content: this.composerText, refers: [] }),
      });
      if (!res.ok) {
        window.notice.show({ message: '回答の登録に失敗しました', type: 'error' });
        return;
      }
      this.composerText = '';
      await this.fetchQuestion();
    },

    async addMemo() {
      const res = await fetch(`/api/v1/questions/${this.question.uuid}/memos`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ content: this.composerText }),
      });
      if (!res.ok) {
        window.notice.show({ message: 'メモの追加に失敗しました', type: 'error' });
        return;
      }
      this.composerText = '';
      await this.fetchQuestion();
    },

    async completeQuestion() {
      const res = await fetch(`/api/v1/questions/${this.question.uuid}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ complete: true }),
      });
      if (!res.ok) {
        window.notice.show({ message: '完了処理に失敗しました', type: 'error' });
        return;
      }
      this.question = Question.fromJSON(await res.json());
      window.notice.show({ message: '質問を完了しました', type: 'success' });
    },

    statusLabel,
    statusBadge,
  }));
});
