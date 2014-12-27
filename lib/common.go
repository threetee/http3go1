package common

import (
  "code.google.com/p/gorilla/mux"
  "encoding/json"
  "errors"
  "net/http"
  "net/url"
  "time"
  godis "github.com/simonz05/godis/redis"
  "log"
)

const (
  // special key in redis, that is our global counter
  COUNTER = "__counter__"
  HTTP    = "http"
)

var (
  redis        *godis.Client
  filenotfound string
)

type Redirect struct {
  Key          string
  SourceUrl    string
  TargetUrl    string
  CreationDate int64
  Clicks       int64
}

// Converts the Redirect to JSON.
func (r Redirect) Json() []byte {
  b, _ := json.Marshal(r)
  return b
}

// Creates a new Redirect instance. The Given key, sourceurl and targeturl will
// be used. Clicks will be set to 0 and CreationDate to time.Nanoseconds()
func NewRedirect(key, sourceurl, targeturl string) *Redirect {
  redir := new(Redirect)
  redir.CreationDate = time.Now().UnixNano()
  redir.Key = key
  redir.TargetUrl = targeturl
  redir.SourceUrl = sourceurl
  redir.Clicks = 0
  return redir
}

// stores a new Redirect for the given key, sourceurl and targeturl. Existing
// ones with the same url will be overwritten
func store(key, sourceurl, targeturl string) *Redirect {
  redir := NewRedirect(key, sourceurl, targeturl)
  go redis.Hset(redir.Key, "TargetUrl", redir.TargetUrl)
  go redis.Hset(redir.Key, "SourceUrl", redir.SourceUrl)
  go redis.Hset(redir.Key, "CreationDate", redir.CreationDate)
  go redis.Hset(redir.Key, "Clicks", redir.Clicks)
  return redir
}

// loads the Redirect for the given key. If the key is
// not found, os.Error is returned.
func load(key string) (*Redirect, error) {
  if ok, _ := redis.Hexists(key, "SourceUrl"); ok {
    redir := new(Redirect)
    redir.Key = key
    reply, _ := redis.Hmget(key, "TargetUrl", "SourceUrl", "CreationDate", "Clicks")
    redir.TargetUrl, redir.SourceUrl, redir.CreationDate, redir.Clicks =
    reply.Elems[0].Elem.String(), reply.Elems[1].Elem.String(),
    reply.Elems[2].Elem.Int64(), reply.Elems[3].Elem.Int64()
    return redir, nil
  }
  return nil, errors.New("unknown key: " + key)
}

// //Returns a json array with information about the last shortened urls. If data
// // is a valid integer, that's the amount of data it will return, otherwise
// // a maximum of 10 entries will be returned.
// func latest(w http.ResponseWriter, r *http.Request) {
//   data := mux.Vars(r)["data"]
//   howmany, err := strconv.ParseInt(data, 10, 64)
//   if err != nil {
//     howmany = 10
//   }
//   c, _ := redis.Get(COUNTER)
//
//   last := c.Int64()
//   upTo := (last - howmany)
//
//   w.Header().Set("Content-Type", "application/json")
//
//   var redirs = []*Redirect{}
//
//   for i := last; i > upTo && i > 0; i -= 1 {
//     redir, err := load(Encode(i))
//     if err == nil {
//       redirs = append(redirs, redir)
//     }
//   }
//   s, _ := json.Marshal(redirs)
//   w.Write(s)
// }

// function to translate a URL and redirect
func Resolve(w http.ResponseWriter, r *http.Request) {
  scheme := mux.Vars(r)["scheme"]
  host := mux.Vars(r)["host"]
  path := mux.Vars(r)["path"]
  u, err := url.Parse("")
  if err != nil {
    log.Fatal(err)
  }
  u.Scheme = scheme
  u.Host = host
  u.Path = path

  redir, err := load(u.String())
  if err == nil {
    go redis.Hincrby(redir.Key, "Clicks", 1)
    http.Redirect(w, r, redir.TargetUrl, http.StatusMovedPermanently)
  } else {
    http.Redirect(w, r, filenotfound, http.StatusMovedPermanently)
  }
}
