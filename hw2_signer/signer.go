package main

import (
	"sort"
	"strconv"
	"strings"
	"sync"
)

func asyncDataSignerCrc32(data string, out chan string, wg *sync.WaitGroup) {
	defer func() {
		wg.Done()
	}()
	out <- DataSignerCrc32(data)
}

func SingleHash(in, out chan interface{}) {
	for inputVal := range in {
		data := strconv.Itoa(inputVal.(int))
		wg := &sync.WaitGroup{}

		outCrc32FirstPart := make(chan string)
		wg.Add(1)
		go asyncDataSignerCrc32(data, outCrc32FirstPart, wg)

		outCrc32SecondPart := make(chan string)
		wg.Add(1)
		go asyncDataSignerCrc32(DataSignerMd5(data), outCrc32SecondPart, wg)

		wg.Wait()

		firstPart := <-outCrc32FirstPart
		secondPart := <-outCrc32SecondPart

		out <- firstPart + "~" + secondPart
	}
}

func MultiHash(in, out chan interface{}) {
	for inputVal := range in {
		// TODO: все что внутри в цикла можно вынести в горутины
		var result string
		outCrc32 := make(chan string, 6)
		wg := &sync.WaitGroup{}

		for i := 0; i < 6; i++ {
			wg.Add(1)
			go asyncDataSignerCrc32(strconv.Itoa(i)+inputVal.(string), outCrc32, wg)
		}

		wg.Wait()
		close(outCrc32)
		//TODO: тут нужно сортировать то что пришло из канала
		for val := range outCrc32 {
			result += val
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
