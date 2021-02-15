package middleware

import "time"
import "net/http"
import "sync"
import "strconv"

type Metadata struct {
	count      int
	timeExpire time.Time
}

// ServerHandler is recorded for server usages
type ServerHandler struct {
	mu        sync.Mutex
	visit     map[string]Metadata
	Ratelimit int
}

// NewMiddleware creates a handler
func NewMiddleware(ratelimit int) *ServerHandler {
	mw := &ServerHandler{
		visit:     make(map[string]Metadata),
		Ratelimit: ratelimit,
	}
	return mw
}

func (h *ServerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	t := GetCurrentTime(r)
	addr := GetIP(r)
	var data Metadata
	h.mu.Lock()
	if data, exist := h.visit[addr]; !exist {
		h.visit[addr] = Metadata{
			count:      1,
			timeExpire: GetNextExpireTime(t),
		}
	} else {
		// check whether visit times reach limit
		if data.count >= h.Ratelimit {
			after := t.After(data.timeExpire) || t.Equal(data.timeExpire)
			if after {
				h.visit[addr] = Metadata{
					count:      1,
					timeExpire: GetNextExpireTime(t),
				}
			} else {
				http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
				h.mu.Unlock()
				return
			}
		} else {
			// TODO: golang cannot assign to struct field in map, is there idiomatic way to implement?
			h.visit[addr] = Metadata{
				count:      data.count + 1,
				timeExpire: data.timeExpire,
			}
		}
	}
	data = h.visit[addr]

	h.mu.Unlock()

	// Add header
	w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(h.Ratelimit-data.count))
	w.Header().Set("X-RateLimit-Reset", data.timeExpire.Sub(t).String())
}

// GetCurrentTime return current time, Header.Date if in test mode
func GetCurrentTime(r *http.Request) time.Time {
	t := r.Header.Get("Date")
	if t != "" {
		// format: 2021-01-01 12:00:00 +0800 CST
		tm, _ := time.Parse(time.RFC3339, t)
		return tm
	}
	return time.Now()
}

// GetNextExpireTime return time round to next hour
func GetNextExpireTime(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour()+1, 0, 0, 0, t.Location())
}

// GetIP get client IP address
// 1. Check Header "X-Forwarded-For"
// 2. Check RemoteAddress
func GetIP(r *http.Request) string {
	addr := r.Header.Get("X-Forwarded-For")
	if addr != "" {
		return addr
	}
	return r.RemoteAddr
}