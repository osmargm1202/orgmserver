package email

import (
	"fmt"
	"net/smtp"
	"orgmserver/notification"
	"orgmserver/utils"
	"os"
	"time"
)

type EmailService struct {
	appName  string
	host     string
	port     int
	user     string
	password string
	to       string
	debug    bool
	pending  []notification.Event
}

func NewEmailService(appName string, host string, port int, user, password, to string, debug bool) *EmailService {
	return &EmailService{
		appName:  appName,
		host:     host,
		port:     port,
		user:     user,
		password: password,
		to:       to,
		debug:    debug,
	}
}

// SendStartupEmail envía correo cuando el servicio inicia
func (e *EmailService) SendStartupEmail(ip string) error {
	subject := fmt.Sprintf("Servidor %s Iniciado", e.appName)

	body := fmt.Sprintf(`Servidor %s iniciado correctamente.

Estado: Funcionando y Activo
IP Externa: %s
Fecha/Hora: %s

El servicio está monitoreando la conexión a internet cada minuto.`,
		e.appName, ip, time.Now().Format("2006-01-02 15:04:05"))

	return e.sendEmail(subject, body)
}

// SendReconnectionEmail envía correo cuando se restaura la conexión
func (e *EmailService) SendReconnectionEmail(ip string, duration time.Duration) error {
	subject := fmt.Sprintf("Conexión Restaurada - %s", e.appName)

	minutes := int(duration.Minutes())
	seconds := int(duration.Seconds()) % 60

	body := fmt.Sprintf(`Conexión a internet restaurada.

IP Externa: %s
Duración de desconexión: %d minutos y %d segundos
Fecha/Hora de restauración: %s

El servicio continúa monitoreando la conexión.`,
		ip, minutes, seconds, time.Now().Format("2006-01-02 15:04:05"))

	return e.sendEmail(subject, body)
}

// SendIPChangeEmail envía correo cuando cambia la IP externa
func (e *EmailService) SendIPChangeEmail(newIP string, oldIP string) error {
	subject := fmt.Sprintf("Cambio de IP Externa - %s", e.appName)

	body := fmt.Sprintf(`Se ha detectado un cambio en la IP externa.

IP Anterior: %s
IP Nueva: %s
Fecha/Hora: %s

El servicio continúa monitoreando la conexión.`,
		oldIP, newIP, time.Now().Format("2006-01-02 15:04:05"))

	return e.sendEmail(subject, body)
}

func (e *EmailService) sendEmail(subject, body string) error {
	if os.Getenv("NOTIFICATION_DIR") != "" {
		e.pending = append(e.pending, notification.Event{Time: time.Now(), Kind: subject, Detail: body})
		return e.RetryPending()
	}

	utils.WriteLog(fmt.Sprintf("[EMAIL] Intentando enviar correo a %s: %s", e.to, subject), e.debug)

	addr := fmt.Sprintf("%s:%d", e.host, e.port)
	auth := smtp.PlainAuth("", e.user, e.password, e.host)

	msg := []byte(fmt.Sprintf("To: %s\r\n"+
		"Subject: %s\r\n"+
		"MIME-Version: 1.0\r\n"+
		"Content-Type: text/plain; charset=UTF-8\r\n"+
		"\r\n"+
		"%s\r\n", e.to, subject, body))

	err := smtp.SendMail(addr, auth, e.user, []string{e.to}, msg)
	if err != nil {
		utils.WriteLog(fmt.Sprintf("[EMAIL] Error enviando correo: %v", err), e.debug)
		return fmt.Errorf("error enviando correo: %w", err)
	}

	utils.WriteLog(fmt.Sprintf("[EMAIL] Correo enviado exitosamente a %s", e.to), e.debug)
	return nil
}

// RetryPending retries local queue failures on every monitor tick. Events already
// published remain on disk until orgmdns confirms SMTP delivery.
func (e *EmailService) RetryPending() error {
	dir := os.Getenv("NOTIFICATION_DIR")
	if dir == "" {
		return nil
	}
	for len(e.pending) > 0 {
		if err := notification.Enqueue(dir, e.pending[0]); err != nil {
			return fmt.Errorf("no se pudo guardar el aviso para el resumen: %w", err)
		}
		e.pending = e.pending[1:]
	}
	return nil
}
