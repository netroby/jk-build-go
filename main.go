package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

func callJob(url string, crumb string) {
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		fmt.Println("Failed to call job : ", url)
		return
	}
	fmt.Println("ci_user:", ci_user, " ci_token", ci_token)
	req.SetBasicAuth(ci_user, ci_token)
	req.Header.Add("Jenkins-Crumb", crumb)

	client := &http.Client{}
	resp, _ := client.Do(req)
	fmt.Println("Result code: ", resp.StatusCode)
}

var (
	ci_user  string
	ci_token string
)

func main() {

	pwd, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	config_file := pwd + "/config.json"
	if _, err := os.Stat(config_file); os.IsNotExist(err) {
		fmt.Println("config.json file not exists at the work directory")
		os.Exit(1)
	}

	config_data, err := ioutil.ReadFile(config_file)
	if err != nil {
		fmt.Println(err)
		fmt.Println("Can not read config file")
		os.Exit(1)
	}
	var cf interface{}
	err = json.Unmarshal(config_data, &cf)
	if err != nil {
		fmt.Println(err)
		fmt.Println("Can not read config file, json parse error")
		os.Exit(1)
	}
	config := cf.(map[string]interface{})

	ci_user = config["ci_user"].(string)
	ci_token = config["ci_token"].(string)
	fmt.Println("Got ci_user: ", ci_user, " ci_token:", ci_token)

	r := gin.Default()
	r.GET("/", func(c *gin.Context) {
		urls := c.DefaultQuery("urls", "")
		if len(urls) == 0 {
			c.JSON(200, gin.H{
				"message": "urls empty",
			})
			return
		}
		urls_data, err := base64.StdEncoding.DecodeString(urls)
		if err != nil {
			fmt.Println("urls decode base64 failed")
			fmt.Println("failed:", err)
			return
		}

		fmt.Println("Decoded urls data", string(urls_data))
		url_arr := strings.Split(string(urls_data), "|")
		fmt.Println("url_arr->", url_arr)
		first_url := url_arr[0]
		fmt.Println("first_url->", first_url)

		if strings.Contains(first_url, "/job/") == false {
			c.JSON(200, gin.H{
				"message": "job url not valid",
			})
			return

		}

		fua := strings.Split(first_url, "/job/")

		crurl := fua[0] + "/crumbIssuer/api/json"
		fmt.Println("crumb url", crurl)
		req, err := http.NewRequest("GET", crurl, nil)
		if err != nil {
			c.JSON(412, gin.H{
				"message": "Failed to got crumb",
			})

			return
		}
		req.SetBasicAuth(ci_user, ci_token)
		client := &http.Client{}
		resp, _ := client.Do(req)
		body, _ := ioutil.ReadAll(resp.Body)

		fmt.Println(string(body))

		var f interface{}
		err = json.Unmarshal(body, &f)
		if err != nil {
			c.JSON(412, gin.H{
				"message": "Failed to got crumb due to json error",
			})

			return
		}
		m := f.(map[string]interface{})

		crumb_key := m["crumb"].(string)
		fmt.Println("Crumb key is -> ", crumb_key)

		for _, v := range url_arr {
			fmt.Println("Now call job: ", v)
			callJob(v, crumb_key)
		}

		c.JSON(200, gin.H{
			"message": "hello world",
		})
	})
	fmt.Println("Listing on :18009")
	r.Run(":18009")
}
