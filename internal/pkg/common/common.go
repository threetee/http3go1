package common

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/golang/glog"
	"github.com/gorilla/mux"
	"github.com/kelseyhightower/envconfig"
	godis "github.com/simonz05/godis/redis"
)

const (
	// special key in redis, that is our global counter
	COUNTER = "__counter__"
	HTTP    = "http"
)

type RedisConf struct {
	Prefix string
	Host   string
	DB     string
	Pass   string
}

type Redirect struct {
	Key          string `json:"key"`
	SourceUrl    string `json:"source_url"`
	TargetUrl    string `json:"target_url"`
	CreationDate int64  `json:"creation_date"`
	Clicks       int64  `json:"clicks"`
}

var (
	redis        *godis.Client
	filenotfound string
	redisconf    RedisConf
)

// Converts the Redirect to JSON.
func (r Redirect) Json() []byte {
	b, _ := json.Marshal(r)
	return b
}

func constructRedirKey(url string) string {
	return redisconf.Prefix + ":redirect:" + url
}

func constructRedirsSetName() string {
	return redisconf.Prefix + ":redirects"
}

// Creates a new Redirect instance. The Given key, sourceurl and targeturl will
// be used. Clicks will be set to 0 and CreationDate to time.Nanoseconds()
func NewRedirect(sourceurl, targeturl string) *Redirect {
	redir := new(Redirect)
	redir.Key = constructRedirKey(sourceurl)
	redir.CreationDate = time.Now().UnixNano()
	redir.TargetUrl = targeturl
	redir.SourceUrl = sourceurl
	redir.Clicks = 0
	return redir
}

// stores a new Redirect for the given key, sourceurl and targeturl. Existing
// ones with the same url will be overwritten
// TODO: consider using MULTI to ensure data integrity
func store(sourceurl, targeturl string) *Redirect {
	redir := NewRedirect(sourceurl, targeturl)
	redirsSet := constructRedirsSetName()
	go redis.Hset(redir.Key, "TargetUrl", redir.TargetUrl)
	go redis.Hset(redir.Key, "SourceUrl", redir.SourceUrl)
	go redis.Hset(redir.Key, "CreationDate", redir.CreationDate)
	go redis.Hset(redir.Key, "Clicks", redir.Clicks)
	go redis.Srem(redirsSet, redir.SourceUrl)
	go redis.Sadd(redirsSet, redir.SourceUrl)
	return redir
}

// loads the Redirect for the given url. If the key is
// not found, os.Error is returned.
func load(url string) (*Redirect, error) {
	glog.Infof("Loading redirect for: %s", url)
	key := constructRedirKey(url)
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

//Returns a json array with information about all redirects.
// TODO: allow number of entries and offset to be passed in for pagination
func ListRedirects(w http.ResponseWriter, r *http.Request) {
	glog.Infof("Loading redirects")
	w.Header().Set("Content-Type", "application/json")
	var redirs = []*Redirect{}

	redirsSetName := constructRedirsSetName()
	rs, _ := redis.Smembers(redirsSetName)
	for _, r := range rs.Elems {
		redir, err := load(r.Elem.String())
		if err == nil {
			glog.Infof("Appending redirect")
			redirs = append(redirs, redir)
		} else {
			glog.Fatal(err)
		}
	}

	s, _ := json.Marshal(redirs)
	w.Write(s)
}

// Creates a new redirect
func CreateRedirect(w http.ResponseWriter, r *http.Request) {
	glog.Infof("Creating redirect")
	w.Header().Set("Content-Type", "application/json")

	decoder := json.NewDecoder(r.Body)
	var redir Redirect
	err := decoder.Decode(&redir)
	if err != nil {
		glog.Fatal(err)
	}

	s, _ := json.Marshal(store(redir.SourceUrl, redir.TargetUrl))
	w.Write(s)
}

// healthcheck
func Healthcheck(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "ok")
}

// lookup a URL and redirect
func Resolve(w http.ResponseWriter, r *http.Request) {
	glog.Infof("request: %+v", r)
	glog.Infof("request mux vars: %+v", mux.Vars(r))

	scheme := mux.Vars(r)["scheme"]
	host := mux.Vars(r)["host"]
	path := mux.Vars(r)["path"]
	u, err := url.Parse("")
	if err != nil {
		glog.Fatal(err)
	}
	u.Scheme = scheme
	u.Host = host
	u.Path = path

	if u.Scheme == "" {
		u.Scheme = "http"
	}

	glog.Infof("url: %s", u.String())

	redir, err := load(u.String())
	if err == nil {
		glog.Infof("Found source URL %s, redirecting to target URL %s", redir.Key, redir.TargetUrl)
		go redis.Hincrby(redir.Key, "Clicks", 1)
		http.Redirect(w, r, redir.TargetUrl, http.StatusMovedPermanently)
	} else {
		glog.Infof("Error: %s. Redirecting to default target URL %s", err, filenotfound)
		http.Redirect(w, r, filenotfound, http.StatusMovedPermanently)
	}
}

func Init(defaultTarget string) {
	glog.Info("Initializing")
	err := envconfig.Process("redis", &redisconf)

	if redisconf.Host == "" {
		redisconf.Host = "tcp:localhost:6379"
	}
	if redisconf.DB == "" {
		redisconf.DB = "0"
	}
	if redisconf.Prefix == "" {
		redisconf.Prefix = "h3g1"
	}
	glog.Infof("redis host:" + redisconf.Host)
	glog.Infof("redis DB:" + redisconf.DB)
	glog.Infof("redis prefix:" + redisconf.Prefix)

	host := redisconf.Host
	db, err := strconv.Atoi(redisconf.DB)
	if err != nil {
		glog.Fatal(err)
	}
	passwd := redisconf.Pass

	redis = godis.New(host, db, passwd)

	if defaultTarget == "" {
		filenotfound = "http://www.google.com"
	} else {
		filenotfound = defaultTarget
	}

}
