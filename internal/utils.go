package internal

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
)

func PrintRes(resp *http.Response) {
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error(fmt.Sprintf("error reading response body: %v", err.Error()))
		return
	}
	bodyString := string(bodyBytes)
	fmt.Println("HTML response body:")
	fmt.Println(bodyString)
}

func WriteJSONError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func WriteJSONResponse(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
