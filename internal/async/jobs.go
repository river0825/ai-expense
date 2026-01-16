package async

import (
	"github.com/google/uuid"
)

// CategorySuggestionJob creates a job for category suggestion
func CategorySuggestionJob(description string, userID string) *Job {
	return &Job{
		ID:       uuid.New().String(),
		Type:     JobTypeCategorySuggestion,
		Priority: PriorityNormal,
		Payload: map[string]interface{}{
			"description": description,
			"user_id":     userID,
		},
	}
}

// NotificationJob creates a job for sending notifications
func NotificationJob(userID string, notificationType string, message string) *Job {
	return &Job{
		ID:       uuid.New().String(),
		Type:     JobTypeNotification,
		Priority: PriorityLow,
		Payload: map[string]interface{}{
			"user_id":             userID,
			"notification_type":   notificationType,
			"message":             message,
		},
	}
}

// MetricsUpdateJob creates a job for updating metrics
func MetricsUpdateJob(userID string, expenseID string, amount float64) *Job {
	return &Job{
		ID:       uuid.New().String(),
		Type:     JobTypeMetricsUpdate,
		Priority: PriorityLow,
		Payload: map[string]interface{}{
			"user_id":    userID,
			"expense_id": expenseID,
			"amount":     amount,
		},
	}
}

// DataExportJob creates a job for data export
func DataExportJob(userID string, format string, startDate string, endDate string) *Job {
	return &Job{
		ID:       uuid.New().String(),
		Type:     JobTypeDataExport,
		Priority: PriorityNormal,
		Payload: map[string]interface{}{
			"user_id":    userID,
			"format":     format,
			"start_date": startDate,
			"end_date":   endDate,
		},
	}
}

// AIParseExpenseJob creates a job for AI-based expense parsing
func AIParseExpenseJob(userID string, messageText string, messengerType string) *Job {
	return &Job{
		ID:       uuid.New().String(),
		Type:     JobTypeAIParseExpense,
		Priority: PriorityNormal,
		Payload: map[string]interface{}{
			"user_id":       userID,
			"message_text":  messageText,
			"messenger_type": messengerType,
		},
	}
}
