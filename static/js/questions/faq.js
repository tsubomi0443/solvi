import { SummaryListItem } from "../model/question.js";
import { pagerMethods } from "../pagination.js";

const VIEW_MODE_KEY = "solvi.faq.viewMode";

function excerpt(text = "", max = 80) {
    const trimmed = (text || "").trim();
    if (trimmed.length <= max) return trimmed;
    return trimmed.slice(0, max) + "…";
}

document.addEventListener("alpine:init", () => {
    Alpine.data("solviFAQ", () => ({
        summaries: [],
        filter: "",
        viewMode: "table",
        selectedTags: [],
        isAdmin: window.solviIsAdmin === "true",
        selectedSummary: null,
        showDetailModal: false,
        showDeleteModal: false,
        deleting: false,
        ...pagerMethods("visibleSummaries"),

        init() {
            const savedView = localStorage.getItem(VIEW_MODE_KEY);
            if (savedView === "table" || savedView === "card") {
                this.viewMode = savedView;
            }

            const el = document.getElementById("summaries-json");
            if (el) {
                const json = JSON.parse(el.textContent) || [];
                this.summaries = Array.from(json).map((dto) =>
                    SummaryListItem.fromJSON(dto),
                );
            }
            this.$nextTick(() => {
                if (typeof lucide !== "undefined") lucide.createIcons();
            });

            document.addEventListener("update-faq", (e) => {
                this.applyFAQUpdate(e.detail);
                this.$nextTick(() => {
                    if (typeof lucide !== "undefined") lucide.createIcons();
                });
            });
            this.bindPager(["filter", "selectedTags"]);
        },

        applyFAQUpdate(detail) {
            if (!detail) return;
            const summary = detail.summary;
            const summaryUUID = summary?.uuid;
            if (detail.supportStatus === "done" && summaryUUID) {
                const item = SummaryListItem.fromJSON(summary);
                const idx = this.summaries.findIndex(
                    (s) => s.uuid === summaryUUID,
                );
                if (idx >= 0) {
                    this.summaries = [
                        ...this.summaries.slice(0, idx),
                        item,
                        ...this.summaries.slice(idx + 1),
                    ];
                } else {
                    this.summaries = [item, ...this.summaries];
                }
                if (
                    this.selectedSummary?.uuid === summaryUUID &&
                    this.showDetailModal
                ) {
                    this.selectedSummary = item;
                }
                return;
            }
            if (!summaryUUID) return;
            this.removeSummaryByUUID(summaryUUID);
        },

        removeSummaryByUUID(summaryUUID) {
            this.summaries = this.summaries.filter(
                (s) => s.uuid !== summaryUUID,
            );
            if (this.selectedSummary?.uuid === summaryUUID) {
                this.showDetailModal = false;
                this.showDeleteModal = false;
                this.selectedSummary = null;
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
            this.$nextTick(() => {
                if (typeof lucide !== "undefined") lucide.createIcons();
            });
        },

        availableTags() {
            const tags = new Set();
            for (const item of this.summaries) {
                for (const tag of item.tags || []) tags.add(tag);
            }
            return Array.from(tags).sort((a, b) => a.localeCompare(b, "ja"));
        },

        visibleSummaries() {
            let items = [...this.summaries];

            const keyword = this.filter.trim().toLowerCase();
            if (keyword) {
                items = items.filter(
                    (item) =>
                        item.title?.toLowerCase().includes(keyword) ||
                        item.content?.toLowerCase().includes(keyword) ||
                        item.answer?.toLowerCase().includes(keyword),
                );
            }

            if (this.selectedTags.length > 0) {
                items = items.filter((item) =>
                    (item.tags || []).some((t) =>
                        this.selectedTags.includes(t),
                    ),
                );
            }

            return items;
        },

        openDetail(item) {
            this.selectedSummary = item;
            this.showDetailModal = true;
            this.$nextTick(() => {
                if (typeof lucide !== "undefined") lucide.createIcons();
            });
        },

        closeDetail() {
            this.showDetailModal = false;
            this.selectedSummary = null;
        },

        openDeleteModal(item, event) {
            if (event) event.stopPropagation();
            this.selectedSummary = item;
            this.showDeleteModal = true;
        },

        closeDeleteModal() {
            if (this.deleting) return;
            this.showDeleteModal = false;
            if (!this.showDetailModal) {
                this.selectedSummary = null;
            }
        },

        async confirmDelete() {
            if (this.deleting || !this.selectedSummary?.uuid) return;
            this.deleting = true;
            try {
                const res = await fetch(
                    `/api/v1/faqs/${this.selectedSummary.uuid}`,
                    { method: "DELETE" },
                );
                if (!res.ok) {
                    const msg = await res
                        .json()
                        .catch(() => ({ error: "削除に失敗しました" }));
                    window.notice?.show({
                        message: msg.error || "削除に失敗しました",
                        type: "error",
                    });
                    return;
                }
                const uuid = this.selectedSummary.uuid;
                this.summaries = this.summaries.filter((s) => s.uuid !== uuid);
                this.showDeleteModal = false;
                this.showDetailModal = false;
                this.selectedSummary = null;
                window.notice?.show({
                    message: "FAQを削除しました",
                    type: "success",
                });
            } finally {
                this.deleting = false;
            }
        },

        excerpt,
    }));
});
