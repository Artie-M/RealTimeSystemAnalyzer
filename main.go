package main

import (
	"math"
)

type Task struct {
	Name     string `json:"name"`
	Hardness string `json:"hardness"`
	C        int    `json:"c"`
	D        int    `json:"d"`
	T        int    `json:"t"`
}

type SystemConfig struct {
	Tasks      []Task `json:"tasks"`
	NumCPU     int    `json:"num_cpu"`
	HasNetwork bool   `json:"has_network"`
}

func slack(task Task) int {
	return task.D - task.C
}

func is_feasible(task Task) bool {
	return slack(task) >= 0
}

func classify_system(tasks []Task) string {
	for _, t := range tasks {
		if t.Hardness == "жёсткое" {
			return "жёсткого реального времени"
		}
	}
	return "мягкого реального времени"
}

func critical_task(tasks []Task) (Task, bool) {
	minSlack := math.MaxInt32
	var critTask Task
	found := false

	for _, t := range tasks {
		if t.Hardness == "жёсткое" {
			s := slack(t)
			if s < minSlack {
				minSlack = s
				critTask = t
				found = true
			}
		}
	}
	return critTask, found
}

func required_reaction_time(tasks []Task) int {
	minD := math.MaxInt32
	hasHard := false

	for _, t := range tasks {
		if t.Hardness == "жёсткое" {
			if t.D < minD {
				minD = t.D
			}
			hasHard = true
		}
	}

	if !hasHard {
		for _, t := range tasks {
			if t.D < minD {
				minD = t.D
			}
		}
	}

	if minD == math.MaxInt32 {
		return 0
	}
	return minD
}

func classify_architecture(n_cpu int, has_network bool) string {
	if has_network {
		return "распределённая"
	}
	if n_cpu > 1 {
		return "многопроцессорная"
	}
	return "однопроцессорная"
}
