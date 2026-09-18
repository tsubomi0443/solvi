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
        isAdmin: window.solviIsAdmin === "true",
        isMyselfQuestion: false,
        canViewAll: window.solviCanViewAll === "true",
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
        showDeleteModal: false,
        showDeleteQuestionModal: false,
        showReferModal: false,
        currentShowList: "",
        showReferList: false,
        showMemoList: false,
        hideChatRefers: false,
        hideChatMemos: false,
        referRows: [{ name: "", url: "" }],
        savingRefers: false,
        deleteTarget: null,
        deletingItem: false,
        deletingQuestion: false,
        showDoneModal: false,
        doneSummaryContent: "",
        doneSummaryAnswer: "",
        doneSelectedReferUuids: [],
        submittingDone: false,

        init() {
            const qEl = document.getElementById("question-json");
            const uEl = document.getElementById("user-json");
            if (qEl) {
                this.question = Question.fromJSON(
                    JSON.parse(qEl.textContent || "{}"),
                );
            }
            if (uEl) {
                this.currentUser = User.fromJSON(
                    JSON.parse(uEl.textContent || "{}"),
                );
            }
            this.isMyselfQuestion =
                this.question.questionUserUuid === this.currentUser.uuid;
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
            if (this.canViewAll) {
                (this.question.memos || []).forEach((m) => {
                    items.push({ ...m, kind: "memo" });
                });
            }
            (this.question.refers || []).forEach((r) => {
                items.push({
                    uuid: r.uuid,
                    kind: "refer",
                    name: r.name,
                    url: r.url,
                    userUuid: r.userUuid,
                    createdAt: r.createdAt,
                    userName: "引用情報",
                });
            });
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

        canDeleteItem(item) {
            const isMyselfItem = this.isSelfItem(item);
            if (this.isAdmin || isMyselfItem) return true;
            if (
                item.kind !== "answer" &&
                item.kind !== "memo" &&
                item.kind !== "refer"
            )
                return false;
            return this.isSupporter && isMyselfItem;
        },

        deleteModalMessage() {
            if (!this.deleteTarget) return "";
            if (this.deleteTarget.kind === "memo") {
                return "このメモを削除しますか？";
            }
            if (this.deleteTarget.kind === "refer") {
                return "この引用情報を削除しますか？";
            }
            return "この回答を削除しますか？";
        },

        openDeleteModal(item) {
            this.deleteTarget = item;
            this.showDeleteModal = true;
        },

        closeDeleteModal() {
            if (this.deletingItem) return;
            this.showDeleteModal = false;
            this.deleteTarget = null;
        },

        async confirmDelete() {
            if (!this.deleteTarget || this.deletingItem) return;
            const { kind, uuid } = this.deleteTarget;
            let path;
            if (kind === "memo") {
                path = `/api/v1/questions/${this.question.uuid}/memos/${uuid}`;
            } else if (kind === "refer") {
                path = `/api/v1/questions/${this.question.uuid}/refers/${uuid}`;
            } else {
                path = `/api/v1/questions/${this.question.uuid}/answers/${uuid}`;
            }
            this.deletingItem = true;
            try {
                const res = await fetch(path, { method: "DELETE" });
                if (!res.ok) {
                    const msg = await res
                        .json()
                        .catch(() => ({ error: "削除に失敗しました" }));
                    window.notice.show({
                        message: msg.error || "削除に失敗しました",
                        type: "error",
                    });
                    return;
                }
                this.showDeleteModal = false;
                this.deleteTarget = null;
                await this.fetchQuestion();
                window.notice.show({
                    message: "削除しました",
                    type: "success",
                });
            } finally {
                this.deletingItem = false;
            }
        },

        isShown(item) {
            if (item.kind === "refer" && this.hideChatRefers) {
                return false;
            }
            if (item.kind === "memo" && this.hideChatMemos) {
                return false;
            }
            if (item.kind === "refer") {
                return true;
            }

            return (
                this.isSelfItem(item) ||
                item.kind === "answer" ||
                this.canViewAll
            );
        },

        openDeleteQuestionModal() {
            this.showDeleteQuestionModal = true;
        },

        closeDeleteQuestionModal() {
            if (this.deletingQuestion) return;
            this.showDeleteQuestionModal = false;
        },

        async confirmDeleteQuestion() {
            if (this.deletingQuestion) return;
            this.deletingQuestion = true;
            try {
                const res = await fetch(
                    `/api/v1/questions/${this.question.uuid}`,
                    {
                        method: "DELETE",
                    },
                );
                if (!res.ok) {
                    const msg = await res
                        .json()
                        .catch(() => ({ error: "削除に失敗しました" }));
                    window.notice.show({
                        message: msg.error || "削除に失敗しました",
                        type: "error",
                    });
                    return;
                }
                window.location.href = "/";
            } finally {
                this.deletingQuestion = false;
            }
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
            if (this.savingMeta) return false;
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
                    return false;
                }
                this.question = Question.fromJSON(await res.json());
                this.syncMetaFields();
                window.notice.show({
                    message: "更新しました",
                    type: "success",
                });
                return true;
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
            if (this.editStatus === "done") {
                this.openDoneModal();
                return;
            }
            await this.updateQuestion({ status: this.editStatus });
        },

        openDoneModal() {
            const sum = this.question.summary;
            this.doneSummaryContent = sum?.content || "";
            this.doneSummaryAnswer = sum?.answer || "";
            const currentRefers = this.question.refers || [];
            if (
                sum &&
                Array.isArray(sum.references) &&
                sum.references.length > 0
            ) {
                const matchedUuids = [];
                sum.references.forEach((savedRef) => {
                    const match = currentRefers.find(
                        (r) =>
                            r.name === savedRef.name && r.url === savedRef.url,
                    );
                    if (match && match.uuid) {
                        matchedUuids.push(match.uuid);
                    }
                });
                this.doneSelectedReferUuids = matchedUuids;
            } else {
                this.doneSelectedReferUuids = [];
            }
            this.submittingDone = false;
            this.showDoneModal = true;
            this.$nextTick(() => {
                this.$refs.doneSummaryContentInput?.focus();
            });
        },

        closeDoneModal() {
            if (this.submittingDone) return;
            this.showDoneModal = false;
            this.editStatus = this.question.supportStatus || "pending";
        },

        canSubmitDone() {
            if (this.submittingDone) return false;
            if (!this.doneSummaryContent.trim()) return false;
            if (!this.doneSummaryAnswer.trim()) return false;
            const refers = this.question.refers || [];
            if (refers.length > 0 && this.doneSelectedReferUuids.length === 0) {
                return false;
            }
            return true;
        },

        async submitDone() {
            if (!this.canSubmitDone()) return;
            this.submittingDone = true;
            try {
                const payload = {
                    status: "done",
                    summary: {
                        content: this.doneSummaryContent.trim(),
                        answer: this.doneSummaryAnswer.trim(),
                        referUuids: this.doneSelectedReferUuids.slice(),
                    },
                };
                const ok = await this.updateQuestion(payload);
                if (ok !== false) {
                    this.showDoneModal = false;
                } else {
                    this.editStatus = this.question.supportStatus || "pending";
                }
            } finally {
                this.submittingDone = false;
            }
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

        openReferModal() {
            this.referRows = [{ name: "", url: "" }];
            this.showReferModal = true;
            this.$nextTick(() => {
                if (typeof lucide !== "undefined") lucide.createIcons();
            });
        },

        closeReferModal() {
            if (this.savingRefers) return;
            this.showReferModal = false;
        },

        addReferRow() {
            this.referRows.push({ name: "", url: "" });
        },

        removeReferRow(index) {
            if (this.referRows.length <= 1) return;
            this.referRows.splice(index, 1);
        },

        validReferRows() {
            return this.referRows.filter(
                (row) => row.name.trim() || row.url.trim(),
            );
        },

        canSubmitRefers() {
            const rows = this.validReferRows();
            if (!rows.length) return false;
            return rows.every((row) => row.name.trim() && row.url.trim());
        },

        toggleReferList() {
            this.showMemoList = false;
            this.showReferList = !this.showReferList;
            this.setCurrentShowList();
        },

        toggleMemoList() {
            this.showReferList = false;
            this.showMemoList = !this.showMemoList;
            this.setCurrentShowList();
        },

        showMemoFirstLine(memo = "") {
            const nlineIdx = memo.indexOf("\n");
            if (nlineIdx === -1) return memo;
            return memo.slice(0, nlineIdx);
        },

        isExistsMemoNextLine(memo = "") {
            return memo.indexOf("\n") !== -1;
        },

        setCurrentShowList() {
            if (this.showReferList) this.currentShowList = "refer";
            else if (this.showMemoList) this.currentShowList = "memo";
            else this.currentShowList = "";
        },

        async submitRefers() {
            if (this.savingRefers || !this.canSubmitRefers()) return;
            const refers = this.validReferRows().map((row) => ({
                name: row.name.trim(),
                url: row.url.trim(),
            }));
            this.savingRefers = true;
            try {
                const res = await fetch(
                    `/api/v1/questions/${this.question.uuid}/refers`,
                    {
                        method: "POST",
                        headers: { "Content-Type": "application/json" },
                        body: JSON.stringify({ refers }),
                    },
                );
                if (!res.ok) {
                    const msg = await res.json().catch(() => ({
                        error: "引用情報の登録に失敗しました",
                    }));
                    window.notice.show({
                        message: msg.error || "引用情報の登録に失敗しました",
                        type: "error",
                    });
                    return;
                }
                this.showReferModal = false;
                await this.fetchQuestion();
                this.scrollChatToBottom();
                window.notice.show({
                    message: "引用情報を登録しました",
                    type: "success",
                });
            } finally {
                this.savingRefers = false;
            }
        },

        statusLabel,
        statusBadge,
    }));
});
