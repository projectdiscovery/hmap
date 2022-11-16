# hmap

Hybrid memory/disk map

## Simple usage example

```go
func main() {
	var wg sync.WaitGroup
	wg.Add(1)
	go normal(&wg)
	wg.Wait()
}

func normal(wg *sync.WaitGroup) {
	defer wg.Done()
	hm, err := hybrid.New(hybrid.DefaultOptions)
	if err != nil {
		log.Fatal(err)
	}
	defer hm.Close()
	err2 := hm.Set("a", []byte("b"))
	if err2 != nil {
		log.Fatal(err2)
	}
	v, ok := hm.Get("a")
	if ok {
		log.Println(v)
	}
}
```
