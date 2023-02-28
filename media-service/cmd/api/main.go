package main

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"sync"

	"github.com/bernardn38/socialsphere/image-service/application"
)

func main() {
	var wg sync.WaitGroup
	go func() {
		fmt.Println(http.ListenAndServe(":6060", nil))
	}()
	wg.Add(1) // pprof - so we won't exit prematurely
	go application.New().Run()
	wg.Wait()
}
