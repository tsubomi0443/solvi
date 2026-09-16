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

export class QuestionListItem {
    constructor({
        uuid = "",
        title = "",
        supportStatus = "",
        isRequireHumanSupport = false,
        questionUserName = "",
        questionUserDepartment = "",
        answerDue = "",
        tags = [],
    } = {}) {
        this.uuid = uuid;
        this.title = title;
        this.supportStatus = supportStatus;
        this.isRequireHumanSupport = isRequireHumanSupport;
        this.questionUserName = questionUserName;
        this.questionUserDepartment = questionUserDepartment;
        this.answerDue = answerDue;
        this.tags = tags;
    }

    static fromJSON(dto) {
        if (!dto) return new QuestionListItem();
        return new QuestionListItem({
            uuid: dto.uuid ?? "",
            title: dto.title ?? "",
            supportStatus: dto.supportStatus ?? "",
            isRequireHumanSupport: Boolean(dto.isRequireHumanSupport),
            questionUserName: dto.questionUserName ?? "",
            questionUserDepartment: dto.questionUserDepartment ?? "",
            answerDue: dto.answerDue ?? "",
            tags: Array.isArray(dto.tags) ? dto.tags.slice() : [],
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
        });
    }
}
