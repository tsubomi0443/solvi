import { QuestionListItem } from "../model/question.js";

const VIEW_MODE_KEY = "solvi.home.viewMode";
const STATUS_OPTIONS = [
    { value: "pending", label: "未対応" },
    { value: "supporting", label: "対応中" },
    { value: "done", label: "完了" },
];

function statusLabel(s) {
    return { pending: "未対応", supporting: "対応中", done: "完了" }[s] || s;
}

function statusBadge(s) {
    return (
        {
            pending: "badge-ghost",
            supporting: "badge-warning",
            done: "badge-success",
        }[s] || "badge-ghost"
    );
}

function toListItem(dto) {
    if (dto instanceof QuestionListItem) return dto;
    return QuestionListItem.fromJSON(dto);
}

document.addEventListener("alpine:init", () => {
    Alpine.data("solviHome", () => ({
        questions: [],
        filter: "",
        viewMode: "table",
        sortDir: "asc",
        selectedTags: [],
        selectedStatuses: [],
        supportKind: "all",
        isSupporter: window.solviIsSupporter === "true",
        isAdmin: window.solviIsAdmin === "true",
        canViewAll: window.solviCanViewAll === "true",
        statusOptions: STATUS_OPTIONS,

        init() {
            const savedView = localStorage.getItem(VIEW_MODE_KEY);
            if (savedView === "table" || savedView === "card") {
                this.viewMode = savedView;
            }

            const el = document.getElementById("questions-json");
            if (el) {
                const json = JSON.parse(el.textContent) || [];
                this.questions = Array.from(json).map((dto) =>
                    QuestionListItem.fromJSON(dto),
                );
            }
            if (typeof lucide !== "undefined") lucide.createIcons();
            document.addEventListener("create-question", (e) => {
                this.upsertQuestion(e.detail);
            });
            document.addEventListener("update-question", (e) => {
                this.upsertQuestion(e.detail);
            });
        },

        upsertQuestion(detail) {
            if (!detail?.uuid) return;
            const idx = this.questions.findIndex((q) => q.uuid === detail.uuid);
            const item = toListItem(
                idx >= 0 ? { ...this.questions[idx], ...detail } : detail,
            );
            if (idx >= 0) {
                this.questions = [
                    ...this.questions.slice(0, idx),
                    item,
                    ...this.questions.slice(idx + 1),
                ];
            } else {
                this.questions = [item, ...this.questions];
            }
        },

        setViewMode(mode) {
            if (mode !== "table" && mode !== "card") return;
            this.viewMode = mode;
            localStorage.setItem(VIEW_MODE_KEY, mode);
        },

        toggleTag(tag) {
            if (this.selectedTags.includes(tag)) {
                this.selectedTags = this.selectedTags.filter((t) => t !== tag);
            } else {
                this.selectedTags = [...this.selectedTags, tag];
            }
        },

        toggleStatus(status) {
            if (this.selectedStatuses.includes(status)) {
                this.selectedStatuses = this.selectedStatuses.filter(
                    (s) => s !== status,
                );
            } else {
                this.selectedStatuses = [...this.selectedStatuses, status];
            }
        },

        availableTags() {
            const tags = new Set();
            for (const q of this.questions) {
                for (const tag of q.tags || []) tags.add(tag);
            }
            return Array.from(tags).sort((a, b) => a.localeCompare(b, "ja"));
        },

        visibleQuestions() {
            let items = [...this.questions];

            const keyword = this.filter.trim().toLowerCase();
            if (keyword) {
                items = items.filter(
                    (item) =>
                        item.title?.toLowerCase().includes(keyword) ||
                        (item.tags || []).some((t) =>
                            t.toLowerCase().includes(keyword),
                        ),
                );
            }

            if (this.selectedTags.length > 0) {
                items = items.filter((item) =>
                    (item.tags || []).some((t) =>
                        this.selectedTags.includes(t),
                    ),
                );
            }

            if (this.selectedStatuses.length > 0) {
                items = items.filter((item) =>
                    this.selectedStatuses.includes(item.supportStatus),
                );
            }

            if (this.supportKind === "human") {
                items = items.filter((item) => item.isRequireHumanSupport);
            } else if (this.supportKind === "ai") {
                items = items.filter((item) => !item.isRequireHumanSupport);
            }

            items.sort((a, b) => this.compareDue(a.answerDue, b.answerDue));
            return items;
        },

        compareDue(a, b) {
            const aTime = a ? Date.parse(a) : NaN;
            const bTime = b ? Date.parse(b) : NaN;
            const aMissing = Number.isNaN(aTime);
            const bMissing = Number.isNaN(bTime);

            if (aMissing && bMissing) return 0;
            if (aMissing) return 1;
            if (bMissing) return -1;

            const diff = aTime - bTime;
            return this.sortDir === "desc" ? -diff : diff;
        },

        formatDue(iso) {
            if (!iso) return "期限未設定";
            const date = new Date(iso);
            if (Number.isNaN(date.getTime())) return "期限未設定";
            return date.toLocaleDateString("ja-JP", {
                year: "numeric",
                month: "2-digit",
                day: "2-digit",
            });
        },

        statusLabel,
        statusBadge,
    }));
});
