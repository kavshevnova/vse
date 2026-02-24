package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/honeynil/honey-task/maps"
	"github.com/honeynil/honey-task/pointers"
	"github.com/honeynil/honey-task/slices"
	"github.com/honeynil/honey-task/structs"
)

type Topic struct {
	Name        string
	Tasks       map[int]func()
	Count       int
	Description string
	IsFileBased bool // для тем где задачи в отдельных файлах
}

var topics = map[string]Topic{
	"slices": {
		Name:        "slices",
		Tasks:       slices.GetTasks(),
		Count:       slices.Count,
		Description: "Слайсы в go",
	},
	"pointers": {
		Name:        "pointers",
		Tasks:       pointers.GetTasks(),
		Count:       pointers.Count,
		Description: "Указатели в go",
	},
	"maps": {
		Name:        "maps",
		Tasks:       maps.GetTasks(),
		Count:       maps.Count,
		Description: "Maps в go",
	},
	"structs": {
		Name:        "structs",
		Tasks:       structs.GetTasks(),
		Count:       structs.Count,
		Description: "Структуры в go",
	},
	"interface": {
		Name:        "interface",
		Tasks:       nil,
		Count:       30,
		Description: "Интерфейсы в go (реализация паттернов и систем)",
		IsFileBased: true,
	},
	"concurrency": {
		Name:        "concurrency",
		Tasks:       nil,
		Count:       30,
		Description: "Конкурентность в go (goroutines, channels, sync)",
		IsFileBased: true,
	},
}

func main() {
	if len(os.Args) < 2 {
		printHelp()
		return
	}

	topicName := os.Args[1]

	if topicName == "list" {
		listTopics()
		return
	}

	topic, ok := topics[topicName]
	if !ok {
		fmt.Printf("Неизвестная тема: %s\n\n", topicName)
		printHelp()
		return
	}

	if len(os.Args) < 3 {
		printTopicHelp(topic)
		return
	}

	runTasks(topic, os.Args[2:])
}

func printHelp() {
	fmt.Println("Go Practice Tasks - Практические задачи по Go")
	fmt.Println("\nИспользование:")
	fmt.Println("  go run main.go <тема> <задачи>")
	fmt.Println("\nКоманды:")
	fmt.Println("  go run main.go list                    - список доступных тем")
	fmt.Println("  go run main.go <тема>                  - справка по теме")
	fmt.Println("  go run main.go <тема> <номер>          - запустить задачу")
	fmt.Println("  go run main.go <тема> <номера>         - запустить несколько задач")
	fmt.Println("  go run main.go <тема> all              - запустить все задачи")
	fmt.Println("\nПримеры:")
	fmt.Println("  go run main.go slices 1                - запустить задачу 1 по слайсам")
	fmt.Println("  go run main.go slices 1 5 10           - запустить задачи 1, 5 и 10")
	fmt.Println("  go run main.go slices all              - все задачи по слайсам")
	fmt.Println("\nДоступные темы:")
	for name, topic := range topics {
		fmt.Printf("  %-10s - %s (%d задач)\n", name, topic.Description, topic.Count)
	}
}

func listTopics() {
	fmt.Printf("Доступные темы:\n")
	for name, topic := range topics {
		fmt.Printf("  %-10s - %s (%d задач)\n", name, topic.Description, topic.Count)
	}
	fmt.Println("\nИспользуйте: go run main.go <тема> для подробностей")
}

func printTopicHelp(topic Topic) {
	fmt.Printf("%s\n\n", topic.Description)
	fmt.Printf("Доступно задач: %d\n\n", topic.Count)

	if topic.IsFileBased {
		fmt.Println("Задачи находятся в отдельных файлах.")
		fmt.Println("\nИспользование:")
		fmt.Printf("  cd %s/\n", topic.Name)
		fmt.Printf("  go run task001.go              - запустить задачу 1\n")
		fmt.Printf("  go run task002.go              - запустить задачу 2\n")
		fmt.Println("\nИли используйте:")
		fmt.Printf("  go run main.go %s <номер>      - автоматически запустит файл задачи\n", topic.Name)
		fmt.Println("\nПримеры:")
		fmt.Printf("  go run main.go %s 1\n", topic.Name)
		fmt.Printf("  go run main.go %s 5\n", topic.Name)
		fmt.Printf("\nСмотрите %s/README.md для подробного списка задач.\n", topic.Name)
	} else {
		fmt.Println("Использование:")
		fmt.Printf("  go run main.go %s <номер>          - запустить одну задачу\n", topic.Name)
		fmt.Printf("  go run main.go %s 1 5 10           - запустить несколько задач\n", topic.Name)
		fmt.Printf("  go run main.go %s all              - запустить все задачи\n", topic.Name)
		fmt.Println("\nПримеры:")
		fmt.Printf("  go run main.go %s 1\n", topic.Name)
		fmt.Printf("  go run main.go %s 1 5 10\n", topic.Name)
	}
}

func runTasks(topic Topic, args []string) {
	if topic.IsFileBased {
		runFileBasedTasks(topic, args)
		return
	}

	for _, arg := range args {
		num, err := strconv.Atoi(arg)
		if err != nil || num < 1 || num > topic.Count {
			fmt.Printf("Неверный номер задачи: %s (доступны 1-%d)\n", arg, topic.Count)
			continue
		}

		fmt.Println("\n" + strings.Repeat("=", 50))
		fmt.Printf("ЗАДАЧА %d\n", num)
		fmt.Println(strings.Repeat("=", 50) + "\n")

		if task, ok := topic.Tasks[num]; ok {
			task()
		}
	}
}

func runFileBasedTasks(topic Topic, args []string) {
	for _, arg := range args {
		num, err := strconv.Atoi(arg)
		if err != nil || num < 1 || num > topic.Count {
			fmt.Printf("Неверный номер задачи: %s (доступны 1-%d)\n", arg, topic.Count)
			continue
		}

		taskDir := fmt.Sprintf("%s/task%03d", topic.Name, num)

		fmt.Println("\n" + strings.Repeat("=", 50))
		fmt.Printf("ЗАДАЧА %d - %s/main.go\n", num, taskDir)
		fmt.Println(strings.Repeat("=", 50) + "\n")

		cmd := exec.Command("go", "run", "main.go")
		cmd.Dir = taskDir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin

		if err := cmd.Run(); err != nil {
			fmt.Printf("Ошибка при запуске задачи: %v\n", err)
		}
	}
}
