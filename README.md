# Mundial 2010 — Backend

> API REST para gestionar documentales y series sobre el Mundial de Sudáfrica 2010.

🔗 **Repositorio del frontend:** [mundial2010-frontend](https://github.com/zoomjosue/mundial2010-frontend)


> La aplicación permite explorar, crear, editar y calificar documentales y series sobre el Mundial 2010, con paginación, búsqueda, filtros de ordenamiento y subida de imágenes.


## Cómo correr el proyecto localmente

### Requisitos previos

- Docker Desktop instalado y corriendo

### Pasos (todo en un solo comando)

El backend incluye un script de arranque que **automáticamente clona el frontend** y levanta todos los servicios (PostgreSQL, backend y frontend) con Docker Compose.

**En Linux / macOS:**

```bash
git clone https://github.com/zoomjosue/mundial2010-backend.git
cd mundial2010-backend
chmod +x start.sh
./start.sh
```

**En Windows (PowerShell):**

```powershell
git clone https://github.com/zoomjosue/mundial2010-backend.git
cd mundial2010-backend
.\start.ps1
```

Eso es todo. El script clona el frontend automáticamente si no existe, y Docker Compose construye y levanta los tres servicios.

### URLs una vez corriendo

 Servicio    URL                              |

 Frontend    http://localhost:3000            
 Backend     http://localhost:8080            
 Swagger UI  http://localhost:8080/swagger    
 Health      http://localhost:8080/health     

##  Challenges implementados

### API y Backend

 Challenge: 
    Spec de OpenAPI/Swagger escrita y precisa (YAML) 
    Swagger UI corriendo y siendo servido desde el backend 
    Códigos HTTP correctos (201 al crear, 204 al eliminar, 404, 400)
    Validación server-side con respuestas de error en JSON descriptivas
    Paginación en `GET /series` con `?page=` y `?limit=` 
    Búsqueda por nombre con `?q=` 
    Ordenamiento con `?sort=` y `?order=asc\|desc` 

### Challenges opcionales

 Challenge

    Exportar lista a CSV — generado desde JavaScript sin librerías.
    Sistema de rating — tabla propia, endpoints REST, visible en el cliente.
    Permite subir imágenes.

---

## Endpoints principales

```
GET    /series                    → Listar (paginación, búsqueda, ordenamiento)
POST   /series                    → Crear serie
GET    /series/:id                → Obtener por ID
PUT    /series/:id                → Actualizar
DELETE /series/:id                → Eliminar (204)
POST   /series/:id/image          → Subir imagen
GET    /series/:id/rating         → Ver calificaciones
POST   /series/:id/rating         → Agregar calificación
DELETE /series/:id/rating/:rid    → Eliminar calificación (204)
GET    /swagger                   → Swagger UI
GET    /health                    → Health check
```

---

## Reflexión técnica

**¿Usaría Go de nuevo para este tipo de proyecto?**

Sí, sin dudarlo. Usar Go con `net/http` puro y sin frameworks como Gin fue interesante, por que al principio parece más complicado o que se necesitan más código que Express o FastAPI, pero el resultado es increible por que no se necesitan dependencias externas en runtime y con tiempos de respuesta muy bajos. 

La parte más dificil fue el routing manual para paths como /series/:id/rating/:ratingId, por que como Go no tiene parámetros de ruta nativos, hubo que parsear strings.Split manualmente.

PostgreSQL facilito las migraciones automáticas al arrancar, el seed de datos inicial y las restricciones, hicieron que la base de datos fuera confiable desde el día uno.

Docker Compose con el script de arranque fue un deatalle que queria añadir, para que se pudiera levantar simplemente con un comando y facilitarnos la vida al, cualquier persona puede clonar el repo y tener todo corriendo en minutos sin instalar Go ni PostgreSQL.

**¿Lo repetiríamos?** 
Sí, aunque para un proyecto más grande probablemente agregaría un router o se cambiaria para manejar los path de una mejor forma.