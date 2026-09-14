import { QuestionListItem } from "../model/question.js";

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

document.addEventListener("alpine:init", () => {
    Alpine.data("solviHome", () => ({
        questions: [],
        filter: "",
        isSupporter: Boolean(window.solviIsSupporter),

        init() {
            const el = document.getElementById("questions-json");
            if (el) {
                const json = JSON.parse(el.textContent) || [];
                console.log(json);
                this.questions = Array.from(json).map((dto) =>
                    QuestionListItem.fromJSON(dto),
                );
            }
            if (typeof lucide !== "undefined") lucide.createIcons();
            document.addEventListener("create-question", (e) => {
                if (!this.isSupporter || !e.detail) return;
                this.questions = [
                    QuestionListItem.fromJSON(e.detail),
                    ...this.questions,
                ];
            });
            document.addEventListener("update-question", (e) => {
                if (!e.detail?.uuid) return;
                const idx = this.questions.findIndex(
                    (q) => q.uuid === e.detail.uuid,
                );
                if (idx >= 0)
                    this.questions[idx] = QuestionListItem.fromJSON({
                        ...this.questions[idx],
                        ...e.detail,
                    });
            });
        },

        filteredQuestions() {
            const q = this.filter.trim().toLowerCase();
            if (!q) return this.questions;
            return this.questions.filter(
                (item) =>
                    item.title?.toLowerCase().includes(q) ||
                    (item.tags || []).some((t) => t.toLowerCase().includes(q)),
            );
        },

        statusLabel,
        statusBadge,
    }));
});
