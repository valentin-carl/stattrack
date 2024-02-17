package monitor

import (
	"log"
	"time"
    "math"

	cpustat "github.com/mackerelio/go-osstat/cpu"
	"github.com/valentin-carl/stattrack/pkg/measurements"
)

func Monitor(ticker <-chan time.Time, out chan<- measurements.Measurement, mT measurements.MeasurementType) error {

	var (
		err  error
		prev measurements.Measurement
	)

	for range ticker {
		switch mT {
		case 0:
			{
				log.Printf("measurementType %d\n", mT)

				curr, err := cpu(prev)
				if err != nil {
					log.Panicln("go-osstats wasn't able to retrieve cpu measurements")
				}

				out <- curr

                prev = curr
			}
		case 1:
			{
				log.Printf("measurementType %d\n", mT)

			}
		case 2:
			{
				log.Printf("measurementType %d\n", mT)

			}
		case 3:
			{
				log.Printf("measurementType %d\n", mT)

			}
		}
	}
	log.Printf("monitor %d is done\n", mT)
	return err
}

func cpu(previous measurements.Measurement) (measurements.Measurement, error) {

    if previous == nil {

        log.Println("previous measurement is nil, cannot compute relative values")

        curr, err := cpustat.Get()
        if err != nil {
            return nil, err
        }

        // return without relative values to be able to calculate them in the next iteration
        return measurements.CPUMeasurement{
            Timestamp: time.Now().UnixMilli(),
            User: curr.User,
            System: curr.System,
            Idle: curr.Idle,
            Nice: curr.Nice,
            Total: curr.Total,
            Userp: math.NaN(),
            Systp: math.NaN(),
            Idlep: math.NaN(),
        }, nil
    }

	prev, ok := previous.(measurements.CPUMeasurement)
	if !ok {
		log.Panicln("type assertion failed: tried measurement.Measurement -> measurement.CPUMeasurement")
	}

	var result measurements.CPUMeasurement

	timestamp := time.Now().UnixMilli()

	curr, err := cpustat.Get()
	if err != nil {
		log.Println("Error:", err.Error())
		return result, err
	}

	tDiff := float64(curr.Total - prev.Total)
	userp := (float64(curr.User-prev.User) / tDiff) * 100
	systp := (float64(curr.System-prev.System) / tDiff) * 100
	idlep := (float64(curr.Idle-prev.Idle) / tDiff) * 100

	result = measurements.CPUMeasurement{
		Timestamp: timestamp,
		User:      curr.User,
		System:    curr.System,
		Idle:      curr.Idle,
		Nice:      curr.Nice,
		Total:     curr.Total,
		Userp:     userp,
		Systp:     systp,
		Idlep:     idlep,
	}

	return result, nil
}

func mem(prev *any) any {
    // TODO
    return 0
}

func nettx(prev *any) any {
    // TODO
    return 0
}

func netrx(prev *any) any {
    // TODO
    return 0
}

