package middleware

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// TODO: remove after debugging
var _ = fmt.Printf

// RedisClient defines a redis client calls
type RedisClient interface {
	RateDel(string) error
	RateEvalSha(string, []string, ...interface{}) (interface{}, error)
	RateScriptLoad(string) (string, error)
}

type abstractLimiter interface {
	getLimit(key string, policy ...int) ([]interface{}, error)
	removeLimit(key string) error
}

// Limiter struct.
type Limiter struct {
	abstractLimiter
}

// Options for Limiter
type Options struct {
	IPLimit  int           // The limit visit times, default is 1000
	Duration time.Duration // Count duration for no policy, default is 1 Minute.
	// Prefix   string        // Redis key prefix, default is "LIMIT:".
	Client RedisClient // Use a redis client for limiter, if omit, it will use a memory limiter.
}

// NumofResult is set for assertion
const NumofResult = 2

// Result includes
// 1. remaining count for specific ip
// 2. reset time for specific ip
type Result struct {
	Remaining int
	Reset     time.Time
}

type redisLimiter struct {
	sha1     string
	ipLimit  string
	duration string
	rc       RedisClient
}

// New returns a Limiter instance with given options.
// If options.Client omit, the limiter is a memory limiter
func New(opts Options) *Limiter {
	if opts.IPLimit <= 0 {
		opts.IPLimit = 1000
	}
	if opts.Duration <= 0 {
		opts.Duration = time.Hour
	}
	if opts.Client == nil {
		return newMemLimiter(&opts)
	}
	return newRedisLimiter(&opts)
}

func newRedisLimiter(opts *Options) *Limiter {
	sha1, err := opts.Client.RateScriptLoad(rateLimitScript)
	if err != nil {
		panic(err)
	}
	r := &redisLimiter{
		rc:       opts.Client,
		sha1:     sha1,
		ipLimit:  strconv.FormatInt(int64(opts.IPLimit), 10),
		duration: strconv.FormatInt(int64(opts.Duration/time.Millisecond), 10),
	}
	return &Limiter{r}
}

func (r *redisLimiter) getLimit(key string, policy ...int) ([]interface{}, error) {
	keys := []string{key}
	args := make([]interface{}, 3, 3)
	args[0] = genTimestamp()
	args[1] = r.ipLimit
	args[2] = r.duration
	res, err := r.rc.RateEvalSha(r.sha1, keys, args...)
	if err != nil {
		return nil, err
	}
	arr, ok := res.([]interface{})
	if ok && len(arr) == NumofResult {
		return arr, nil
	}
	return nil, errors.New("Invalid result")
}

func genTimestamp() string {
	time := time.Now().UnixNano() / 1e6
	return strconv.FormatInt(time, 10)
}

// Get get a result for ip
func (l *Limiter) Get(ip string) (Result, error) {
	var result Result
	res, err := l.getLimit(ip)
	if err != nil {
		return result, err
	}
	result = Result{}

	switch res[1].(type) {
	case time.Time:
		result.Remaining = res[0].(int)
		result.Reset = res[1].(time.Time)
	default:
		result.Remaining = int(res[0].(int64))
		timestamp := res[1].(int64)
		sec := timestamp / 1e3
		nsec := (timestamp - sec*1e3) * 1e6
		result.Reset = time.Unix(sec, nsec)
	}

	return result, nil
}

func (r *redisLimiter) removeLimit(key string) error {
	return r.rc.RateDel(key)
}

// Remove remove limiter record for id
func (l *Limiter) Remove(ip string) error {
	return l.removeLimit(ip)
}

func isNoScriptErr(err error) bool {
	return strings.HasPrefix(
		err.Error(), "NOSCRIPT ")
}
