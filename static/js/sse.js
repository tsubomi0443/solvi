import { QuestionListItem } from "./model/question.js";
import { Question } from "./model/question.js";
import { User } from "./model/user.js";

const SOLVI_SSE_EVENTS = [
    "time-tick",
    "create-question",
    "create-content",
    "create-answer",
    "create-memo",
    "create-refer",
    "update-question",
    "update-user",
];

const QUESTION_DETAIL_EVENTS = new Set([
    "create-content",
    "create-answer",
    "create-memo",
    "create-refer",
    "update-question",
]);

function parseEventDetail(name, data) {
    let parsed = data;
    try {
        parsed = JSON.parse(data);
    } catch {
        return data;
    }
    if (name === "create-question") return QuestionListItem.fromJSON(parsed);
    if (QUESTION_DETAIL_EVENTS.has(name)) return Question.fromJSON(parsed);
    if (name === "update-user") return User.fromJSON(parsed);
    return parsed;
}

document.addEventListener("DOMContentLoaded", () => {
    if (!window.solviSSEEnabled) return;
    const es = new EventSource("/sse");
    SOLVI_SSE_EVENTS.forEach((name) => {
        es.addEventListener(name, (event) => {
            const detail = parseEventDetail(name, event.data);
            document.dispatchEvent(new CustomEvent(name, { detail }));
        });
    });
});
