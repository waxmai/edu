package core

import "github.com/gin-gonic/gin"

// NewContextForTest creates a core.Context from a gin.Context for tests.
func NewContextForTest(ctx *gin.Context) Context {
	return newContext(ctx)
}
