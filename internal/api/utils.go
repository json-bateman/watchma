package api

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
)

func printHtmlRes(resp *http.Response) {
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("error reading response body: %v", err)
		return
	}
	bodyString := string(bodyBytes)
	fmt.Println("HTML response body:")
	fmt.Println(bodyString)
}
