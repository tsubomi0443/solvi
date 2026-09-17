package bedrock

import (
	"context"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	brtypes "github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
)

func TestBedrockClient_UnconfiguredReturnsDefault(t *testing.T) {
	os.Unsetenv("AWS_BEDROCK_KNOWLEDGEBASE_ID")
	os.Unsetenv("AWS_LL_MODEL_ARN")

	client, err := NewBedrockClient(context.Background())
	if err != nil {
		t.Fatalf("NewBedrockClient failed: %v", err)
	}

	res, err := client.AnswerQuestion("給与について", "締め日はいつですか？")
	if err != nil {
		t.Fatalf("AnswerQuestion failed: %v", err)
	}
	if res.Content == "" {
		t.Fatal("expected fallback content when unconfigured")
	}
}

func TestExtractText_Branches(t *testing.T) {
	t.Run("nil output", func(t *testing.T) {
		if text := extractText(nil); text != "" {
			t.Fatalf("expected empty string for nil, got %q", text)
		}
		if text := extractText(&bedrockruntime.ConverseOutput{}); text != "" {
			t.Fatalf("expected empty string for nil Output field, got %q", text)
		}
	})

	t.Run("valid message with content block text", func(t *testing.T) {
		val := "hello bedrock"
		out := &bedrockruntime.ConverseOutput{
			Output: &brtypes.ConverseOutputMemberMessage{
				Value: brtypes.Message{
					Content: []brtypes.ContentBlock{
						&brtypes.ContentBlockMemberText{Value: val},
					},
				},
			},
		}
		if text := extractText(out); text != "hello bedrock" {
			t.Fatalf("unexpected extracted text: %q", text)
		}
	})
}
