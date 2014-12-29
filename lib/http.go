package common

import (
  "net/http"
  "time"
)

func HttpInterceptor(router http.Handler) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
    startTime := time.Now()

    router.ServeHTTP(w, req)

    finishTime := time.Now()
    elapsedTime := finishTime.Sub(startTime)

    switch req.Method {
      case "GET":
      // We may not always want to StatusOK, but for the sake of
      // this example we will
      logAccess(w, req, elapsedTime)
      case "POST":
      // here we might use http.StatusCreated
    }

  })
}
