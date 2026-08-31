// Package httpx は小さな JSON レスポンスヘルパ。
package httpx

import (
	"encoding/json"
	"net/http"
)

// JSON は v を JSON で書き出す。CORS とキャッシュ方針もここで統一する。
func JSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Cache-Control", "s-maxage=600, stale-while-revalidate=86400")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// Error は {"error": msg} 形式で返す。
func Error(w http.ResponseWriter, status int, msg string) {
	JSON(w, status, map[string]string{"error": msg})
}
