package persistence

import (
	"context"
	"encoding/csv"
	_ "encoding/csv"
	"io/fs"
	"log"
	"os"
	"path"
	"strings"

	"github.com/valentin-carl/stattrack/pkg/measurements"
)

type CSVBackend struct {
	ctx context.Context
    values <-chan measurements.Measurement
    mType measurements.MeasurementType
    writer csv.Writer
}

func NewCSVBackend(
    ctx context.Context, 
    values <-chan measurements.Measurement,
    outdir string, 
    mType measurements.MeasurementType,
) (*CSVBackend, error) {

	log.Println("creating new CSV backend")

    c := &CSVBackend{
        ctx: ctx,
        values: values,
        mType: mType,
    }

    err := os.MkdirAll(outdir, fs.ModePerm)
    if err != nil {
        log.Println("error occurred while trying to create output directory", err.Error())
        return nil, err
    }

    fpath := path.Join(outdir, measurements.GetFileName(mType))
    file, err := os.Create(fpath)
    if err != nil {
        log.Println("error occurred while trying to create output file")
        return nil, err
    }

    s, err := file.Stat()
    if err != nil {
        log.Println("error occurred while trying to create output file")
        return nil, err
    } else {
        log.Printf("output file %s created with mod %s\n", s.Name(), s.Mode().String())
    }

    c.writer = *csv.NewWriter(file)

    return c, nil
}

func (c *CSVBackend) Start() error {

    log.Printf("csv backend for %d starting\n", measurements.MeasurementType(c.mType))

    defer c.writer.Flush()

    var err error

    // write csv title 
    err = c.writer.Write(measurements.GetColumnNames(c.mType))
    if err != nil {
        log.Println("error while trying to write column names for type", c.mType)
        return err
    }
    
    // read + store values 
	for {
		select {
        case value := <-c.values:
			{
                toStr := func(m measurements.Measurement) string {
                    return strings.Join(m.Record(), ",")
                }
                log.Println("CSV backend: received value ", toStr(value))
                c.writer.Write(value.Record())
			}
		case <-c.ctx.Done():
			{
				log.Println("CSV backend: context cancelled, quitting ...")
                // TODO write/close files and so on ...
				goto TheEnd
			}
		}
	}

TheEnd:
	log.Println("CSV backend done")

    return err
}
