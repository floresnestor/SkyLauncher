// main.go
package main

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gen2brain/beeep"
	"github.com/getlantern/systray"
	"github.com/lxn/walk"
	decl "github.com/lxn/walk/declarative" // <-- MODIFICADO
	"golang.design/x/clipboard"
)

// --- Constantes y Estructuras ---
const (
	remoteURL             = "https://genesys.bancopatagonia.net.ar/GenesysNotify/genesysCloud.html"
	configFileName        = "config.txt"
	initialClipboardState = "ESTADOINICIAL"
)

type Config struct {
	Path1, Path2, Prefix, Suffix string
	MonitorTimeoutSeconds        int
	MinTokenLength               int
}

var (
	logWindow   *walk.MainWindow
	logView     *walk.TextEdit
	logMessages = make(chan string, 100)
	once        sync.Once
)

func main() {
	systray.Run(onReady, onExit)
}

func onReady() {
	systray.SetTitle("Genesys Sky Launcher")
	systray.SetTooltip("Iniciando...")
	mShowLog := systray.AddMenuItem("Ver Log en Tiempo Real", "Muestra la ventana de debug")
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Salir", "Cierra la aplicación")

	go func() {
		for {
			select {
			case <-mShowLog.ClickedCh:
				showLogWindow()
			case <-mQuit.ClickedCh:
				systray.Quit()
				return
			}
		}
	}()

	go runLauncher()
}

func onExit() {
	log.Println("Aplicación cerrada.")
}

func runLauncher() {
	setupLogging()

	config, err := loadConfig(configFileName)
	if err != nil {
		log.Printf("Error de configuración: %v", err)
		beeep.Alert("Error de Configuración", err.Error(), "")
		systray.Quit()
		return
	}
	log.Println("Configuración cargada exitosamente.")
	systray.SetTooltip("Verificando rutas de acceso...")

	if err := checkPathsExist(config); err != nil {
		log.Printf("Error de acceso a rutas: %v", err)
		beeep.Alert("Error de Acceso", err.Error(), "")
		systray.Quit()
		return
	}
	log.Println("Acceso a los ejecutables verificado correctamente.")

	systray.SetTooltip("Verificando conexión de red...")
	if !isURLAccessible() {
		log.Println("Fallo en el chequeo de conectividad a la URL. El servidor no es accesible.")
		beeep.Alert("Error de Red", "No se puede acceder al concentrador desplegado dentro del banco. Verifique su conexión de red o VPN.", "")
		systray.Quit()
		return
	}
	log.Println("Chequeo de conectividad a la URL exitoso.")

	log.Println("Lanzando aplicación de path1 en segundo plano...")
	if err := launchDetached(config.Path1); err != nil {
		log.Printf("Error al lanzar path1: %v", err)
		beeep.Alert("Error al lanzar Path1", err.Error(), "")
		systray.Quit()
		return
	}
	log.Println("Aplicación de path1 lanzada.")

	clipboard.Write(clipboard.FmtText, []byte(initialClipboardState))
	log.Printf("Portapapeles inicializado con la bandera: '%s'", initialClipboardState)

	openBrowser(remoteURL)
	log.Println("Navegador Edge lanzado.")
	systray.SetTooltip("Esperando que se copie el token...")
	beeep.Notify("Esperando Token", "Por favor, completa el login para copiar el token.", "")

	monitorClipboard(config)

	systray.Quit()
}

func isURLAccessible() bool {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Transport: tr,
		Timeout:   8 * time.Second,
	}
	req, err := http.NewRequest("GET", remoteURL, nil)
	if err != nil {
		log.Printf("Error al crear la petición de chequeo de red: %v", err)
		return false
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Fallo al contactar la URL: %v", err)
		return false
	}
	resp.Body.Close()
	return true
}

func checkPathsExist(config Config) error {
	log.Printf("Verificando acceso a: %s", config.Path1)
	if _, err := os.Stat(config.Path1); os.IsNotExist(err) {
		return fmt.Errorf("no se pudo encontrar el archivo de path1: %s. Asegúrese de que la unidad de red esté montada", config.Path1)
	}

	if config.Path2 != "" {
		log.Printf("Verificando acceso a: %s", config.Path2)
		if _, err := os.Stat(config.Path2); os.IsNotExist(err) {
			return fmt.Errorf("no se pudo encontrar el archivo de path2: %s", config.Path2)
		}
	}
	return nil
}

func launchDetached(path string) error {
	cmdDir := filepath.Dir(path)
	cmdFile := filepath.Base(path)
	cmd := exec.Command(cmdFile)
	cmd.Dir = cmdDir
	return cmd.Start()
}

func monitorClipboard(config Config) {
	log.Printf("Comenzando monitoreo del portapapeles (Timeout: %d segundos, Longitud Mínima: %d)", config.MonitorTimeoutSeconds, config.MinTokenLength)
	timeout := time.After(time.Duration(config.MonitorTimeoutSeconds) * time.Second)
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			log.Println("Tiempo de espera agotado.")
			beeep.Alert("Tiempo Agotado", fmt.Sprintf("El token no fue copiado en %d segundos.", config.MonitorTimeoutSeconds), "")
			return
		case <-ticker.C:
			clipboardText := strings.TrimSpace(string(clipboard.Read(clipboard.FmtText)))

			if clipboardText != initialClipboardState && len(clipboardText) >= config.MinTokenLength {
				token := clipboardText
				log.Printf("¡Cambio válido detectado! Token crudo capturado (longitud %d): [%s]", len(token), token)

				formattedToken := config.Prefix + token + config.Suffix
				clipboard.Write(clipboard.FmtText, []byte(formattedToken))
				log.Printf("Portapapeles actualizado con el token formateado: [%s]", formattedToken)
				beeep.Notify("Token Procesado", "El token ha sido formateado.", "")

				executeFinalCommand(config)
				return
			}
		}
	}
}

func executeFinalCommand(config Config) {
	if config.Path2 != "" {
		log.Printf("Ejecutando el comando final: %s", config.Path2)
		if err := launchDetached(config.Path2); err != nil {
			msg := fmt.Sprintf("No se pudo ejecutar el comando final: %s\nError: %v", config.Path2, err)
			log.Println(msg)
			beeep.Alert("Error en Comando Final", msg, "")
		} else {
			log.Printf("Comando final ejecutado correctamente.")
		}
	}
}

func loadConfig(filename string) (Config, error) {
	var config Config
	var configPath string

	// NUEVO: Comprobar si 'filename' ya es una ruta absoluta.
	if filepath.IsAbs(filename) {
		// Si es absoluta (como en los tests), usarla directamente.
		configPath = filename
	} else {
		// Si no (uso normal), buscarla junto al ejecutable.
		exePath, _ := os.Executable()
		configPath = filepath.Join(filepath.Dir(exePath), filename)
	}

	file, err := os.Open(configPath)
	if err != nil {
		return config, fmt.Errorf("no se pudo encontrar '%s'", configPath)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key, value := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
			switch key {
			case "path1":
				config.Path1 = value
			case "path2":
				config.Path2 = value
			case "prefix":
				config.Prefix = value
			case "suffix":
				config.Suffix = value
			case "monitor_timeout_seconds":
				config.MonitorTimeoutSeconds, _ = strconv.Atoi(value)
			case "min_token_length":
				config.MinTokenLength, _ = strconv.Atoi(value)
			}
		}
	}

	if config.MonitorTimeoutSeconds <= 0 {
		config.MonitorTimeoutSeconds = 600
	}
	if config.MinTokenLength <= 0 {
		config.MinTokenLength = 50
	}
	if config.Path1 == "" || config.Prefix == "" || config.Suffix == "" {
		return config, fmt.Errorf("el archivo '%s' debe contener al menos 'path1', 'prefix' y 'suffix'", filename)
	}
	return config, nil
}

type writerFunc func(p []byte) (n int, err error)

func (wf writerFunc) Write(p []byte) (n int, err error) { return wf(p) }

func setupLogging() {
	logFileName := fmt.Sprintf("genesysSkyLauncher_%s.log", time.Now().Format("2006-01-02"))
	exePath, err := os.Executable()
	if err != nil {
		log.Fatalf("No se pudo obtener la ruta del ejecutable: %v", err)
	}
	logDir := filepath.Dir(exePath)
	fullLogPath := filepath.Join(logDir, logFileName)

	logFile, err := os.OpenFile(fullLogPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf("Error fatal al abrir el archivo de log en %s: %v", fullLogPath, err)
	}

	multiWriter := io.MultiWriter(logFile, writerFunc(func(p []byte) (int, error) {
		msg := string(p)
		select {
		case logMessages <- msg:
		default:
		}
		return len(p), nil
	}))

	log.SetFlags(log.Ldate | log.Ltime)
	log.SetOutput(multiWriter)
	log.Println("==================================================")
	log.Printf("Iniciando aplicación. El log se encuentra en: %s", fullLogPath)
}

// MODIFICADO: Se añade el prefijo 'decl.' a todos los widgets
func showLogWindow() {
	once.Do(func() {
		decl.MainWindow{
			AssignTo: &logWindow,
			Title:    "Log en Tiempo Real - Genesys Sky Launcher",
			Size:     decl.Size{Width: 800, Height: 600},
			Layout:   decl.VBox{},
			Children: []decl.Widget{
				decl.TextEdit{
					AssignTo: &logView,
					ReadOnly: true,
					VScroll:  true,
				},
			},
		}.Create()
		go func() {
			for msg := range logMessages {
				if logView != nil && logView.Text() != "" {
					logView.AppendText(strings.ReplaceAll(msg, "\n", "\r\n"))
				}
			}
		}()
	})
	logWindow.Show()
	logWindow.SetFocus()
}

func openBrowser(url string) {
	edgePath := `C:\Program Files (x86)\Microsoft\Edge\Application\msedge.exe`
	cmd := exec.Command(edgePath, url)
	if err := cmd.Start(); err != nil {
		log.Printf("Error al lanzar Edge: %v", err)
	}
}
