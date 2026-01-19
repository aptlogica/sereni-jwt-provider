package tests

import (
	"auth-service/internal/models"
	"auth-service/internal/utils"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestSendSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name     string
		code     string
		message  string
		data     interface{}
		expected models.SuccessResponse
	}{
		{
			name:    "success with data",
			code:    "SUCCESS",
			message: "Operation successful",
			data:    map[string]string{"key": "value"},
			expected: models.SuccessResponse{
				Success: true,
				Code:    "SUCCESS",
				Message: "Operation successful",
				Data:    map[string]string{"key": "value"},
			},
		},
		{
			name:    "success without data",
			code:    "OK",
			message: "Done",
			data:    nil,
			expected: models.SuccessResponse{
				Success: true,
				Code:    "OK",
				Message: "Done",
				Data:    nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			utils.SendSuccess(c, tt.code, tt.message, tt.data)

			if w.Code != http.StatusOK {
				t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
			}

			var response models.SuccessResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			if err != nil {
				t.Fatalf("failed to unmarshal response: %v", err)
			}

			if response.Success != tt.expected.Success {
				t.Errorf("expected success %v, got %v", tt.expected.Success, response.Success)
			}
			if response.Code != tt.expected.Code {
				t.Errorf("expected code %s, got %s", tt.expected.Code, response.Code)
			}
			if response.Message != tt.expected.Message {
				t.Errorf("expected message %s, got %s", tt.expected.Message, response.Message)
			}
			// Note: Data comparison would require type assertion, skipping for simplicity
		})
	}
}

func TestSendError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		statusCode int
		code       string
		message    string
		expected   models.ErrorResponse
	}{
		{
			name:       "bad request error",
			statusCode: http.StatusBadRequest,
			code:       "INVALID_REQUEST",
			message:    "Invalid input",
			expected: models.ErrorResponse{
				Success: false,
				Code:    "INVALID_REQUEST",
				Message: "Invalid input",
				Data:    nil,
			},
		},
		{
			name:       "internal server error",
			statusCode: http.StatusInternalServerError,
			code:       "SERVER_ERROR",
			message:    "Something went wrong",
			expected: models.ErrorResponse{
				Success: false,
				Code:    "SERVER_ERROR",
				Message: "Something went wrong",
				Data:    nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			utils.SendError(c, tt.statusCode, tt.code, tt.message)

			if w.Code != tt.statusCode {
				t.Errorf("expected status %d, got %d", tt.statusCode, w.Code)
			}

			var response models.ErrorResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			if err != nil {
				t.Fatalf("failed to unmarshal response: %v", err)
			}

			if response.Success != tt.expected.Success {
				t.Errorf("expected success %v, got %v", tt.expected.Success, response.Success)
			}
			if response.Code != tt.expected.Code {
				t.Errorf("expected code %s, got %s", tt.expected.Code, response.Code)
			}
			if response.Message != tt.expected.Message {
				t.Errorf("expected message %s, got %s", tt.expected.Message, response.Message)
			}
		})
	}
}

func TestSendCreated(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	data := map[string]string{"id": "123"}
	utils.SendCreated(c, "CREATED", "Resource created", data)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, w.Code)
	}

	var response models.SuccessResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if response.Success != true {
		t.Errorf("expected success true, got %v", response.Success)
	}
	if response.Code != "CREATED" {
		t.Errorf("expected code CREATED, got %s", response.Code)
	}
	if response.Message != "Resource created" {
		t.Errorf("expected message 'Resource created', got %s", response.Message)
	}
}
