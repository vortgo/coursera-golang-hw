package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
)

func asyncDataSignerCrc32(data string, resultBox *sync.Map, resultIndex int, wg *sync.WaitGroup) {
	defer func() {
		wg.Done()
	}()

	resultBox.Store(resultIndex, DataSignerCrc32(data))
	fmt.Printf("Calc index %v\n", resultIndex)
}

func SingleHash(in, out chan interface{}) {
	for inputVal := range in {
		var Crc32MapResult = sync.Map{}

		data := strconv.Itoa(inputVal.(int))
		wg := &sync.WaitGroup{}

		wg.Add(1)
		go asyncDataSignerCrc32(data, &Crc32MapResult, 1, wg)

		wg.Add(1)
		go asyncDataSignerCrc32(DataSignerMd5(data), &Crc32MapResult, 2, wg)

		wg.Wait()

		firstPart, _ := Crc32MapResult.Load(1)
		secondPart, _ := Crc32MapResult.Load(2)

		out <- firstPart.(string) + "~" + secondPart.(string)
		fmt.Println(`out first hash`)
	}
}

func MultiHash(in, out chan interface{}) {
	for inputVal := range in {
		result := calcMultiHash(inputVal)
		out <- result
	}
}

func calcMultiHash(inputVal interface{}) string {
	var result string
	var Crc32MapResult = sync.Map{}
	outCrc32 := make(chan string, 6)
	wg := &sync.WaitGroup{}
	for i := 0; i < 6; i++ {
		wg.Add(1)
		go asyncDataSignerCrc32(strconv.Itoa(i)+inputVal.(string), &Crc32MapResult, i, wg)
	}
	wg.Wait()
	close(outCrc32)

	for i := 0; i < 6; i++ {
		val, _ := Crc32MapResult.Load(i)
		result += val.(string)
	}
	return result
}

func CombineResults(in, out chan interface{}) {
	var inputResults []string
	for inputVal := range in {
		inputResults = append(inputResults, inputVal.(string))
	}
	sort.Strings(inputResults)

	out <- strings.Join(inputResults, "_")
}

func pipelineWorker(job job, in, out chan interface{}, wg *sync.WaitGroup) {
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
		go pipelineWorker(job, in, out, wg)
		in = out
	}

	wg.Wait()
}
