package client

import (
	"fmt"
	"io"
	"net/http"
)

func ExampleClient() {
	httpClient := Client()
	response, err := httpClient.Get("http://detectportal.firefox.com")
	if err != nil {
		fmt.Println("Error:", err.Error())
		return
	}
	if response.StatusCode != http.StatusOK {
		fmt.Println("Error:", "status %s", response.Status)
		return
	}
	content, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Println("Error:", err.Error())
		return
	}
	_ = response.Body.Close()
	fmt.Println(string(content))
	// Output: success
}
