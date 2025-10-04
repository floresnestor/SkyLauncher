// main_test.go
package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"golang.design/x/clipboard"
)

// --- Test para loadConfig ---

func TestLoadConfig(t *testing.T) {
	// Caso 1: Archivo de configuración válido y completo
	t.Run("ValidConfig", func(t *testing.T) {
		content := `
path1=C:\App1\run.exe
path2=C:\App2\start.exe
prefix=PRE-
suffix=-SUF
monitor_timeout_seconds=300
min_token_length=40
`
		// Crear un archivo de config temporal para el test
		tmpFile, err := os.CreateTemp("", "config_*.txt")
		if err != nil {
			t.Fatalf("No se pudo crear el archivo temporal: %v", err)
		}
		defer os.Remove(tmpFile.Name()) // Limpiar después del test

		tmpFile.WriteString(content)
		tmpFile.Close()

		config, err := loadConfig(tmpFile.Name())
		if err != nil {
			t.Errorf("loadConfig falló con un archivo válido: %v", err)
		}

		if config.Path1 != `C:\App1\run.exe` {
			t.Errorf("Path1 incorrecto: se obtuvo %s, se esperaba %s", config.Path1, `C:\App1\run.exe`)
		}
		if config.MonitorTimeoutSeconds != 300 {
			t.Errorf("MonitorTimeoutSeconds incorrecto: se obtuvo %d, se esperaba 300", config.MonitorTimeoutSeconds)
		}
		if config.MinTokenLength != 40 {
			t.Errorf("MinTokenLength incorrecto: se obtuvo %d, se esperaba 40", config.MinTokenLength)
		}
	})

	// Caso 2: Archivo con valores por defecto
	t.Run("DefaultValues", func(t *testing.T) {
		content := `
path1=C:\App1\run.exe
prefix=PRE-
suffix=-SUF
`
		tmpFile, err := os.CreateTemp("", "config_*.txt")
		if err != nil {
			t.Fatalf("No se pudo crear el archivo temporal: %v", err)
		}
		defer os.Remove(tmpFile.Name())
		tmpFile.WriteString(content)
		tmpFile.Close()

		config, err := loadConfig(tmpFile.Name())
		if err != nil {
			t.Errorf("loadConfig falló con un archivo válido: %v", err)
		}

		if config.MonitorTimeoutSeconds != 600 { // Valor por defecto
			t.Errorf("Valor por defecto de MonitorTimeoutSeconds incorrecto: se obtuvo %d, se esperaba 600", config.MonitorTimeoutSeconds)
		}
		if config.MinTokenLength != 50 { // Valor por defecto
			t.Errorf("Valor por defecto de MinTokenLength incorrecto: se obtuvo %d, se esperaba 50", config.MinTokenLength)
		}
	})

	// Caso 3: Archivo no existente
	t.Run("FileNotExists", func(t *testing.T) {
		_, err := loadConfig("archivo_inexistente.txt")
		if err == nil {
			t.Errorf("Se esperaba un error al cargar un archivo inexistente, pero no se obtuvo ninguno.")
		}
	})
}

// --- Test para checkPathsExist ---

func TestCheckPathsExist(t *testing.T) {
	// Crear un directorio temporal para nuestros archivos de prueba
	tmpDir := t.TempDir()

	// Crear archivos falsos para simular los ejecutables
	path1 := filepath.Join(tmpDir, "app1.exe")
	path2 := filepath.Join(tmpDir, "app2.exe")
	os.WriteFile(path1, []byte(""), 0644)
	os.WriteFile(path2, []byte(""), 0644)

	// Caso 1: Ambos paths existen
	t.Run("BothPathsExist", func(t *testing.T) {
		config := Config{Path1: path1, Path2: path2}
		err := checkPathsExist(config)
		if err != nil {
			t.Errorf("Se esperaba que no hubiera error ya que ambos archivos existen, pero se obtuvo: %v", err)
		}
	})

	// Caso 2: Path1 no existe
	t.Run("Path1NotExists", func(t *testing.T) {
		config := Config{Path1: filepath.Join(tmpDir, "no_existe.exe"), Path2: path2}
		err := checkPathsExist(config)
		if err == nil {
			t.Errorf("Se esperaba un error ya que path1 no existe, pero no se obtuvo ninguno.")
		}
	})

	// Caso 3: Path2 no existe
	t.Run("Path2NotExists", func(t *testing.T) {
		config := Config{Path1: path1, Path2: filepath.Join(tmpDir, "no_existe.exe")}
		err := checkPathsExist(config)
		if err == nil {
			t.Errorf("Se esperaba un error ya que path2 no existe, pero no se obtuvo ninguno.")
		}
	})
}

// --- Tests de Integración ---

// Test de Integración para la Conectividad de Red
func TestIntegration_URLAccessibility(t *testing.T) {
	// Este test requiere una conexión real a la red corporativa/VPN.
	t.Log("--- INICIANDO TEST DE INTEGRACIÓN: ACCESO A URL ---")
	t.Log("Requisito: Debes estar conectado a la VPN del banco.")

	if !isURLAccessible() {
		t.Error("isURLAccessible() devolvió 'false'. El test falló.")
		t.Log("Asegúrate de estar conectado a la red corporativa e inténtalo de nuevo.")
	} else {
		t.Log("isURLAccessible() devolvió 'true'. ¡El test pasó exitosamente!")
	}
}

// Test de Integración Interactivo para el Flujo del Portapapeles
func TestIntegration_ClipboardWorkflow(t *testing.T) {
	t.Log("--- INICIANDO TEST DE INTEGRACIÓN: FLUJO DEL PORTAPAPELES (INTERACTIVO) ---")

	// 1. Configuración del Test
	config := Config{
		Prefix: "SKYTEST#;",
		Suffix: ";#TESTSKY",
	}

	// 2. Escribir el estado inicial en el portapapeles
	clipboard.Write(clipboard.FmtText, []byte(initialClipboardState))
	t.Logf("Portapapeles inicializado con la bandera: '%s'", initialClipboardState)

	// 3. Instrucción para el usuario
	testToken := "ESTE_ES_UN_TOKEN_DE_PRUEBA_MANUAL"
	t.Logf("\n\n*** ACCIÓN REQUERIDA ***\n")
	t.Logf("Por favor, copia el siguiente texto en tu portapapeles (Ctrl+C):")
	t.Logf("===> %s <===", testToken)
	t.Logf("\nTienes 15 segundos para hacerlo...\n\n")

	// 4. Monitoreo del cambio y verificación
	timeout := time.After(15 * time.Second)
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			t.Fatal("TEST FALLIDO: Tiempo de espera agotado. No se detectó el cambio en el portapapeles.")
			return
		case <-ticker.C:
			clipboardText := strings.TrimSpace(string(clipboard.Read(clipboard.FmtText)))

			// Primero, esperamos a que el usuario copie el token de prueba
			if clipboardText == testToken {
				t.Log("Paso 1/2: Se detectó el token de prueba copiado por el usuario.")

				// Simular la lógica de monitorClipboard
				formattedToken := config.Prefix + clipboardText + config.Suffix
				clipboard.Write(clipboard.FmtText, []byte(formattedToken))

				t.Log("Paso 2/2: El token ha sido formateado y reescrito en el portapapeles.")

				// Verificación final
				finalContent := string(clipboard.Read(clipboard.FmtText))
				if finalContent == formattedToken {
					t.Logf("VERIFICACIÓN EXITOSA: El portapapeles ahora contiene: '%s'", finalContent)
					return // ¡Test exitoso!
				} else {
					t.Fatalf("TEST FALLIDO: Se esperaba '%s' en el portapapeles, pero se encontró '%s'", formattedToken, finalContent)
				}
			}
		}
	}
}
