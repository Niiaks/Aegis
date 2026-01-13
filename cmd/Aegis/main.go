package main

import (
	"fmt"
	"sync"
)

func main() {
 var wg sync.WaitGroup

 for i := 0; i < 3; i++ {
  defer fmt.Println("defer:", i)

  wg.Add(1)
  go func(n int) {
   defer wg.Done()
   fmt.Println("goroutine:", n)
  }(i)
 }

 wg.Wait()
}