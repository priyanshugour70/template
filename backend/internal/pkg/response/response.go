package response

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/your-org/your-service/internal/pkg/errors"
)

var istLocation = func() *time.Location {
	loc, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		return time.FixedZone("IST", 5*60*60+30*60)
	}
	return loc
}()

type Body struct {
	Success   bool        `json:"success"`
	Data      interface{} `json:"data,omitempty"`
	Error     *ErrPayload `json:"error,omitempty"`
	Message   string      `json:"message,omitempty"`
	Timestamp string      `json:"timestamp,omitempty"`
}

type ErrPayload struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

type Pagination struct {
	Page       int  `json:"page"`
	Limit      int  `json:"limit"`
	Total      int  `json:"total"`
	TotalPages int  `json:"totalPages"`
	HasNext    bool `json:"hasNext"`
	HasPrev    bool `json:"hasPrev"`
}

func JSON(c *gin.Context, status int, data interface{}, message string) {
	c.JSON(status, Body{
		Success:   true,
		Data:      data,
		Message:   message,
		Timestamp: time.Now().In(istLocation).Format(time.RFC3339),
	})
}

func OK(c *gin.Context, data interface{}, message ...string) {
	msg := ""
	if len(message) > 0 {
		msg = message[0]
	}
	JSON(c, http.StatusOK, data, msg)
}

func Created(c *gin.Context, data interface{}, message ...string) {
	msg := ""
	if len(message) > 0 {
		msg = message[0]
	}
	JSON(c, http.StatusCreated, data, msg)
}

func Accepted(c *gin.Context, data interface{}, message ...string) {
	msg := ""
	if len(message) > 0 {
		msg = message[0]
	}
	JSON(c, http.StatusAccepted, data, msg)
}

func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

func Fail(c *gin.Context, status int, code, message string, details map[string]interface{}) {
	c.JSON(status, Body{
		Success: false,
		Error: &ErrPayload{
			Code:    code,
			Message: message,
			Details: details,
		},
		Timestamp: time.Now().In(istLocation).Format(time.RFC3339),
	})
}

func Error(c *gin.Context, err error) {
	err = errors.NormalizeMissingTableError(err)
	ts := time.Now().In(istLocation).Format(time.RFC3339)
	if ae, ok := err.(*errors.AppError); ok {
		payload := &ErrPayload{Code: ae.Code, Message: ae.Message}
		if len(ae.Details) > 0 {
			payload.Details = ae.Details
		}
		c.JSON(statusFromCode(ae.Code), Body{
			Success:   false,
			Error:     payload,
			Timestamp: ts,
		})
		return
	}
	c.JSON(http.StatusInternalServerError, Body{
		Success:   false,
		Error:     &ErrPayload{Code: errors.CodeInternalError, Message: "internal server error"},
		Timestamp: ts,
	})
}

func PaginatedOK(c *gin.Context, data interface{}, page, limit, total int) {
	totalPages := 1
	if limit > 0 {
		totalPages = (total + limit - 1) / limit
	}
	if totalPages < 1 {
		totalPages = 1
	}
	out := gin.H{
		"success": true,
		"data":    data,
		"pagination": gin.H{
			"page":       page,
			"limit":      limit,
			"total":      total,
			"totalPages": totalPages,
			"hasNext":    page < totalPages,
			"hasPrev":    page > 1,
		},
		"timestamp": time.Now().In(istLocation).Format(time.RFC3339),
	}
	c.JSON(http.StatusOK, out)
}

func statusFromCode(code string) int {
	switch code {
	case errors.CodeNotFound:
		return http.StatusNotFound
	case errors.CodeValidationError, errors.CodeInvalidInput, errors.CodeBadRequest:
		return http.StatusBadRequest
	case errors.CodeUnauthorized, errors.CodeInvalidToken, errors.CodeTokenExpired, errors.CodeInvalidCredentials:
		return http.StatusUnauthorized
	case errors.CodeForbidden:
		return http.StatusForbidden
	case errors.CodeAlreadyExists:
		return http.StatusConflict
	case errors.CodeServiceUnavailable, errors.CodeSchemaNotReady:
		return http.StatusServiceUnavailable
	case errors.CodeTimeout, errors.CodeRateLimitExceeded:
		return http.StatusTooManyRequests
	default:
		return http.StatusInternalServerError
	}
}
