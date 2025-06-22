# JWT Token Management

Das JWT-Modul bietet eine sichere und flexible Lösung für die Erstellung und Verarbeitung von JSON Web Tokens in Go-Anwendungen.

## Features

- ✅ Sichere Token-Generierung mit HMAC-SHA256
- ✅ Flexible Ablaufzeiten (Standard: 24 Stunden)
- ✅ Benutzerdefinierte Claims
- ✅ Token-Parsing und -Validierung
- ✅ Einfache API mit Standardwerten

## Installation

```go
import "github.com/Talk-Point/go-webtoolkit/pkg/v2/jwt"
```

## Grundlegende Verwendung

### Token erstellen

```go
// Standard-Token (24 Stunden gültig)
token, err := jwt.NewJwtToken("your-secret-key", map[string]interface{}{
    "user_id": "123",
    "role":    "admin",
    "email":   "user@example.com",
})
if err != nil {
    log.Fatal(err)
}
```

### Token mit benutzerdefinierter Gültigkeit

```go
// Token mit 2 Stunden Gültigkeit
token, err := jwt.NewJwtTokenExtended("your-secret-key", map[string]interface{}{
    "user_id": "123",
    "role":    "admin",
}, 2*time.Hour)
if err != nil {
    log.Fatal(err)
}
```

### Token parsen und validieren

```go
// Claims extrahieren
claims := []string{"user_id", "role", "email"}
values, err := jwt.ParseJwtToken("your-secret-key", tokenString, claims)
if err != nil {
    log.Printf("Token ungültig: %v", err)
    return
}

// Werte zugreifen (in der Reihenfolge der claims)
userID := values[0]  // user_id
role := values[1]    // role
email := values[2]   // email
```

## API Referenz

### NewJwtToken

```go
func NewJwtToken(secret string, data map[string]interface{}) (string, error)
```

Erstellt einen neuen JWT-Token mit Standard-Gültigkeit von 24 Stunden.

**Parameter:**
- `secret`: Geheimer Schlüssel für die Token-Signierung
- `data`: Map mit benutzerdefinierten Claims

**Rückgabe:**
- `string`: Signierter JWT-Token
- `error`: Fehler bei der Token-Erstellung

### NewJwtTokenExtended

```go
func NewJwtTokenExtended(secret string, data map[string]interface{}, valid time.Duration) (string, error)
```

Erstellt einen JWT-Token mit benutzerdefinierter Gültigkeit.

**Parameter:**
- `secret`: Geheimer Schlüssel für die Token-Signierung
- `data`: Map mit benutzerdefinierten Claims
- `valid`: Gültigkeitsdauer des Tokens

**Rückgabe:**
- `string`: Signierter JWT-Token
- `error`: Fehler bei der Token-Erstellung

### ParseJwtToken

```go
func ParseJwtToken(secret string, token string, extract []string) ([]string, error)
```

Parst und validiert einen JWT-Token und extrahiert spezifische Claims.

**Parameter:**
- `secret`: Geheimer Schlüssel für die Token-Validierung
- `token`: JWT-Token-String
- `extract`: Liste der zu extrahierenden Claim-Namen

**Rückgabe:**
- `[]string`: Werte der extrahierten Claims (in der Reihenfolge der `extract`-Liste)
- `error`: Fehler bei der Token-Validierung oder beim Parsing

## Erweiterte Verwendung

### Token-Struktur

Jeder erstellte Token enthält automatisch folgende Standard-Claims:

```go
{
    "exp": 1234567890,  // Ablaufzeit (Unix Timestamp)
    "iat": 1234567890,  // Ausgestellt am (Unix Timestamp)
    // ... benutzerdefinierte Claims
}
```

### Middleware-Integration

```go
func JWTMiddleware(secret string) gin.HandlerFunc {
    return func(c *gin.Context) {
        tokenString := c.GetHeader("Authorization")
        if tokenString == "" {
            c.JSON(401, gin.H{"error": "Authorization header required"})
            c.Abort()
            return
        }
        
        // "Bearer " Präfix entfernen
        if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
            tokenString = tokenString[7:]
        }
        
        // Token validieren
        claims := []string{"user_id", "role"}
        values, err := jwt.ParseJwtToken(secret, tokenString, claims)
        if err != nil {
            c.JSON(401, gin.H{"error": "Invalid token"})
            c.Abort()
            return
        }
        
        // Claims im Context speichern
        c.Set("user_id", values[0])
        c.Set("role", values[1])
        c.Next()
    }
}
```

### Fehlerbehandlung

```go
token, err := jwt.ParseJwtToken(secret, tokenString, []string{"user_id"})
if err != nil {
    switch {
    case strings.Contains(err.Error(), "token is expired"):
        // Token abgelaufen
        return errors.New("Token ist abgelaufen")
    case strings.Contains(err.Error(), "invalid token"):
        // Ungültiger Token
        return errors.New("Ungültiger Token")
    default:
        // Anderer Fehler
        return fmt.Errorf("Token-Fehler: %v", err)
    }
}
```

## Best Practices

### Sicherheit

1. **Geheimschlüssel**: Verwenden Sie starke, zufällige Secrets (mindestens 32 Zeichen)
2. **Secret-Management**: Speichern Sie Secrets niemals im Code, nutzen Sie Umgebungsvariablen
3. **Token-Rotation**: Implementieren Sie regelmäßige Token-Erneuerung
4. **HTTPS**: Übertragen Sie Tokens nur über sichere Verbindungen

### Performance

1. **Caching**: Cachen Sie häufig verwendete Claims
2. **Parsing**: Parsen Sie nur benötigte Claims
3. **Validation**: Validieren Sie Tokens zentral in Middleware

### Fehlerbehandlung

```go
// Robuste Token-Validierung
func ValidateToken(secret, tokenString string) (*UserClaims, error) {
    claims := []string{"user_id", "role", "email"}
    values, err := jwt.ParseJwtToken(secret, tokenString, claims)
    if err != nil {
        return nil, fmt.Errorf("token validation failed: %w", err)
    }
    
    // Prüfen ob alle Claims vorhanden sind
    if values[0] == "" {
        return nil, errors.New("user_id claim missing")
    }
    
    return &UserClaims{
        UserID: values[0],
        Role:   values[1],
        Email:  values[2],
    }, nil
}
```

## Konfiguration

### Umgebungsvariablen

```bash
# JWT-Konfiguration
JWT_SECRET=your-super-secret-key-here
JWT_EXPIRY_HOURS=24
```

### Konfiguration im Code

```go
type JWTConfig struct {
    Secret string
    Expiry time.Duration
}

func NewJWTService(config JWTConfig) *JWTService {
    return &JWTService{
        secret: config.Secret,
        expiry: config.Expiry,
    }
}

func (j *JWTService) CreateToken(claims map[string]interface{}) (string, error) {
    return jwt.NewJwtTokenExtended(j.secret, claims, j.expiry)
}
```

## Häufige Anwendungsfälle

### Benutzer-Authentifizierung

```go
// Login-Handler
func LoginHandler(c *gin.Context) {
    // ... Benutzer validieren ...
    
    token, err := jwt.NewJwtToken(os.Getenv("JWT_SECRET"), map[string]interface{}{
        "user_id":   user.ID,
        "email":     user.Email,
        "role":      user.Role,
        "tenant_id": user.TenantID,
    })
    if err != nil {
        c.JSON(500, gin.H{"error": "Token creation failed"})
        return
    }
    
    c.JSON(200, gin.H{"token": token})
}
```

### API-Autorisierung

```go
// Protected Handler
func GetUserProfile(c *gin.Context) {
    userID := c.GetString("user_id")  // Aus JWT Middleware
    
    // Benutzerprofil laden...
    profile, err := userService.GetProfile(userID)
    if err != nil {
        c.JSON(404, gin.H{"error": "User not found"})
        return
    }
    
    c.JSON(200, profile)
}
```

## Fehlermeldungen

| Fehler | Bedeutung | Lösung |
|--------|-----------|---------|
| `token is expired` | Token ist abgelaufen | Neuen Token anfordern |
| `invalid token` | Token ist ungültig | Token-Format prüfen |
| `signature is invalid` | Signatur stimmt nicht überein | Secret prüfen |
| `invalid token claims` | Claims können nicht gelesen werden | Token-Struktur prüfen |

## Testing

```go
func TestJWTToken(t *testing.T) {
    secret := "test-secret"
    claims := map[string]interface{}{
        "user_id": "123",
        "role":    "admin",
    }
    
    // Token erstellen
    token, err := jwt.NewJwtToken(secret, claims)
    assert.NoError(t, err)
    assert.NotEmpty(t, token)
    
    // Token parsen
    values, err := jwt.ParseJwtToken(secret, token, []string{"user_id", "role"})
    assert.NoError(t, err)
    assert.Equal(t, "123", values[0])
    assert.Equal(t, "admin", values[1])
}
```