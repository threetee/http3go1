package main

import (
  "net/http"
  "flag"
  "path"

  "github.com/kelseyhightower/envconfig"
  "github.com/gorilla/mux"
  // godis "github.com/simonz05/godis/redis"
  "github.com/threetee/http3go1/lib"
  "github.com/golang/glog"
)

// AdminConf struct represents the app's configuration.
type AdminConf struct {
  Debug    bool
  Host     string
  Port     string
  StaticDir string
}

// adminConf variable holds the app's configuration.
var adminConf AdminConf

func static(w http.ResponseWriter, r *http.Request) {
  fname := mux.Vars(r)["fileName"]
  // empty means, we want to serve the index file. Due to a bug in http.serveFile
  // the file cannot be called index.html, anything else is fine.
  if fname == "" {
    fname = "index.htm"
  }
  staticDir := adminConf.StaticDir
  staticFile := path.Join(staticDir, fname)
  if common.FileExists(staticFile) {
    http.ServeFile(w, r, staticFile)
  }
}

func main() {
  flag.Parse()
  defer glog.Flush()

  err := envconfig.Process("admin", &adminConf)
  if err != nil {
    glog.Fatalf("Couldn't load environment: %s", err)
  }

  if adminConf.Host == "" {
    glog.Infof("Host not set, using default of 0.0.0.0")
    adminConf.Host = "0.0.0.0"
  }
  if adminConf.Port == "" {
    glog.Infof("Port not set, using default of 9001")
    adminConf.Port = "9001"
  }

  common.Init("")

  glog.Infof("Config: %+v", adminConf)

  router := mux.NewRouter()
  // router.HandleFunc("/add/{url:(.*$)}", store)
  // router.HandleFunc("/delete/{url:(.*$)}", shorten)
  // router.HandleFunc("/urls", create).Methods("POST")

  // router.HandleFunc("/{short:([a-zA-Z0-9]+$)}", resolve)
  // router.HandleFunc("/{short:([a-zA-Z0-9]+)\\+$}", info)
  // router.HandleFunc("/info/{short:[a-zA-Z0-9]+}", info)
  router.HandleFunc("/latest/{data:[0-9]+}", common.Latest)

  router.HandleFunc("/{fileName:(.*$)}", static)

  http.Handle("/", common.HttpInterceptor(router))

  listen := adminConf.Host
  port := adminConf.Port
  addr := listen + ":" + port

  http.ListenAndServe(addr, nil)
}
