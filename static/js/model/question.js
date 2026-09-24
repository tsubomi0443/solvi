export class Timeline {
    constructor({
        uuid = "",
        content = "",
        userUuid = "",
        userName = "",
        createdAt = "",
    } = {}) {
        this.uuid = uuid;
        this.content = content;
        this.userUuid = userUuid;
        this.userName = userName;
        this.createdAt = createdAt;
    }

    static fromJSON(dto) {
        if (!dto) return new Timeline();
        return new Timeline({
            uuid: dto.uuid ?? "",
            content: dto.content ?? "",
            userUuid: dto.userUuid ?? "",
            userName: dto.userName ?? "",
            createdAt: dto.createdAt ?? "",
        });
    }
}

export class Refer {
    constructor({
        uuid = "",
        name = "",
        url = "",
        userUuid = "",
        createdAt = "",
    } = {}) {
        this.uuid = uuid;
        this.name = name;
        this.url = url;
        this.userUuid = userUuid;
        this.createdAt = createdAt;
    }

    static fromJSON(dto) {
        if (!dto) return new Refer();
        return new Refer({
            uuid: dto.uuid ?? "",
            name: dto.name ?? "",
            url: dto.url ?? "",
            userUuid: dto.userUuid ?? "",
            createdAt: dto.createdAt ?? "",
        });
    }
}

export class QuestionSummaryReference {
    constructor({ uuid = "", name = "", url = "" } = {}) {
        this.uuid = uuid;
        this.name = name;
        this.url = url;
    }

    static fromJSON(dto) {
        if (!dto) return new QuestionSummaryReference();
        return new QuestionSummaryReference({
            uuid: dto.uuid ?? "",
            name: dto.name ?? "",
            url: dto.url ?? "",
        });
    }
}

export class QuestionSummary {
    constructor({
        uuid = "",
        title = "",
        content = "",
        answer = "",
        references = [],
    } = {}) {
        this.uuid = uuid;
        this.title = title;
        this.content = content;
        this.answer = answer;
        this.references = references;
    }

    static fromJSON(dto) {
        if (!dto) return null;
        return new QuestionSummary({
            uuid: dto.uuid ?? "",
            title: dto.title ?? "",
            content: dto.content ?? "",
            answer: dto.answer ?? "",
            references: (dto.references || []).map((r) =>
                QuestionSummaryReference.fromJSON(r),
            ),
        });
    }
}

export class SummaryListItem {
    constructor({
        uuid = "",
        title = "",
        content = "",
        answer = "",
        tags = [],
        references = [],
        createdAt = "",
        createdDate = "",
        createdTime = "",
    } = {}) {
        this.uuid = uuid;
        this.title = title;
        this.content = content;
        this.answer = answer;
        this.tags = tags;
        this.references = references;
        this.createdAt = createdAt;
        this.createdDate = createdDate;
        this.createdTime = createdTime;
    }

    static fromJSON(dto) {
        if (!dto) return new SummaryListItem();
        return new SummaryListItem({
            uuid: dto.uuid ?? "",
            title: dto.title ?? "",
            content: dto.content ?? "",
            answer: dto.answer ?? "",
            tags: Array.isArray(dto.tags) ? dto.tags.slice() : [],
            references: (dto.references || []).map((r) =>
                QuestionSummaryReference.fromJSON(r),
            ),
            createdAt: dto.createdAt ?? "",
            createdDate: dto.createdDate ?? "",
            createdTime: dto.createdTime ?? "",
        });
    }
}

export class QuestionListItem {
    constructor({
        uuid = "",
        title = "",
        content = "",
        supportStatus = "",
        isRequireHumanSupport = false,
        questionUserUUID = "",
        questionUserName = "",
        questionUserDepartment = "",
        answerDue = "",
        answerDueDate = "",
        tags = [],
        createdAt = "",
        createdDate = "",
        createdTime = "",
    } = {}) {
        this.uuid = uuid;
        this.title = title;
        this.content = content;
        this.supportStatus = supportStatus;
        this.isRequireHumanSupport = isRequireHumanSupport;
        this.questionUserUUID = questionUserUUID;
        this.questionUserName = questionUserName;
        this.questionUserDepartment = questionUserDepartment;
        this.answerDue = answerDue;
        this.answerDueDate = answerDueDate;
        this.tags = tags;
        this.createdAt = createdAt;
        this.createdDate = createdDate;
        this.createdTime = createdTime;
    }

    static fromJSON(dto) {
        if (!dto) return new QuestionListItem();
        return new QuestionListItem({
            uuid: dto.uuid ?? "",
            title: dto.title ?? "",
            content: dto.content ?? "",
            supportStatus: dto.supportStatus ?? "",
            isRequireHumanSupport: Boolean(dto.isRequireHumanSupport),
            questionUserUUID: dto.questionUserUUID ?? "",
            questionUserName: dto.questionUserName ?? "",
            questionUserDepartment: dto.questionUserDepartment ?? "",
            answerDueDate: dto.answerDueDate ?? "",
            answerDue: dto.answerDue ?? "",
            tags: Array.isArray(dto.tags) ? dto.tags.slice() : [],
            createdAt: dto.createdAt ?? "",
            createdDate: dto.createdDate ?? "",
            createdTime: dto.createdTime ?? "",
        });
    }
}

export class Question {
    constructor({
        uuid = "",
        title = "",
        supportStatus = "",
        isRequireHumanSupport = false,
        answerDue = "",
        questionUserUuid = "",
        questionUserName = "",
        questionUserDepartment = "",
        tags = [],
        contents = [],
        answers = [],
        memos = [],
        refers = [],
        summary = null,
    } = {}) {
        this.uuid = uuid;
        this.title = title;
        this.supportStatus = supportStatus;
        this.isRequireHumanSupport = isRequireHumanSupport;
        this.answerDue = answerDue;
        this.questionUserUuid = questionUserUuid;
        this.questionUserName = questionUserName;
        this.questionUserDepartment = questionUserDepartment;
        this.tags = tags;
        this.contents = contents;
        this.answers = answers;
        this.memos = memos;
        this.refers = refers;
        this.summary = summary;
    }

    static fromJSON(dto) {
        if (!dto) return new Question();
        return new Question({
            uuid: dto.uuid ?? "",
            title: dto.title ?? "",
            supportStatus: dto.supportStatus ?? "",
            // TODO: AI機能実装後に修正
            isRequireHumanSupport: Boolean(dto.isRequireHumanSupport),
            answerDue: dto.answerDue ?? "",
            questionUserUuid: dto.questionUserUuid ?? "",
            questionUserName: dto.questionUserName ?? "",
            questionUserDepartment: dto.questionUserDepartment ?? "",
            tags: Array.isArray(dto.tags) ? dto.tags.slice() : [],
            contents: (dto.contents || []).map((item) =>
                Timeline.fromJSON(item),
            ),
            answers: (dto.answers || []).map((item) => Timeline.fromJSON(item)),
            memos: (dto.memos || []).map((item) => Timeline.fromJSON(item)),
            refers: (dto.refers || []).map((item) => Refer.fromJSON(item)),
            summary: QuestionSummary.fromJSON(dto.summary),
        });
    }
}
