.PHONY: all gui console clean test help

BINARY_GUI=lintvision_gui
BINARY_CONSOLE=lintvision_console
GO_FILES_GUI=gui.go main_gui.go
GO_FILES_CONSOLE=main_console.go

all: gui

gui:
	@echo "Сборка GUI версии..."
	go build -o $(BINARY_GUI) $(GO_FILES_GUI)
	@echo "GUI версия собрана: $(BINARY_GUI)"


console:
	@echo "Сборка консольной версии..."
	go build -o $(BINARY_CONSOLE) $(GO_FILES_CONSOLE)
	@echo "Консольная версия собрана: $(BINARY_CONSOLE)"


run-gui: gui
	@echo "Запуск GUI версии..."
	./$(BINARY_GUI)

run-console: console
	@echo "Запуск консольной версии..."
	./$(BINARY_CONSOLE) -path . -out test_results.json


test:
	@echo "Запуск тестов..."
	go test ./...

clean:
	@echo "Очистка..."
	rm -f $(BINARY_GUI) $(BINARY_CONSOLE) test_results.json
	@echo "Очистка завершена"

deps:
	@echo "Обновление зависимостей..."
	go mod tidy
	go get -u fyne.io/fyne/v2@latest
	@echo "Зависимости обновлены"


check-deps:
	@echo "Проверка зависимостей..."
	go mod verify
	@echo "Зависимости проверены"

help:
	@echo "Доступные команды:"
	@echo "  all          - собрать все версии"
	@echo "  gui          - собрать только GUI версию"
	@echo "  console      - собрать только консольную версию"
	@echo "  run-gui      - собрать и запустить GUI версию"
	@echo "  run-console  - собрать и запустить консольную версию"
	@echo "  test         - запустить тесты"
	@echo "  clean        - очистить собранные файлы"
	@echo "  deps         - обновить зависимости"
	@echo "  check-deps   - проверить зависимости"
	@echo "  help         - показать эту справку"
