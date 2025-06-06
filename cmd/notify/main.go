package main

import (
	"fmt"
	"os/exec"
	"time"

	"github.com/gen2brain/beeep"
)

func main() {
	title := "Go Notify"
	message := "Test de notification système"
	icon := "" // Optionnel, chemin vers une icône .png

	fmt.Println("Tentative d'envoi de notification...")

	// 1. Essayer d'envoyer la notif
	err := beeep.Notify(title, message, icon)
	if err != nil {
		fmt.Println("Erreur lors de l'envoi de la notif:", err)
		notifyFallback()
		return
	}

	// 2. Pause pour que l'utilisateur voie la notif (sinon ça va trop vite)
	time.Sleep(2 * time.Second)

	fmt.Println("Notification envoyée. Si rien ne s'est affiché, ouverture des réglages...")
	notifyFallback()
}

// Fonction fallback si la notif échoue ou est silencieuse
func notifyFallback() {
	// 1. Alerte bloquante via AppleScript
	alert := `display alert "Activez les notifications pour cette app dans Réglages > Notifications"`
	_ = exec.Command("osascript", "-e", alert).Run()

	// 2. Ouvre directement les réglages macOS pour les notifications
	_ = exec.Command("open", "x-apple.systempreferences:com.apple.preference.notifications").Run()
}
