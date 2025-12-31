package main

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func main() {
	myApp := app.New()
	myApp.Settings().SetTheme(theme.DarkTheme())
	window := myApp.NewWindow("RKN SIMULATOR: RU-STORE")
	window.Resize(fyne.NewSize(850, 1000))

	// --- СОСТОЯНИЕ ---
	budget := 10 // Хардкорный старт
	anger := 0.0
	corruption := 0
	days := 1
	isGameOver := false
	isRKN := true
	vpnActive := false
	
	blockSpeed := 40 * time.Millisecond
	angerMitigation := 1.0
	autoAccept := false
	goldenParachute := false

	apps := []string{"Telegram", "YouTube", "Discord", "Steam", "Minecraft", "TikTok", "Roblox", "Instagram", "Facebook", "Twitter"}
	banned := make(map[string]bool)
	slowed := make(map[string]bool)

	// --- UI ЭЛЕМЕНТЫ ---
	header := widget.NewLabel(fmt.Sprintf("ДЕНЬ %d | РОЛЬ: АДМИНИСТРАТОР", days))
	budgetLabel := widget.NewLabel(fmt.Sprintf("БЮДЖЕТ: %d ₽", budget))
	angerBar := widget.NewProgressBar()
	orderLabel := widget.NewLabel("ОЖИДАНИЕ ПРИКАЗОВ...")
	log := widget.NewMultiLineEntry()
	log.Disable()

	iconDisp := container.NewCenter()
	progress := widget.NewProgressBar()
	var target string
	var currentOrder string
	var timeLeft int

	btnAccept := widget.NewButtonWithIcon("ПРИНЯТЬ", theme.ConfirmIcon(), nil)
	btnBribe := widget.NewButtonWithIcon("ВЗЯТКА", theme.ContentAddIcon(), nil)
	btnAccept.Hide()
	btnBribe.Hide()

	// --- ФУНКЦИИ ---
	checkStatus := func() {
		if isGameOver { return }
		if budget < -150000 {
			isGameOver = true
			window.SetContent(container.NewCenter(widget.NewLabel("ВАС УВОЛИЛИ: БЮДЖЕТ ПУСТ")))
		}
		if anger >= 1.0 {
			if goldenParachute {
				goldenParachute = false
				anger = 0.5
				angerBar.SetValue(0.5)
				log.SetText(log.Text + "[!] ПАРАШЮТ: Гнев снижен до 50%!\n")
			} else {
				isGameOver = true
				window.SetContent(container.NewCenter(widget.NewLabel("РЕВОЛЮЦИЯ: ВЫ ПРОИГРАЛИ")))
			}
		}
	}

	// --- МАГАЗИН ---
	shopBtn := widget.NewButtonWithIcon("МАГАЗИН", theme.SettingsIcon(), func() {
		sWindow := myApp.NewWindow("ГОСЗАКУПКИ")
		sWindow.SetContent(container.NewVBox(
			widget.NewButton("ИИ-Цензор (150к) - Автоприем", func() {
				if budget >= 150000 { budget -= 150000; autoAccept = true; sWindow.Close() }
			}),
			widget.NewButton("Парашют (200к) - 2-й шанс", func() {
				if budget >= 200000 { budget -= 200000; goldenParachute = true; sWindow.Close() }
			}),
		))
		sWindow.Show()
	})

	// --- ИГРОВЫЕ ЦИКЛЫ ---
	go func() { // Время и расходы
		for !isGameOver {
			time.Sleep(30 * time.Second)
			days++
			budget -= 10000
			budgetLabel.SetText(fmt.Sprintf("БЮДЖЕТ: %d ₽", budget))
			header.SetText(fmt.Sprintf("ДЕНЬ %d | РОЛЬ: %s", days, map[bool]string{true: "АДМИНИСТРАТОР", false: "ЮЗЕР"}[isRKN]))
			checkStatus()
		}
	}()

	go func() { // Задания
		for !isGameOver {
			time.Sleep(time.Duration(rand.Intn(10)+8) * time.Second)
			if currentOrder == "" {
				if autoAccept {
					currentOrder = apps[rand.Intn(len(apps))]
					timeLeft = 30
					log.SetText(log.Text + "[ИИ] Цель: " + currentOrder + "\n")
				} else if isRKN {
					btnAccept.Show()
					time.Sleep(5 * time.Second)
					btnAccept.Hide()
				}
			}
		}
	}()

	// --- КНОПКИ ДЕЙСТВИЯ ---
	btnBlock := widget.NewButton("БАН", func() {
		if target == "" || banned[target] || isGameOver { return }
		go func() {
			for i := 0.0; i <= 1.0; i += 0.05 { progress.SetValue(i); time.Sleep(blockSpeed) }
			banned[target] = true
			if target == currentOrder { budget += 110000; currentOrder = ""; orderLabel.SetText("УСПЕХ") }
			anger += 0.18 * angerMitigation
			angerBar.SetValue(anger)
			budgetLabel.SetText(fmt.Sprintf("БЮДЖЕТ: %d ₽", budget))
			checkStatus()
		}()
	})

	btnOpen := widget.NewButton("ЗАПУСТИТЬ", func() {
		if target == "" { return }
		res := "[OK] Работает"
		if banned[target] && !vpnActive { res = "[X] ЗАБАНЕНО" }
		log.SetText(log.Text + res + ": " + target + "\n")
	})

	modeBtn := widget.NewButton("РОЛЬ", func() { isRKN = !isRKN })

	btnAccept.OnTapped = func() {
		currentOrder = apps[rand.Intn(len(apps))]
		timeLeft = 30
		btnAccept.Hide()
		go func() {
			for timeLeft > 0 && currentOrder != "" {
				time.Sleep(time.Second)
				timeLeft--
				orderLabel.SetText(fmt.Sprintf("ЦЕЛЬ: %s (%dс)", currentOrder, timeLeft))
			}
		}()
	}

	btnBribe.OnTapped = func() {
		budget += 85000; anger += 0.12; corruption++
		budgetLabel.SetText(fmt.Sprintf("БЮДЖЕТ: %d ₽", budget))
		btnBribe.Hide(); checkStatus()
	}

	combo := widget.NewSelect(apps, func(s string) {
		target = s
		iconDisp.Objects = nil
		img := canvas.NewImageFromFile(fmt.Sprintf("assets/%s.png", strings.ToLower(s)))
		img.FillMode = canvas.ImageFillContain
		img.SetMinSize(fyne.NewSize(120, 120))
		iconDisp.Add(img)
	})

	rknBox := container.NewGridWithColumns(2, btnBlock, widget.NewButton("МЕДЛЕННО", func() {
		slowed[target] = true; anger += 0.05; angerBar.SetValue(anger); checkStatus()
	}))
	userBox := container.NewVBox(btnOpen, widget.NewCheck("VPN", func(v bool) { vpnActive = v }))

	go func() {
		for !isGameOver {
			if isRKN { rknBox.Show(); userBox.Hide() } else { rknBox.Hide(); userBox.Show() }
			time.Sleep(200 * time.Millisecond)
		}
	}()

	window.SetContent(container.NewPadded(container.NewVBox(
		container.NewCenter(header),
		container.NewHBox(budgetLabel, layout.NewSpacer(), modeBtn, shopBtn),
		angerBar,
		widget.NewCard("ПРИКАЗЫ", "", container.NewVBox(orderLabel, container.NewHBox(btnAccept, btnBribe))),
		widget.NewCard("ОБЪЕКТ", "", container.NewVBox(combo, iconDisp, progress, rknBox, userBox)),
		log,
	)))

	window.ShowAndRun()
}