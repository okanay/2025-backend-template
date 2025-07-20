package StaticRoutesHandler

import (
	"github.com/gin-gonic/gin"
)

type Test struct {
	Name string `json:"name" validate:"required,min=2,max=8"`
}

func (h *Handler) TestValidate(c *gin.Context) {
	var test Test
	if err := h.validationService.Validate(c, &test); err != nil {
		return
	}

	c.JSON(200, gin.H{
		"Status": "Test Completed",
	})
}
