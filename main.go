package main

import (
	"github.com/gin-gonic/gin"
	"net"
)

func main() {
	r := gin.Default()
	var ip string
	if addrs, err := net.InterfaceAddrs(); err == nil {
		for _, address := range addrs {
			// 检查ip地址判断是否回环地址
			if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					ip =  ipnet.IP.String()
				}

			}
		}
	}
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"version": "2.0",
			"ip":      ip,
		})
	})
	r.Run("0.0.0.0:8000")
}
