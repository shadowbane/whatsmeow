package middleware

import (
	"github.com/julienschmidt/httprouter"
	"go.uber.org/zap"
	"net/http"
)

func Auth(next httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		zap.S().Infof("[AUTH MIDDLEWARE]\t%s - %s %s %s", r.RemoteAddr, r.Proto, r.Method, r.URL.RequestURI())
		next(w, r, p)
	}
}
