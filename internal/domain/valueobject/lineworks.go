package valueobject

type LineWorksEvent int

const (
	LineWorksEventReceived LineWorksEvent = 1
	LineWorksEventAnswered LineWorksEvent = 2
	LineWorksEventReopened LineWorksEvent = 3
)

type LineWorksDestination int

const (
	LineWorksDestinationChannel LineWorksDestination = 1
	LineWorksDestinationUser    LineWorksDestination = 2
)

type LineWorksJobStatus int

const (
	LineWorksJobPending LineWorksJobStatus = 1
	LineWorksJobSending LineWorksJobStatus = 2
	LineWorksJobSent    LineWorksJobStatus = 3
	LineWorksJobFailed  LineWorksJobStatus = 4
)

type LineWorksErrorKind int

const (
	LineWorksErrorNone                LineWorksErrorKind = 0
	LineWorksErrorRecipientUnresolved LineWorksErrorKind = 1
	LineWorksErrorBotUnavailable      LineWorksErrorKind = 2
	LineWorksErrorInvalidDestination  LineWorksErrorKind = 3
	LineWorksErrorRateLimited         LineWorksErrorKind = 4
	LineWorksErrorTransient           LineWorksErrorKind = 5
	LineWorksErrorUnknownOutcome      LineWorksErrorKind = 6
)

type LineWorksAttemptOutcome int

const (
	LineWorksAttemptSuccess     LineWorksAttemptOutcome = 1
	LineWorksAttemptRateLimited LineWorksAttemptOutcome = 2
	LineWorksAttemptTransient   LineWorksAttemptOutcome = 3
	LineWorksAttemptUnknown     LineWorksAttemptOutcome = 4
	LineWorksAttemptPermanent   LineWorksAttemptOutcome = 5
)

const LineWorksMaxSendAttempts = 8
