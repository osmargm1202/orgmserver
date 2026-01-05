package monitor

import (
	"fmt"
	"orgmserver/config"
	"orgmserver/email"
	"orgmserver/healthcheck"
	"orgmserver/utils"
	"time"
)

type Monitor struct {
	config            *config.Config
	emailService      *email.EmailService
	healthcheckService *healthcheck.HealthcheckService
	stateFilePath     string
	monitorInterval   time.Duration
	isConnected       bool
	disconnectTime    time.Time
	debug             bool
}

func NewMonitor(
	cfg *config.Config,
	emailSvc *email.EmailService,
	debug bool,
) *Monitor {
	healthcheckSvc := healthcheck.NewHealthcheckService(cfg.HealthcheckURL, debug)

	return &Monitor{
		config:            cfg,
		emailService:      emailSvc,
		healthcheckService: healthcheckSvc,
		stateFilePath:     cfg.StateFilePath,
		monitorInterval:   cfg.MonitorInterval,
		isConnected:       true,
		debug:             debug,
	}
}

// Start inicia el loop de monitoreo
func (m *Monitor) Start() error {
	utils.WriteLog("[MONITOR] Iniciando loop de monitoreo", m.debug)

	ticker := time.NewTicker(m.monitorInterval)
	defer ticker.Stop()

	// Primera verificación inmediata
	m.checkConnection()

	for {
		select {
		case <-ticker.C:
			m.checkConnection()
		}
	}
}

// checkConnection verifica la conexión a internet
func (m *Monitor) checkConnection() {
	utils.WriteLog("[MONITOR] Verificando conexión a internet", m.debug)

	// Intentar obtener IP externa para verificar conexión
	ip, err := utils.GetExternalIP()
	
	if err != nil {
		// No hay conexión
		if m.isConnected {
			// Acabamos de perder la conexión
			m.handleDisconnection()
		}
		utils.WriteLog(fmt.Sprintf("[MONITOR] Sin conexión a internet: %v", err), m.debug)
		return
	}

	// Hay conexión
	if !m.isConnected {
		// Acabamos de recuperar la conexión
		m.handleReconnection(ip)
	} else {
		// Conexión estable, verificar cambio de IP y actualizar estado
		m.checkIPChange(ip)
		m.updateState(ip)
	}

	// Enviar healthcheck si está configurado (en goroutine para no bloquear)
	go func() {
		if err := m.healthcheckService.SendHealthcheck(); err != nil {
			utils.WriteLog(fmt.Sprintf("[MONITOR] Error en healthcheck: %v", err), m.debug)
		}
	}()

	m.isConnected = true
}

// handleDisconnection maneja cuando se pierde la conexión
func (m *Monitor) handleDisconnection() {
	utils.WriteLog("[MONITOR] Conexión perdida", m.debug)
	m.isConnected = false
	m.disconnectTime = time.Now()

	// Actualizar estado
	state, err := utils.LoadState(m.stateFilePath)
	if err != nil {
		utils.WriteLog(fmt.Sprintf("[MONITOR] Error cargando estado: %v", err), m.debug)
		state = &utils.State{
			StartTime: time.Now(),
		}
	}

	state.IsConnected = false
	state.LastDisconnected = m.disconnectTime

	if err := utils.SaveState(m.stateFilePath, state); err != nil {
		utils.WriteLog(fmt.Sprintf("[MONITOR] Error guardando estado: %v", err), m.debug)
	}
}

// handleReconnection maneja cuando se recupera la conexión
func (m *Monitor) handleReconnection(ip string) {
	utils.WriteLog("[MONITOR] Conexión restaurada", m.debug)
	
	// Calcular duración de desconexión desde el estado guardado
	state, err := utils.LoadState(m.stateFilePath)
	var duration time.Duration
	
	if err == nil && !state.LastDisconnected.IsZero() {
		duration = time.Since(state.LastDisconnected)
	} else {
		// Si no hay timestamp de desconexión, usar el tiempo desde que detectamos la desconexión
		duration = time.Since(m.disconnectTime)
	}
	
	// Enviar correo de reconexión (solo si hubo desconexión real, no reinicio manual)
	if duration > 0 {
		if err := m.emailService.SendReconnectionEmail(ip, duration); err != nil {
			utils.WriteLog(fmt.Sprintf("[MONITOR] Error enviando correo de reconexión: %v", err), m.debug)
		}
	}

	// Actualizar estado
	state, err = utils.LoadState(m.stateFilePath)
	if err != nil {
		state = &utils.State{
			StartTime: time.Now(),
		}
	}

	state.IsConnected = true
	state.LastConnected = time.Now()
	state.LastDisconnected = time.Time{} // Limpiar desconexión
	state.LastIP = ip // Guardar la nueva IP

	if err := utils.SaveState(m.stateFilePath, state); err != nil {
		utils.WriteLog(fmt.Sprintf("[MONITOR] Error guardando estado: %v", err), m.debug)
	}

	m.isConnected = true
}

// checkIPChange verifica si la IP ha cambiado y envía notificación
func (m *Monitor) checkIPChange(newIP string) {
	state, err := utils.LoadState(m.stateFilePath)
	if err != nil {
		utils.WriteLog(fmt.Sprintf("[MONITOR] Error cargando estado para verificar IP: %v", err), m.debug)
		return
	}

	// Si hay una IP anterior y es diferente a la nueva, hubo un cambio
	if state.LastIP != "" && state.LastIP != newIP {
		utils.WriteLog(fmt.Sprintf("[MONITOR] Cambio de IP detectado: %s -> %s", state.LastIP, newIP), m.debug)
		
		// Enviar correo de cambio de IP
		if err := m.emailService.SendIPChangeEmail(newIP, state.LastIP); err != nil {
			utils.WriteLog(fmt.Sprintf("[MONITOR] Error enviando correo de cambio de IP: %v", err), m.debug)
		}
	}
}

// updateState actualiza el estado cuando hay conexión estable
func (m *Monitor) updateState(ip string) {
	state, err := utils.LoadState(m.stateFilePath)
	if err != nil {
		state = &utils.State{
			StartTime: time.Now(),
		}
	}

	state.IsConnected = true
	state.LastConnected = time.Now()
	state.LastIP = ip // Guardar la IP actual

	if err := utils.SaveState(m.stateFilePath, state); err != nil {
		utils.WriteLog(fmt.Sprintf("[MONITOR] Error guardando estado: %v", err), m.debug)
	}
}


