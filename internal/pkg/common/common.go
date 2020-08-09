package common

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	goredis "github.com/go-redis/redis/v8"
	"github.com/golang/glog"
	"github.com/gorilla/mux"
	"github.com/kelseyhightower/envconfig"
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
	redis  *goredis.Client
	filenotfound string
	redisconf    RedisConf
)

var ctx = context.Background()

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
	go redis.HSet(ctx, redir.Key, "TargetUrl", redir.TargetUrl)
	go redis.HSet(ctx, redir.Key, "SourceUrl", redir.SourceUrl)
	go redis.HSet(ctx, redir.Key, "CreationDate", redir.CreationDate)
	go redis.HSet(ctx, redir.Key, "Clicks", redir.Clicks)
	go redis.SRem(ctx, redirsSet, redir.SourceUrl)
	go redis.SAdd(ctx, redirsSet, redir.SourceUrl)
	return redir
}

// loads the Redirect for the given url. If the key is
// not found, os.Error is returned.
func load(url string) (*Redirect, error) {
	glog.Infof("Loading redirect for: %s", url)
	key := constructRedirKey(url)
	glog.Infof("Redis key: %s", key)

	if ok, _ := redis.HExists(ctx, key, "SourceUrl").Result(); ok {
		redir := new(Redirect)
		redir.Key = key
		reply, _ := redis.HMGet(ctx, key, "TargetUrl", "SourceUrl", "CreationDate", "Clicks").Result()
		target := reply[0].(string)
		source := reply[1].(string)
		creationDate, _ := strconv.ParseInt(reply[2].(string), 10, 64)
		var clicks int64 = 0
		if (reply[3] != nil) {
			clicks, _ = strconv.ParseInt(reply[3].(string), 10, 64)
		}
		redir.TargetUrl, redir.SourceUrl, redir.CreationDate, redir.Clicks = target, source, creationDate, clicks
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
	rs, _ := redis.SMembers(ctx, redirsSetName).Result()
	for _, r := range rs {
		redir, err := load(r)
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
	glog.Infof("running healthcheck")
	fmt.Fprintf(w, "ok")
}

// lookup a URL and redirect
func Resolve(w http.ResponseWriter, r *http.Request) {
	glog.Infof("request: %+v", r)
	glog.Infof("request mux vars: %+v", mux.Vars(r))

	u := r.URL

	if u.Scheme == "" {
		u.Scheme = "http"
	}
	if u.Host == "" {
		u.Host = r.Host
	}

	// glog.Infof("scheme: %s", u.Scheme)
	// glog.Infof("host: %s", u.Host)
	glog.Infof("path: %s", u.Path)

	glog.Infof("url: %s", u.String())

	redir, err := load(u.String())
	if err == nil {
		glog.Infof("Found source URL %s, redirecting to target URL %s", redir.Key, redir.TargetUrl)
		go redis.HIncrBy(ctx, redir.Key, "Clicks", 1)
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
		redisconf.Host = "localhost:6379"
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

	redis = goredis.NewClient(&goredis.Options{
		Addr: 		host,
		DB: 			db,
		Password: passwd,
	})

	if defaultTarget == "" {
		filenotfound = "http://www.google.com"
	} else {
		filenotfound = defaultTarget
	}

}
