package main

import (
	"os"
	"fmt"
	"bufio"
	"strings"

	"user/core/filehandler"
	"user/uifunc"
)

var isSeeding bool = false

func main() {

	// Инициализация PublicHouse (ph)
    ph := &filehandler.PublicHouse{}
	filehandler.EnsureFileExists()
    
    fmt.Println("--- Приложение BitTorrent Клиент/Сидер ---")

    reader := bufio.NewReader(os.Stdin)

    for {
        fmt.Println("\n--- Статус: " + getStatusString() + " ---") // <-- Отображаем статус
        fmt.Println("Выберите действие:")
        fmt.Println("1. Создать мета-файл и начать сидировать (CreateMetaFile)")
        
        // Динамическое меню:
        if !isSeeding {
            fmt.Println("2. Запустить сид-сервер (Seed)")
        } else {
            fmt.Println("2. Остановить сид-сервер (Stop Seed)") // <-- Новый пункт меню
        }
        
        fmt.Println("3. Скачать файл (Download)")
        fmt.Println("4. Выход")
        fmt.Print("Введите номер (1-4): ")

        input, _ := reader.ReadString('\n')
        input = strings.TrimSpace(input)

        switch input {
        case "1":
            handleCreateMetaFile(reader, ph)
        case "2":
            if !isSeeding {
                handleStartSeed(ph) // <-- Запуск
            } else {
                handleStopSeed()    // <-- Остановка
            }
        case "3":
            handleDownload(reader, ph)
        case "4":
            fmt.Println("👋 Выход из программы.")
            // Возможно, здесь нужно добавить логику для корректной остановки сервера, 
            // если он запущен.
            return
        default:
            fmt.Println("❌ Неверный ввод. Пожалуйста, введите число от 1 до 4.")
        }
    }

}	

// --- Обработчики Функций ---

func handleCreateMetaFile(reader *bufio.Reader, ph *filehandler.PublicHouse) {
    fmt.Print("Введите полный путь к файлу для сидирования (e.g., /path/to/file.txt): ")
    filePath, _ := reader.ReadString('\n')
    filePath = strings.TrimSpace(filePath)

    fmt.Print("Введите IP/адрес трекера (e.g., http://tracker.com:8080): ")
    trackerIP, _ := reader.ReadString('\n')
    trackerIP = strings.TrimSpace(trackerIP)

    if filePath != "" && trackerIP != "" {
        uifunc.CreateMetaFile(filePath, trackerIP, ph)
        fmt.Println("✅ Запрос на создание мета-файла и начало сидирования отправлен.")
    } else {
        fmt.Println("❌ Путь к файлу и IP трекера не могут быть пустыми.")
    }
}

func getStatusString() string {
    if isSeeding {
        return "🟢 СЕРВЕР РАБОТАЕТ"
    }
    return "🔴 СЕРВЕР ОСТАНОВЛЕН"
}

func handleStartSeed(ph *filehandler.PublicHouse) {
    uifunc.StartSeederBackground(ph) // Вызываем функцию, которая запускает в горутине
    isSeeding = true          // Обновляем состояние
}

func handleStopSeed() {
    uifunc.StopSeeder() // Вызываем функцию остановки
    isSeeding = false // Обновляем состояние (предполагая успешную остановку)
}

func handleDownload(reader *bufio.Reader, ph *filehandler.PublicHouse) {
    fmt.Print("Введите полный путь к мета-файлу (.meta) для скачивания (e.g., /path/to/file.meta): ")
    metaFilePath, _ := reader.ReadString('\n')
    metaFilePath = strings.TrimSpace(metaFilePath)

    if metaFilePath != "" {
        uifunc.Download(metaFilePath, ph)
    } else {
        fmt.Println("❌ Путь к мета-файлу не может быть пустым.")
    }
}

