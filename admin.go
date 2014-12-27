package main

import (
  "fmt"
  "net/http"
  "os"
  "github.com/gorilla/mux"
  godis "github.com/simonz05/godis/redis"
  "github.com/threetee/http3go1/lib"
)

func static(w http.ResponseWriter, r *http.Request) {
  fname := mux.Vars(r)["fileName"]
  // empty means, we want to serve the index file. Due to a bug in http.serveFile
  // the file cannot be called index.html, anything else is fine.
  if fname == "" {
    fname = "index.htm"
  }
  staticDir := config.GetStringDefault("static-directory", "")
  staticFile := path.Join(staticDir, fname)
  if fileExists(staticFile) {
    http.ServeFile(w, r, staticFile)
  }
}

func main() {
  router := mux.NewRouter()
  router.HandleFunc("/add/{url:(.*$)}", store)
  router.HandleFunc("/delete/{url:(.*$)}", shorten)

  router.HandleFunc("/{short:([a-zA-Z0-9]+$)}", resolve)
  router.HandleFunc("/{short:([a-zA-Z0-9]+)\\+$}", info)
  router.HandleFunc("/info/{short:[a-zA-Z0-9]+}", info)
  router.HandleFunc("/latest/{data:[0-9]+}", latest)

  router.HandleFunc("/{fileName:(.*$)}", static)

  listen := "0.0.0.0"
  port := os.Getenv("ADMIN_PORT")
  s := &http.Server{
    Addr:    listen + ":" + port,
    Handler: router,
  }
  s.ListenAndServe()
}
