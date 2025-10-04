// main_test.go
package main

import (
	"os"
	"path/filepath"
	"testing"
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

// NOTA: Funciones como monitorClipboard, openBrowser, o las relacionadas con la GUI (showLogWindow)
// no se testean con tests unitarios porque dependen de interacciones del usuario, el sistema operativo
// o la red. Para ellas se usan tests de integración o manuales.
