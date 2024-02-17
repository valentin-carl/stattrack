package measurements

import (
    "fmt"
)

type MeasurementType uint

const (
    CPU = iota
    MEM
    NET_RX
    NET_TX
)

type Measurement interface {
    Record() []string
    // TODO add RecordRaw() with correct types to store in DB once SQLite support is added
}

type CPUMeasurement struct {
    Timestamp                       int64   // unix timestamp of measurement
	User, System, Idle, Nice, Total uint64  // raw values
	Userp, Systp, Idlep             float64 // percentage calculated with last measurement
}

func (c CPUMeasurement) Record() []string {
    return []string{
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
}

type MemoryMeasurement struct {
	Timestamp                                                                  int64   // unix timestamp of measurement
	Free, Total, Active, Cached, Inactive, SwapFree, SwapTotal, SwapUsed, Used uint64  // values in bytes
	Freep                                                                      float64 // freep => free/total * 100
}

func (m *MemoryMeasurement) Record() string {
    // TODO
    return ""
}

type NetworkMeasurement struct {
	Timestamp    int64
	Measurements map[string]struct {
		RxBytes, TxBytes uint64 // bytes received/transmitted since the previous measurement
	}
}

func (n *NetworkMeasurement) Record() string {
    // TODO
    return ""
}
