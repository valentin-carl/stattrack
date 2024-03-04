package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path"
	"sync"
	"time"

	"github.com/VividCortex/multitick"
	"github.com/fatih/color"
	"github.com/google/uuid"
	"github.com/valentin-carl/stattrack/pkg/measurements"
	"github.com/valentin-carl/stattrack/pkg/monitor"
	"github.com/valentin-carl/stattrack/pkg/persistence"
)

func main() {

	log.Println("stattrack started")

	// read command line flags
	var types measurements.MeasurementTypes
	flag.Var(&types, "m", "measurement type [0=cpu|1=mem|2=net]. Can occur multiple times for measuring different stats simultaneously.")

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

	/* create the backends */
	var err error

	backends := make(map[measurements.MeasurementType]persistence.Backend)
	channels := make(map[measurements.MeasurementType]chan measurements.Measurement)
	outdir := fmt.Sprintf("%s-%s", "./output", uuid.New().String())
	outdir = path.Join(*directoryPtr, outdir)

	log.Println(color.GreenString(outdir))

	// create the backends
	switch *formatPtr {
	case "csv":
		{
			for _, mType := range types {

				mType := measurements.MeasurementType(mType)

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
			}
		}
	default:
		{
			log.Panicf("didn't get valid output format %s\n", *formatPtr)
		}
	}

	// wait group for both, monitors and backends
	var wg sync.WaitGroup

	/* start the backends */

	for i := range types {
		i := i
		go func() {
			log.Println("starting backend for type", types[i], i)
			wg.Add(1)
			backends[measurements.MeasurementType(types[i])].Start()
			wg.Done()
			log.Println("goroutine for backend for type", i, "is done")
		}()
	}

	/* start the monitors */

	var ticker = multitick.NewTicker(time.Second, 0)

	for i := range types {
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
	wg.Wait() // waits until all monitors & backends are done

	// program over :-)
	log.Println("thank you for recording your os stats with deutsche bahn")
}
