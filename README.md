# StatTrack

**StatTrack** is a small program for recording a computer's CPU utilization, memory usage, and the amount of incoming and outgoing traffic.
Measurements are made once per second, and data can be stored either in csv-format or in a sqlite database.

## Installation

To install StatTrack, you need `make` and `go`. 
Next, clone this repository and run 

```shell
make build
```

in the repository's root.

## Usage

StatTrack has four options that can be set by the user.
- `-d`: sets the output directory in which the data will be stored. StatTrack will create the directory if it doesn't exist.
- `-m`: sets which statistics to track. The following options are available.
    - `0`: CPU utilization
    - `1`: memory usage
    - `2`: bytes transmitted and received
  
    It is possible to set multiple values by repeating the flag with different values, i.e., `-d 0 -d 1 -d 2`.
- `-o`: sets the output type. The available are `csv` and `sqlite`.
- `-t`: sets the duration in seconds.

## Extending StatTrack 

New statistics can be added by creating a new `MeasurementType` in `pkg/measurements/measurement.go` and adjust the code where there is a switch on the `MeasurementType`.
