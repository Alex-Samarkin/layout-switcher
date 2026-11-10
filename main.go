// "github.com/godbus/dbus/v5"
package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"slices"
	"time"

	"github.com/godbus/dbus/v5"
)

const (
	enChars = "qwertyuiop[]asdfghjkl;'zxcvbnm,.QWERTYUIOP{}ASDFGHJKL:\"ZXCVBNM<>`~"
	ruChars = "йцукенгшщзхъфывапролджэячсмитьбюЙЦУКЕНГШЩЗХЪФЫВАПРОЛДЖЭЯЧСМИТЬБЮёЁ"
)

// Функция трансляции текста в обе стороны
func transliterate(text string) string {
	runes := []rune(text)
	enRunes := []rune(enChars)
	ruRunes := []rune(ruChars)

	for i, r := range runes {
		if idx := slices.Index(enRunes, r); idx != -1 {
			runes[i] = ruRunes[idx]
		} else if idx := slices.Index(ruRunes, r); idx != -1 {
			runes[i] = enRunes[idx]
		}
	}
	return string(runes)
}

// Мягкий короткий системный звук уведомления (~0.3 секунды)
func playSound() {
	_ = exec.Command("paplay", "/usr/share/sounds/freedesktop/stereo/message.oga").Run()
}

// Показывает всплывающее уведомление с результатом конвертации
func notifyResult(text string) {
	body := text
	if r := []rune(text); len(r) > 80 {
		body = string(r[:80]) + "…"
	}
	_ = exec.Command("notify-send", "--app-name", "layout-switcher", "Конвертация", body).Run()
}

// Читает выделенный текст (primary selection) через wl-paste.
// Таймаут нужен, потому что при пустом выделении wl-paste может ждать появления данных.
func readPrimarySelection() string {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	out, err := exec.CommandContext(ctx, "wl-paste", "--primary", "--no-newline").Output()
	if err != nil {
		return ""
	}
	return string(out)
}

func main() {
	// Флаг -buffer_only: конвертировать только содержимое буфера обмена, не читая выделение
	bufferOnly := flag.Bool("buffer_only", false, "конвертировать только содержимое буфера обмена")
	flag.Parse()

	// 1. Подключаемся к D-Bus для чтения буфера Klipper (нужен и для синхронизации в конце)
	conn, err := dbus.ConnectSessionBus()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка D-Bus: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	klipper := conn.Object("org.kde.klipper", "/klipper")

	// 2. Сначала пробуем прочитать выделенный текст (primary selection)
	//    если выделения нет (или задан -buffer_only) — берём текущий буфер Klipper
	var currentText string
	if !*bufferOnly {
		currentText = readPrimarySelection()
	}
	if currentText == "" {
		_ = klipper.Call("org.kde.klipper.klipper.getClipboardContents", 0).Store(&currentText)
	}
	if currentText == "" {
		return
	}

	// 3. Транслитерируем текст
	convertedText := transliterate(currentText)

	// Если переводить нечего, просто выходим
	if convertedText == currentText {
		return
	}

	// 4. Прямая запись в буфер обмена Wayland через wl-copy
	// Флаг --trim-newline убирает лишние переносы строк при записи
	cmd := exec.Command("wl-copy", "--trim-newline")
	cmd.Stdin = bytes.NewBufferString(convertedText)
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка записи через wl-copy: %v\n", err)
		return
	}

	// Синхронизируем изменение с Klipper, чтобы текст появился в истории
	_ = klipper.Call("org.kde.klipper.klipper.setClipboardContents", 0, convertedText).Store()

	// Мягкий звук успешной конвертации
	playSound()

	// Всплывающее уведомление с результатом
	notifyResult(convertedText)
}
