package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/rfxxfy/LintVision/logging"
	"github.com/rfxxfy/LintVision/parseurl"
	"github.com/rfxxfy/LintVision/stats"
)

type LintVisionGUI struct {
	app            fyne.App
	mainWindow     fyne.Window
	pathEntry      *widget.Entry
	urlEntry       *widget.Entry
	outputEntry    *widget.Entry
	logConfigEntry *widget.Entry
	progressBar    *widget.ProgressBar
	statusLabel    *widget.Label
	resultText     *widget.Entry
	isAnalyzing    bool
	cancelFunc     context.CancelFunc
}

func NewLintVisionGUI() *LintVisionGUI {
	gui := &LintVisionGUI{
		app: app.New(),
	}

	gui.mainWindow = gui.app.NewWindow("LintVision - Анализ кода")
	gui.mainWindow.Resize(fyne.NewSize(800, 600))

	gui.setupUI()
	return gui
}

func (g *LintVisionGUI) setupUI() {
	g.pathEntry = widget.NewEntry()
	g.pathEntry.SetPlaceHolder("Путь к директории для анализа (например: . или ~/projects)")
	g.pathEntry.SetText(".")

	g.urlEntry = widget.NewEntry()
	g.urlEntry.SetPlaceHolder("GitHub URL репозитория (например: https://github.com/user/repo)")

	g.outputEntry = widget.NewEntry()
	g.outputEntry.SetPlaceHolder("Путь к файлу для сохранения результата (опционально)")

	g.logConfigEntry = widget.NewEntry()
	g.logConfigEntry.SetPlaceHolder("Путь к конфигу логгера (опционально)")

	g.progressBar = widget.NewProgressBar()
	g.progressBar.Hide()

	g.statusLabel = widget.NewLabel("Готов к анализу")
	g.statusLabel.Alignment = fyne.TextAlignCenter

	g.resultText = widget.NewMultiLineEntry()
	g.resultText.SetPlaceHolder("Результаты анализа появятся здесь...")
	g.resultText.Disable()

	selectPathBtn := widget.NewButton("Выбрать директорию", g.selectDirectory)
	analyzeGitHubBtn := widget.NewButton("Анализ GitHub", g.runGitHubAnalysis)
	selectOutputBtn := widget.NewButton("Выбрать файл вывода", g.selectOutputFile)
	selectLogConfigBtn := widget.NewButton("Выбрать конфиг логгера", g.selectLogConfig)
	analyzeBtn := widget.NewButton("Запустить анализ", g.runAnalysis)
	cancelBtn := widget.NewButton("Отменить", g.cancelAnalysis)

	pathContainer := container.NewBorder(nil, nil, widget.NewLabel("Директория:"), selectPathBtn, g.pathEntry)
	urlContainer := container.NewBorder(nil, nil, widget.NewLabel("GitHub URL:"), analyzeGitHubBtn, g.urlEntry)
	outputContainer := container.NewBorder(nil, nil, widget.NewLabel("Файл вывода:"), selectOutputBtn, g.outputEntry)
	logConfigContainer := container.NewBorder(nil, nil, widget.NewLabel("Конфиг логгера:"), selectLogConfigBtn, g.logConfigEntry)

	controlsContainer := container.NewVBox(
		pathContainer,
		urlContainer,
		outputContainer,
		logConfigContainer,
		container.NewHBox(analyzeBtn, cancelBtn),
		g.progressBar,
		g.statusLabel,
	)

	content := container.NewBorder(
		controlsContainer,
		nil,
		nil,
		nil,
		container.NewBorder(
			widget.NewLabel("Результаты анализа:"),
			nil,
			nil,
			nil,
			g.resultText,
		),
	)

	g.mainWindow.SetContent(content)
}

func (g *LintVisionGUI) selectDirectory() {
	dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
		if err != nil {
			dialog.ShowError(err, g.mainWindow)
			return
		}
		if uri != nil {
			g.pathEntry.SetText(uri.Path())
		}
	}, g.mainWindow)
}

func (g *LintVisionGUI) selectOutputFile() {
	dialog.ShowFileSave(func(writer fyne.URIWriteCloser, err error) {
		if err != nil {
			dialog.ShowError(err, g.mainWindow)
			return
		}
		if writer != nil {
			g.outputEntry.SetText(writer.URI().Path())
			writer.Close()
		}
	}, g.mainWindow)
}

func (g *LintVisionGUI) selectLogConfig() {
	dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil {
			dialog.ShowError(err, g.mainWindow)
			return
		}
		if reader != nil {
			g.logConfigEntry.SetText(reader.URI().Path())
			reader.Close()
		}
	}, g.mainWindow)
}

func (g *LintVisionGUI) runAnalysis() {
	if g.isAnalyzing {
		dialog.ShowError(fmt.Errorf("Анализ уже выполняется. Дождитесь завершения."), g.mainWindow)
		return
	}

	g.resultText.SetText("")
	g.statusLabel.SetText("Подготовка к анализу...")

	path := g.pathEntry.Text
	output := g.outputEntry.Text
	logConfig := g.logConfigEntry.Text

	if path == "" {
		dialog.ShowError(fmt.Errorf("Укажите директорию для анализа"), g.mainWindow)
		return
	}

	expandedPath, err := g.expandPath(path)
	if err != nil {
		dialog.ShowError(fmt.Errorf("Ошибка в пути: %v", err), g.mainWindow)
		return
	}

	g.isAnalyzing = true
	g.progressBar.Show()
	g.progressBar.SetValue(0.1)
	g.statusLabel.SetText("Загрузка конфигурации логгера...")

	if logConfig != "" {
		if err := logging.LoadConfig(logConfig); err != nil {
			dialog.ShowError(fmt.Errorf("ошибка загрузки конфига логгера: %v", err), g.mainWindow)
			g.progressBar.Hide()
			g.statusLabel.SetText("Ошибка загрузки конфига")
			return
		}
	}

	g.progressBar.SetValue(0.3)
	g.statusLabel.SetText("Запуск анализа...")

	ctx, cancel := context.WithCancel(context.Background())
	g.cancelFunc = cancel

	go func() {
		defer func() {
			g.isAnalyzing = false
			g.cancelFunc = nil
		}()

		g.progressBar.SetValue(0.5)
		g.statusLabel.SetText("Анализируем код...")

		select {
		case <-ctx.Done():
			return
		default:
		}

		result, err := stats.AnalyzeAndSave(expandedPath, output)
		if err != nil {
			g.progressBar.Hide()
			g.statusLabel.SetText("Ошибка анализа")
			dialog.ShowError(fmt.Errorf("Анализ не удался: %v", err), g.mainWindow)
			return
		}

		select {
		case <-ctx.Done():
			return
		default:
		}

		g.progressBar.SetValue(1.0)
		g.statusLabel.SetText("Анализ завершен успешно!")

		resultText := g.formatResults(result, output)
		g.resultText.SetText(resultText)

		g.progressBar.Hide()
	}()
}

func (g *LintVisionGUI) cancelAnalysis() {
	if g.isAnalyzing && g.cancelFunc != nil {
		g.cancelFunc()
		g.isAnalyzing = false
		g.progressBar.Hide()
		g.statusLabel.SetText("Анализ отменен")
	}
}

func (g *LintVisionGUI) expandPath(path string) (string, error) {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		path = filepath.Join(home, path[2:])
	}

	path = os.ExpandEnv(path)
	return path, nil
}

func (g *LintVisionGUI) formatResults(stats stats.ProjectStats, outputPath string) string {
	var result strings.Builder

	result.WriteString("Анализ завершен успешно!\n\n")

	if outputPath != "" {
		result.WriteString(fmt.Sprintf("Результаты сохранены в: %s\n\n", outputPath))
	}

	result.WriteString("=== ОБЩАЯ СТАТИСТИКА ===\n")
	result.WriteString(fmt.Sprintf("Всего файлов: %d\n", len(stats.Files)))
	result.WriteString(fmt.Sprintf("Скрытых файлов: %d\n", stats.HiddenFiles))
	result.WriteString(fmt.Sprintf("Скрытых директорий: %d\n", stats.HiddenDirs))
	result.WriteString(fmt.Sprintf("Нескрытых директорий: %d\n\n", stats.NonHiddenDirs))

	result.WriteString("=== СТАТИСТИКА ПО КАТЕГОРИЯМ ===\n")
	for category, count := range stats.CategoryCounts {
		result.WriteString(fmt.Sprintf("%s: %d файлов\n", category, count))
	}
	result.WriteString("\n")

	if len(stats.Files) > 0 {
		result.WriteString("=== ДЕТАЛЬНАЯ СТАТИСТИКА ===\n")
		for _, file := range stats.Files {
			result.WriteString(fmt.Sprintf("📁 %s\n", filepath.Base(file.Path)))
			result.WriteString(fmt.Sprintf("   Тип: %s (%s)\n", file.Category, file.Ext))
			result.WriteString(fmt.Sprintf("   Строк: %d (код: %d, комментарии: %d, пустые: %d)\n",
				file.LinesTotal, file.LinesCode, file.LinesComments, file.LinesBlank))
			result.WriteString("\n")
		}
	}

	return result.String()
}

func (g *LintVisionGUI) runGitHubAnalysis() {
	if g.isAnalyzing {
		dialog.ShowError(fmt.Errorf("Анализ уже выполняется. Дождитесь завершения."), g.mainWindow)
		return
	}

	g.resultText.SetText("")
	g.statusLabel.SetText("Валидация GitHub URL...")

	url := g.urlEntry.Text
	output := g.outputEntry.Text
	logConfig := g.logConfigEntry.Text

	if url == "" {
		dialog.ShowError(fmt.Errorf("укажите GitHub URL репозитория"), g.mainWindow)
		return
	}

	validationResult := parseurl.ValidateGitHubURL(url)
	if !validationResult.IsValid {
		var errorMsg strings.Builder
		errorMsg.WriteString(fmt.Sprintf("❌ %s\n\n", validationResult.Error))
		errorMsg.WriteString(fmt.Sprintf("URL: %s\n\n", url))

		if len(validationResult.Suggestions) > 0 {
			errorMsg.WriteString("💡 Предложения по исправлению:\n")
			for i, suggestion := range validationResult.Suggestions {
				errorMsg.WriteString(fmt.Sprintf("%d. %s\n", i+1, suggestion))
			}
		}

		dialog.ShowError(fmt.Errorf("%s", errorMsg.String()), g.mainWindow)
		return
	}

	g.statusLabel.SetText("GitHub URL валиден. Подготовка к анализу...")

	g.isAnalyzing = true
	g.progressBar.Show()
	g.progressBar.SetValue(0.1)
	g.statusLabel.SetText("Загрузка конфигурации логгера...")

	if logConfig != "" {
		if err := logging.LoadConfig(logConfig); err != nil {
			dialog.ShowError(fmt.Errorf("ошибка загрузки конфига логгера: %v", err), g.mainWindow)
			g.progressBar.Hide()
			g.statusLabel.SetText("Ошибка загрузки конфига")
			return
		}
	}

	g.progressBar.SetValue(0.2)
	g.statusLabel.SetText("Клонирование репозитория...")

	ctx, cancel := context.WithCancel(context.Background())
	g.cancelFunc = cancel

	go func() {
		defer func() {
			g.isAnalyzing = false
			g.cancelFunc = nil
		}()

		g.progressBar.SetValue(0.4)
		g.statusLabel.SetText("Анализируем GitHub репозитория...")

		select {
		case <-ctx.Done():
			return
		default:
		}

		result, err := parseurl.AnalyzeRepoFromURL(url)
		if err != nil {
			g.progressBar.Hide()
			g.statusLabel.SetText("Ошибка анализа GitHub репозитория")

			errorMsg := err.Error()
			if strings.Contains(errorMsg, "репозиторий не найден") {
				dialog.ShowError(fmt.Errorf("❌ Репозиторий не найден!\n\nURL: %s\n\nВозможные причины:\n• Репозиторий не существует\n• Опечатка в названии\n• Репозиторий был удален", url), g.mainWindow)
			} else if strings.Contains(errorMsg, "закрытый или требует аутентификации") {
				dialog.ShowError(fmt.Errorf("🔒 Репозиторий закрытый!\n\nURL: %s\n\nВозможные причины:\n• Приватный репозиторий\n• Требуется авторизация\n• Нет доступа", url), g.mainWindow)
			} else if strings.Contains(errorMsg, "нет доступа") {
				dialog.ShowError(fmt.Errorf("🚫 Нет доступа к репозиторию!\n\nURL: %s\n\nВозможные причины:\n• Репозиторий приватный\n• Требуются права доступа\n• Репозиторий заблокирован", url), g.mainWindow)
			} else if strings.Contains(errorMsg, "timed out") {
				dialog.ShowError(fmt.Errorf("⏰ Превышено время ожидания!\n\nURL: %s\n\nВозможные причины:\n• Медленное интернет-соединение\n• GitHub недоступен\n• Репозиторий слишком большой", url), g.mainWindow)
			} else {
				dialog.ShowError(fmt.Errorf("❌ Ошибка анализа GitHub репозитория!\n\nURL: %s\n\nОшибка: %s", url, errorMsg), g.mainWindow)
			}
			return
		}

		select {
		case <-ctx.Done():
			return
		default:
		}

		g.progressBar.SetValue(0.8)
		g.statusLabel.SetText("Сохранение результатов...")

		if output != "" {
			if err := stats.SaveStats(result, output); err != nil {
				dialog.ShowError(fmt.Errorf("Ошибка сохранения результатов: %v", err), g.mainWindow)
			}
		}

		select {
		case <-ctx.Done():
			return
		default:
		}

		g.progressBar.SetValue(1.0)
		g.statusLabel.SetText("Анализ GitHub репозитория завершен успешно!")

		resultText := g.formatResults(result, output)
		g.resultText.SetText(resultText)

		g.progressBar.Hide()
	}()
}

func (g *LintVisionGUI) Run() {
	g.mainWindow.ShowAndRun()
}
