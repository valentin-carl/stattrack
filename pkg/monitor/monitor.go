package monitor

import (
	"context"
	"errors"
	"log"
	"math"
	"time"

	cpustat "github.com/mackerelio/go-osstat/cpu"
	memstat "github.com/mackerelio/go-osstat/memory"
	netstat "github.com/mackerelio/go-osstat/network"

	"github.com/valentin-carl/stattrack/pkg/measurements"
)

func Monitor(ctx context.Context, ticker <-chan time.Time, out chan<- measurements.Measurement, mT measurements.MeasurementType) error {

	var (
		err  error
		prev []measurements.Measurement
	)

	log.Printf("monitor of measurementType %d starting\n", mT)

	for {
		select {
		case <-ticker:
			{

				log.Printf("monitor %d: getting measurement\n", mT)

				curr, err := getMeasurements(prev, mT)
				if err != nil {
					log.Panicf("go-osstats wasn't able to retrieve os measurements of type %d\n", mT)
				}

				// send all current measurements
				// it's a slice because there could be multiple network interfaces
				for _, mm := range curr {
					log.Printf("monitor %d: sending message\n", mT)

					// FIXME see issue #2
					mm := mm
					go func() {
						// TODO problem: if the backend crashes, this creates a lot of goroutines that try to send something through the channel!!!
						out <- mm
					}()
				}

				prev = curr
			}
		case <-ctx.Done():
			{
				log.Printf("monitor %d: context was cancelled\n", mT)
				goto TheEnd
			}
		}
	}

TheEnd:
	log.Printf("monitor %d is done\n", mT)

	return err
}

func getMeasurements(previous []measurements.Measurement, mT measurements.MeasurementType) ([]measurements.Measurement, error) {

	switch mT {
	case measurements.CPU:
		{
			return cpu(previous)
		}
	case measurements.MEM:
		{
			return mem(previous)
		}
	case measurements.NET:
		{
			return net(previous)
		}
	}

	log.Panicln("invalid measurement type")
	return nil, errors.New("invalid measurement type")
}

func cpu(previous []measurements.Measurement) ([]measurements.Measurement, error) {

	if previous == nil || len(previous) == 0 {

		log.Println("no previous measurements, cannot compute relative values")

		curr, err := cpustat.Get()
		if err != nil {
			return nil, err
		}

		// return without relative values to be able to calculate them in the next iteration
		return []measurements.Measurement{measurements.CPUMeasurement{
			Timestamp: time.Now().Unix(),
			User:      curr.User,
			System:    curr.System,
			Idle:      curr.Idle,
			Nice:      curr.Nice,
			Total:     curr.Total,
			Userp:     math.NaN(),
			Systp:     math.NaN(),
			Idlep:     math.NaN(),
		}}, nil
	}

	prev_cpu := previous[0] // for CPU, the slice will one contain one element at this point

	prev, ok := prev_cpu.(measurements.CPUMeasurement)
	if !ok {
		log.Panicln("type assertion failed: tried measurement.Measurement -> measurement.CPUMeasurement")
	}

	var result measurements.CPUMeasurement

	timestamp := time.Now().Unix()

	curr, err := cpustat.Get()
	if err != nil {
		log.Println("Error:", err.Error())
		return []measurements.Measurement{result}, err
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

	return []measurements.Measurement{result}, nil
}

func mem(previous []measurements.Measurement) ([]measurements.Measurement, error) {

	// `previous` is not required to calculate memory stats

	timestamp := time.Now().Unix()

	curr, err := memstat.Get()
	if err != nil {
		log.Println("Error:", err.Error())
		var result measurements.MemoryMeasurement
		return []measurements.Measurement{result}, err
	}

	freep := float64(curr.Free) / float64(curr.Total) * 100

	return []measurements.Measurement{measurements.MemoryMeasurement{
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
	}}, nil
}

func net(previous []measurements.Measurement) ([]measurements.Measurement, error) {

	// helper
	toMap := func(mms []measurements.Measurement) map[string]measurements.NetworkMeasurement {
		res := make(map[string]measurements.NetworkMeasurement)
		for _, m := range mms {
			current, ok := m.(measurements.NetworkMeasurement)
			if !ok {
				log.Panicln("invalid measurement type")
			}
			res[current.Source.Name] = current
		}
		return res
	}

	prev := toMap(previous)

	current, err := netstat.Get()
	if err != nil {
		log.Println("something went wrong while trying to retrieve network stats")
		return []measurements.Measurement{}, err
	}

	result := make([]measurements.Measurement, len(current))

	for i, curr := range current {

		var m measurements.NetworkMeasurement

		prevm, ok := prev[curr.Name]
		if ok {
			//            log.Printf("found previous value for interface %s\n", curr.Name)
			m = measurements.NetworkMeasurement{
				Timestamp: time.Now().Unix(),
				Interface: curr.Name,
				RxBytes:   curr.RxBytes - prevm.RxBytes,
				TxBytes:   curr.TxBytes - prevm.TxBytes,
				Source:    curr,
			}
		} else {
			log.Printf("didn't find previous value for interface %s\n", m.Interface)
			// TODO check if using absolute values here creates weird data
			//  => possible fix: don't store the first iteration of network measurements
			m = measurements.NetworkMeasurement{
				Timestamp: time.Now().Unix(),
				Interface: curr.Name,
				RxBytes:   curr.RxBytes,
				TxBytes:   curr.TxBytes,
				Source:    curr,
			}
		}

		result[i] = m
	}

	return result, nil
}
