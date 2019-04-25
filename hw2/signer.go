package main

import (
	"sort"
	"strconv"
	"strings"
	"sync"
)

func ExecutePipeline(jobs ...job) {
	var in = make(chan interface{}, MaxInputDataLen)

	wg := &sync.WaitGroup{}
	for _, job := range jobs {
		out := make(chan interface{}, MaxInputDataLen)
		wg.Add(1)
		go pipelineWorker(job, in, out, wg)
		in = out
	}

	wg.Wait()
}

func SingleHash(in, out chan interface{}) {
	mu := &sync.Mutex{}
	wg := &sync.WaitGroup{}
	for inputVal := range in {
		wg.Add(1)
		go calcSingleHash(inputVal, out, wg, mu)
	}
	wg.Wait()
}

func MultiHash(in, out chan interface{}) {
	wg := &sync.WaitGroup{}
	for inputVal := range in {
		wg.Add(1)
		go calcMultiHash(inputVal, out, wg)
	}
	wg.Wait()
}

func CombineResults(in, out chan interface{}) {
	var inputResults []string
	for inputVal := range in {
		inputResults = append(inputResults, inputVal.(string))
	}
	sort.Strings(inputResults)

	out <- strings.Join(inputResults, "_")
}

func calcSingleHash(inputVal interface{}, out chan interface{}, wgParent *sync.WaitGroup, mu *sync.Mutex) {
	defer func() {
		wgParent.Done()
	}()

	var Crc32MapResult = sync.Map{}
	data := strconv.Itoa(inputVal.(int))
	wg := &sync.WaitGroup{}

	wg.Add(1)
	go asyncDataSignerCrc32(data, &Crc32MapResult, 1, wg)
	wg.Add(1)
	go asyncDataSignerCrc32(safeCalcDataSignerMd5(data, mu), &Crc32MapResult, 2, wg)
	wg.Wait()

	firstPart, _ := Crc32MapResult.Load(1)
	secondPart, _ := Crc32MapResult.Load(2)
	out <- firstPart.(string) + "~" + secondPart.(string)
}

func calcMultiHash(inputVal interface{}, out chan interface{}, wgParent *sync.WaitGroup) {
	defer func() {
		wgParent.Done()
	}()

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
	out <- result
}

func pipelineWorker(job job, in, out chan interface{}, wg *sync.WaitGroup) {
	defer wg.Done()
	job(in, out)
	close(out)
}

func asyncDataSignerCrc32(data string, resultBox *sync.Map, resultIndex int, wg *sync.WaitGroup) {
	defer func() {
		wg.Done()
	}()

	resultBox.Store(resultIndex, DataSignerCrc32(data))
}

func safeCalcDataSignerMd5(data string, mu *sync.Mutex) string {
	mu.Lock()
	result := DataSignerMd5(data)
	mu.Unlock()
	return result
}
