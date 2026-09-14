import { User } from "../model/user.js";

document.addEventListener("alpine:init", () => {
    Alpine.data("solviSetting", () => ({
        loading: false,
        form: { name: "", email: "", icon: "" },

        init() {
            const el = document.getElementById("user-json");
            if (el) {
                const u = User.fromJSON(JSON.parse(el.textContent || "{}"));
                this.form = {
                    name: u.name || "",
                    email: u.email || "",
                    icon: u.icon || "",
                };
            }
            if (typeof lucide !== "undefined") lucide.createIcons();
        },

        async onIconChange(e) {
            const file = e.target.files?.[0];
            if (!file) return;

            if (!file.type.startsWith("image/")) {
                window.notice.show({
                    message: "画像ファイルを選択してください",
                    type: "error",
                });
                e.target.value = "";
                return;
            }

            this.loading = true;
            try {
                const fd = new FormData();
                fd.append("icon", file);
                const res = await fetch("/api/v1/setting/icon", {
                    method: "POST",
                    body: fd,
                });
                if (!res.ok) {
                    window.notice.show({
                        message: "アイコンの保存に失敗しました",
                        type: "error",
                    });
                    return;
                }
                window.notice.show({
                    message: "アイコンを更新しました",
                    type: "success",
                });
                location.reload();
            } catch {
                window.notice.show({
                    message: "サーバへ接続できませんでした",
                    type: "error",
                });
            } finally {
                this.loading = false;
                e.target.value = "";
            }
        },

        async deleteIcon() {
            if (this.loading) return;

            this.loading = true;
            try {
                const res = await fetch("/api/v1/setting/icon", {
                    method: "DELETE",
                });
                if (!res.ok) {
                    window.notice.show({
                        message: "アイコン削除に失敗しました",
                        type: "error",
                    });
                    return;
                }
                window.notice.show({
                    message: "アイコンを削除しました",
                    type: "success",
                });
                location.reload();
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
