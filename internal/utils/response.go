// Copyright (c) 2026 Aptlogica Technologies Private Limited
// SPDX-License-Identifier: MIT
// Websites: https://www.aptlogica.com | https://www.serenibase.com
// Support: support@aptlogica.com | support@serenibase.com

package utils

import (
	"github.com/aptlogica/sereni-jwt-provider/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

// SendSuccess sends a successful response
func SendSuccess(c *gin.Context, code string, message string, data interface{}) {
	response := models.SuccessResponse{
		Success: true,
		Code:    code,
		Message: message,
		Data:    data,
	}
	c.JSON(http.StatusOK, response)
}

// SendError sends an error response
func SendError(c *gin.Context, statusCode int, code string, message string) {
	response := models.ErrorResponse{
		Success: false,
		Code:    code,
		Message: message,
		Data:    nil,
	}
	c.JSON(statusCode, response)
}

// SendCreated sends a created response
func SendCreated(c *gin.Context, code string, message string, data interface{}) {
	response := models.SuccessResponse{
		Success: true,
		Code:    code,
		Message: message,
		Data:    data,
	}
	c.JSON(http.StatusCreated, response)
}
