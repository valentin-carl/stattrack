package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/VividCortex/multitick"
	"github.com/google/uuid"
	"github.com/valentin-carl/stattrack/pkg/measurements"
	"github.com/valentin-carl/stattrack/pkg/monitor"
	"github.com/valentin-carl/stattrack/pkg/persistence"
)

// usage:
// `stattrack -t=60 -o=csv -d="/Users/valentincarl/Code/Go/stattrack/data"`
// - t: duration in seconds. Stattrack will run indefinitely if left empty but can be stopped with SIGINT
// - o: output format. Available options: [csv|sqlite]
// - d: output directory. stattrack will create a new subdirectory of that in which the measurement are stored. Defaults to ".". Can be absolute or relative.
func main() {

	// TODO add support for setting the frequency

	log.Println("stattrack started")

	// read command line flags
    var types measurements.MeasurementTypes
    flag.Var(&types, "m", "measurement type [cpu|mem|net]. Can occur multiple times for measuring different stats simultaneously.")

	durationPtr := flag.Int("t", -1, "measurement duration in seconds")
	formatPtr := flag.String("o", "csv", "output format [csv|sqlite]")
	directoryPtr := flag.String("d", ".", "output directory")

	flag.Parse() // ends the program if input is invalid

	log.Printf("%d %s %s", *durationPtr, *formatPtr, *directoryPtr)
    log.Println(types)

	// these tell the main goroutine when it's time to stop
	timer := time.NewTimer(time.Duration(*durationPtr) * time.Second)
	interrupt := make(chan os.Signal)
	signal.Notify(interrupt, os.Interrupt)

	// this tells the monitors when it's time to stop
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	// create + start the backend
	var err error

    // TODO read these from CLI inputs
    /*
    var types []measurements.MeasurementType = []measurements.MeasurementType{
        measurements.CPU,
        measurements.MEM,
        measurements.NET,
    }
    */

	backends := make(map[measurements.MeasurementType]persistence.Backend)
    channels := make(map[measurements.MeasurementType]chan measurements.Measurement)
    outdir := fmt.Sprintf("%s-%s", "./output", uuid.New().String())

    // create + start backends
    // TODO start multiple!!
	switch *formatPtr {
	case "csv":
		{
            for _, mType := range types {

                log.Println("MEASUREMENT TYPE", mType)
                
                // channel through which monitor and backend communicate
                channels[mType] = make(chan measurements.Measurement)

                // backend
                backends[mType], err = persistence.NewCSVBackend(
                    ctx,
                    channels[mType],
                    outdir,
                    mType,
                )
                if err != nil {
                    log.Panicln("cannot create CSV backend for measurement type", mType)
                }

                log.Printf("main goroutine is starting backend %d\n", mType)
                go backends[mType].Start()
            }
		}
	case "sqlite":
		{
            for _, mType := range types {

                log.Println("MEASUREMENT TYPE", mType)
                
                // channel through which monitor and backend communicate
                channels[mType] = make(chan measurements.Measurement)

                // backend
                backends[mType], err = persistence.NewSqliteBackend(
                    ctx,
                    channels[mType],
                    outdir,
                    mType,
                    "data.db",
                )
                if err != nil {
                    log.Panicln("cannot create CSV backend for measurement type", mType)
                }

                log.Printf("main goroutine is starting backend %d\n", mType)
                go backends[mType].Start()
            }
		}
	default:
		{
			log.Panicf("didn't get valid output format %s\n", *formatPtr)
		}
	}

	// start the monitors

	var ticker = multitick.NewTicker(time.Second, 0)
	var wg sync.WaitGroup

	for i := 0; i < len(types); i++ {
		i := i
		go func() {
            log.Println("starting monitor for type", types[i])
			wg.Add(1)
            mType := measurements.MeasurementType(types[i])
			monitor.Monitor(ctx, ticker.Subscribe(), channels[mType], mType)
			log.Printf("monitor %d: calling `wg.Done()`\n", i)
			wg.Done()
		}()
	}

	// wait for timer/interrupt
	// and cancel the context
    log.Println("main goroutine waiting for interrupt or timer to end")
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
		}
	}

TheFinishLine:
	log.Println("stopping monitors ...")
	cancel()
	wg.Wait()

    // FIXME why does the wait group thing not wait until all files are written?
    log.Println("writing measurements ...")
    time.Sleep(5 * time.Second)

	// program over :-)
	log.Println("thank you for recording your os stats with deutsche bahn")
}
