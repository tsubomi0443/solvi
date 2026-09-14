import { Question } from "../model/question.js";

document.addEventListener("alpine:init", () => {
    Alpine.data("solviQuestionNew", () => ({
        loading: false,
        form: {
            title: "",
            content: "",
            tagsText: "",
            answerDue: "",
            isRequireHumanSupport: false,
        },

        init() {
            if (typeof lucide !== "undefined") lucide.createIcons();
        },

        async submit() {
            this.loading = true;
            try {
                const tags = this.form.tagsText
                    .split(",")
                    .map((t) => t.trim())
                    .filter(Boolean);
                const res = await fetch("/api/v1/questions", {
                    method: "POST",
                    headers: { "Content-Type": "application/json" },
                    body: JSON.stringify({
                        title: this.form.title,
                        content: this.form.content,
                        tags,
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
