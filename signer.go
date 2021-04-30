package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
)

// сюда писать код

func ExecutePipeline(jobs ...job) {
	wg := &sync.WaitGroup{}
	in := make(chan interface{})
	for _, j := range jobs {
		wg.Add(1)
		out := make(chan interface{})
		go worker(in, out, j, wg)
		in = out
	}
	wg.Wait()
}
func worker(in, out chan interface{}, j job, wg *sync.WaitGroup) {
	defer wg.Done()
	defer close(out)
	j(in, out)
}
func SingleHash(in, out chan interface{}) {
	wg := &sync.WaitGroup{}
	mu := &sync.Mutex{}
	for input := range in {
		wg.Add(1)
		go func(input interface{}) {
			defer wg.Done()
			data := strconv.Itoa(input.(int))
			mu.Lock()
			md5 := DataSignerMd5(data)
			mu.Unlock()
			crc32chan := make(chan string)
			go func(crc32chan chan string, data string) {
				crc32chan <- DataSignerCrc32(data)
			}(crc32chan, data)
			crc32md5chan := make(chan string)
			go func(crc32md5chan chan string, md5 string) {
				crc32md5chan <- DataSignerCrc32(md5)
			}(crc32md5chan, md5)
			crc32 := <-crc32chan
			crc32md5 := <-crc32md5chan
			fmt.Printf("%s SingleHash data %s\n", data, data)
			fmt.Printf("%s SingleHash md5(data) %s\n", data, md5)
			fmt.Printf("%s SingleHash crc32(md5(data)) %s\n", data, crc32md5)
			fmt.Printf("%s SingleHash crc32(data) %s\n", data, crc32)
			fmt.Printf("%s SingleHash result %s\n", data, crc32+"~"+crc32md5)
			result := crc32 + "~" + crc32md5
			out <- result
		}(input)
	}
	wg.Wait()
}
func MultiHash(in, out chan interface{}) {
	wg := &sync.WaitGroup{}
	for input := range in {
		array := make([]string, 6)
		wg.Add(1)
		go func(input interface{}) {
			defer wg.Done()
			waitG := &sync.WaitGroup{}
			for th := 0; th<6; th++ {
				waitG.Add(1)
				data :=  strconv.Itoa(th) + input.(string)
				go func(input string, th int) {
					defer waitG.Done()
					tmp := DataSignerCrc32(data)
					array[th] = tmp
					fmt.Printf("%s MultiHash: crc32(th+data)) %d %s\n", input, th, tmp)
				}(input.(string), th)
			}
			waitG.Wait()
			result := strings.Join(array, "")
			fmt.Printf("%s MultiHash result: %s\n", input, result)
			out <- result
		}(input)
	}
	wg.Wait()
}
func CombineResults(in, out chan interface{}) {
	var tmp []string
	for input := range in {
		tmp = append(tmp, input.(string))
	}
	sort.Strings(tmp)
	result := strings.Join(tmp, "_")
	fmt.Printf("CombineResults \n%s\n", result)
	out <- result
}

