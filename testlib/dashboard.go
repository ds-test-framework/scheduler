package testlib

import "github.com/gin-gonic/gin"

func (srv *TestingServer) SetupRouter(router *gin.Engine) {

}

func (srv *TestingServer) Name() string {
	return "Testlib"
}
