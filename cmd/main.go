package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"time"

	"github.com/VividCortex/multitick"
	"github.com/valentin-carl/stattrack/pkg/measurements"
	"github.com/valentin-carl/stattrack/pkg/monitor"
)

/// usage:
/// `stattrack -t=60 -o=csv -d="/Users/valentincarl/Code/Go/stattrack/data"`
/// - t: duration in seconds. Stattrack will run indefinitely if left empty but can be stopped with SIGINT
/// - o: output format. Available options: [csv|sqlite]
/// - d: output directory. stattrack will create a new subdirectory of that in which the measurement are stored. Defaults to ".". Can be absolute or relative.
func main() {

	log.Println("stattrack started")

	// TODO add support for choosing which measurements should be done (default all)
	// TODO add support for setting the frequency

	// read command line flags
	durationPtr := flag.Int("t", -1, "measurement duration in seconds")
	formatPtr := flag.String("o", "csv", "output format [csv|sqlite]")
	directoryPtr := flag.String("d", ".", "output directory")

	flag.Parse() // ends the program if input is invalid
	log.Printf("%d %s %s", *durationPtr, *formatPtr, *directoryPtr)

	// these tell the main goroutine when it's time to stop
	timer := time.NewTimer(time.Duration(*durationPtr) * time.Second)
	interrupt := make(chan os.Signal)
	signal.Notify(interrupt, os.Interrupt)

	// this tells the monitors when it's time to stop
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	// create + start the monitors
	var (
		ticker                               = multitick.NewTicker(time.Second, 0)
		out    chan measurements.Measurement = make(chan measurements.Measurement)
	)

	defer close(out)

	var wg sync.WaitGroup

	for i := 0; i < 3; i++ {
		i := i
		go func() {
			wg.Add(1)
			monitor.Monitor(ctx, ticker.Subscribe(), out, measurements.MeasurementType(i))
			log.Printf("monitor %d: calling `wg.Done()`\n", i)
			wg.Done()
		}()
	}

	// wait for timer/interrupt
	// and cancel the context
	for {
		select {
		case <-timer.C:
			{
				log.Println("timer over, quitting ...")
				goto TheFinishLine
			}
		case <-interrupt:
			{
				log.Println("main goroutine interrupted, quitting ...")
				goto TheFinishLine
			}
		case msg := <-out:
			{
				log.Printf("received measurement: %s\n", strings.Join(msg.Record(), " | "))
			}
		}
	}

TheFinishLine:
	log.Println("stopping monitors ...")
	cancel()
	wg.Wait()

	// program over :-)
	log.Println("thank you for recording your os stats with deutsche bahn")
}
