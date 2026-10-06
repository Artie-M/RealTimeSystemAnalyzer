package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"text/tabwriter"
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

func deadline_type(task Task) string {
	if task.D == task.T {
		return "неявный"
	}
	if task.D < task.T {
		return "ограниченный"
	}
	return "произвольный"
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

func utilization(tasks []Task) float64 {
	u := 0.0
	for _, t := range tasks {
		u += float64(t.C) / float64(t.T)
	}
	return u
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

func main() {
	fileData, err := os.ReadFile("system.json")
	if err != nil {
		fmt.Println("Ошибка чтения файла:", err)
		return
	}

	var sys SystemConfig
	err = json.Unmarshal(fileData, &sys)
	if err != nil {
		fmt.Println("Ошибка парсинга JSON:", err)
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)

	// Разделители колонок — символы '\t'
	fmt.Fprintln(w, "Задача\tКласс\tC (мс)\tD (мс)\tT (мс)\tL (мс)\tВыполнимо\tТип дедлайна")
	fmt.Fprintln(w, "------\t-----\t------\t------\t------\t------\t---------\t------------")

	allFeasible := true
	for _, t := range sys.Tasks {
		s := slack(t)
		f := is_feasible(t)
		fStr := "Да"
		if !f {
			fStr = "Нет"
			allFeasible = false
		}
		dT := deadline_type(t)

		fmt.Fprintf(w, "%s\t%s\t%d\t%d\t%d\t%d\t%s\t%s\n",
			t.Name, t.Hardness, t.C, t.D, t.T, s, fStr, dT)
	}

	w.Flush()

	fmt.Println("\n--- ЗАКЛЮЧЕНИЕ ---")

	sysClass := classify_system(sys.Tasks)
	fmt.Printf("Класс системы: %s\n", sysClass)

	crit, hasCrit := critical_task(sys.Tasks)
	if hasCrit {
		fmt.Printf("Критическая задача: %s (L = %d мс)\n", crit.Name, slack(crit))
	} else {
		fmt.Println("Критическая задача: нет (отсутствуют жёсткие задачи)")
	}

	reqReaction := required_reaction_time(sys.Tasks)
	fmt.Printf("Требуемое время реакции R_треб: %d мс\n", reqReaction)

	util := utilization(sys.Tasks)
	uCheck := "≤ 1 (выполняется)"
	if util > 1 {
		uCheck = "> 1 (не выполняется)"
	}
	fmt.Printf("Коэффициент загрузки U: %.2f %s\n", util, uCheck)

	feasStr := "все дедлайны выполнимы"
	if !allFeasible {
		feasStr = "не все дедлайны выполнимы"
	}
	fmt.Printf("Общая выполнимость: %s\n", feasStr)

	arch := classify_architecture(sys.NumCPU, sys.HasNetwork)
	fmt.Printf("Архитектура: %s\n", arch)
}
