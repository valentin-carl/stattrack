package measurements

import (
	"errors"
	"fmt"
	"log"
	"math"
	"strconv"

	"github.com/fatih/color"
	netstat "github.com/mackerelio/go-osstat/network"
)

type MeasurementType uint

const (
	CPU MeasurementType = iota // TODO does adding the type here break stuff?
	MEM
	NET
)

type MeasurementTypes []MeasurementType

func (m *MeasurementTypes) String() string {
	var res string
	for _, n := range *m {
		res += fmt.Sprintf("%d, ", n)
	}
	return res
}

func (m *MeasurementTypes) Set(value string) error {
	n, err := strconv.Atoi(value)
	if err != nil {
		log.Println("error while trying append MeasurementType value")
		return err
	}
	*m = append(*m, MeasurementType(n))
	return nil
}

// TODO delete if not used anymore
//
//	(but double check first)
func (m *MeasurementType) String() string {
	switch *m {
	case CPU:
		return "CPU"
	case MEM:
		return "memory"
	case NET:
		return "network"
	default:
		panic("no such measurement type")
	}
}

type Measurement interface {
	Record() ([]string, error)
}

func GetColumnNames(mType MeasurementType) []string {
	switch mType {
	case CPU:
		return []string{
			"timestamp",
			"user",
			"system",
			"idle",
			"nice",
			"total",
			"userp",
			"systemp",
			"idlep",
		}
	case MEM:
		return []string{
			"timestamp",
			"free",
			"total",
			"active",
			"cached",
			"inactive",
			"swapFree",
			"swapTotal",
			"swapUsed",
			"used",
			"freep",
		}
	case NET:
		return []string{
			"timestamp",
			"name",
			"RxBytes",
			"TxBytes",
		}
	default:
		log.Panicln("unknown measurement type")
	}
	return nil
}

func GetFileName(mType MeasurementType) string {
	switch mType {
	case CPU:
		return "cpu"
	case MEM:
		return "memory"
	case NET:
		return "network"
	default:
		log.Panicln("unknown measurement type")
	}
	return ""
}

type CPUMeasurement struct {
	Timestamp                       int64   // unix timestamp of measurement
	User, System, Idle, Nice, Total uint64  // raw values
	Userp, Systp, Idlep             float64 // percentage calculated with last measurement
}

func (c CPUMeasurement) Record() ([]string, error) {

	anyEq := func(iter []any, to any) bool {
		for _, x := range iter {
			if x == to {
				return true
			}
		}
		return false
	}

	if anyEq([]any{c.Userp, c.Systp, c.Idlep}, math.NaN()) {
		// TODO use color.YellowString and log instead?
		msg := "found NaN in CPU measurements"
		color.Yellow(msg)
		return nil, errors.New(msg)
	}

	res := []string{
		fmt.Sprintf("%d", c.Timestamp),
		fmt.Sprintf("%d", c.User),
		fmt.Sprintf("%d", c.System),
		fmt.Sprintf("%d", c.Idle),
		fmt.Sprintf("%d", c.Nice),
		fmt.Sprintf("%d", c.Total),
		fmt.Sprintf("%.4f", c.Userp),
		fmt.Sprintf("%.4f", c.Systp),
		fmt.Sprintf("%.4f", c.Idlep),
	}

	return res, nil
}

type MemoryMeasurement struct {
	Timestamp                                                                  int64   // unix timestamp of measurement
	Free, Total, Active, Cached, Inactive, SwapFree, SwapTotal, SwapUsed, Used uint64  // values in bytes
	Freep                                                                      float64 // freep => free/total * 100
}

func (m MemoryMeasurement) Record() ([]string, error) {
	return []string{
		fmt.Sprintf("%d", m.Timestamp),
		fmt.Sprintf("%d", m.Free),
		fmt.Sprintf("%d", m.Total),
		fmt.Sprintf("%d", m.Active),
		fmt.Sprintf("%d", m.Cached),
		fmt.Sprintf("%d", m.Inactive),
		fmt.Sprintf("%d", m.SwapFree),
		fmt.Sprintf("%d", m.SwapTotal),
		fmt.Sprintf("%d", m.SwapUsed),
		fmt.Sprintf("%d", m.Used),
		fmt.Sprintf("%f", m.Freep),
	}, nil
}

type NetworkMeasurement struct {
	Timestamp        int64
	Interface        string        // TODO create multiple NetworkMeasurement structs in `monitor.go`, one per interface
	RxBytes, TxBytes uint64        // bytes received/transmitted since the previous measurement
	Source           netstat.Stats // to calculate when stored as previous
}

func (n NetworkMeasurement) Record() ([]string, error) {
	return []string{
		fmt.Sprintf("%d", n.Timestamp),
		fmt.Sprintf("'%s'", n.Interface),
		fmt.Sprintf("%d", n.RxBytes),
		fmt.Sprintf("%d", n.TxBytes),
	}, nil
}
