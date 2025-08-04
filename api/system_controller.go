package api

import (
	"net/http"
)

type SystemController struct {
}

func NewSystemController() *SystemController {
	return &SystemController{}
}

func (s *SystemController) apiCheck(w http.ResponseWriter, r *http.Request) {
	//log.Println("apiCheck")
	//if err := s.connCheck(); err != nil {
	//	_ = response.ServeJSON(w, http.StatusInternalServerError, err.Error(), nil)
	//	return
	//}
	//response.ServeJSONData(w, "ok", http.StatusOK)
	//return
}
