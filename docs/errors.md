# Error Handling System

Das Error-Handling-System bietet eine strukturierte und konsistente Methode zur Behandlung von Fehlern in Go-Web-APIs mit automatischem HTTP-Status-Code-Mapping.

## Features

- ✅ Typisierte Error-Strukturen für verschiedene HTTP-Status-Codes
- ✅ Automatische Validierungsfehler-Behandlung
- ✅ Konsistente Error-Response-Formate
- ✅ JSON-serialisierbare Error-Strukturen
- ✅ Integration mit `go-playground/validator`

## Installation

```go
import "github.com/Talk-Point/go-webtoolkit/pkg/v2/errors"
```

## Error-Typen

### Verfügbare Error-Typen

```go
// 404 Not Found
type ErrorNotFound struct {
    ErrorDetail
}

// 400 Bad Request
type ErrorBadRequest struct {
    ErrorDetail
}

// 409 Conflict
type ErrorAlreadyExists struct {
    ErrorDetail
}

// 401 Unauthorized
type ErrorUnauthorized struct {
    ErrorDetail
}

// 401 Unauthorized (spezifisch für inaktive User)
type ErrorUserNotActive struct {
    ErrorDetail
}

// 403 Forbidden (spezifisch für Salechannel)
type ErrorSalechannelNotAllowed struct {
    ErrorDetail
}
```

### Error Detail Struktur

```go
type ErrorDetail struct {
    Resource string `json:"resource"`
    Field    string `json:"field"`
    Value    string `json:"value"`
    Message  string `json:"message"`
}
```

## Grundlegende Verwendung

### Error erstellen und werfen

```go
// Not Found Error
notFoundErr := &errors.ErrorNotFound{
    ErrorDetail: errors.ErrorDetail{
        Resource: "User",
        Field:    "id",
        Value:    "123",
        Message:  "User with ID 123 not found",
    },
}

// Already Exists Error
existsErr := &errors.ErrorAlreadyExists{
    ErrorDetail: errors.ErrorDetail{
        Resource: "User",
        Field:    "email",
        Value:    "user@example.com",
        Message:  "User with email user@example.com already exists",
    },
}
```

### Error-Response generieren

```go
func HandleError(err error) (int, interface{}) {
    errorResponse, statusCode := errors.NewErrorResponse(err)
    return statusCode, errorResponse
}
```

## HTTP-Handler Integration

### Gin Framework

```go
func GetUser(c *gin.Context) {
    userID := c.Param("id")
    
    user, err := userService.GetByID(userID)
    if err != nil {
        // Automatische Error-Response-Generierung
        errorResponse, statusCode := errors.NewErrorResponse(err)
        c.JSON(statusCode, errorResponse)
        return
    }
    
    c.JSON(200, user)
}

func CreateUser(c *gin.Context) {
    var user User
    if err := c.ShouldBindJSON(&user); err != nil {
        // Validierungsfehler automatisch behandeln
        errorResponse, statusCode := errors.NewErrorResponse(err)
        c.JSON(statusCode, errorResponse)
        return
    }
    
    createdUser, err := userService.Create(user)
    if err != nil {
        errorResponse, statusCode := errors.NewErrorResponse(err)
        c.JSON(statusCode, errorResponse)
        return
    }
    
    c.JSON(201, createdUser)
}
```

### Standard HTTP

```go
func userHandler(w http.ResponseWriter, r *http.Request) {
    // ... User-Logic ...
    
    if err != nil {
        errorResponse, statusCode := errors.NewErrorResponse(err)
        
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(statusCode)
        json.NewEncoder(w).Encode(errorResponse)
        return
    }
    
    // Success response...
}
```

## Validierungsintegration

Das System integriert automatisch mit `go-playground/validator`:

### Struct Validation

```go
type CreateUserRequest struct {
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,min=8"`
    Name     string `json:"name" validate:"required,max=100"`
    Age      int    `json:"age" validate:"min=18,max=120"`
}

func CreateUserHandler(c *gin.Context) {
    var req CreateUserRequest
    
    if err := c.ShouldBindJSON(&req); err != nil {
        errorResponse, statusCode := errors.NewErrorResponse(err)
        c.JSON(statusCode, errorResponse)
        return
    }
    
    // Validation
    validate := validator.New()
    if err := validate.Struct(req); err != nil {
        // Automatische Validierungsfehler-Behandlung
        errorResponse, statusCode := errors.NewErrorResponse(err)
        c.JSON(statusCode, errorResponse)
        return
    }
    
    // ... User erstellen ...
}
```

### Unterstützte Validation Tags

Das System erkennt und behandelt folgende Validation-Tags automatisch:

| Tag | Beschreibung | Fehlermeldung |
|-----|--------------|---------------|
| `required` | Feld ist erforderlich | "is required." |
| `email` | Gültige E-Mail-Adresse | "is not a valid email." |
| `eqfield` | Feld muss anderem Feld entsprechen | "does not match the other field." |
| `min` | Minimale Länge/Wert | "must be at least X characters long." |
| `max` | Maximale Länge/Wert | "must be at most X characters long." |
| `uuid` | Gültige UUID | "is not a valid UUID." |
| `url` | Gültige URL | "is not a valid URL." |

## Service Layer Integration

### Repository Pattern

```go
type UserRepository struct {
    // ... 
}

func (r *UserRepository) GetByID(id string) (*User, error) {
    user, err := r.db.FindByID(id)
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, &errors.ErrorNotFound{
                ErrorDetail: errors.ErrorDetail{
                    Resource: "User",
                    Field:    "id",
                    Value:    id,
                    Message:  fmt.Sprintf("User with ID %s not found", id),
                },
            }
        }
        return nil, err
    }
    return user, nil
}

func (r *UserRepository) Create(user *User) error {
    // Prüfen ob User bereits existiert
    existing, _ := r.GetByEmail(user.Email)
    if existing != nil {
        return &errors.ErrorAlreadyExists{
            ErrorDetail: errors.ErrorDetail{
                Resource: "User",
                Field:    "email",
                Value:    user.Email,
                Message:  fmt.Sprintf("User with email %s already exists", user.Email),
            },
        }
    }
    
    return r.db.Create(user)
}
```

### Service Layer

```go
type UserService struct {
    repo UserRepository
}

func (s *UserService) Authenticate(email, password string) (*User, error) {
    user, err := s.repo.GetByEmail(email)
    if err != nil {
        return nil, err
    }
    
    if !user.Active {
        return nil, &errors.ErrorUserNotActive{
            ErrorDetail: errors.ErrorDetail{
                Resource: "User",
                Field:    "active",
                Value:    "false",
                Message:  "User account is not active",
            },
        }
    }
    
    if !s.validatePassword(user.Password, password) {
        return nil, &errors.ErrorUnauthorized{
            ErrorDetail: errors.ErrorDetail{
                Resource: "Authentication",
                Field:    "credentials",
                Value:    "",
                Message:  "Invalid credentials",
            },
        }
    }
    
    return user, nil
}
```

## Error Response Format

### Single Error

```json
{
    "message": "Resource not found",
    "errors": [
        {
            "resource": "User",
            "field": "id",
            "value": "123",
            "message": "User with id 123 not found"
        }
    ]
}
```

### Validation Errors

```json
{
    "message": "Validation failed",
    "errors": [
        {
            "resource": "Validation",
            "field": "Email",
            "value": "",
            "message": "Email is required."
        },
        {
            "resource": "Validation",
            "field": "Password",
            "value": "",
            "message": "Password must be at least 8 characters long."
        }
    ]
}
```

## Erweiterte Verwendung

### Custom Error Handler

```go
func CustomErrorHandler(err error) (int, interface{}) {
    // Custom Error-Behandlung vor Standard-Handler
    switch e := err.(type) {
    case *MyCustomError:
        return 422, gin.H{
            "error": "Custom error occurred",
            "details": e.Details,
        }
    default:
        // Fallback zu Standard-Handler
        return errors.NewErrorResponse(err)
    }
}
```

### Middleware für Error-Handling

```go
func ErrorMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Next()
        
        // Nach Request-Verarbeitung prüfen auf Errors
        if len(c.Errors) > 0 {
            err := c.Errors.Last().Err
            errorResponse, statusCode := errors.NewErrorResponse(err)
            c.JSON(statusCode, errorResponse)
        }
    }
}
```

### Logging Integration

```go
func LoggingErrorHandler(err error) (int, interface{}) {
    errorResponse, statusCode := errors.NewErrorResponse(err)
    
    // Error loggen
    log.WithFields(log.Fields{
        "error":       err.Error(),
        "status_code": statusCode,
        "type":        fmt.Sprintf("%T", err),
    }).Error("Request error occurred")
    
    return statusCode, errorResponse
}
```

## Best Practices

### 1. Konsistente Error-Erstellung

```go
// Schlecht - Inkonsistente Error-Messages
return errors.New("user not found")

// Gut - Strukturierte Error-Erstellung
return &errors.ErrorNotFound{
    ErrorDetail: errors.ErrorDetail{
        Resource: "User",
        Field:    "id",
        Value:    userID,
        Message:  fmt.Sprintf("User with ID %s not found", userID),
    },
}
```

### 2. Error-Wrapping für Context

```go
func (s *UserService) UpdateProfile(userID string, updates map[string]interface{}) error {
    user, err := s.repo.GetByID(userID)
    if err != nil {
        return fmt.Errorf("failed to get user for profile update: %w", err)
    }
    
    // ... update logic ...
    
    if err := s.repo.Update(user); err != nil {
        return fmt.Errorf("failed to update user profile: %w", err)
    }
    
    return nil
}
```

### 3. Defensive Programmierung

```go
func SafeGetUser(id string) (*User, error) {
    if id == "" {
        return nil, &errors.ErrorBadRequest{
            ErrorDetail: errors.ErrorDetail{
                Resource: "User",
                Field:    "id",
                Value:    "",
                Message:  "User ID is required",
            },
        }
    }
    
    return userRepo.GetByID(id)
}
```

## Testing

```go
func TestErrorHandling(t *testing.T) {
    // Test Not Found Error
    err := &errors.ErrorNotFound{
        ErrorDetail: errors.ErrorDetail{
            Resource: "User",
            Field:    "id",
            Value:    "123",
        },
    }
    
    errorResponse, statusCode := errors.NewErrorResponse(err)
    
    assert.Equal(t, 404, statusCode)
    assert.Equal(t, "Resource not found", errorResponse.Message)
    assert.Len(t, errorResponse.Errors, 1)
    assert.Equal(t, "User", errorResponse.Errors[0].Resource)
}

func TestValidationError(t *testing.T) {
    validate := validator.New()
    
    type TestStruct struct {
        Required string `validate:"required"`
        Email    string `validate:"email"`
    }
    
    test := TestStruct{
        Required: "",
        Email:    "invalid-email",
    }
    
    err := validate.Struct(test)
    assert.Error(t, err)
    
    errorResponse, statusCode := errors.NewErrorResponse(err)
    assert.Equal(t, 400, statusCode)
    assert.Equal(t, "Validation failed", errorResponse.Message)
}
```

## HTTP Status Code Mapping

| Error Type | HTTP Status | Verwendung |
|------------|-------------|------------|
| `ErrorNotFound` | 404 | Ressource nicht gefunden |
| `ErrorBadRequest` | 400 | Ungültige Anfrage/Validierung |
| `ErrorAlreadyExists` | 409 | Ressource existiert bereits |
| `ErrorUnauthorized` | 401 | Nicht autorisiert |
| `ErrorUserNotActive` | 401 | Benutzer nicht aktiv |
| `ErrorSalechannelNotAllowed` | 403 | Salechannel nicht erlaubt |
| Validation Errors | 400 | Validierungsfehler |
| Unbekannte Errors | 500 | Interne Serverfehler |