package testlib

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// SetupRouter for setting up the dashboard routes implements DashboardRouter
func (srv *TestingServer) SetupRouter(router *gin.RouterGroup) {
	router.GET("/testcases", srv.handleTestCases)
	router.GET("/testcase/:name", srv.handleTestCase)
}

// Name implements DashboardRouter
func (srv *TestingServer) Name() string {
	return "Testlib"
}

type testCaseResponse struct {
	Name    string `json:"name"`
	Timeout string `json:"timeout"`
	Handler string `json:"handler"`
}

func (srv *TestingServer) handleTestCases(c *gin.Context) {
	responses := make([]*testCaseResponse, len(srv.testCases))
	i := 0
	for _, t := range srv.testCases {
		responses[i] = &testCaseResponse{
			Name:    t.Name,
			Timeout: t.Timeout.String(),
			Handler: t.Handler.Name(),
		}
		i++
	}
	c.JSON(http.StatusOK, gin.H{"testcases": responses})
}

func (srv *TestingServer) handleTestCase(c *gin.Context) {
	name, ok := c.Params.Get("name")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing param `name`"})
		return
	}
	response := gin.H{}
	testcase, ok := srv.testCases[name]
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "no such testcase"})
		return
	}
	response["testcase"] = &testCaseResponse{
		testcase.Name,
		testcase.Timeout.String(),
		testcase.Handler.Name(),
	}
	report, ok := srv.ReportStore.GetReport(name)
	if ok {
		response["report"] = report
	}
	c.JSON(http.StatusOK, response)
}
