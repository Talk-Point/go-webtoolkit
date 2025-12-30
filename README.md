# Go-WebToolkit

<div align="center">

![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?style=for-the-badge&logo=go)
[![License](https://img.shields.io/github/license/Talk-Point/go-webtoolkit?style=for-the-badge)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/Talk-Point/go-webtoolkit?style=for-the-badge)](https://goreportcard.com/report/github.com/Talk-Point/go-webtoolkit)

Ein umfassendes und vielseitiges Go-Utility-Toolkit, speziell entwickelt für die Entwicklung robuster Web-APIs bei Talk-Point. Diese Bibliothek bietet eine Sammlung von bewährten Komponenten und Utilities, die die Go-Entwicklung beschleunigen und standardisieren.

</div>

## 🚀 Installation

```bash
go get github.com/Talk-Point/go-webtoolkit@latest
```

## 📚 Features Übersicht

Das Go-WebToolkit bietet eine umfassende Sammlung von Modulen für moderne Go-Webentwicklung:

### 🔐 JWT-Token Management
Sichere JWT-Token-Erstellung und -Verarbeitung mit flexiblen Ablaufzeiten.
- Token-Generierung mit benutzerdefinierten Claims
- Token-Parsing und -Validierung
- Erweiterte Konfiguration für Ablaufzeiten

**[→ Detaillierte JWT Dokumentation](docs/jwt.md)**

### 🛡️ Error Handling System
Strukturiertes Error-Management mit HTTP-Status-Code-Mapping und Validierung.
- Typisierte Error-Strukturen für verschiedene HTTP-Status-Codes
- Automatische Validierungsfehler-Behandlung
- Konsistente Error-Response-Formate

**[→ Detaillierte Error Handling Dokumentation](docs/errors.md)**

### 🔍 Advanced Query System
Flexibles System für URL-basierte Queries und Filterung.
- URL-Parameter zu Filter-Objekten
- Unterstützung für komplexe Operatoren (eq, gt, lt, contains, etc.)
- Firestore-Integration mit automatischer Operator-Konvertierung
- Sortierung und Paginierung

**[→ Detaillierte Query System Dokumentation](docs/query.md)**

### 🗄️ Generic Repository Pattern
Type-safe Repository-Pattern für Firestore mit Generics.
- CRUD-Operationen mit automatischem ID-Management
- Paginierte Abfragen mit Next/Previous-Unterstützung
- Transaktionale Erstellung mit Duplikatsprüfung
- Flexible Query-Integration

**[→ Detaillierte Repository Dokumentation](docs/repository.md)**

### 🌐 URL Builder Utilities
Typsichere URL-Konstruktion mit Parameter-Handling.
- Fluent API für URL-Erstellung
- Automatische Parameter-Enkodierung
- Platzhalter-Ersetzung für dynamische URLs
- Type-Safe Parameter-Handling

**[→ Detaillierte URL Builder Dokumentation](docs/urllib.md)**

### 🔎 Advanced Search Engine
Leistungsstarke Such- und Filter-Engine für komplexe Abfragen.
- Textbasierte Such-Queries mit SQL-Operatoren
- Unterstützung für verschiedene Datentypen
- Spalten-Aliasing und Mapping
- Automatische SQL-Generierung

**[→ Detaillierte Advanced Search Dokumentation](docs/advanced-search.md)**

### ⚙️ Task Manager (v2 & v3)
Google Cloud Tasks Integration für Hintergrund-Jobs.

#### TaskManager v2 (Legacy)
- Einfache Task-Erstellung und -Verwaltung
- Queue-Management (Erstellen, Pausieren, Fortsetzen)
- Test-Mode für Entwicklung

#### TaskManager v3 (Empfohlen)
- Verbesserte API mit funktionalen Optionen
- Erweiterte Queue-Konfiguration
- Delay-Unterstützung für zeitgesteuerte Tasks
- Bessere Error-Behandlung und Logging

**[→ Detaillierte TaskManager v2 Dokumentation](docs/taskmanager-v2.md)**  
**[→ Detaillierte TaskManager v3 Dokumentation](docs/taskmanager-v3.md)**

## 🏗️ Architektur

Das Toolkit folgt modernen Go-Patterns und Best Practices:

- **Generics**: Typsicherheit wo möglich (Go 1.18+)
- **Interface-basiert**: Testbare und erweiterbare Designs
- **Context-aware**: Ordnungsgemäße Context-Verbreitung
- **Error-first**: Explizite Error-Behandlung
- **Dependency Injection**: Flexible Konfiguration

## 🚀 Quick Start

```go
package main

import (
    "context"
    "log"
    
    "github.com/Talk-Point/go-webtoolkit/pkg/v2/jwt"
    "github.com/Talk-Point/go-webtoolkit/pkg/v2/query"
    "github.com/Talk-Point/go-webtoolkit/pkg/v2/urllib"
)

func main() {
    // JWT Token erstellen
    token, err := jwt.NewJwtToken("secret", map[string]interface{}{
        "user_id": "123",
        "role":    "admin",
    })
    if err != nil {
        log.Fatal(err)
    }
    
    // URL mit Parametern erstellen
    url := urllib.Url("https://api.example.com/users/:id", map[string]interface{}{
        "id": "123",
    }).AddParam("active", true).String()
    
    // Query-Parameter parsen
    filters, _ := query.NewFilterFromUrlString("name=John&age__gt=18&active=true")
    
    log.Printf("Token: %s", token)
    log.Printf("URL: %s", url)
    log.Printf("Filters: %+v", filters)
}
```

## 🛠️ Verwendung in Projekten

### 1. Basis-Setup

```go
// go.mod
module your-project

go 1.25

require github.com/Talk-Point/go-webtoolkit v1.x.x
```

### 2. Import-Pattern

```go
import (
    // Basis-Module
    "github.com/Talk-Point/go-webtoolkit/pkg/v2/errors"
    "github.com/Talk-Point/go-webtoolkit/pkg/v2/jwt"
    
    // Repository und Query
    "github.com/Talk-Point/go-webtoolkit/pkg/v2/repository"
    "github.com/Talk-Point/go-webtoolkit/pkg/v2/query"
    
    // Utilities
    "github.com/Talk-Point/go-webtoolkit/pkg/v2/urllib"
    "github.com/Talk-Point/go-webtoolkit/pkg/v2/advanced_search"
    
    // TaskManager (neueste Version)
    "github.com/Talk-Point/go-webtoolkit/pkg/taskmanager/v3/taskmanager"
)
```

## 🔧 Konfiguration

Viele Module unterstützen flexible Konfiguration über Options-Pattern:

```go
// TaskManager mit Konfiguration
tm, err := taskmanager.NewTaskManager(&taskmanager.TaskManagerOptions{
    Project:  "my-project",
    Location: "europe-west1",
    AuthKey:  "api-key",
    BaseUrl:  "https://api.example.com",
    QueueOpts: &taskmanager.QueueOptions{
        MaxDispatchesPerSecond:  10,
        MaxConcurrentDispatches: 5,
        MaxAttempts:            3,
    },
})
```

## 🧪 Testing

Das Toolkit ist darauf ausgelegt, testbar zu sein:

```go
// Test-Varianten für verschiedene Module
func TestExample(t *testing.T) {
    // TaskManager Test-Mode
    os.Setenv("TEST_MODE", "true")
    
    // Repository mit Test-Firestore-Client
    testRepo := repository.NewFirebaseRepository[MyEntity, MyEntity](testClient, "TestEntity")
    
    // Weitere Test-Utilities...
}
```

## 📋 Anforderungen

- **Go Version**: 1.25+ (für Generics-Support)
- **Google Cloud**: Firestore und Cloud Tasks (für entsprechende Module)
- **Dependencies**: Automatisch über `go mod` verwaltet

## 🤝 Beitrag

Dieses Toolkit wird aktiv bei Talk-Point entwickelt und gewartet. 

### Entwicklung

```bash
# Repository klonen
git clone https://github.com/Talk-Point/go-webtoolkit.git

# Dependencies installieren
go mod download

# Tests ausführen
go test ./...

# Build überprüfen
go build ./...
```

## 📄 Lizenz

[Lizenz-Details in LICENSE-Datei](LICENSE)

## 🆘 Support

- **Dokumentation**: Vollständige Dokumentation in `/docs`
- **Issues**: GitHub Issues für Bug-Reports und Feature-Requests
- **Internal**: Talk-Point interne Kanäle für Support

---

**Entwickelt mit ❤️ bei Talk-Point für moderne Go-Webentwicklung**