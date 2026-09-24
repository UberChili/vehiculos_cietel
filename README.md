# Vehículos CIETEL

Aplicación web para llevar el control de la flotilla de vehículos de CIETEL:
qué vehículos hay, quién tiene cada uno, dónde están, y el historial de
servicios y reparaciones de cada uno (con sus costos).

Se usa desde el navegador, en computadora o en celular. Todo lo que se
captura aquí es la base para, más adelante, sacar reportes de cuánto cuesta
la flotilla, qué vehículos conviene reemplazar y qué tanto tiempo pasan
descompuestos. Por eso importa capturar bien cada campo: **un dato que no se
captura hoy no se puede analizar mañana.**

## Contenido

- [Uso de la aplicación](#uso-de-la-aplicación)
  - [Página principal](#página-principal)
  - [Vehículos](#vehículos)
  - [Técnicos](#técnicos)
  - [Registros: servicios y reparaciones](#registros-servicios-y-reparaciones)
  - [Preguntas frecuentes sobre los registros](#preguntas-frecuentes-sobre-los-registros)
  - [Capturar el historial de un vehículo](#capturar-el-historial-de-un-vehículo)
- [Para desarrolladores](#para-desarrolladores)

---

# Uso de la aplicación

Al entrar, el navegador pide un usuario y una contraseña. Pídeselos a quien
administra el sistema.

## Página principal

Muestra la **lista de vehículos**, ordenada por ubicación. En celular se ve
como tarjetas y en computadora como tabla. Toca cualquier vehículo para ver
su página.

- **Buscar:** el cuadro de búsqueda filtra por placas, marca, modelo, técnico
  o ubicación. No importan mayúsculas ni acentos ("nunez" encuentra "Núñez").
- **👤 Técnicos / 🚙 Vehículos:** cambia entre la lista de vehículos y la lista
  de técnicos (con el vehículo que tiene cada uno).
- **+ Agregar:** agrega un vehículo o un técnico, según la lista que estés
  viendo.

### Las etiquetas "Servicio" y "Bandas"

Cada vehículo muestra dos etiquetas que dicen, de un vistazo, si le toca
servicio o cambio de bandas:

| Etiqueta | Significado |
|---|---|
| 🟢 **Al día** | Todavía no le toca. |
| 🟡 **Próximo** | Le toca en 15 días o menos. |
| 🔴 **Vencido** | Ya le tocaba. |
| ⚪ **Sin registro** | No hay ningún registro para calcularlo. |

Cómo se calculan:
- **Servicio:** a partir del registro de tipo *Servicio* más reciente. Toca
  cada **8 meses**.
- **Bandas:** a partir del registro más reciente (servicio o reparación) que
  mencione **"banda" o "bandas"** en su descripción. Toca cada **6 meses**.

> Los intervalos de 8 y 6 meses son **provisionales**: todavía falta
> acordarlos. Si cambian, se ajustan en el sistema y las etiquetas se
> recalculan solas.

## Vehículos

### Agregar un vehículo

| Campo | Qué poner |
|---|---|
| **Placas** \* | Las placas tal cual (máximo 9 caracteres). Se guardan en mayúsculas. No puede haber dos vehículos con las mismas placas; para eso no cuentan guiones ni espacios ("ABC-123" y "ABC 123" se consideran iguales). |
| **Marca** \* | Se elige de la lista. |
| **Modelo** \* | Se elige de la lista. Solo aparecen los modelos de la marca elegida. |
| **Año** \* | Año del modelo, de 2010 al año siguiente al actual. |
| **Asignado a** | El técnico que usa el vehículo, o *Sin asignar*. Cada técnico puede tener **un solo** vehículo. |
| **Ubicación actual** | Se elige de la lista (ciudad y estado). Opcional. |
| **Foto** | Opcional. En celular se puede tomar en el momento. JPG, PNG, WEBP o GIF, hasta 20 MB. |

\* Campo obligatorio.

Si falta una marca, un modelo o una ciudad en las listas, pídele a quien
administra el sistema que la agregue. Las listas son fijas a propósito, para
que no aparezcan "Chevrolet", "chevrolet" y "Cevrolet" como tres marcas
distintas.

### Reasignar un vehículo

Como un técnico no puede tener dos vehículos, para pasarle un vehículo a
alguien que ya tiene otro, primero edita ese otro vehículo y déjalo
*Sin asignar*.

### Página del vehículo

Muestra sus datos, la foto, la fecha del último servicio, de la última
reparación y del último cambio de bandas, y la lista de todos sus registros,
del más reciente al más antiguo. Desde aquí se **edita** el vehículo y se
**agregan registros**.

### Editar o eliminar un vehículo

En **Editar** se cambia cualquier dato. La foto se puede reemplazar o quitar.

**Eliminar vehículo** (al final de la página de edición) borra el vehículo,
**todos sus registros y su foto**. No se puede deshacer. Úsalo solo para
vehículos dados de alta por error: si un vehículo se vendió o dejó de usarse,
eliminarlo borra su historial, que sirve para comparar costos.

## Técnicos

### Agregar o editar un técnico

Solo **nombre** y **apellido**. No importa cómo se escriban las mayúsculas:
se guardan todos igual y se muestran con mayúscula inicial ("juan PÉREZ" se
muestra como "Juan Pérez").

El vehículo de cada técnico **se asigna desde el vehículo** (campo
*Asignado a*), no desde el técnico.

### Dar de baja, reactivar o eliminar

Al final de la página de edición de cada técnico:

- **Dar de baja:** para quien ya no trabaja con nosotros. Su vehículo queda
  *Sin asignar* y ya no aparece en la lista para asignarle vehículos, pero
  **su historial se conserva**: los servicios y reparaciones que se hicieron
  mientras tenía un vehículo siguen a su nombre. En la lista de técnicos
  aparece atenuado, con "(dado de baja)".
- **Reactivar técnico:** deshace la baja.
- **Eliminar técnico:** solo para técnicos dados de alta por error. No se
  puede eliminar a quien ya tiene registros a su nombre; a esos se les da de
  baja.

## Registros: servicios y reparaciones

Un registro es cada vez que a un vehículo se le hace algo. Se agrega desde la
página del vehículo con **+ Agregar**. Al tocar un registro se ve completo, y
desde ahí se puede **editar** o **eliminar**.

### ¿Servicio o reparación?

- **Servicio:** mantenimiento **planeado**, para que el vehículo no se
  descomponga: servicio general, afinación, cambio de aceite y filtros,
  cambio de bandas preventivo.
- **Reparación:** arreglar algo que **ya falló o se dañó**: frenos que ya no
  sirven, clutch, una banda que se rompió, un golpe, una llanta ponchada.

Regla práctica: si se hizo porque "ya tocaba", es servicio; si se hizo porque
"algo pasó", es reparación.

### Los campos, uno por uno

**Tipo** \* — Servicio o Reparación (ver arriba).

**Fecha** \* — El día en que se **terminó** el trabajo (cuando el vehículo
quedó listo). Formato dd-mm-aaaa. Basta con escribir los números: los guiones
se ponen solos. Si el trabajo tomó varios días, los días que el vehículo
estuvo parado van en *Días sin poder usarse*.

**Causa de la reparación** \* (solo en reparaciones) — Por qué falló:

| Causa | Cuándo elegirla | Ejemplos |
|---|---|---|
| **Desgaste normal** | La pieza se acabó por uso y edad, como a cualquier vehículo. | Balatas gastadas después de mucho tiempo, batería vieja, amortiguadores cansados. |
| **Mal uso o descuido** | Falló por cómo se manejó o se cuidó el vehículo. | Clutch quemado por manejarlo "arrastrando", motor dañado por no revisar el aceite, no avisar de una falla a tiempo y que empeorara. |
| **Accidente o golpe** | Un choque, golpe o incidente. | Golpes en carrocería, llanta rota por un bache fuerte, espejo roto. |
| **No se sabe** | No está claro. | Cuando de verdad no hay forma de saberlo. |

Este dato es el que permite distinguir un vehículo viejo que se está acabando
de uno al que no se cuida, así que conviene que lo elija **quien autoriza o
paga la reparación**, no el técnico que usa el vehículo.

**Descripción** \* — Qué se hizo, con el detalle útil: piezas cambiadas, qué
taller o mecánico lo hizo, y cualquier cosa fuera de lo común. Si se
cambiaron bandas, **escribe la palabra "bandas"** (por ejemplo, "cambio de
bandas"): así el sistema sabe cuándo fue el último cambio.

**Costo** — Lo que se **pagó en total**, en pesos: refacciones + mano de obra,
con IVA. Sin signo de pesos ni comas (por ejemplo `1850.50`).
- Si no se sabe cuánto costó, **déjalo vacío**.
- Si no costó nada (por ejemplo, una garantía), pon **0**.

Vacío y 0 significan cosas distintas: vacío es "no sabemos", 0 es "fue
gratis".

**Kilometraje** — El número que marca el odómetro **cuando el vehículo llegó
al taller** (o cuando se descompuso), en kilómetros enteros, sin comas ni
decimales.
- Si el odómetro **no funciona**, marca **"El odómetro no funciona"** y deja
  el número vacío. No lo estimes: un número inventado es peor que no tener
  el dato.
- Si simplemente no se revisó, déjalo vacío sin marcar la casilla.

**Días sin poder usarse** — Cuántos **días hábiles** el vehículo no se pudo
usar para trabajar por este servicio o reparación.
- Si se hizo sin afectar el trabajo (por ejemplo, en la tarde después de la
  jornada), pon **0**.
- Medio día o más cuenta como **1** día.
- Si no se sabe, déjalo vacío. Igual que con el costo, vacío es "no
  sabemos" y 0 es "no se perdió ningún día".

**Técnico a cargo en ese momento** — **No se captura**: el sistema guarda
solo quién tenía asignado el vehículo al momento de guardar el registro, y se
muestra en la página del registro. Así, aunque el vehículo cambie de manos
después, el registro sigue a nombre de quien lo tenía. Solo se guarda en
registros con fecha de los **últimos 30 días** (ver
[por qué](#por-qué-un-registro-no-tiene-técnico-a-cargo)).

## Preguntas frecuentes sobre los registros

### ¿El kilometraje es el de cuando llegó al taller o el de cuando salió?

**El de cuando llegó** (o el de cuando se descompuso, si llegó en grúa). En la
práctica la diferencia es de unos pocos kilómetros de prueba, así que lo
importante es **hacerlo siempre igual**. Se eligió la llegada porque lo que
interesa es con cuántos kilómetros falló o le tocó el servicio, y porque es
el momento en que alguien está viendo el tablero.

### ¿Por qué me pide confirmar el kilometraje antes de guardar?

Porque el número no cuadra con los demás registros de ese vehículo: es menor
que uno de una fecha anterior, o mayor que uno de una fecha posterior. Casi
siempre es un error de dedo; revísalo. Si el número es correcto (por
ejemplo, porque le cambiaron el odómetro), confirma y explícalo en la
descripción. Es solo un aviso: no impide guardar.

### ¿Y si el odómetro marca millas?

Captura siempre **kilómetros**: multiplica las millas por 1.609 y redondea.
Anótalo en la descripción ("odómetro en millas").

### ¿Y si el odómetro funciona a veces?

Si en ese momento marca un número creíble, captúralo. Si no, marca "El
odómetro no funciona". Cuando se arregle o se cambie el odómetro, anótalo en
la descripción del registro de esa reparación.

### Si en la misma visita al taller se hizo un servicio y una reparación, ¿qué capturo?

**Dos registros**: uno de tipo Servicio y otro de tipo Reparación, cada uno
con su costo (sepáralo si la nota lo permite). Los días sin poder usarse
ponlos **solo en uno** de los dos, para no contarlos doble.

### Le prestamos otro vehículo al técnico mientras arreglaban el suyo, ¿cuento los días?

Sí: el campo cuenta los días que **este vehículo** no se pudo usar. Anota en
la descripción que el técnico tuvo otro vehículo mientras tanto.

### Se cambiaron las bandas pero la etiqueta "Bandas" no cambió

La descripción tiene que incluir la palabra **"banda" o "bandas"**. Edita el
registro y agrégala (por ejemplo, "Servicio general con cambio de bandas").

Ojo: cualquier descripción con esa palabra cuenta, incluso "no se cambiaron
las bandas". Evita mencionarlas si no se cambiaron.

### ¿Por qué un registro no tiene "Técnico a cargo"?

Por alguna de estas razones:
- El vehículo estaba *Sin asignar* cuando se guardó el registro.
- La fecha del registro es de hace **más de 30 días**. Es historial capturado
  tarde, y quien tiene el vehículo hoy no necesariamente lo tenía entonces.
  Para no culpar a la persona equivocada, el sistema prefiere dejarlo vacío.
- El registro se capturó antes de que existiera este campo.

### ¿Quién tiene que elegir la causa de una reparación?

Idealmente **quien autoriza o paga la reparación**, no el técnico que usa el
vehículo. Si no está claro, "No se sabe" es mejor que adivinar.

## Capturar el historial de un vehículo

Al dar de alta un vehículo que ya tiene historia, captura sus servicios y
reparaciones anteriores como registros normales:

- **El orden no importa.** Se pueden capturar del más viejo al más nuevo o al
  revés; la lista siempre se ordena por fecha.
- **Si no se sabe el día exacto**, usa el día 1 de ese mes y escribe "fecha
  aproximada" en la descripción.
- **Captura el kilometraje** que aparezca en las notas o facturas, si lo hay.
  El aviso de kilometraje compara cada número con los registros anteriores y
  posteriores, así que también ayuda a detectar errores en el historial.
- Los registros de hace más de 30 días se guardan **sin técnico a cargo**, a
  propósito (ver arriba).

---

# Para desarrolladores

## Requisitos y cómo correrlo

- **Go**, en la versión indicada en `go.mod`.
- **Un compilador de C** (gcc), porque el driver de SQLite
  (`github.com/mattn/go-sqlite3`) usa cgo.

```sh
go run .
```

La aplicación queda en <http://127.0.0.1:8081>. Escucha **solo en
127.0.0.1**, a propósito: en el servidor, nginx es el único que le habla
(y es quien pide la contraseña). Si el puerto está ocupado, el programa
termina con el error en lugar de fallar en silencio.

Al arrancar crea, si no existen, el archivo `vehicles.db` con sus tablas y,
con la primera foto, la carpeta `uploads/`. Ambos están en `.gitignore`: los
datos no viven en el repositorio.

## Estructura

| Archivo | Qué contiene |
|---|---|
| `main.go` | Rutas, funciones de plantilla y arranque del servidor. |
| `handlers.go` | Un handler por página o acción. |
| `database.go` | Creación de tablas y consultas de vehículos y registros. |
| `technician.go` | Consultas de técnicos (alta, baja, eliminación). |
| `vehicle.go` | Tipos `Vehicle` y `Record`, y las **listas fijas**. |
| `helpers.go` | Validación, fotos, fechas, costos, etiquetas de estado. |
| `templates/` | Una plantilla HTML por página, cada una con su propio CSS. |
| `notes.org` | Notas de trabajo y respaldo del esquema de la base de datos. |

## Listas fijas y constantes

Lo que más probablemente haya que ajustar:

| Qué | Dónde |
|---|---|
| Marcas y sus modelos (`ModelsByMaker`) | `vehicle.go` |
| Ciudades por estado (`CitiesByState`) | `vehicle.go` |
| Causas de reparación (`RepairCauses`) | `vehicle.go` |
| Días para el técnico a cargo (`TechnicianSnapshotMaxDays`, 30) | `vehicle.go` |
| Intervalo de servicio y de bandas (`ServiceIntervalMonths`, `BandasIntervalMonths`) | `helpers.go` |
| Anticipación de "Próximo" (`DueSoonDays`, 15) | `helpers.go` |

## Base de datos

SQLite, un solo archivo: `vehicles.db`. Tablas `technicians`, `vehicles` y
`records`; el esquema completo está en `InitDBandCreateOrOpenTables`
(`database.go`), con una copia de respaldo en `notes.org`.

Convenciones:
- **Fechas** como texto ISO (`AAAA-MM-DD`), para que ordenar y comparar
  fechas funcione como texto. Se muestran como dd-mm-aaaa.
- **Costos** en **centavos** (`INTEGER`): $1,850.50 se guarda como `185050`,
  para que las sumas sean exactas.
- **Nombres de técnicos** en minúsculas; se capitalizan al mostrarlos.
- **Datos opcionales desconocidos** como `NULL`, nunca como 0 (kilometraje,
  días sin poder usarse, costo).

### Cambios al esquema

`InitDBandCreateOrOpenTables` usa `CREATE TABLE IF NOT EXISTS`: crea las
tablas en una base de datos nueva, pero **nunca modifica una tabla que ya
existe**. Por eso, cada cambio al esquema necesita dos cosas:

1. Cambiar el `CREATE TABLE` en `database.go`, para las bases de datos nuevas.
2. Aplicar el cambio a mano en cada base de datos existente (la local y la del
   servidor): primero en una copia, y con un respaldo antes de tocar la real
   (`sqlite3 vehicles.db ".backup respaldo.db"`; nunca `cp` de una base de
   datos en uso).

Agregar una columna es sencillo (`ALTER TABLE ... ADD COLUMN ...`). Cambiar el
tipo de una columna requiere reconstruir la tabla.

## Compilar para el servidor

```sh
go build -o vehiculos_cietel .
```

El binario se llama `vehiculos_cietel` porque así lo espera el servicio de
systemd del servidor. El procedimiento de despliegue está documentado en las
notas de despliegue (fuera de este repositorio).

## Rutas

| Método | Ruta | Qué hace |
|---|---|---|
| GET | `/` | Lista de vehículos (y de técnicos, con `/#tecnicos`) |
| GET, POST | `/vehiculos/new` | Agregar vehículo |
| GET | `/vehiculos/{id}` | Página del vehículo |
| GET, POST | `/vehiculos/{id}/editar` | Editar vehículo |
| POST | `/vehiculos/{id}/eliminar` | Eliminar vehículo (con sus registros) |
| GET, POST | `/vehiculos/{id}/nuevo-registro` | Agregar registro |
| GET | `/vehiculos/{id}/registro/{record_id}` | Ver registro |
| GET, POST | `/vehiculos/{id}/registro/{record_id}/editar` | Editar registro |
| POST | `/vehiculos/{id}/registro/{record_id}/eliminar` | Eliminar registro |
| GET, POST | `/tecnicos/nuevo` | Agregar técnico |
| GET, POST | `/tecnicos/{id}/editar` | Editar técnico |
| POST | `/tecnicos/{id}/baja` | Dar de baja |
| POST | `/tecnicos/{id}/reactivar` | Reactivar |
| POST | `/tecnicos/{id}/eliminar` | Eliminar (solo si no tiene registros) |
| GET | `/uploads/*` | Fotos de los vehículos |
