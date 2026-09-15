import { Question } from "../model/question.js";
import { User } from "../model/user.js";

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

const STATUS_OPTIONS = [
    { value: "pending", label: "未対応" },
    { value: "supporting", label: "対応中" },
    { value: "done", label: "完了" },
];

document.addEventListener("alpine:init", () => {
    Alpine.data("solviDetail", () => ({
        question: {},
        currentUser: {},
        isSupporter: window.solviIsSupporter === "true",
        composerText: "",
        editTitle: "",
        editStatus: "",
        editAnswerDue: "",
        editRequireHuman: false,
        newTag: "",
        statusOptions: STATUS_OPTIONS,
        savingMeta: false,
        showScrollToBottom: false,
        chatAtBottom: true,
        lastTimelineCount: 0,
        userIconMap: {},

        init() {
            const qEl = document.getElementById("question-json");
            const uEl = document.getElementById("user-json");
            if (qEl)
                this.question = Question.fromJSON(
                    JSON.parse(qEl.textContent || "{}"),
                );
            if (uEl)
                this.currentUser = User.fromJSON(
                    JSON.parse(uEl.textContent || "{}"),
                );
            this.syncMetaFields();
            if (typeof lucide !== "undefined") lucide.createIcons();
            [
                "create-content",
                "create-answer",
                "create-memo",
                "create-refer",
                "update-question",
            ].forEach((ev) => {
                document.addEventListener(ev, (e) => {
                    if (e.detail?.uuid === this.question.uuid)
                        this.fetchQuestion();
                });
            });
            this.$nextTick(() => {
                this.lastTimelineCount = this.chatTimelineItems().length;
                this.scrollChatToBottom(false);
                if (typeof lucide !== "undefined") lucide.createIcons();
            });
        },

        onChatScroll() {
            const el = this.$refs.chatBox;
            if (!el) return;
            const threshold = 80;
            this.chatAtBottom =
                el.scrollHeight - el.scrollTop - el.clientHeight < threshold;
            if (this.chatAtBottom) {
                this.showScrollToBottom = false;
            }
        },

        scrollChatToBottom(smooth = true) {
            const el = this.$refs.chatBox;
            if (!el) return;
            el.scrollTo({
                top: el.scrollHeight,
                behavior: smooth ? "smooth" : "auto",
            });
            this.showScrollToBottom = false;
            this.chatAtBottom = true;
        },

        afterChatUpdate(wasAtBottom) {
            this.$nextTick(() => {
                const count = this.chatTimelineItems().length;
                const hasNew = count > this.lastTimelineCount;
                this.lastTimelineCount = count;
                if (hasNew && !wasAtBottom) {
                    this.showScrollToBottom = true;
                } else if (wasAtBottom) {
                    this.scrollChatToBottom(false);
                }
                if (typeof lucide !== "undefined") lucide.createIcons();
            });
        },

        syncMetaFields() {
            this.editTitle = this.question.title || "";
            this.editStatus = this.question.supportStatus || "pending";
            this.editAnswerDue = this.toDateInputValue(this.question.answerDue);
            this.editRequireHuman = Boolean(
                this.question.isRequireHumanSupport,
            );
        },

        initial(name) {
            return (name || "?").slice(0, 1);
        },

        chatTimelineItems() {
            const items = [];
            (this.question.contents || []).forEach((c) => {
                items.push({ ...c, kind: "content" });
            });
            (this.question.answers || []).forEach((a) => {
                items.push({
                    ...a,
                    kind: "answer",
                    refers: this.refersForAnswer(a),
                });
            });
            if (this.isSupporter) {
                (this.question.memos || []).forEach((m) => {
                    items.push({ ...m, kind: "memo" });
                });
            }
            return items.sort(
                (a, b) => new Date(a.createdAt) - new Date(b.createdAt),
            );
        },

        async chatTimelineUserIcon(uuid) {
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

        refersForAnswer(answer) {
            return (this.question.refers || []).filter(() => false);
        },

        isSelfItem(item) {
            return Boolean(
                item.userUuid &&
                this.currentUser.uuid &&
                item.userUuid === this.currentUser.uuid,
            );
        },

        isShown(item) {
            return (
                this.isSelfItem(item) ||
                item.kind === "answer" ||
                this.isSupporter
            );
        },

        chatBubbleClass(kind) {
            if (kind === "memo") return "chat-bubble-accent";
            if (kind === "content") return "chat-bubble-secondary";
            return "";
        },

        formatDate(v) {
            if (!v) return "";
            try {
                return new Date(v).toLocaleString("ja-JP");
            } catch {
                return v;
            }
        },

        formatDue(v) {
            if (!v) return "—";
            const date = new Date(v);
            if (Number.isNaN(date.getTime())) return "期限未設定";
            return date.toLocaleDateString("ja-JP", {
                year: "numeric",
                month: "2-digit",
                day: "2-digit",
            });
        },

        toDateInputValue(iso) {
            if (!iso) return "";
            const date = new Date(iso);
            if (Number.isNaN(date.getTime())) return "";
            return new Intl.DateTimeFormat("en-CA", {
                timeZone: "Asia/Tokyo",
            }).format(date);
        },

        async fetchQuestion() {
            const wasAtBottom = this.chatAtBottom;
            const res = await fetch(`/api/v1/questions/${this.question.uuid}`);
            if (!res.ok) return;
            this.question = Question.fromJSON(await res.json());
            this.syncMetaFields();
            this.afterChatUpdate(wasAtBottom);
        },

        async updateQuestion(payload) {
            if (this.savingMeta) return;
            this.savingMeta = true;
            try {
                const res = await fetch(
                    `/api/v1/questions/${this.question.uuid}`,
                    {
                        method: "PUT",
                        headers: { "Content-Type": "application/json" },
                        body: JSON.stringify(payload),
                    },
                );
                if (!res.ok) {
                    const msg = await res
                        .json()
                        .catch(() => ({ error: "更新に失敗しました" }));
                    window.notice.show({
                        message: msg.error || "更新に失敗しました",
                        type: "error",
                    });
                    this.syncMetaFields();
                    return;
                }
                this.question = Question.fromJSON(await res.json());
                this.syncMetaFields();
                window.notice.show({
                    message: "更新しました",
                    type: "success",
                });
            } finally {
                this.savingMeta = false;
            }
        },

        async saveTitle() {
            const title = this.editTitle.trim();
            if (!title || title === this.question.title) return;
            await this.updateQuestion({ title });
        },

        async saveStatus() {
            if (
                !this.editStatus ||
                this.editStatus === this.question.supportStatus
            )
                return;
            await this.updateQuestion({ status: this.editStatus });
        },

        async saveAnswerDue() {
            const current = this.toDateInputValue(this.question.answerDue);
            if (this.editAnswerDue === current) return;
            if (!this.editAnswerDue) return;
            await this.updateQuestion({ answerDue: this.editAnswerDue });
        },

        async saveRequireHuman() {
            if (
                this.editRequireHuman ===
                Boolean(this.question.isRequireHumanSupport)
            )
                return;
            await this.updateQuestion({
                isRequireHumanSupport: this.editRequireHuman,
            });
        },

        async saveTags(tags) {
            await this.updateQuestion({ tags });
        },

        async addTag() {
            const tag = this.newTag.trim();
            if (!tag) return;
            if ((this.question.tags || []).includes(tag)) {
                this.newTag = "";
                return;
            }
            this.newTag = "";
            await this.saveTags([...(this.question.tags || []), tag]);
        },

        async removeTag(tag) {
            const tags = (this.question.tags || []).filter((t) => t !== tag);
            await this.saveTags(tags);
        },

        async appendContent() {
            const res = await fetch(
                `/api/v1/questions/${this.question.uuid}/contents`,
                {
                    method: "POST",
                    headers: { "Content-Type": "application/json" },
                    body: JSON.stringify({ content: this.composerText }),
                },
            );
            if (!res.ok) {
                window.notice.show({
                    message: "追記に失敗しました",
                    type: "error",
                });
                return;
            }
            this.composerText = "";
            await this.fetchQuestion();
            this.scrollChatToBottom();
        },

        async addAnswer() {
            const res = await fetch(
                `/api/v1/questions/${this.question.uuid}/answers`,
                {
                    method: "POST",
                    headers: { "Content-Type": "application/json" },
                    body: JSON.stringify({
                        content: this.composerText,
                        refers: [],
                    }),
                },
            );
            if (!res.ok) {
                window.notice.show({
                    message: "回答の登録に失敗しました",
                    type: "error",
                });
                return;
            }
            this.composerText = "";
            await this.fetchQuestion();
            this.scrollChatToBottom();
        },

        async addMemo() {
            const res = await fetch(
                `/api/v1/questions/${this.question.uuid}/memos`,
                {
                    method: "POST",
                    headers: { "Content-Type": "application/json" },
                    body: JSON.stringify({ content: this.composerText }),
                },
            );
            if (!res.ok) {
                window.notice.show({
                    message: "メモの追加に失敗しました",
                    type: "error",
                });
                return;
            }
            this.composerText = "";
            await this.fetchQuestion();
            this.scrollChatToBottom();
        },

        statusLabel,
        statusBadge,
    }));
});
