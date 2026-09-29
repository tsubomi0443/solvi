package lineworks_usecase

import (
	"fmt"
	"strings"
)

const dueDigestLimit = 1000

// DueDigestItem は定期通知の1行分。
type DueDigestItem struct {
	Title     string
	AskerName string
	DueDate   string
	Link      string
}

func DueDigestMessage(todayItems, tomorrowItems []DueDigestItem) string {
	if len(todayItems) == 0 && len(tomorrowItems) == 0 {
		return "【定期通知】\n期日に近づいている問い合わせはありませんでした。"
	}

	var b strings.Builder
	b.WriteString("【定期通知】")
	if len(todayItems) > 0 {
		appendDueBlock(&b, fmt.Sprintf("期日当日の問い合わせが、%d件あります。", len(todayItems)), true, todayItems)
	}
	if len(tomorrowItems) > 0 {
		appendDueBlock(&b, fmt.Sprintf("期日が明日に迫った問い合わせが、%d件あります。", len(tomorrowItems)), false, tomorrowItems)
	}
	return trimToLimit(b.String(), dueDigestLimit)
}

func appendDueBlock(b *strings.Builder, header string, includePrompt bool, items []DueDigestItem) {
	if b.Len() > 0 {
		b.WriteByte('\n')
	}
	b.WriteString(header)
	if includePrompt {
		b.WriteByte('\n')
		b.WriteString("内容を確認し回答してください。")
	}
	added := 0
	for _, item := range items {
		line := "\n" + dueDigestLine(item)
		if len(b.String())+len(line) > dueDigestLimit {
			break
		}
		b.WriteString(line)
		added++
	}
	if added < len(items) {
		b.WriteString(fmt.Sprintf("\n…他 %d 件", len(items)-added))
	}
}

func dueDigestLine(item DueDigestItem) string {
	name := strings.TrimSpace(item.AskerName)
	if name == "" {
		name = "不明"
	}
	return "・" + strings.TrimSpace(item.Title) + " " + name + " " + item.DueDate + " " + item.Link
}

func trimToLimit(body string, limit int) string {
	if len(body) <= limit {
		return body
	}
	return strings.TrimRight(body[:limit], "\n")
}
