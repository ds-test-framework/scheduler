package apiserver

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type DashboardRouter interface {
	Name() string
	SetupRouter(*gin.RouterGroup)
}

func (srv *APIServer) HandleDashboardName(c *gin.Context) {
	if srv.dashboard == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "no dashboard router set",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"name": srv.dashboard.Name(),
	})
}

func (srv *APIServer) HandleDashboard(c *gin.Context) {

}
