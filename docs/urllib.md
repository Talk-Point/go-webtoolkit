# URL Builder Utilities

Das urllib-Modul bietet eine typsichere und elegante Lösung für die Konstruktion von URLs mit Parameter-Handling, automatischer Enkodierung und Platzhalter-Ersetzung.

## Features

- ✅ Fluent API für URL-Erstellung
- ✅ Automatische Parameter-Enkodierung  
- ✅ Platzhalter-Ersetzung für dynamische URLs
- ✅ Type-Safe Parameter-Handling für verschiedene Datentypen
- ✅ Builder-Pattern für einfache Verkettung
- ✅ Null/Empty-Value-Handling

## Installation

```go
import "github.com/Talk-Point/go-webtoolkit/pkg/v2/urllib"
```

## Grundlegende Verwendung

### Einfache URL-Erstellung

```go
// Basis-URL ohne Parameter
url := urllib.Url("https://api.example.com/users").String()
// Ergebnis: "https://api.example.com/users"

// URL mit Query-Parametern
url := urllib.Url("https://api.example.com/users").
    AddParam("active", true).
    AddParam("limit", 50).
    String()
// Ergebnis: "https://api.example.com/users?active=true&limit=50"
```

### Platzhalter-Ersetzung

```go
// URL mit Platzhaltern
url := urllib.Url("https://api.example.com/users/:id/posts/:postId", map[string]interface{}{
    "id":     "123",
    "postId": "456",
}).String()
// Ergebnis: "https://api.example.com/users/123/posts/456"

// Mit zusätzlichen Query-Parametern
url := urllib.Url("https://api.example.com/users/:id", map[string]interface{}{
    "id": "123",
}).
    AddParam("include", "profile").
    AddParam("format", "json").
    String()
// Ergebnis: "https://api.example.com/users/123?include=profile&format=json"
```

## Unterstützte Datentypen

### Parameter-Typen

```go
url := urllib.Url("https://api.example.com/search").
    AddParam("query", "golang").              // string
    AddParam("active", true).                 // bool
    AddParam("limit", 25).                    // int
    AddParam("userId", int64(12345)).         // int64
    AddParam("score", 98.5).                  // float64
    AddParam("tags", []string{"go", "web"}).  // wird zu "go,web" 
    String()
```

### Automatische Typ-Konvertierung

```go
// Verschiedene Integer-Typen
url := urllib.Url("https://api.example.com/data").
    AddParam("int8", int8(127)).
    AddParam("int16", int16(32767)).
    AddParam("int32", int32(2147483647)).
    AddParam("int64", int64(9223372036854775807)).
    AddParam("uint", uint(42)).
    AddParam("uint8", uint8(255)).
    String()

// Float-Typen
url := urllib.Url("https://api.example.com/data").
    AddParam("float32", float32(3.14)).
    AddParam("float64", float64(2.718281828)).
    String()
```

## Bulk Parameter-Handling

### AddParams für mehrere Parameter

```go
params := map[string]interface{}{
    "name":     "John Doe",
    "age":      30,
    "active":   true,
    "score":    95.5,
    "tags":     []string{"admin", "user"},
}

url := urllib.Url("https://api.example.com/users").
    AddParams(params).
    String()
// Ergebnis: "https://api.example.com/users?active=true&age=30&name=John+Doe&score=95.500000&tags=admin%2Cuser"
```

### Kombinierung von AddParam und AddParams

```go
url := urllib.Url("https://api.example.com/search").
    AddParams(map[string]interface{}{
        "category": "tech",
        "limit":    20,
    }).
    AddParam("sort", "date").
    AddParam("order", "desc").
    String()
```

## HTTP-Client Integration

### Standard HTTP Client

```go
func FetchUser(userID string) (*User, error) {
    url := urllib.Url("https://api.example.com/users/:id", map[string]interface{}{
        "id": userID,
    }).
        AddParam("include", "profile").
        AddParam("format", "json").
        String()
    
    resp, err := http.Get(url)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    var user User
    err = json.NewDecoder(resp.Body).Decode(&user)
    return &user, err
}
```

### API Client Pattern

```go
type APIClient struct {
    BaseURL string
    APIKey  string
}

func (c *APIClient) buildURL(endpoint string, pathParams map[string]interface{}) *urllib.BaseUrl {
    return urllib.Url(c.BaseURL+endpoint, pathParams).
        AddParam("api_key", c.APIKey)
}

func (c *APIClient) GetUsers(filters map[string]interface{}) ([]User, error) {
    url := c.buildURL("/users", nil).
        AddParams(filters).
        String()
    
    // HTTP request...
    return users, nil
}

func (c *APIClient) GetUser(userID string, includeProfile bool) (*User, error) {
    url := c.buildURL("/users/:id", map[string]interface{}{
        "id": userID,
    }).
        AddParam("include_profile", includeProfile).
        String()
    
    // HTTP request...
    return user, nil
}
```

## Gin Framework Integration

### Route URL-Generierung

```go
type URLBuilder struct {
    scheme string
    host   string
}

func NewURLBuilder(scheme, host string) *URLBuilder {
    return &URLBuilder{scheme: scheme, host: host}
}

func (b *URLBuilder) UserURL(userID string) string {
    return urllib.Url(fmt.Sprintf("%s://%s/users/:id", b.scheme, b.host), map[string]interface{}{
        "id": userID,
    }).String()
}

func (b *URLBuilder) UserPostsURL(userID string, filters map[string]interface{}) string {
    return urllib.Url(fmt.Sprintf("%s://%s/users/:id/posts", b.scheme, b.host), map[string]interface{}{
        "id": userID,
    }).
        AddParams(filters).
        String()
}

// Usage in Handler
func GetUser(c *gin.Context) {
    userID := c.Param("id")
    
    // ... user logic ...
    
    builder := NewURLBuilder("https", "api.example.com")
    
    response := gin.H{
        "user": user,
        "links": gin.H{
            "self":  builder.UserURL(userID),
            "posts": builder.UserPostsURL(userID, map[string]interface{}{
                "limit": 10,
            }),
        },
    }
    
    c.JSON(200, response)
}
```

## API Referenz

### BaseUrl Struktur

```go
type BaseUrl struct {
    Base   string                // Basis-URL
    Params map[string]string     // Query-Parameter
}
```

### Funktionen

#### Url

```go
func Url(base string, params ...map[string]interface{}) *BaseUrl
```

Erstellt eine neue BaseUrl-Instanz mit optionalen Pfad-Parametern.

**Parameter:**
- `base`: Basis-URL mit optionalen Platzhaltern (`:param`)
- `params`: Optional - Map mit Pfad-Parametern für Platzhalter-Ersetzung

#### AddParam

```go
func (b *BaseUrl) AddParam(key string, value interface{}) *BaseUrl
```

Fügt einen einzelnen Query-Parameter hinzu.

**Parameter:**
- `key`: Parameter-Name
- `value`: Parameter-Wert (unterstützt verschiedene Typen)

**Unterstützte Typen:**
- `string`
- `bool` 
- `int`, `int8`, `int16`, `int32`, `int64`
- `uint`, `uint8`, `uint16`, `uint32`, `uint64`
- `float32`, `float64`
- Andere Typen werden mit `fmt.Sprintf("%v")` konvertiert

#### AddParams

```go
func (b *BaseUrl) AddParams(params map[string]interface{}) *BaseUrl
```

Fügt mehrere Query-Parameter aus einer Map hinzu.

#### String

```go
func (b *BaseUrl) String() string
```

Generiert die finale URL-String mit allen Parametern.

## Erweiterte Verwendung

### URL-Templates

```go
type URLTemplate struct {
    template string
    baseURL  string
}

func NewURLTemplate(baseURL, template string) *URLTemplate {
    return &URLTemplate{
        baseURL:  baseURL,
        template: template,
    }
}

func (t *URLTemplate) Build(pathParams map[string]interface{}, queryParams map[string]interface{}) string {
    return urllib.Url(t.baseURL+t.template, pathParams).
        AddParams(queryParams).
        String()
}

// Usage
userTemplate := NewURLTemplate("https://api.example.com", "/users/:id")
userURL := userTemplate.Build(
    map[string]interface{}{"id": "123"},  // path params
    map[string]interface{}{"include": "profile"},  // query params
)
```

### Conditional Parameters

```go
func BuildSearchURL(query string, filters map[string]interface{}) string {
    url := urllib.Url("https://api.example.com/search").
        AddParam("q", query)
    
    // Nur hinzufügen wenn Wert vorhanden
    if category, ok := filters["category"]; ok && category != "" {
        url = url.AddParam("category", category)
    }
    
    if minPrice, ok := filters["min_price"]; ok {
        url = url.AddParam("min_price", minPrice)
    }
    
    if maxPrice, ok := filters["max_price"]; ok {
        url = url.AddParam("max_price", maxPrice)
    }
    
    return url.String()
}
```

### Builder-Pattern Wrapper

```go
type SearchURLBuilder struct {
    url *urllib.BaseUrl
}

func NewSearchURLBuilder(baseURL string) *SearchURLBuilder {
    return &SearchURLBuilder{
        url: urllib.Url(baseURL),
    }
}

func (b *SearchURLBuilder) Query(q string) *SearchURLBuilder {
    b.url = b.url.AddParam("q", q)
    return b
}

func (b *SearchURLBuilder) Category(category string) *SearchURLBuilder {
    b.url = b.url.AddParam("category", category)
    return b
}

func (b *SearchURLBuilder) PriceRange(min, max float64) *SearchURLBuilder {
    b.url = b.url.AddParam("min_price", min).AddParam("max_price", max)
    return b
}

func (b *SearchURLBuilder) Sort(field string, ascending bool) *SearchURLBuilder {
    direction := "desc"
    if ascending {
        direction = "asc"
    }
    return b.url.AddParam("sort", field).AddParam("order", direction)
}

func (b *SearchURLBuilder) Build() string {
    return b.url.String()
}

// Usage
searchURL := NewSearchURLBuilder("https://api.example.com/search").
    Query("golang books").
    Category("programming").
    PriceRange(10.0, 50.0).
    Sort("price", true).
    Build()
```

## Best Practices

### 1. Konsistente Parameter-Benennung

```go
// Gut - Einheitliche Benennung
url := urllib.Url("https://api.example.com/users").
    AddParam("user_id", userID).
    AddParam("include_profile", true).
    AddParam("sort_by", "created_at").
    String()

// Schlecht - Inkonsistente Benennung
url := urllib.Url("https://api.example.com/users").
    AddParam("userId", userID).
    AddParam("includeProfile", true).
    AddParam("sort", "created_at").
    String()
```

### 2. Validation und Sanitization

```go
func BuildUserURL(userID string, options map[string]interface{}) (string, error) {
    if userID == "" {
        return "", errors.New("userID is required")
    }
    
    url := urllib.Url("https://api.example.com/users/:id", map[string]interface{}{
        "id": userID,
    })
    
    // Validate und sanitize options
    for key, value := range options {
        switch key {
        case "limit":
            if limit, ok := value.(int); ok && limit > 0 && limit <= 1000 {
                url = url.AddParam("limit", limit)
            }
        case "include":
            if include, ok := value.(string); ok && include != "" {
                url = url.AddParam("include", include)
            }
        }
    }
    
    return url.String(), nil
}
```

### 3. Environment-spezifische URLs

```go
type Config struct {
    APIBaseURL string
    APIVersion string
}

func (c *Config) BuildAPIURL(endpoint string, pathParams map[string]interface{}) *urllib.BaseUrl {
    fullEndpoint := fmt.Sprintf("/%s%s", c.APIVersion, endpoint)
    return urllib.Url(c.APIBaseURL+fullEndpoint, pathParams)
}

// Usage
config := &Config{
    APIBaseURL: os.Getenv("API_BASE_URL"),
    APIVersion: "v1",
}

userURL := config.BuildAPIURL("/users/:id", map[string]interface{}{
    "id": "123",
}).
    AddParam("format", "json").
    String()
```

## Testing

```go
func TestURLBuilder(t *testing.T) {
    // Test einfache URL
    url := urllib.Url("https://api.example.com/users").String()
    assert.Equal(t, "https://api.example.com/users", url)
    
    // Test mit Parametern
    url = urllib.Url("https://api.example.com/users").
        AddParam("active", true).
        AddParam("limit", 50).
        String()
    assert.Contains(t, url, "active=true")
    assert.Contains(t, url, "limit=50")
    
    // Test Platzhalter-Ersetzung
    url = urllib.Url("https://api.example.com/users/:id", map[string]interface{}{
        "id": "123",
    }).String()
    assert.Equal(t, "https://api.example.com/users/123", url)
    
    // Test Parameter-Enkodierung
    url = urllib.Url("https://api.example.com/search").
        AddParam("query", "hello world").
        String()
    assert.Contains(t, url, "query=hello+world")
}

func TestParameterTypes(t *testing.T) {
    url := urllib.Url("https://api.example.com/test").
        AddParam("str", "text").
        AddParam("bool", true).
        AddParam("int", 42).
        AddParam("float", 3.14).
        String()
    
    assert.Contains(t, url, "str=text")
    assert.Contains(t, url, "bool=true")
    assert.Contains(t, url, "int=42")
    assert.Contains(t, url, "float=3.14")
}
```

## Fehlerbehebung

### Häufige Probleme

1. **Leere Parameter werden übersprungen**
   ```go
   // Empty strings werden nicht hinzugefügt
   url := urllib.Url("https://api.example.com").
       AddParam("empty", "").  // Wird ignoriert
       String()
   ```

2. **Platzhalter nicht ersetzt**
   ```go
   // Platzhalter werden nur ersetzt wenn pathParams übergeben werden
   url := urllib.Url("https://api.example.com/users/:id").  // :id bleibt
       String()
   // Besser:
   url := urllib.Url("https://api.example.com/users/:id", map[string]interface{}{
       "id": "123",
   }).String()
   ```

3. **Parameter-Enkodierung**
   ```go
   // Automatische URL-Enkodierung
   url := urllib.Url("https://api.example.com").
       AddParam("query", "hello world & more").  // Wird zu "hello+world+%26+more"
       String()
   ```