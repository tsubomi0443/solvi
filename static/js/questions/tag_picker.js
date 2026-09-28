import { Tag } from "../model/tag.js";

function sortTags(names) {
    return names.sort((a, b) => a.localeCompare(b, "ja"));
}

export function tagPickerState() {
    return {
        availableTags: [],
        selectedTags: [],
        tagMenuOpen: false,
        tagQuery: "",
        tagsLoaded: false,
        showNewTagModal: false,
        newTagName: "",

        tagPickerDisabled() {
            return false;
        },

        includeAvailableTags(names) {
            const set = new Set(this.availableTags);
            for (const name of names || []) {
                const trimmed = String(name || "").trim();
                if (trimmed) set.add(trimmed);
            }
            this.availableTags = sortTags([...set]);
        },

        async loadAvailableTags() {
            const names = new Set(this.availableTags);

            const addName = (name) => {
                const trimmed = String(name || "").trim();
                if (trimmed) names.add(trimmed);
            };

            try {
                const res = await fetch("/api/v1/tags");
                if (res.ok) {
                    const body = await res.json();
                    if (Array.isArray(body)) {
                        for (const dto of body) addName(Tag.fromJSON(dto).name);
                        this.availableTags = sortTags([...names]);
                        this.tagsLoaded = true;
                        return;
                    }
                }
            } catch {
                // タグAPIが使えない場合は質問一覧から集める
            }

            try {
                const res = await fetch("/api/v1/questions");
                if (res.ok) {
                    const body = await res.json();
                    if (Array.isArray(body)) {
                        for (const item of body) {
                            for (const tag of item.tags || []) addName(tag);
                        }
                    }
                }
            } catch {
                // 一覧が取れなくても手入力での追加は続ける
            }
            this.availableTags = sortTags([...names]);
            this.tagsLoaded = true;
        },

        filteredTagOptions() {
            const q = this.tagQuery.trim().toLocaleLowerCase("ja");
            if (!q) return this.availableTags;
            return this.availableTags.filter((tag) =>
                tag.toLocaleLowerCase("ja").includes(q),
            );
        },

        tagListMessage() {
            if (!this.tagsLoaded) return "読み込み中";
            if (this.tagQuery.trim()) return "該当するタグがありません";
            return "タグがありません";
        },

        applyTagSelection(tags) {
            this.selectedTags = tags;
        },

        toggleTagMenu() {
            if (this.tagPickerDisabled()) return;
            if (this.tagMenuOpen) {
                this.closeTagMenu();
                return;
            }
            this.tagMenuOpen = true;
            this.$nextTick(() => this.$refs.tagSearch?.focus());
        },

        closeTagMenu() {
            this.tagMenuOpen = false;
            this.tagQuery = "";
        },

        toggleTag(tag) {
            if (this.tagPickerDisabled()) return;
            if (this.selectedTags.includes(tag)) {
                return this.removeSelectedTag(tag);
            }
            return this.applyTagSelection([...this.selectedTags, tag]);
        },

        removeSelectedTag(tag) {
            if (this.tagPickerDisabled()) return;
            return this.applyTagSelection(
                this.selectedTags.filter((t) => t !== tag),
            );
        },

        findExistingTag(name) {
            const key = name.trim().toLocaleLowerCase("ja");
            return this.availableTags.find(
                (tag) => tag.toLocaleLowerCase("ja") === key,
            );
        },

        openNewTagModal() {
            if (this.tagPickerDisabled()) return;
            this.closeTagMenu();
            this.newTagName = "";
            this.showNewTagModal = true;
            this.$nextTick(() => this.$refs.newTagInput?.focus());
        },

        closeNewTagModal() {
            this.showNewTagModal = false;
            this.newTagName = "";
        },

        onTagPickerEscape() {
            if (this.showNewTagModal) {
                this.closeNewTagModal();
                return;
            }
            this.closeTagMenu();
        },

        addCustomTag() {
            const name = this.newTagName.trim();
            if (!name || this.tagPickerDisabled()) return;
            const existing = this.findExistingTag(name);
            const tag = existing || name;
            if (!existing) {
                this.includeAvailableTags([name]);
            } else if (existing !== name) {
                window.notice.show({
                    message: "同じ名前のタグがあるため、既存のタグを選択しました",
                    type: "info",
                });
            }
            this.closeNewTagModal();
            if (!this.selectedTags.includes(tag)) {
                return this.applyTagSelection([...this.selectedTags, tag]);
            }
        },
    };
}
