package lineworks_usecase

import (
	"strings"
	"unicode/utf8"
)

const summaryLimit = 200

func QuestionLink(baseURL, questionUUID string) string {
	return strings.TrimRight(strings.TrimSpace(baseURL), "/") + "/questions/" + questionUUID
}

func Summarize(content string) string {
	content = strings.TrimSpace(content)
	if utf8.RuneCountInString(content) <= summaryLimit {
		return content
	}
	runes := []rune(content)
	return string(runes[:summaryLimit]) + "…"
}

func ReceivedMessage(questionUUID, askerName, title, summary, link string) string {
	return "【新規受付 #" + questionUUID + "】\n" +
		"質問者: " + askerName + "\n" +
		"件名: " + title + "\n" +
		"内容要約: " + summary + "\n" +
		link
}

func AnsweredMessage(questionUUID, link string) string {
	return "【回答完了 #" + questionUUID + "】\n" +
		"回答が完了しました。内容は次のページで確認できます。\n" +
		link
}

func ReopenedMessage(questionUUID, comment, assignee, link string) string {
	return "【再対応依頼 #" + questionUUID + "】\n" +
		"追加コメント: " + comment + "\n" +
		"前回担当: " + assignee + "\n" +
		link
}
