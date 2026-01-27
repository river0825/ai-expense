package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/riverlin/aiexpense/internal/usecase"
)

type ReportHandler struct {
	generateReportUC *usecase.GenerateReportUseCase
	jwtSecret        []byte
}

func NewReportHandler(generateReportUC *usecase.GenerateReportUseCase) *ReportHandler {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "default-secret-do-not-use-in-prod"
	}

	return &ReportHandler{
		generateReportUC: generateReportUC,
		jwtSecret:        []byte(secret),
	}
}

// GetReportSummary retrieves the expense report summary
func (h *ReportHandler) GetReportSummary(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 1. Get token from Query Param, Header, or Cookie
	tokenString := r.URL.Query().Get("token")
	if tokenString == "" {
		authHeader := r.Header.Get("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			tokenString = strings.TrimPrefix(authHeader, "Bearer ")
		}
	}
	if tokenString == "" {
		cookie, err := r.Cookie("report_token")
		if err == nil {
			tokenString = cookie.Value
		}
	}

	if tokenString == "" {
		h.writeResponse(w, http.StatusUnauthorized, &Response{Status: "error", Error: "Missing authentication token"})
		return
	}

	// 2. Validate Token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return h.jwtSecret, nil
	})

	if err != nil || !token.Valid {
		h.writeResponse(w, http.StatusUnauthorized, &Response{Status: "error", Error: "Invalid or expired token"})
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		h.writeResponse(w, http.StatusUnauthorized, &Response{Status: "error", Error: "Invalid token claims"})
		return
	}

	userID, ok := claims["sub"].(string)
	if !ok || userID == "" {
		h.writeResponse(w, http.StatusUnauthorized, &Response{Status: "error", Error: "Invalid user ID in token"})
		return
	}

	// 3. Generate Report
	startDateStr := r.URL.Query().Get("start_date")
	endDateStr := r.URL.Query().Get("end_date")

	var report *usecase.ExpenseReport
	var reportErr error

	if startDateStr != "" && endDateStr != "" {
		// Parse dates
		startDate, err1 := time.Parse("2006-01-02", startDateStr)
		endDate, err2 := time.Parse("2006-01-02", endDateStr)

		if err1 != nil || err2 != nil {
			h.writeResponse(w, http.StatusBadRequest, &Response{Status: "error", Error: "Invalid date format. Use YYYY-MM-DD"})
			return
		}

		// Adjust end date to end of day if it's just a date
		endDate = endDate.Add(24*time.Hour - time.Nanosecond)

		report, reportErr = h.generateReportUC.Execute(ctx, &usecase.ReportRequest{
			UserID:     userID,
			ReportType: "custom",
			StartDate:  startDate,
			EndDate:    endDate,
		})
	} else {
		// Default to monthly
		report, reportErr = h.generateReportUC.GenerateMonthlyReport(ctx, userID)
	}

	if reportErr != nil {
		h.writeResponse(w, http.StatusInternalServerError, &Response{Status: "error", Error: reportErr.Error()})
		return
	}

	h.writeResponse(w, http.StatusOK, &Response{Status: "success", Data: report})
}

func (h *ReportHandler) writeResponse(w http.ResponseWriter, status int, resp *Response) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(resp)
}
