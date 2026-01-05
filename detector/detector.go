package detector

import (
	"fmt"
	"orgmserver/utils"
	"os"
	"strings"
	"time"
)

type CauseType string

const (
	CauseNormal      CauseType = "normal"
	CausePowerLoss   CauseType = "power_loss"
	CauseInternetLoss CauseType = "internet_loss"
)

type Detector struct {
	stateFilePath string
	startTime     time.Time
	debug         bool
}

func NewDetector(stateFilePath string, debug bool) *Detector {
	return &Detector{
		stateFilePath: stateFilePath,
		startTime:     time.Now(),
		debug:         debug,
	}
}

// DetectStartupCause detecta la causa del inicio del servicio
func (d *Detector) DetectStartupCause() (CauseType, error) {
	utils.WriteLog("[DETECTOR] Detectando causa del inicio del servicio", d.debug)

	// Método 1: Verificar archivo de estado
	stateExists, stateTime, err := d.checkStateFile()
	if err != nil {
		utils.WriteLog(fmt.Sprintf("[DETECTOR] Error verificando archivo de estado: %v", err), d.debug)
	}

	// Método 2: Verificar uptime del sistema
	systemUptime, err := d.getSystemUptime()
	if err != nil {
		utils.WriteLog(fmt.Sprintf("[DETECTOR] Error obteniendo uptime del sistema: %v", err), d.debug)
	}

	utils.WriteLog(fmt.Sprintf("[DETECTOR] Estado archivo existe: %v, Tiempo: %v", stateExists, stateTime), d.debug)
	utils.WriteLog(fmt.Sprintf("[DETECTOR] Uptime del sistema: %v", systemUptime), d.debug)
	utils.WriteLog(fmt.Sprintf("[DETECTOR] Tiempo de inicio del proceso: %v", d.startTime), d.debug)

	// Si no existe archivo de estado, es inicio normal o pérdida de energía
	if !stateExists {
		// Si el uptime del sistema es muy corto (< 5 minutos), probablemente pérdida de energía
		if systemUptime < 5*time.Minute {
			utils.WriteLog("[DETECTOR] Causa detectada: PÉRDIDA DE ENERGÍA (sin archivo de estado + uptime corto)", d.debug)
			return CausePowerLoss, nil
		}
		utils.WriteLog("[DETECTOR] Causa detectada: INICIO NORMAL (sin archivo de estado + uptime normal)", d.debug)
		return CauseNormal, nil
	}

	// Si existe archivo de estado, verificar cuándo fue la última actualización
	timeSinceStateUpdate := time.Since(stateTime)
	
	// Si el archivo de estado es muy antiguo (> 1 hora), probablemente pérdida de energía
	if timeSinceStateUpdate > 1*time.Hour {
		// Verificar si el uptime del sistema es menor que el tiempo desde la última actualización
		if systemUptime < timeSinceStateUpdate {
			utils.WriteLog("[DETECTOR] Causa detectada: PÉRDIDA DE ENERGÍA (archivo antiguo + uptime menor)", d.debug)
			return CausePowerLoss, nil
		}
	}

	// Si el archivo de estado es reciente pero el proceso se reinició, podría ser pérdida de internet
	// Si el uptime del sistema es mayor que el tiempo desde la última actualización del estado,
	// significa que el sistema no se reinició, pero el proceso sí
	if systemUptime > timeSinceStateUpdate && timeSinceStateUpdate < 10*time.Minute {
		utils.WriteLog("[DETECTOR] Causa detectada: PÉRDIDA DE INTERNET (archivo reciente + proceso reiniciado)", d.debug)
		return CauseInternetLoss, nil
	}

	// Por defecto, inicio normal
	utils.WriteLog("[DETECTOR] Causa detectada: INICIO NORMAL", d.debug)
	return CauseNormal, nil
}

// checkStateFile verifica si existe el archivo de estado y su timestamp
func (d *Detector) checkStateFile() (bool, time.Time, error) {
	state, err := utils.LoadState(d.stateFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, time.Time{}, nil
		}
		return false, time.Time{}, err
	}

	// Retornar el tiempo más reciente entre LastConnected y StartTime
	var latestTime time.Time
	if state.LastConnected.After(state.StartTime) {
		latestTime = state.LastConnected
	} else {
		latestTime = state.StartTime
	}

	return true, latestTime, nil
}

// getSystemUptime obtiene el uptime del sistema
func (d *Detector) getSystemUptime() (time.Duration, error) {
	// En Linux, leer /proc/uptime
	data, err := os.ReadFile("/proc/uptime")
	if err != nil {
		// Si no está disponible (no es Linux o no se puede leer), usar una aproximación
		// basada en el tiempo de inicio del proceso
		utils.WriteLog("[DETECTOR] No se pudo leer /proc/uptime, usando aproximación", d.debug)
		return time.Since(d.startTime), nil
	}

	// /proc/uptime contiene dos valores: uptime total y tiempo idle
	// Solo necesitamos el primero
	fields := strings.Fields(string(data))
	if len(fields) == 0 {
		return time.Since(d.startTime), nil
	}

	var uptimeSeconds float64
	_, err = fmt.Sscanf(fields[0], "%f", &uptimeSeconds)
	if err != nil {
		return time.Since(d.startTime), nil
	}

	return time.Duration(uptimeSeconds) * time.Second, nil
}

// GetCauseDescription retorna una descripción legible de la causa
func GetCauseDescription(cause CauseType) string {
	switch cause {
	case CausePowerLoss:
		return "PÉRDIDA DE ENERGÍA"
	case CauseInternetLoss:
		return "PÉRDIDA DE INTERNET"
	default:
		return "INICIO NORMAL"
	}
}

