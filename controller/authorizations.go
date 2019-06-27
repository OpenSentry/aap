package controller

import (
  "github.com/gin-gonic/gin"
  "net/http"
)

func GetAuthorizations(c *gin.Context) {
  c.JSON(http.StatusOK, gin.H{
    "message": "pong",
  })
}

func PostAuthorizations(c *gin.Context) {
  c.JSON(http.StatusOK, gin.H{
    "message": "pong",
  })
}

func PutAuthorizations(c *gin.Context) {
  c.JSON(http.StatusOK, gin.H{
    "message": "pong",
  })
}
