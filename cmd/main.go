package main

import (
	"flag"
	"log"
	"time"

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

    // read command line flags
    durationPtr := flag.Int("t", -1, "measurement duration in seconds")
    formatPtr := flag.String("o", "csv", "output format [csv|sqlite]")
    directoryPtr := flag.String("d", ".", "output directory")

    flag.Parse() // ends the program if input is invalid
    log.Printf("%d %s %s", *durationPtr, *formatPtr, *directoryPtr)

    // 
    // test cpu monitor
    //

    var (
        ticker <-chan time.Time = time.NewTicker(time.Second).C
        out chan measurements.Measurement = make(chan measurements.Measurement)
    )

    go monitor.Monitor(ticker, out, 0)

    for m := range out {
        log.Println(m.Record())
    }
}
