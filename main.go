package main

import "github.com/gin-gonic/gin"

func main() {
	r := gin.Default()
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"version": "1.0",
		})
	})
	r.Run("0.0.0.0:8000")
}
