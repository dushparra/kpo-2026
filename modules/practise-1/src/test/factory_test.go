package test

import (
	"os/exec"
	"strings"
	"testing"
)

func TestFactorySellsCarsInQueueOrderAndLiquidatesRemainingStock(t *testing.T) {
	command := exec.Command("go", "run", ".")
	command.Dir = "../main"

	outputBytes, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("не удалось запустить программу: %v\n%s", err, outputBytes)
	}

	output := string(outputBytes)
	expectedFragments := []string{
		"VIN: 1, размер педалей: 1",
		"VIN: 2, размер педалей: 2",
		"VIN: 3, размер педалей: 3",
		"VIN: 4, размер педалей: 4",
		"Покупатель: Вася, VIN: 1",
		"Покупатель: Вова, VIN: 2",
		"Покупатель: Света, VIN: 3",
		"На складе нет автомобилей",
	}

	for _, fragment := range expectedFragments {
		if !strings.Contains(output, fragment) {
			t.Errorf("в выводе нет строки %q\nполный вывод:\n%s", fragment, output)
		}
	}

	afterSaleParts := strings.Split(output, "Автомобили после продажи")
	if len(afterSaleParts) != 2 {
		t.Fatalf("не найден раздел с результатом продажи")
	}

	afterSale := afterSaleParts[1]
	if strings.Contains(afterSale, "VIN: 4") {
		t.Error("непроданный автомобиль VIN 4 должен быть ликвидирован")
	}
}
