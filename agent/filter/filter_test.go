package filter

import (
	"net/http"
	"testing"
)

func TestChain_Next(t *testing.T) {
	const filterAmount = 5
	for amount := 0; amount < 5; amount++ {
		filters = filters[:0]
		step := 0
		for idx := 0; idx < amount; idx++ {
			RegisterFilter(func(filterIndex int) func(writer http.ResponseWriter, request *http.Request, chain *Chain) {
				return func(writer http.ResponseWriter, request *http.Request, chain *Chain) {
					if filterIndex != step {
						t.Error("未第一步执行")
					}
					step++
					chain.Next(writer, request)
				}
			}(idx))
		}
		chain := newChain(func(writer http.ResponseWriter, request *http.Request) {
			if step != amount {
				t.Error("次数错误")
			}
			step++
		})
		chain.ServeHTTP(nil, nil)
		if step != (amount + 1) {
			t.Error("handler未执行")
		}
		step = 0
		chain.ServeHTTP(nil, nil)
		if step != (amount + 1) {
			t.Error("handler未执行")
		}
	}
}
