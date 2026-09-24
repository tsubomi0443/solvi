import { SummaryListItem } from "../model/question.js";

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
            if (typeof lucide !== "undefined") lucide.createIcons();
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

        availableTags() {
            const tags = new Set();
            for (const item of this.summaries) {
                for (const tag of item.tags || []) tags.add(tag);
            }
            console.log(tags)
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
