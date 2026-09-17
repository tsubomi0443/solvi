package question_usecase

// RunAIForTest exposes internal runAI for usecase unit tests.
func RunAIForTest(uc *QuestionUsecase, questionUUID, title, content string) {
	uc.runAI(questionUUID, title, content)
}
