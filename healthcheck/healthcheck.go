package healthcheck

import (
	"fmt"
	"net/http"
	"orgmserver/utils"
	"time"
)

type HealthcheckService struct {
	url    string
	client *http.Client
	debug  bool
}

func NewHealthcheckService(url string, debug bool) *HealthcheckService {
	return &HealthcheckService{
		url: url,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
		debug: debug,
	}
}

// SendHealthcheck envía una solicitud HTTP al healthcheck URL
func (h *HealthcheckService) SendHealthcheck() error {
	if h.url == "" {
		return nil // No configurado, no hacer nada
	}

	utils.WriteLog(fmt.Sprintf("[HEALTHCHECK] Enviando healthcheck a %s", h.url), h.debug)

	req, err := http.NewRequest("GET", h.url, nil)
	if err != nil {
		utils.WriteLog(fmt.Sprintf("[HEALTHCHECK] Error creando request: %v", err), h.debug)
		return fmt.Errorf("error creando request: %w", err)
	}

	// Agregar User-Agent
	req.Header.Set("User-Agent", "ORGMServer-Healthcheck/1.0")

	resp, err := h.client.Do(req)
	if err != nil {
		utils.WriteLog(fmt.Sprintf("[HEALTHCHECK] Error enviando request: %v", err), h.debug)
		return fmt.Errorf("error enviando healthcheck: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		utils.WriteLog(fmt.Sprintf("[HEALTHCHECK] Healthcheck exitoso: status %d", resp.StatusCode), h.debug)
		return nil
	}

	utils.WriteLog(fmt.Sprintf("[HEALTHCHECK] Healthcheck recibió status code: %d", resp.StatusCode), h.debug)
	return fmt.Errorf("healthcheck recibió status code: %d", resp.StatusCode)
}

