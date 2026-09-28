package api

import (
	"solvi/internal/application/converter"
	outputmodel "solvi/internal/application/model/output_model"
)

func faqUpdateFromDetail(q *outputmodel.QuestionDetailOutput) outputmodel.FAQUpdateOutput {
	if q == nil {
		return outputmodel.FAQUpdateOutput{}
	}
	return converter.QuestionDetailToFAQUpdate(*q)
}

func faqUpdateRemoved(summaryUUID string) outputmodel.FAQUpdateOutput {
	return outputmodel.FAQUpdateOutput{
		SupportStatus: "pending",
		Summary:       &outputmodel.SummaryListItemOutput{UUID: summaryUUID},
	}
}
