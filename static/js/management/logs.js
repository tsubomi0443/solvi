document.addEventListener('alpine:init', () => {
  Alpine.data('solviManagementLogs', () => ({
    dates: [],
    selectedDate: '',
    showModal: false,
    modalPhase: 'encrypting',
    issuing: false,
    saving: false,
    issueMode: '',
    downloadKey: '',
    downloadLink: '',
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

    openEncryptingModal(mode) {
      this.issueMode = mode;
      this.modalPhase = 'encrypting';
      this.downloadKey = '';
      this.downloadLink = '';
      this.password = '';
      this.showModal = true;
      if (typeof lucide !== 'undefined') lucide.createIcons();
    },

    closeModal() {
      if (this.issuing || this.saving) return;
      this.showModal = false;
      this.modalPhase = 'encrypting';
      this.downloadKey = '';
      this.downloadLink = '';
      this.password = '';
      this.issueMode = '';
      this.zipFilename = 'solvi-logs.zip';
    },

    async startDownloadAll() {
      if (this.issuing || this.saving) return;
      this.zipFilename = 'solvi-logs-all.zip';
      this.openEncryptingModal('all');
      this.issuing = true;
      try {
        const res = await fetch('/api/v1/log/all');
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
        window.notice.show({ message: '暗号化の開始に失敗しました', type: 'error' });
        this.closeModal();
      } finally {
        this.issuing = false;
      }
    },

    async startDownloadByDate() {
      if (this.issuing || this.saving || !this.selectedDate) return;
      const compact = this.selectedDate.replace(/-/g, '');
      this.zipFilename = `solvi-logs-${compact}.zip`;
      this.openEncryptingModal('date');
      this.issuing = true;
      try {
        const res = await fetch(`/api/v1/log/date/?date=${encodeURIComponent(this.selectedDate)}`);
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
        window.notice.show({ message: '暗号化の開始に失敗しました', type: 'error' });
        this.closeModal();
      } finally {
        this.issuing = false;
      }
    },

    applyIssueResponse(data) {
      this.password = data.password || '';
      this.downloadKey = data.key || data.uuid || '';
      this.downloadLink = data.downloadLink || '';
    },

    resolveDownloadUrl() {
      if (this.downloadLink) return this.downloadLink;
      if (this.downloadKey) return `/api/v1/log/download/${encodeURIComponent(this.downloadKey)}`;
      return '';
    },

    async saveZip() {
      const url = this.resolveDownloadUrl();
      if (!url || this.saving) return;
      this.saving = true;
      try {
        const res = await fetch(url);
        if (!res.ok) {
          const msg = await this.parseError(res);
          window.notice.show({ message: msg, type: 'error' });
          return;
        }
        const base64 = await this.extractBase64(res);
        if (!base64) {
          window.notice.show({ message: 'ZIP データの取得に失敗しました', type: 'error' });
          return;
        }
        this.triggerDownload(base64, this.zipFilename);
        window.notice.show({ message: 'ZIP ファイルを保存しました', type: 'success' });
      } catch {
        window.notice.show({ message: 'ZIP の保存に失敗しました', type: 'error' });
      } finally {
        this.saving = false;
      }
    },

    async extractBase64(res) {
      const contentType = res.headers.get('content-type') || '';
      if (contentType.includes('application/json')) {
        const data = await res.json();
        return data.zip || data.data || data.content || '';
      }
      const text = await res.text();
      return text.trim();
    },

    triggerDownload(base64, filename) {
      const bin = Uint8Array.from(atob(base64), (c) => c.charCodeAt(0));
      const blob = new Blob([bin], { type: 'application/zip' });
      const objectUrl = URL.createObjectURL(blob);
      const anchor = document.createElement('a');
      anchor.href = objectUrl;
      anchor.download = filename;
      anchor.click();
      URL.revokeObjectURL(objectUrl);
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
