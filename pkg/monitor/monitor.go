package monitor

import (
	"errors"
	"log"
	"math"
	"time"

	cpustat "github.com/mackerelio/go-osstat/cpu"
	memstat "github.com/mackerelio/go-osstat/memory"
	"github.com/valentin-carl/stattrack/pkg/measurements"
)

func Monitor(ticker <-chan time.Time, out chan<- measurements.Measurement, mT measurements.MeasurementType) error {

	var (
		err  error
		prev measurements.Measurement
	)

	for range ticker {

		log.Printf("measurementType %d\n", mT)

		curr, err := getMeasurement(prev, mT)
		if err != nil {
			log.Panicln("go-osstats wasn't able to retrieve cpu measurements")
		}

		out <- curr
		prev = curr
	}

	log.Printf("monitor %d is done\n", mT)
	return err
}

func getMeasurement(previous measurements.Measurement, mT measurements.MeasurementType) (measurements.Measurement, error) {

	// TODO add network types

	switch mT {
	case measurements.CPU:
		{
			return cpu(previous)
		}
	case measurements.MEM:
		{
			return mem(previous)
		}
	}

	log.Panicln("invalid measurement type")
	return nil, errors.New("invalid measurement type")
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
			Timestamp: time.Now().Unix(),
			User:      curr.User,
			System:    curr.System,
			Idle:      curr.Idle,
			Nice:      curr.Nice,
			Total:     curr.Total,
			Userp:     math.NaN(),
			Systp:     math.NaN(),
			Idlep:     math.NaN(),
		}, nil
	}

	prev, ok := previous.(measurements.CPUMeasurement)
	if !ok {
		log.Panicln("type assertion failed: tried measurement.Measurement -> measurement.CPUMeasurement")
	}

	var result measurements.CPUMeasurement

	timestamp := time.Now().Unix()

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

func mem(previous measurements.Measurement) (measurements.Measurement, error) {

    // `previous` is not required to calculate memory stats

	timestamp := time.Now().Unix()

	curr, err := memstat.Get()
	if err != nil {
		log.Println("Error:", err.Error())
        var result measurements.Measurement
		return result, err
	}

    freep := float64(curr.Free)/float64(curr.Total) * 100

	return measurements.MemoryMeasurement{
		Timestamp: timestamp,
		Free:      curr.Free,
		Total:     curr.Total,
		Active:    curr.Active,
		Cached:    curr.Cached,
		Inactive:  curr.Inactive,
		SwapFree:  curr.SwapFree,
		SwapUsed:  curr.SwapUsed,
		SwapTotal: curr.SwapTotal,
		Used:      curr.Used,
		Freep:     freep,
	}, nil
}

func nettx(prev *any) any {
	// TODO
	return 0
}

func netrx(prev *any) any {
	// TODO
	return 0
}
