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
        sortDir: "due_asc",
        selectedDues: [],
        selectedTags: [],
        selectedStatuses: [],
        supportKind: "all",
        isSupporter: window.solviIsSupporter === "true",
        isAdmin: window.solviIsAdmin === "true",
        canViewAll: window.solviCanViewAll === "true",
        statusOptions: STATUS_OPTIONS,
        userIconMap: {},

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
            document.addEventListener("delete-question", (e) => {
                this.deleteQuestion(e.detail);
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

        deleteQuestion(detail) {
            if (!detail?.uuid) return;
            const idx = this.questions.findIndex((q) => q.uuid === detail.uuid);
            if (idx >= 0) {
                this.questions = [
                    ...this.questions.slice(0, idx),
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

        toggleDue(due) {
            if (this.selectedDues.includes(due)) {
                this.selectedDues = this.selectedDues.filter((t) => t !== due);
            } else {
                this.selectedDues = [...this.selectedDues, due];
            }
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

        availableDues() {
            const dues = new Set();
            for (const q of this.questions.sort((a, b) =>
                this.compareDate("asc", a.answerDue, b.answerDue),
            )) {
                if (q.answerDueDate === "") continue;
                dues.add(q.answerDueDate);
            }

            return Array.from(dues).sort((a, b) => {
                this.parseYYYYMMDD(a).getTime() -
                    this.parseYYYYMMDD(b).getTime();
            });
        },

        /**
         * @returns {Date}
         */
        parseYYYYMMDD(v) {
            const [year, month, day] = v.split("/").map(Number);
            return new Date(year, month - 1, day);
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

            if (this.selectedDues.length > 0) {
                items = items.filter((item) =>
                    this.selectedDues.includes(item.answerDueDate),
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

            if (this.sortDir.split(".").length > 1) {
                const key = this.sortDir.split(".")[0];
                const order = this.sortDir.split(".")[1];
                if (key === "due") {
                    items.sort((a, b) =>
                        this.compareDate(a.answerDue, b.answerDue),
                    );
                } else if (key === "created") {
                    items.sort((a, b) =>
                        this.compareDate(order, a.createdAt, b.createdAt),
                    );
                }
            }

            return items;
        },

        compareDate(order = "", a, b) {
            const aTime = a ? Date.parse(a) : NaN;
            const bTime = b ? Date.parse(b) : NaN;
            const aMissing = Number.isNaN(aTime);
            const bMissing = Number.isNaN(bTime);

            if (aMissing && bMissing) return 0;
            if (aMissing) return 1;
            if (bMissing) return -1;

            const diff = aTime - bTime;
            return order === "desc" ? -diff : diff;
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

        initial(name) {
            return (name || "?").slice(0, 1);
        },

        showContentFirstLine(questionContent = "") {
            const nlineIdx = questionContent.indexOf("\n");
            if (nlineIdx === -1) return questionContent;
            return questionContent.slice(0, nlineIdx);
        },

        async userIcon(uuid) {
            if (uuid in this.userIconMap) {
                return this.userIconMap[uuid];
            }

            this.userIconMap[uuid] = (async () => {
                try {
                    const res = await fetch(`/api/v1/user/icon/${uuid}`, {
                        method: "GET",
                    });
                    if (!res.ok) {
                        throw new Error("ユーザアイコンの取得に失敗しました");
                    }
                    const data = await res.json();
                    return data["icon"];
                } catch (err) {
                    delete this.userIconMap[uuid];
                    console.error(err);
                    throw err;
                }
            })();

            return this.userIconMap[uuid];
        },

        statusLabel,
        statusBadge,
    }));
});
