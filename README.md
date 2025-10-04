# ⚙️ Genesys Sky Launcher

Aplicación en Go para automatizar la ejecución de Sky y PureCloud, obtener access_token para el websocket que conecta con PureCloud y dejarlo en el portapapeles. La herramienta se encarga de verificar el acceso a la red, lanzar aplicaciones preliminares (SKY se ejecuta entes que el resto por su demora en levantar), monitorear el copiado de un token, formatearlo según reglas predefinidas y ejecutar comando que levanta un servidor http que se comunica con PureCloud y utiliza el token obtenido previamente.

---

## 📄 Configuración (`config.txt`)

Para que la aplicación funcione, se necesita un archivo `config.txt` en la misma carpeta que el `.exe`. Este archivo setea todos los parámetros de ejecución.

```ini
# --- Rutas de Comandos ---
# Aplicación que se ejecuta en segundo plano ANTES de obtener el token.
# La unidad (ej. F:) debe estar montada o ser accesible.
# Esta unidad depende de un recurso de red y de que el usuario este en la red corporativa
path1=F:\CTI-PC\SkyTrfData.exe

# Aplicación que se ejecuta DESPUÉS de obtener y formatear el token. Esta es la app que desarrollo genesys y que recibe para cada llamada entrante el ANI y el DNI del cliente en la linea
path2=F:\CTI-PC\TQPComm.exe

# --- Configuración del Token ---
# Esto lo utiliza Cesar en la app desarrollada en VB como marcadores en el string que trae el access_token, el ANI y el DNI
# Texto que se añade al principio del token.
prefix=SKYGENESYSCLOUD#;

# Texto que se añade al final del token.
suffix=;#GENESYSCLOUDSKY

# --- Opciones de Monitoreo ---
# Tiempo máximo en segundos que la aplicación esperará por el token. Esta tarea depende de una accion del usuario que debe ingresar a la aplicacion CRM de PureCloud
monitor_timeout_seconds=600

# Longitud mínima que debe tener el token para ser considerado válido.
# Aquí deberia ir el tamaño exacto del access_token. La aplicación setea en el clipboard un string que utiliza como referencia. Cuando cambia por un string del tamaño del token considera que obtuvo dicho dato y lo almacena como tal para luego utilizarlo. Facilmente hackeable
min_token_length=50

----
Preparación del Entorno
Antes de compilar, tenés que asegurarte de tener todas las dependencias (módulos) de Go que el proyecto necesita.

Instalar las herramientas de Go:
Si no lo hiciste antes, instala la herramienta rsrc para manejar los recursos de Windows.

En un terminal ejecutar:
go install [github.com/akavel/rsrc@latest](https://github.com/akavel/rsrc@latest)

Sincronizar Módulos del Proyecto:
Abre una terminal en la carpeta del proyecto y ejecuta el siguiente comando. Este comando leerá el código, buscará todos los import y descargará automáticamente todo lo que falte.

En un terminal ejecutar:
go mod tidy

Compilación

Generar el archivo de recursos de Windows:
Este comando empaqueta el manifiesto para que la ventana se vea correctamente. Asegúrate de tener el archivo genesysSkyLauncher.exe.manifest en la carpeta.

En un terminal ejecutar:
rsrc -manifest genesysSkyLauncher.exe.manifest -o rsrc.syso

Compilar la aplicación:
Este es el comando final que crea el .exe. El flag -ldflags "-H windowsgui" evita que se abra una ventana de consola negra al ejecutar el programa.

En un terminal ejecutar:
go build -ldflags "-H windowsgui" -o genesysSkyLauncher.exe .


Cómo Ejecutar los Tests

Ejecutar todos los tests unitarios:
Este comando correrá las pruebas que no dependen de la red ni de tu intervención.

En un terminal ejecutar:
go test -v

Ejecutar el test de integración de red:
Para este test, necesitás estar conectado a la VPN para que pueda acceder a la URL de Genesys.

en una terminal ejecutar:
go test -v -run TestIntegration_URLAccessibility

Ejecutar el test de integración del portapapeles:
Este test es interactivo. Te pedirá que copies un texto manualmente para verificar que el flujo de formateo funciona.

En una terminal ejecutar:
go test -v -run TestIntegration_ClipboardWorkflow


Flujo de Uso
Una vez compilado, el flujo de la aplicación es el siguiente:

Se ejecuta genesysSkyLauncher.exe.

La aplicación verifica que los archivos en path1 y path2 existan.

Verifica que se puede acceder a la URL de Genesys (chequeo de VPN).

Se lanza la aplicación de path1 en segundo plano.

Se pone la bandera "ESTADOINICIAL" en el portapapeles.

Se abre el navegador Edge para que te autentiques y copies el token.

La aplicación detecta que el portapapeles cambió y que el contenido tiene la longitud mínima requerida.

Se formatea el token con el prefijo y sufijo y se actualiza el portapapeles.

Se ejecuta la aplicación de path2.

La aplicación se cierra.