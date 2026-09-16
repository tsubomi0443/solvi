document.addEventListener('alpine:init', () => {
  Alpine.data('solviManagementLogs', () => ({
    dates: [],
    selectedDate: '',
    showModal: false,
    modalPhase: 'preparing',
    issuing: false,
    saving: false,
    issueMode: '',
    downloadKey: '',
    password: '',
    zipFilename: 'solvi-logs.zip',

    get hasDates() {
      return this.dates.length > 0;
    },

    init() {
      const el = document.getElementById('log-dates-json');
      if (el) {
        this.dates = JSON.parse(el.textContent || '[]');
        if (this.dates.length > 0) {
          this.selectedDate = this.dates[0];
        }
      }
      if (typeof lucide !== 'undefined') lucide.createIcons();
    },

    openPreparingModal(mode) {
      this.issueMode = mode;
      this.modalPhase = 'preparing';
      this.downloadKey = '';
      this.password = '';
      this.showModal = true;
      if (typeof lucide !== 'undefined') lucide.createIcons();
    },

    closeModal() {
      if (this.issuing || this.saving) return;
      this.showModal = false;
      this.modalPhase = 'preparing';
      this.downloadKey = '';
      this.password = '';
      this.issueMode = '';
      this.zipFilename = 'solvi-logs.zip';
    },

    async startDownloadAll() {
      if (this.issuing || this.saving) return;
      this.zipFilename = 'solvi-logs-all.zip';
      this.openPreparingModal('all');
      this.issuing = true;
      try {
        const res = await fetch('/api/v1/log/download/all', { method: 'POST' });
        if (!res.ok) {
          const msg = await this.parseError(res);
          window.notice.show({ message: msg, type: 'error' });
          this.closeModal();
          return;
        }
        const data = await res.json();
        this.applyIssueResponse(data);
        this.modalPhase = 'ready';
        if (typeof lucide !== 'undefined') lucide.createIcons();
      } catch {
        window.notice.show({ message: 'ダウンロード準備に失敗しました', type: 'error' });
        this.closeModal();
      } finally {
        this.issuing = false;
      }
    },

    async startDownloadByDate() {
      if (this.issuing || this.saving || !this.selectedDate) return;
      const compact = this.selectedDate.replace(/-/g, '');
      this.zipFilename = `solvi-logs-${compact}.zip`;
      this.openPreparingModal('date');
      this.issuing = true;
      try {
        const res = await fetch(
          `/api/v1/log/download/date?date=${encodeURIComponent(this.selectedDate)}`,
          { method: 'POST' },
        );
        if (!res.ok) {
          const msg = await this.parseError(res);
          window.notice.show({ message: msg, type: 'error' });
          this.closeModal();
          return;
        }
        const data = await res.json();
        this.applyIssueResponse(data);
        this.modalPhase = 'ready';
        if (typeof lucide !== 'undefined') lucide.createIcons();
      } catch {
        window.notice.show({ message: 'ダウンロード準備に失敗しました', type: 'error' });
        this.closeModal();
      } finally {
        this.issuing = false;
      }
    },

    applyIssueResponse(data) {
      this.password = data.password || '';
      this.downloadKey = data.downloadKey || data.key || '';
      if (data.filename) {
        this.zipFilename = data.filename;
      }
    },

    resolveDownloadUrl() {
      if (!this.downloadKey) return '';
      return `/api/v1/log/download/${encodeURIComponent(this.downloadKey)}`;
    },

    saveZip() {
      const url = this.resolveDownloadUrl();
      if (!url || this.saving) return;
      this.saving = true;

      const anchor = document.createElement('a');
      anchor.href = url;
      anchor.download = this.zipFilename;
      anchor.style.display = 'none';
      document.body.appendChild(anchor);
      anchor.click();
      document.body.removeChild(anchor);

      window.notice.show({ message: 'ZIP ファイルの保存を開始しました', type: 'success' });
      this.saving = false;
    },

    async copyPassword() {
      if (!this.password) return;
      try {
        await navigator.clipboard.writeText(this.password);
        window.notice.show({ message: 'パスワードをコピーしました', type: 'success' });
      } catch {
        window.notice.show({ message: 'コピーに失敗しました', type: 'error' });
      }
    },

    async parseError(res) {
      try {
        const data = await res.json();
        return data.error || '処理に失敗しました';
      } catch {
        return '処理に失敗しました';
      }
    },
  }));
});
