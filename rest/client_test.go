package rest

import (
	"context"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func Test_Gin_server(t *testing.T) {
	r := gin.Default()
	r.GET("/test/get", func(c *gin.Context) {
		c.String(http.StatusOK, "get")
	})
	r.PUT("/test/put", func(c *gin.Context) {
		c.String(http.StatusOK, "put")
	})
	r.POST("/test/post", func(c *gin.Context) {
		c.String(http.StatusOK, "post")
	})
	r.DELETE("/test/delete", func(c *gin.Context) {
		c.String(http.StatusOK, "delete")
	})
	r.Run(":8000")
}

func Test_RESTClient(t *testing.T) {
	c := http.DefaultClient
	host := "http://127.0.0.1:8000"
	hostURL, err := url.Parse(host)
	if err != nil {
		t.Fatalf("parse url error: %s", err.Error())
	}
	t.Logf("host url path: %s, hostURL: %v", hostURL.Path, hostURL)
	restClient, err := NewRESTClient(hostURL, c)
	if err != nil {
		t.Fatalf("new rest client error: %s", err.Error())
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	result := restClient.Get().AbsPath("/test/get").Do(ctx)
	if result.err != nil {
		t.Fatalf("do error: %s", result.err.Error())
	}
	t.Logf("result: %s", string(result.body))
}
