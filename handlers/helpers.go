package handlers

import (
	"encoding/json"
	"net/http"
)

func RespondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	// Đảm bảo đặt charset=utf-8 trong header
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)

	response, err := json.Marshal(payload)
	if err != nil {
		// Xử lý lỗi
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Lỗi khi chuyển đổi dữ liệu"}`))
		return
	}

	w.Write(response)
}
func RespondWithError(w http.ResponseWriter, code int, message string) {
	RespondWithJSON(w, code, map[string]string{"error": message})
}
