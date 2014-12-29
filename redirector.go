package main

import (
  "flag"
  "net/http"
  "github.com/gorilla/mux"
  // "github.com/gorilla/handlers"
  "github.com/kelseyhightower/envconfig"
  "github.com/threetee/http3go1/lib"
  "github.com/golang/glog"
  )

// Conf struct represents the app's configuration.
type RedirConf struct {
  Debug    bool
  Host     string
  Port     string
}

// redirConf variable holds the app's configuration.
var redirConf RedirConf

func main() {
  flag.Parse()
  defer glog.Flush()

  err := envconfig.Process("redirector", &redirConf)
  if err != nil {
    glog.Fatalf("Couldn't load environment: %s", err)
  }

  if redirConf.Host == "" {
    glog.Infof("Host not set, using default of 0.0.0.0")
    redirConf.Host = "0.0.0.0"
  }
  if redirConf.Port == "" {
    glog.Infof("Port not set, using default of 9000")
    redirConf.Port = "9000"
  }

  common.Init()

  glog.Infof("Config: %+v", redirConf)

  router := mux.NewRouter()
  subRouter := router.Schemes("{scheme:(.*)}").Host("{host:(.*)}").Subrouter()
  subRouter.HandleFunc("/{path:([a-zA-Z0-9]+$)}", common.Resolve)
  // router.Handle("/", common.HttpInterceptor(subRouter))
  http.Handle("/", common.HttpInterceptor(subRouter))

  listen := redirConf.Host
  port := redirConf.Port
  addr := listen + ":" + port

  http.ListenAndServe(addr, nil)
}
