package domain

import "fmt"

type ErrorCode string

const (
	ErrorCodeTeamExists  ErrorCode = "TEAM_EXISTS"
	ErrorCodePRExists    ErrorCode = "PR_EXISTS"
	ErrorCodePRMerged    ErrorCode = "PR_MERGED"
	ErrorCodeNotAssigned ErrorCode = "NOT_ASSIGNED"
	ErrorCodeNoCandidate ErrorCode = "NO_CANDIDATE"
	ErrorCodeNotFound    ErrorCode = "NOT_FOUND"
)

type DomainError struct {
	Code    ErrorCode
	Message string
}

func (e *DomainError) Error() string {
	if e.Message == "" {
		return string(e.Code)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func NewDomainError(code ErrorCode, msg string) *DomainError {
	return &DomainError{
		Code:    code,
		Message: msg,
	}
}
