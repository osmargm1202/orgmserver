package main

import (
	"flag"
	"fmt"
	"log"
	"orgmserver/config"
	"orgmserver/email"
	"orgmserver/monitor"
	"orgmserver/utils"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// Parse flags
	debug := flag.Bool("debug", false, "Habilitar logs de debug")
	flag.Parse()

	// Cargar configuración
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Error cargando configuración: %v", err)
	}

	utils.WriteLog(fmt.Sprintf("[MAIN] Iniciando %s", cfg.AppName), *debug)
	utils.WriteLog("[MAIN] Configuración cargada correctamente", *debug)

	// Obtener IP externa
	ip, err := utils.GetExternalIP()
	if err != nil {
		utils.WriteLog("[MAIN] Error obteniendo IP externa, continuando sin IP", *debug)
		ip = "No disponible"
	}

	utils.WriteLog("[MAIN] IP Externa: "+ip, *debug)

	// Inicializar servicio de email
	emailSvc := email.NewEmailService(
		cfg.AppName,
		cfg.SMTPHost,
		cfg.SMTPPort,
		cfg.SMTPUser,
		cfg.SMTPPassword,
		cfg.EmailTo,
		*debug,
	)

	// Enviar correo de inicio (siempre "iniciado")
	if err := emailSvc.SendStartupEmail(ip); err != nil {
		utils.WriteLog("[MAIN] Error enviando correo de inicio: "+err.Error(), *debug)
		// No fatal, continuar ejecución
	}

	// Inicializar estado - limpiar cualquier desconexión previa
	// Esto asegura que si se para e inicia manualmente, no se detecte como pérdida de internet
	state, err := utils.LoadState(cfg.StateFilePath)
	if err != nil {
		utils.WriteLog("[MAIN] Error cargando estado inicial, creando nuevo", *debug)
		state = &utils.State{
			StartTime:     utils.GetCurrentTime(),
			LastConnected: utils.GetCurrentTime(),
			IsConnected:   true,
			LastIP:        ip, // Guardar IP inicial
		}
	} else {
		// Limpiar estado de desconexión al iniciar (reinicio manual)
		state.IsConnected = true
		state.LastConnected = utils.GetCurrentTime()
		state.LastDisconnected = time.Time{} // Limpiar desconexión previa
		state.LastIP = ip // Actualizar IP inicial
	}

	if err := utils.SaveState(cfg.StateFilePath, state); err != nil {
		utils.WriteLog("[MAIN] Error guardando estado inicial: "+err.Error(), *debug)
	}

	// Inicializar monitor
	mon := monitor.NewMonitor(cfg, emailSvc, *debug)

	// Manejar señales para shutdown graceful
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Iniciar monitor en goroutine
	go func() {
		if err := mon.Start(); err != nil {
			utils.WriteLog("[MAIN] Error en monitor: "+err.Error(), *debug)
			log.Fatal(err)
		}
	}()

	utils.WriteLog("[MAIN] Servicio iniciado y monitoreando", *debug)

	// Esperar señal de terminación
	<-sigChan
	utils.WriteLog("[MAIN] Recibida señal de terminación, cerrando...", *debug)

	// Guardar estado final
	state.IsConnected = false
	if err := utils.SaveState(cfg.StateFilePath, state); err != nil {
		utils.WriteLog("[MAIN] Error guardando estado final: "+err.Error(), *debug)
	}

	utils.WriteLog("[MAIN] Servicio detenido", *debug)
}

