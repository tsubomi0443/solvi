export const PAGE_SIZE = 25;

export function totalPagesFor(count, pageSize = PAGE_SIZE) {
    if (count <= 0) return 1;
    return Math.ceil(count / pageSize);
}

export function pageWindow(page, pages) {
    const current = Math.min(Math.max(1, page), pages);
    const numbers = [];
    if (pages <= 7) {
        for (let i = 1; i <= pages; i++) numbers.push(i);
    } else {
        const windowStart = Math.max(2, current - 1);
        const windowEnd = Math.min(pages - 1, current + 1);
        numbers.push(1);
        if (windowStart > 2) numbers.push(null);
        for (let i = windowStart; i <= windowEnd; i++) numbers.push(i);
        if (windowEnd < pages - 1) numbers.push(null);
        numbers.push(pages);
    }
    return numbers.map((n, i) => ({
        key: n == null ? `ellipsis-${i}` : String(n),
        label: n == null ? "…" : String(n),
        page: n,
        ellipsis: n == null,
    }));
}

export function pagerMethods(listFnName) {
    return {
        page: 1,
        pageSize: PAGE_SIZE,

        totalPages() {
            return totalPagesFor(this[listFnName]().length, this.pageSize);
        },

        pagedItems() {
            const items = this[listFnName]();
            const pages = totalPagesFor(items.length, this.pageSize);
            const page = Math.min(Math.max(1, this.page), pages);
            const start = (page - 1) * this.pageSize;
            return items.slice(start, start + this.pageSize);
        },

        pageItems() {
            return pageWindow(this.page, this.totalPages());
        },

        setPage(n) {
            const next = Math.min(Math.max(1, n), this.totalPages());
            if (next === this.page) return;
            this.page = next;
            this.$nextTick(() => {
                const main = this.$root?.querySelector("main");
                if (main && main.scrollHeight > main.clientHeight + 1) {
                    main.scrollTo({ top: 0, behavior: "smooth" });
                    return;
                }
                window.scrollTo({ top: 0, behavior: "smooth" });
            });
        },

        bindPager(filterExprs) {
            for (const expr of filterExprs) {
                this.$watch(expr, () => {
                    this.page = 1;
                });
            }
            this.$watch(`${listFnName}().length`, (count) => {
                const pages = totalPagesFor(count, this.pageSize);
                if (this.page > pages) this.page = pages;
            });
        },
    };
}
