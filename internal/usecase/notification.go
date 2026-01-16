package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// NotificationUseCase handles notifications and reminders
type NotificationUseCase struct {
	// In production, would have notification repository
}

// NewNotificationUseCase creates a new notification use case
func NewNotificationUseCase() *NotificationUseCase {
	return &NotificationUseCase{}
}

// Notification represents a user notification
type Notification struct {
	ID        string                 `json:"id"`
	UserID    string                 `json:"user_id"`
	Type      string                 `json:"type"` // "budget_alert", "recurring_due", "expense_reminder", "report"
	Title     string                 `json:"title"`
	Message   string                 `json:"message"`
	Data      map[string]interface{} `json:"data,omitempty"`
	IsRead    bool                   `json:"is_read"`
	CreatedAt time.Time              `json:"created_at"`
	ReadAt    *time.Time             `json:"read_at,omitempty"`
}

// CreateNotificationRequest represents a request to create a notification
type CreateNotificationRequest struct {
	UserID  string
	Type    string // "budget_alert", "recurring_due", "expense_reminder", "report"
	Title   string
	Message string
	Data    map[string]interface{}
}

// CreateNotificationResponse represents the response after creating a notification
type CreateNotificationResponse struct {
	ID      string
	Message string
}

// CreateNotification creates a new notification
func (u *NotificationUseCase) CreateNotification(ctx context.Context, req *CreateNotificationRequest) (*CreateNotificationResponse, error) {
	if req.UserID == "" || req.Title == "" {
		return nil, fmt.Errorf("user_id and title are required")
	}

	id := uuid.New().String()

	return &CreateNotificationResponse{
		ID:      id,
		Message: fmt.Sprintf("Notification '%s' created", req.Title),
	}, nil
}

// ListNotificationsRequest represents a request to list notifications
type ListNotificationsRequest struct {
	UserID string
	Unread bool // If true, only return unread notifications
	Limit  int
	Offset int
}

// ListNotificationsResponse represents a list of notifications
type ListNotificationsResponse struct {
	Notifications []*Notification `json:"notifications"`
	Total         int             `json:"total"`
	Unread        int             `json:"unread"`
	Message       string          `json:"message"`
}

// ListNotifications retrieves notifications for a user
func (u *NotificationUseCase) ListNotifications(ctx context.Context, req *ListNotificationsRequest) (*ListNotificationsResponse, error) {
	if req.UserID == "" {
		return nil, fmt.Errorf("user_id is required")
	}

	if req.Limit <= 0 {
		req.Limit = 20
	}

	// In production: query from notifications table
	return &ListNotificationsResponse{
		Notifications: make([]*Notification, 0),
		Total:         0,
		Unread:        0,
		Message:       "No notifications found",
	}, nil
}

// MarkAsReadRequest represents a request to mark notification as read
type MarkAsReadRequest struct {
	UserID         string
	NotificationID string
}

// MarkAsReadResponse represents the response after marking as read
type MarkAsReadResponse struct {
	ID      string
	Message string
}

// MarkAsRead marks a notification as read
func (u *NotificationUseCase) MarkAsRead(ctx context.Context, req *MarkAsReadRequest) (*MarkAsReadResponse, error) {
	if req.UserID == "" || req.NotificationID == "" {
		return nil, fmt.Errorf("user_id and notification_id are required")
	}

	// In production: update notification, set read_at = now
	return &MarkAsReadResponse{
		ID:      req.NotificationID,
		Message: "Notification marked as read",
	}, nil
}

// MarkAllAsReadRequest represents a request to mark all notifications as read
type MarkAllAsReadRequest struct {
	UserID string
}

// MarkAllAsReadResponse represents the response after marking all as read
type MarkAllAsReadResponse struct {
	Count   int
	Message string
}

// MarkAllAsRead marks all notifications as read for a user
func (u *NotificationUseCase) MarkAllAsRead(ctx context.Context, req *MarkAllAsReadRequest) (*MarkAllAsReadResponse, error) {
	if req.UserID == "" {
		return nil, fmt.Errorf("user_id is required")
	}

	// In production: update all unread notifications for user
	return &MarkAllAsReadResponse{
		Count:   0,
		Message: "All notifications marked as read",
	}, nil
}

// DeleteNotificationRequest represents a request to delete a notification
type DeleteNotificationRequest struct {
	UserID         string
	NotificationID string
}

// DeleteNotificationResponse represents the response after deletion
type DeleteNotificationResponse struct {
	ID      string
	Message string
}

// DeleteNotification deletes a notification
func (u *NotificationUseCase) DeleteNotification(ctx context.Context, req *DeleteNotificationRequest) (*DeleteNotificationResponse, error) {
	if req.UserID == "" || req.NotificationID == "" {
		return nil, fmt.Errorf("user_id and notification_id are required")
	}

	// In production: verify ownership and delete
	return &DeleteNotificationResponse{
		ID:      req.NotificationID,
		Message: "Notification deleted",
	}, nil
}

// NotificationPreferences represents user notification preferences
type NotificationPreferences struct {
	UserID              string `json:"user_id"`
	BudgetAlerts        bool   `json:"budget_alerts"`
	RecurringReminders  bool   `json:"recurring_reminders"`
	ReportNotifications bool   `json:"report_notifications"`
	ExpenseReminders    bool   `json:"expense_reminders"`
	DailyDigest         bool   `json:"daily_digest"`
	WeeklyReport        bool   `json:"weekly_report"`
}

// GetPreferencesRequest represents a request to get notification preferences
type GetPreferencesRequest struct {
	UserID string
}

// GetPreferencesResponse represents the response with preferences
type GetPreferencesResponse struct {
	Preferences *NotificationPreferences `json:"preferences"`
	Message     string                   `json:"message"`
}

// GetPreferences retrieves notification preferences for a user
func (u *NotificationUseCase) GetPreferences(ctx context.Context, req *GetPreferencesRequest) (*GetPreferencesResponse, error) {
	if req.UserID == "" {
		return nil, fmt.Errorf("user_id is required")
	}

	// Default preferences
	prefs := &NotificationPreferences{
		UserID:              req.UserID,
		BudgetAlerts:        true,
		RecurringReminders:  true,
		ReportNotifications: true,
		ExpenseReminders:    false,
		DailyDigest:         false,
		WeeklyReport:        true,
	}

	return &GetPreferencesResponse{
		Preferences: prefs,
		Message:     "Preferences retrieved",
	}, nil
}

// UpdatePreferencesRequest represents a request to update preferences
type UpdatePreferencesRequest struct {
	UserID              string
	BudgetAlerts        *bool
	RecurringReminders  *bool
	ReportNotifications *bool
	ExpenseReminders    *bool
	DailyDigest         *bool
	WeeklyReport        *bool
}

// UpdatePreferencesResponse represents the response after updating
type UpdatePreferencesResponse struct {
	Preferences *NotificationPreferences `json:"preferences"`
	Message     string                   `json:"message"`
}

// UpdatePreferences updates notification preferences
func (u *NotificationUseCase) UpdatePreferences(ctx context.Context, req *UpdatePreferencesRequest) (*UpdatePreferencesResponse, error) {
	if req.UserID == "" {
		return nil, fmt.Errorf("user_id is required")
	}

	// In production: get current preferences, update specified fields, save
	prefs := &NotificationPreferences{
		UserID: req.UserID,
	}

	if req.BudgetAlerts != nil {
		prefs.BudgetAlerts = *req.BudgetAlerts
	}
	if req.RecurringReminders != nil {
		prefs.RecurringReminders = *req.RecurringReminders
	}
	if req.ReportNotifications != nil {
		prefs.ReportNotifications = *req.ReportNotifications
	}
	if req.ExpenseReminders != nil {
		prefs.ExpenseReminders = *req.ExpenseReminders
	}
	if req.DailyDigest != nil {
		prefs.DailyDigest = *req.DailyDigest
	}
	if req.WeeklyReport != nil {
		prefs.WeeklyReport = *req.WeeklyReport
	}

	return &UpdatePreferencesResponse{
		Preferences: prefs,
		Message:     "Preferences updated successfully",
	}, nil
}
