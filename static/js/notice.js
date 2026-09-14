const NOTICE_STORE_NAME = 'appNotice';
const _noticeTimers = new Map();
const _pendingNoticeOps = [];

function clearNoticeTimer(id) {
  const t = _noticeTimers.get(id);
  if (t != null) {
    clearTimeout(t);
    _noticeTimers.delete(id);
  }
}

function scheduleNoticeTimer(id, durationMs, onExpire) {
  clearNoticeTimer(id);
  if (durationMs == null || Number(durationMs) <= 0) return;
  const t = setTimeout(() => {
    _noticeTimers.delete(id);
    onExpire();
  }, Number(durationMs));
  _noticeTimers.set(id, t);
}

function refreshNoticeIcons() {
  queueMicrotask(() => {
    if (typeof lucide !== 'undefined') lucide.createIcons();
  });
}

function normalizeNoticeType(type) {
  const t = String(type ?? 'info').toLowerCase();
  if (t === 'alert') return 'warning';
  if (['info', 'success', 'warning', 'error'].includes(t)) return t;
  return 'info';
}

function defaultIconForType(type) {
  switch (type) {
    case 'success': return 'circle-check';
    case 'warning': return 'alert-triangle';
    case 'error': return 'circle-x';
    default: return 'info';
  }
}

function alertClassForType(type) {
  switch (type) {
    case 'success': return 'alert alert-success shadow-md';
    case 'warning': return 'alert alert-warning shadow-md';
    case 'error': return 'alert alert-error shadow-md';
    default: return 'alert alert-info shadow-md';
  }
}

function getNoticeStore() {
  if (typeof Alpine === 'undefined') return null;
  try { return Alpine.store(NOTICE_STORE_NAME); } catch { return null; }
}

function enqueueOrRun(fn) {
  const s = getNoticeStore();
  if (s) { fn(); return; }
  _pendingNoticeOps.push(fn);
}

document.addEventListener('alpine:init', () => {
  const DEFAULT_WAIT_MSEC = 4000;
  Alpine.store(NOTICE_STORE_NAME, {
    items: [],
    _seq: 0,
    show(opts = {}) {
      const message = String(opts.message ?? opts.text ?? '').trim();
      if (!message) return null;
      const id = ++this._seq;
      const type = normalizeNoticeType(opts.type);
      const duration = opts.duration === undefined ? DEFAULT_WAIT_MSEC : (opts.duration === null ? 0 : Number(opts.duration));
      this.items.push({
        id, message, type,
        icon: opts.icon ? String(opts.icon) : defaultIconForType(type),
        duration, dismissible: opts.dismissible !== false,
        widthClass: String(opts.widthClass ?? 'w-full max-w-md'),
        leaving: false,
      });
      scheduleNoticeTimer(id, duration, () => this.dismiss(id));
      refreshNoticeIcons();
      return id;
    },
    dismiss(id) {
      clearNoticeTimer(id);
      const item = this.items.find((n) => n.id === id);
      if (!item || item.leaving) return;
      item.leaving = true;
      setTimeout(() => {
        this.items = this.items.filter((n) => n.id !== id);
        refreshNoticeIcons();
      }, 200);
    },
    alertClass(n) { return alertClassForType(normalizeNoticeType(n.type)); },
  });
  while (_pendingNoticeOps.length) _pendingNoticeOps.shift()();
});

window.notice = {
  show(opts) {
    let result = null;
    enqueueOrRun(() => { const s = getNoticeStore(); if (s) result = s.show(opts); });
    return result;
  },
};
