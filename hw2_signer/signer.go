package main

import (
	"sort"
	"strconv"
	"strings"
	"sync"
)

func SingleHash(in, out chan interface{}) {
	for inputVal := range in {
		data := strconv.Itoa(inputVal.(int))
		out <- DataSignerCrc32(data) + "~" + DataSignerCrc32(DataSignerMd5(data))
	}
}

func MultiHash(in, out chan interface{}) {
	for inputVal := range in {
		var result string
		for i := 0; i < 6; i++ {
			result += DataSignerCrc32(strconv.Itoa(i) + inputVal.(string))
		}
		out <- result
	}
}

func CombineResults(in, out chan interface{}) {
	var inputResults []string
	for inputVal := range in {
		inputResults = append(inputResults, inputVal.(string))
	}
	sort.Strings(inputResults)

	out <- strings.Join(inputResults, "_")
}

func worker(job job, in, out chan interface{}, wg *sync.WaitGroup) {
	defer wg.Done()
	job(in, out)
	close(out)
}

func ExecutePipeline(jobs ...job) {
	var (
		in = make(chan interface{})
	)

	wg := &sync.WaitGroup{}
	for _, job := range jobs {
		out := make(chan interface{}, 100)
		wg.Add(1)
		go worker(job, in, out, wg)
		in = out
	}

	wg.Wait()
}
