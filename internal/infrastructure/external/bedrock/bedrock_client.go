package bedrock

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	ext "solvi/internal/domain/interface/external"
	"os"
	"solvi/internal/shared/config"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentruntime"
	agenttypes "github.com/aws/aws-sdk-go-v2/service/bedrockagentruntime/types"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	brtypes "github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
)

type BedrockClient struct {
	kbID      string
	modelARN  string
	agent     *bedrockagentruntime.Client
	converse  *bedrockruntime.Client
}

func NewBedrockClient(ctx context.Context) (*BedrockClient, error) {
	kbID := strings.TrimSpace(os.Getenv("AWS_BEDROCK_KNOWLEDGEBASE_ID"))
	modelARN := strings.TrimSpace(os.Getenv("AWS_LL_MODEL_ARN"))
	if kbID == "" || modelARN == "" {
		return &BedrockClient{}, nil
	}
	cfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(config.GetRegion()))
	if err != nil {
		return nil, err
	}
	return &BedrockClient{
		kbID:     kbID,
		modelARN: modelARN,
		agent:    bedrockagentruntime.NewFromConfig(cfg),
		converse: bedrockruntime.NewFromConfig(cfg),
	}, nil
}

type faqResponse struct {
	Answer     string `json:"answer"`
	References []struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"references"`
}

func (c *BedrockClient) AnswerQuestion(title, content string) (*ext.FAQAnswerResult, error) {
	if c.agent == nil || c.converse == nil {
		return &ext.FAQAnswerResult{
			Content: "（AI回答は未設定です。総務・人事にお問い合わせください。）",
		}, nil
	}
	query := strings.TrimSpace(title + "\n" + content)
	retrieveOut, err := c.agent.Retrieve(context.Background(), &bedrockagentruntime.RetrieveInput{
		KnowledgeBaseId: aws.String(c.kbID),
		RetrievalQuery:  &agenttypes.KnowledgeBaseQuery{Text: aws.String(query)},
	})
	if err != nil {
		return nil, fmt.Errorf("Knowledge Base検索に失敗: %w", err)
	}
	var contextParts []string
	for _, r := range retrieveOut.RetrievalResults {
		if r.Content != nil && r.Content.Text != nil {
			contextParts = append(contextParts, *r.Content.Text)
		}
	}
	prompt := fmt.Sprintf("社内人事・総務FAQアシスタントとして、次の質問に日本語で回答してください。\n質問: %s\n\n参考情報:\n%s\n\nJSON形式で answer と references(name,url配列) を返してください。", query, strings.Join(contextParts, "\n---\n"))
	out, err := c.converse.Converse(context.Background(), &bedrockruntime.ConverseInput{
		ModelId: aws.String(c.modelARN),
		Messages: []brtypes.Message{{
			Role: brtypes.ConversationRoleUser,
			Content: []brtypes.ContentBlock{&brtypes.ContentBlockMemberText{Value: prompt}},
		}},
	})
	if err != nil {
		return nil, fmt.Errorf("Bedrock Converseに失敗: %w", err)
	}
	text := extractText(out)
	var parsed faqResponse
	if err := json.Unmarshal([]byte(text), &parsed); err != nil {
		slog.Warn("bedrock response parse fallback", "text", text)
		return &ext.FAQAnswerResult{Content: text}, nil
	}
	refs := make([]ext.FAQReference, 0, len(parsed.References))
	for _, r := range parsed.References {
		refs = append(refs, ext.FAQReference{Name: r.Name, URL: r.URL})
	}
	return &ext.FAQAnswerResult{Content: parsed.Answer, References: refs}, nil
}

func extractText(out *bedrockruntime.ConverseOutput) string {
	if out == nil || out.Output == nil {
		return ""
	}
	msg, ok := out.Output.(*brtypes.ConverseOutputMemberMessage)
	if !ok {
		return ""
	}
	var b strings.Builder
	for _, block := range msg.Value.Content {
		if t, ok := block.(*brtypes.ContentBlockMemberText); ok {
			b.WriteString(t.Value)
		}
	}
	return b.String()
}
