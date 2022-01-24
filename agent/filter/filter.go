package filter

import "net/http"

var (
	filters []func(http.ResponseWriter, *http.Request, *Chain)
)

type Chain struct {
	idx     int
	handler func(http.ResponseWriter, *http.Request)
}

func (chain *Chain) ServeHTTP(writer http.ResponseWriter, req *http.Request) {
	if len(filters) == 0 {
		chain.handler(writer, req)
		return
	}
	chain.idx = 0
	filters[chain.idx](writer, req, chain)
}

func (chain *Chain) Next(writer http.ResponseWriter, req *http.Request) {
	chain.idx++
	if chain.idx >= len(filters) {
		chain.handler(writer, req)
		return
	}
	filters[chain.idx](writer, req, chain)
}

func newChain(handler func(http.ResponseWriter, *http.Request)) *Chain {
	return &Chain{handler: handler}
}

func RegisterHandler(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	chain := newChain(handler)
	http.Handle(pattern, chain)
}

func RegisterFilter(handler func(http.ResponseWriter, *http.Request, *Chain)) {
	if handler == nil {
		return
	}
	filters = append(filters, handler)
}
