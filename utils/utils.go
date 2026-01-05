package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type State struct {
	LastConnected    time.Time `json:"last_connected"`
	LastDisconnected time.Time `json:"last_disconnected"`
	IsConnected      bool      `json:"is_connected"`
	StartTime        time.Time `json:"start_time"`
}

// GetExternalIP obtiene la IP externa intentando múltiples servicios
func GetExternalIP() (string, error) {
	services := []string{
		"https://api.ipify.org?format=text",
		"https://ifconfig.me/ip",
		"https://icanhazip.com",
		"https://api.ip.sb/ip",
	}

	for _, service := range services {
		ip, err := tryGetIP(service)
		if err == nil && ip != "" {
			log.Printf("[DEBUG] IP externa obtenida desde %s: %s", service, ip)
			return ip, nil
		}
		log.Printf("[DEBUG] Error obteniendo IP desde %s: %v", service, err)
	}

	return "", fmt.Errorf("no se pudo obtener la IP externa desde ningún servicio")
}

func tryGetIP(url string) (string, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	ip := string(body)
	// Limpiar espacios y saltos de línea
	ip = strings.TrimSpace(ip)
	
	if ip == "" {
		return "", fmt.Errorf("respuesta vacía")
	}

	return ip, nil
}

// LoadState carga el estado desde el archivo
func LoadState(filePath string) (*State, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			// Archivo no existe, crear estado inicial
			state := &State{
				LastConnected: time.Now(),
				IsConnected:   true,
				StartTime:     time.Now(),
			}
			return state, nil
		}
		return nil, err
	}

	var state State
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, err
	}

	return &state, nil
}

// SaveState guarda el estado en el archivo
func SaveState(filePath string, state *State) error {
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}

	// Asegurar que el directorio existe
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.WriteFile(filePath, data, 0644)
}

// GetCurrentTime retorna el tiempo actual
func GetCurrentTime() time.Time {
	return time.Now()
}

// WriteLog escribe un mensaje de log con timestamp
func WriteLog(message string, debug bool) {
	if !debug {
		return
	}

	// Asegurar que el directorio logs existe
	if err := os.MkdirAll("logs", 0755); err != nil {
		log.Printf("[ERROR] No se pudo crear directorio logs: %v", err)
		return
	}

	logFile := "logs/orgmserver.log"
	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("[ERROR] No se pudo abrir archivo de log: %v", err)
		return
	}
	defer f.Close()

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	logMessage := fmt.Sprintf("[%s] %s\n", timestamp, message)
	
	if _, err := f.WriteString(logMessage); err != nil {
		log.Printf("[ERROR] No se pudo escribir en log: %v", err)
	}
	
	// También imprimir en consola
	log.Printf("[LOG] %s", message)
}

