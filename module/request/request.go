package request

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"io/ioutil"
)

func GetJson(c *gin.Context) (map[string]interface{},error ){
	jsonString, _ := ioutil.ReadAll(c.Request.Body)
	var data map[string]interface{}
	err := json.Unmarshal(jsonString, &data)
	return data,err
}

