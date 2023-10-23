package controller

import (
    "github.com/gin-gonic/gin"
)


var (
    StatusInvalidAuthParameter = gin.H{"code": 1000, "message": "auth parameter invalid"}
    StatusInvalidParameter = gin.H{"code": 1000, "message": "parameter invalid"}
    StatusTokenNotFound = gin.H{"code": 1001, "message": "token not found"}
    StatusInvalidToken = gin.H{"code": 1001, "message": "token invalid"}
    StatusNoData = gin.H{"code": 2001, "message": "no data"}
    StatusNoPermission = gin.H{"code": 2003, "message": "no permission"}
    StatusFileNotFound = gin.H{"code": 2004, "message": "GetFile no data"}
    StatusStorageNotFound = gin.H{"code": 2004, "message": "GetStorage no data"}
)
