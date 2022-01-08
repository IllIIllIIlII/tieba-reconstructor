package config

type Status int

const (
	Paused Status = iota
	Running
	Finished
	Failed
)
