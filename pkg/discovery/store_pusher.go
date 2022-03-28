package discovery

import (
	"bytes"
	"net/http"
	"time"
)

type storePusher struct {
	addr string
}

type resultHandler func(err error, response *http.Response)

func (p *storePusher) push(buffer *bytes.Buffer, handler resultHandler) {
	go func() {
		c := http.Client{Timeout: time.Second * 5}
		post, err := c.Post(p.addr, `application/json`, buffer)
		if err != nil {
			handler(err, nil)
			return
		}
		defer post.Body.Close()
		handler(err, post)
	}()
}

func NewStorePusher(addr string) *storePusher {
	return &storePusher{addr: buildPusherApiAddress(addr)}
}

func buildPusherApiAddress(addr string) string {
	return addr + ":8080/configuration/stores"
}
