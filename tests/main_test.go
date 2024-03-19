package tests

import (
	"os"
	"testing"
	"time"

	"github.com/cucumber/godog"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestMain(m *testing.M) {
	time.Sleep(time.Second * 3)
	status := godog.TestSuite{
		Name:                "integration",
		ScenarioInitializer: InitializeScenario,
		Options: &godog.Options{
			Format:    "progress", // Замените на "pretty" для лучшего вывода
			Paths:     []string{"features"},
			Randomize: 0, // Последовательный порядок исполнения
		},
	}.Run()

	if st := m.Run(); st > status {
		status = st
	}
	os.Exit(status)
}
