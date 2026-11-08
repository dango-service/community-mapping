# AGENTS.md — community-mapping

Fuente única de reglas para agentes de IA que trabajen en este repositorio.
Lee este fichero completo antes de tocar cualquier código.

---

## 1. Propósito del proyecto

**community-mapping** es un dataset mantenido por la comunidad y un conjunto
de herramientas en Go para mapeo de identidades entre bases de datos de anime y manga
(MyAnimeList, AniList, Kitsu) y límites de capítulos por tomo (volúmenes).

---

## 2. Reglas de estilo Go

- Errores siempre envueltos con `%w` (`fmt.Errorf("op: %w", err)`).
- `context.Context` como primer parámetro en toda función que haga I/O.
- Interfaces pequeñas, definidas donde se consumen.
- Tests table-driven con `t.Run`.
- `CGO_ENABLED=0` en producción.

---

## 3. Convención de commits

Formato: `<type>(<scope>): <description>`
- Tipos: `feat fix docs style refactor perf test build ci chore revert`
- Descripción en inglés, imperativo, minúscula, sin punto, ≤ 72 chars.
- Nunca commitear secretos, tokens ni datos personales.

---

## 4. Cosas PROHIBIDAS

- El uso de emojis está TERMINANTEMENTE PROHIBIDO en toda la base de código, archivos HTML, CSS, JavaScript, Go, commits, logs, documentación y respuestas al usuario. Utilizar exclusivamente texto limpio o iconos vectoriales SVG.
- Incluir scrapers o herramientas de extracción no públicas en el repositorio.
- Añadir campos de sinónimos manuales en `master/*.yaml` (se gestionan vía snapshot `data/titles.json.gz`).
- Añadir dependencias de servicios de piratería o lectura no oficiales (MangaDex, etc.).

---

## 5. Checklist "antes de dar una tarea por terminada"

- [ ] El código compila sin errores (`go build ./...`).
- [ ] Los tests pasan (`go test ./...`).
- [ ] La validación pasa sin errores (`go run ./cmd/mapping validate`).
- [ ] Cero emojis en la base de código, UI, commits y documentación.
- [ ] No hay secretos ni datos personales en el commit.
