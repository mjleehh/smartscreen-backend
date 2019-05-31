package main

import (
	"fmt"
	"github.com/dghubble/sling"
	"github.com/gin-gonic/gin"
	"net/http"
)

type message struct {
	Msg string `json:"msg" binding:"required"`
}

type deviceEntry struct {
	Id      string `json:"id"`
	Name    string `json:"name"`
	Comment string `json:"comment"`
	Msg     string `json:"msg"`
	Ip      string
}

type deviceEntries struct {
	Devices []deviceEntry `json:"devices"`
}

var devices = make([]deviceEntry, 3)

func runDeviceBackend(retval chan error) {
	deviceBackend := gin.Default()
	deviceBackend.PUT("/api/:deviceId", func(c *gin.Context) {
		deviceId, exists := c.Params.Get("deviceId")
		if (!exists) {
			c.Status(http.StatusInternalServerError)
			return
		}

		var body message
		if err := c.BindJSON(&body); err != nil {
			c.Status(http.StatusBadRequest)
			return
		}

		deviceIp := c.ClientIP()
		for idx, device := range devices {
			if device.Id == deviceId {
				devices[idx].Ip = deviceIp
				devices[idx].Msg = body.Msg
				c.Status(http.StatusOK)
				return
			}
		}
		c.Status(http.StatusNotFound)
		return
	})

	fmt.Println("starting device backend on port :3001")
	err := deviceBackend.Run(":3001")
	if err != nil {
		fmt.Println("failed to start backend")
	}
	retval <- err
}

func runClientBackend(retval chan error)  {
	backend := gin.Default()
	backend.GET("/api/devices", func(c *gin.Context) {
		c.JSON(http.StatusOK, &deviceEntries{Devices:devices})
	})

	backend.PUT("/api/:device/message", func(c *gin.Context) {
		deviceId, exists := c.Params.Get("device")
		if (!exists) {
			c.Status(http.StatusInternalServerError)
			return
		}

		var body message
		if err := c.BindJSON(&body); err != nil {
			c.Status(http.StatusBadRequest)
			return
		}

		for _, device := range devices {
			if device.Id == deviceId {
				var client http.Client
				req, err := sling.New().Put("http://" + device.Ip + "/message").BodyJSON(&body).Request()
				if err != nil {
					c.Status(http.StatusInternalServerError)
					return
				}

				if res, err := client.Do(req); err != nil || res.StatusCode != http.StatusOK {
					c.Status(http.StatusBadRequest)
					return
				}

				c.JSON(http.StatusOK, &body)
				return
			}
		}
		c.Status(http.StatusNotFound)
		return
	})

	fmt.Println("starting client backend on port :3000")
	err := backend.Run(":3000")
	if err != nil {
		fmt.Println("failed to start backend")
	}
	retval <- err
}

func main() {
	devices[0] = deviceEntry{
		Id:      "smartscreen-1",
		Name:    "UDLAP DEVICE",
		Comment: "demo device used in lesson",
		Ip:      "192.168.80.103",
	}
	devices[1] = deviceEntry{
		Id:      "smartscreen-2",
		Name:    "SOME OTHER DEVICE",
		Comment: "a second device",
		Ip:      "172.168.10.1",
	}
	devices[2] = deviceEntry{
		Id:      "smartscreen-3",
		Name:    "ANOTHER DEVICE",
		Comment: "a third device",
		Ip:      "172.168.10.24",
	}

	deviceBackendChan := make(chan error)
	go runDeviceBackend(deviceBackendChan)
	clientBackendChan := make(chan error)
	go runClientBackend(clientBackendChan)
	_ = <- deviceBackendChan
	_ = <- clientBackendChan
}