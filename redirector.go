package main

import (
  "flag"
  "time"
  "net/http"
  "github.com/gorilla/mux"
  // "github.com/gorilla/handlers"
  "github.com/kelseyhightower/envconfig"
  "github.com/threetee/http3go1/lib"
  "github.com/golang/glog"
  )

// Conf struct represents the app's configuration.
type Conf struct {
  Debug    bool
  Host     string
  Port     string
}

// conf variable holds the app's configuration.
var conf Conf

func httpInterceptor(router http.Handler) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
    startTime := time.Now()

    router.ServeHTTP(w, req)

    finishTime := time.Now()
    elapsedTime := finishTime.Sub(startTime)

    switch req.Method {
      case "GET":
      // We may not always want to StatusOK, but for the sake of
      // this example we will
      common.LogAccess(w, req, elapsedTime)
      case "POST":
      // here we might use http.StatusCreated
    }

  })
}

func main() {
  flag.Parse()
  defer glog.Flush()

  err := envconfig.Process("redirector", &conf)
  if err != nil {
    glog.Fatalf("Couldn't load environment: %s", err)
  }

  if conf.Host == "" {
    glog.Infof("Host not set, using default of 0.0.0.0")
    conf.Host = "0.0.0.0"
  }
  if conf.Port == "" {
    glog.Infof("Port not set, using default of 9000")
    conf.Port = "9000"
  }

  common.Init()

  glog.Infof("Config: %+v", conf)

  router := mux.NewRouter()
  subRouter := router.Schemes("{scheme:(.*)}").Host("{host:(.*)}").Subrouter()
  subRouter.HandleFunc("/{path:([a-zA-Z0-9]+$)}", common.Resolve)
  // router.Handle("/", httpInterceptor(subRouter))
  http.Handle("/", httpInterceptor(subRouter))

  listen := conf.Host
  port := conf.Port
  addr := listen + ":" + port

  http.ListenAndServe(addr, nil)
}
