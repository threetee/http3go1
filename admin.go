package main

import (
  "net/http"
  "flag"
  "path"
  "time"

  "github.com/kelseyhightower/envconfig"
  "github.com/gorilla/mux"
  // godis "github.com/simonz05/godis/redis"
  "github.com/threetee/http3go1/lib"
  "github.com/golang/glog"
)

// Conf struct represents the app's configuration.
type Conf struct {
  Debug    bool
  Host     string
  Port     string
  StaticDir string
}

// conf variable holds the app's configuration.
var conf Conf

func static(w http.ResponseWriter, r *http.Request) {
  fname := mux.Vars(r)["fileName"]
  // empty means, we want to serve the index file. Due to a bug in http.serveFile
  // the file cannot be called index.html, anything else is fine.
  if fname == "" {
    fname = "index.htm"
  }
  staticDir := conf.StaticDir
  staticFile := path.Join(staticDir, fname)
  if common.FileExists(staticFile) {
    http.ServeFile(w, r, staticFile)
  }
}

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

  err := envconfig.Process("admin", &conf)
  if err != nil {
    glog.Fatalf("Couldn't load environment: %s", err)
  }

  if conf.Host == "" {
    glog.Infof("Host not set, using default of 0.0.0.0")
    conf.Host = "0.0.0.0"
  }
  if conf.Port == "" {
    glog.Infof("Port not set, using default of 9001")
    conf.Port = "9001"
  }

  common.Init()

  glog.Infof("Config: %+v", conf)

  router := mux.NewRouter()
  // router.HandleFunc("/add/{url:(.*$)}", store)
  // router.HandleFunc("/delete/{url:(.*$)}", shorten)
  router.HandleFunc("/urls", create).Methods("POST")

  // router.HandleFunc("/{short:([a-zA-Z0-9]+$)}", resolve)
  // router.HandleFunc("/{short:([a-zA-Z0-9]+)\\+$}", info)
  // router.HandleFunc("/info/{short:[a-zA-Z0-9]+}", info)
  router.HandleFunc("/latest/{data:[0-9]+}", common.Latest)

  router.HandleFunc("/{fileName:(.*$)}", static)

  http.Handle("/", httpInterceptor(router))

  listen := conf.Host
  port := conf.Port
  addr := listen + ":" + port

  http.ListenAndServe(addr, nil)
}
