import { Question } from "../model/question.js";
import { tagPickerState } from "./tag_picker.js";

document.addEventListener("alpine:init", () => {
    Alpine.data("solviQuestionNew", () => ({
        ...tagPickerState(),
        loading: false,
        form: {
            title: "",
            content: "",
            answerDue: "",
            isRequireHumanSupport: true,
        },

        async init() {
            await this.loadAvailableTags();
            if (typeof lucide !== "undefined") lucide.createIcons();
        },

        async submit() {
            this.loading = true;
            try {
                const res = await fetch("/api/v1/questions", {
                    method: "POST",
                    headers: { "Content-Type": "application/json" },
                    body: JSON.stringify({
                        title: this.form.title,
                        content: this.form.content,
                        tags: this.selectedTags,
                        answerDue: this.form.answerDue,
                        isRequireHumanSupport: this.form.isRequireHumanSupport,
                    }),
                });
                if (!res.ok) {
                    const msg = await res
                        .json()
                        .catch(() => ({ error: "作成に失敗しました" }));
                    window.notice.show({
                        message: msg.error || "作成に失敗しました",
                        type: "error",
                    });
                    return;
                }
                const detail = Question.fromJSON(await res.json());
                location.href = `/questions/${detail.uuid}`;
            } catch {
                window.notice.show({
                    message: "サーバへ接続できませんでした",
                    type: "error",
                });
            } finally {
                this.loading = false;
            }
        },
    }));
});
