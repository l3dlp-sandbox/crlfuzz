package runner

import (
	"fmt"
	"strings"
	"sync"

	"dw1.io/crlfuzz/pkg/crlfuzz"
	"dw1.io/crlfuzz/pkg/errors"
	"github.com/logrusorgru/aurora"
)

// New will fuzz target line by line
func New(options *Options) {
	jobs := make(chan string)
	var wg sync.WaitGroup

	for i := 0; i < options.Concurrency; i++ {
		wg.Add(1)
		go func() {
			for target := range jobs {
				options.run(target)
			}
			defer wg.Done()
		}()
	}

	for _, line := range strings.Split(options.Target, "\n") {
		if isURL(line) {
			for _, url := range crlfuzz.GenerateURL(line) {
				jobs <- url
			}
		}
	}

	close(jobs)
	wg.Wait()
}

func (options *Options) run(url string) {
	v, e := crlfuzz.Scan(
		url,
		options.Method,
		options.Data,
		options.Headers,
		options.Proxy,
	)
	if e != nil {
		if !options.Silent {
			if options.Verbose {
				errors.Show(e.Error())
			} else {
				errors.Show(url)
			}
		}
	}
	if v {
		if options.Silent {
			fmt.Println(url)
		} else {
			fmt.Printf("[%s] %s\n", aurora.Green("VLN").String(), aurora.Green(url).String())
		}
	}
}