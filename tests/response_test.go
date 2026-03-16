// Copyright (c) 2026 Aptlogica Technologies Private Limited
// SPDX-License-Identifier: MIT
// Websites: https://www.aptlogica.com | https://www.serenibase.com
// Support: support@aptlogica.com | support@serenibase.com

package tests

import (
	"github.com/aptlogica/sereni-jwt-provider/internal/models"
	"github.com/aptlogica/sereni-jwt-provider/internal/utils"
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

func TestSendSuccess_ComplexData(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name string
		data interface{}
	}{
		{
			name: "array of strings",
			data: []string{"item1", "item2", "item3"},
		},
		{
			name: "nested map",
			data: map[string]interface{}{
				"user": map[string]string{
					"name":  "Test User",
					"email": "test@example.com",
				},
				"count": 42,
			},
		},
		{
			name: "empty map",
			data: map[string]interface{}{},
		},
		{
			name: "string data",
			data: "simple string data",
		},
		{
			name: "numeric data",
			data: 12345,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			utils.SendSuccess(c, "TEST_CODE", "Test message", tt.data)

			if w.Code != http.StatusOK {
				t.Errorf("expected status 200, got %d", w.Code)
			}

			var response models.SuccessResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			if err != nil {
				t.Fatalf("failed to unmarshal response: %v", err)
			}

			if !response.Success {
				t.Error("expected success to be true")
			}
		})
	}
}

func TestSendError_VariousStatusCodes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		statusCode int
		code       string
	}{
		{
			name:       "unauthorized",
			statusCode: http.StatusUnauthorized,
			code:       "UNAUTHORIZED",
		},
		{
			name:       "forbidden",
			statusCode: http.StatusForbidden,
			code:       "FORBIDDEN",
		},
		{
			name:       "not found",
			statusCode: http.StatusNotFound,
			code:       "NOT_FOUND",
		},
		{
			name:       "conflict",
			statusCode: http.StatusConflict,
			code:       "CONFLICT",
		},
		{
			name:       "unprocessable entity",
			statusCode: http.StatusUnprocessableEntity,
			code:       "UNPROCESSABLE",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			utils.SendError(c, tt.statusCode, tt.code, "Test error message")

			if w.Code != tt.statusCode {
				t.Errorf("expected status %d, got %d", tt.statusCode, w.Code)
			}

			var response models.ErrorResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			if err != nil {
				t.Fatalf("failed to unmarshal response: %v", err)
			}

			if response.Success {
				t.Error("expected success to be false for error response")
			}
			if response.Code != tt.code {
				t.Errorf("expected code %s, got %s", tt.code, response.Code)
			}
			if response.Data != nil {
				t.Error("expected data to be nil for error response")
			}
		})
	}
}

func TestSendCreated_VariousDataTypes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name string
		data interface{}
	}{
		{
			name: "with nil data",
			data: nil,
		},
		{
			name: "with map data",
			data: map[string]interface{}{"id": "new-id", "status": "created"},
		},
		{
			name: "with array data",
			data: []int{1, 2, 3, 4, 5},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			utils.SendCreated(c, "RESOURCE_CREATED", "New resource created", tt.data)

			if w.Code != http.StatusCreated {
				t.Errorf("expected status 201, got %d", w.Code)
			}

			var response models.SuccessResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			if err != nil {
				t.Fatalf("failed to unmarshal response: %v", err)
			}

			if !response.Success {
				t.Error("expected success to be true")
			}
			if response.Code != "RESOURCE_CREATED" {
				t.Errorf("expected code RESOURCE_CREATED, got %s", response.Code)
			}
		})
	}
}

func TestResponseHelpers_EmptyStrings(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("SendSuccess with empty strings", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		utils.SendSuccess(c, "", "", nil)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var response models.SuccessResponse
		json.Unmarshal(w.Body.Bytes(), &response)

		if response.Code != "" {
			t.Errorf("expected empty code, got %s", response.Code)
		}
		if response.Message != "" {
			t.Errorf("expected empty message, got %s", response.Message)
		}
	})

	t.Run("SendError with empty strings", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		utils.SendError(c, http.StatusBadRequest, "", "")

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}

		var response models.ErrorResponse
		json.Unmarshal(w.Body.Bytes(), &response)

		if response.Success {
			t.Error("expected success to be false")
		}
	})
}
