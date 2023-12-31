// Package unit provides valid AWS CloudWatch EMF metric units.
package unit

// Unit represents an EMF metric unit.
type Unit string

const (
	None               Unit = "None"
	Seconds            Unit = "Seconds"
	Microseconds       Unit = "Microseconds"
	Milliseconds       Unit = "Milliseconds"
	Bytes              Unit = "Bytes"
	Kilobytes          Unit = "Kilobytes"
	Megabytes          Unit = "Megabytes"
	Gigabytes          Unit = "Gigabytes"
	Terabytes          Unit = "Terabytes"
	Bits               Unit = "Bits"
	Kilobits           Unit = "Kilobits"
	Megabits           Unit = "Megabits"
	Gigabits           Unit = "Gigabits"
	Terabits           Unit = "Terabits"
	Percent            Unit = "Percent"
	Count              Unit = "Count"
	BytesPerSecond     Unit = "Bytes/Second"
	KilobytesPerSecond Unit = "Kilobytes/Second"
	MegabytesPerSecond Unit = "Megabytes/Second"
	GigabytesPerSecond Unit = "Gigabytes/Second"
	TerabytesPerSecond Unit = "Terabytes/Second"
	BitsPerSecond      Unit = "Bits/Second"
	KilobitsPerSecond  Unit = "Kilobits/Second"
	MegabitsPerSecond  Unit = "Megabits/Second"
	GigabitsPerSecond  Unit = "Gigabits/Second"
	TerabitsPerSecond  Unit = "Terabits/Second"
	CountPerSecond     Unit = "Count/Second"
)
