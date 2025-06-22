# Advanced Query System

Das Query-System bietet eine flexible und mächtige Lösung für URL-basierte Queries, Filterung und Sortierung mit nahtloser Firestore-Integration.

## Features

- ✅ URL-Parameter zu Filter-Objekten konvertieren
- ✅ Unterstützung für komplexe Operatoren (eq, gt, lt, contains, etc.)
- ✅ Automatische Firestore-Operator-Konvertierung
- ✅ Sortierung und Paginierung
- ✅ Form-Data-Integration
- ✅ Field-Mapping für flexible Datenstrukturen

## Installation

```go
import "github.com/Talk-Point/go-webtoolkit/pkg/v2/query"
```

## Query-Operatoren

### Verfügbare Operatoren

```go
const (
    Eq            Operator = "eq"            // Gleich
    Eqe           Operator = "eqe"           // Gleich (empty)
    Gt            Operator = "gt"            // Größer als
    Gte           Operator = "gte"           // Größer oder gleich
    Lt            Operator = "lt"            // Kleiner als
    Lte           Operator = "lte"           // Kleiner oder gleich
    Contains      Operator = "contains"      // Enthält (Array)
    ArrayContains Operator = "array-contains" // Array enthält Element
)
```

### Firestore-Mapping

```go
// Automatische Konvertierung zu Firestore-Operatoren
Eq            -> "=="
Gt            -> ">"
Gte           -> ">="
Lt            -> "<"
Lte           -> "<="
Contains      -> "in"
ArrayContains -> "array-contains"
```

## Grundlegende Verwendung

### URL-Parameter zu Filtern

```go
// Einfache Gleichheits-Filter
// URL: /api/users?name=John&status=active
filters, err := query.NewFilterFromUrlString("/api/users?name=John&status=active")
if err != nil {
    log.Fatal(err)
}

// Resultat:
// filters[0] = Filter{Field: "name", Operator: Eq, Value: "John"}
// filters[1] = Filter{Field: "status", Operator: Eq, Value: "active"}
```

### Erweiterte Operatoren

```go
// URL mit Operatoren
// URL: /api/users?age__gt=18&created_at__gte=2023-01-01&status__contains=active,pending
filters, err := query.NewFilterFromUrlString("/api/users?age__gt=18&created_at__gte=2023-01-01")
if err != nil {
    log.Fatal(err)
}

// Resultat:
// filters[0] = Filter{Field: "age", Operator: Gt, Value: "18"}
// filters[1] = Filter{Field: "created_at", Operator: Gte, Value: "2023-01-01"}
```

### Vollständige Query-Optionen

```go
// URL: /api/users?name=John&age__gt=18&limit=20&sort=-created_at&next=abc123
queryOpts, err := query.NewQueryOptionsFromUrlString("/api/users?name=John&age__gt=18&limit=20&sort=-created_at&next=abc123")
if err != nil {
    log.Fatal(err)
}

// Resultat:
// queryOpts.Limit = 20
// queryOpts.OrderBy = "created_at"
// queryOpts.OrderByDirection = Desc
// queryOpts.Next = "abc123"
// queryOpts.Filters = [Filter{Field: "name", Operator: Eq, Value: "John"}, ...]
```

## HTTP-Handler Integration

### Gin Framework

```go
func GetUsers(c *gin.Context) {
    // Query-Parameter automatisch parsen
    queryOpts, err := query.NewQueryOptionsFromUrl(c.Request.URL)
    if err != nil {
        c.JSON(400, gin.H{"error": "Invalid query parameters"})
        return
    }
    
    // Repository-Aufruf mit Query-Optionen
    users, err := userRepo.Get(c.Request.Context(), &queryOpts)
    if err != nil {
        c.JSON(500, gin.H{"error": "Failed to fetch users"})
        return
    }
    
    c.JSON(200, users)
}
```

### Standard HTTP

```go
func usersHandler(w http.ResponseWriter, r *http.Request) {
    queryOpts, err := query.NewQueryOptionsFromUrl(r.URL)
    if err != nil {
        http.Error(w, "Invalid query parameters", http.StatusBadRequest)
        return
    }
    
    users, err := userRepo.Get(r.Context(), &queryOpts)
    if err != nil {
        http.Error(w, "Failed to fetch users", http.StatusInternalServerError)
        return
    }
    
    json.NewEncoder(w).Encode(users)
}
```

## Form-Data Integration

### HTML-Forms zu Query-Optionen

```go
func SearchHandler(c *gin.Context) {
    // Form-Values parsen
    if err := c.Request.ParseForm(); err != nil {
        c.JSON(400, gin.H{"error": "Invalid form data"})
        return
    }
    
    // Field-Mapping für Form-Felder
    fieldMapping := map[string]string{
        "user_name":    "name",
        "user_email":   "email", 
        "created_from": "created_at__gte",
        "created_to":   "created_at__lte",
    }
    
    queryOpts, err := query.NewQueryOptionsFromForm(c.Request.Form, fieldMapping)
    if err != nil {
        c.JSON(400, gin.H{"error": "Invalid search parameters"})
        return
    }
    
    results, err := searchService.Search(c.Request.Context(), &queryOpts)
    if err != nil {
        c.JSON(500, gin.H{"error": "Search failed"})
        return
    }
    
    c.JSON(200, results)
}
```

### HTML-Form Beispiel

```html
<form method="GET" action="/search">
    <input type="text" name="user_name" placeholder="Name">
    <input type="email" name="user_email" placeholder="Email">
    <input type="date" name="created_from" placeholder="Von">
    <input type="date" name="created_to" placeholder="Bis">
    <select name="sort">
        <option value="name__asc">Name (A-Z)</option>
        <option value="name__desc">Name (Z-A)</option>
        <option value="created_at__desc">Neueste zuerst</option>
    </select>
    <button type="submit">Suchen</button>
</form>
```

## Repository Integration

### Firestore Repository

```go
func (r *UserRepository) Get(ctx context.Context, opts *query.QueryOptions) (*PaginationResult, error) {
    q := r.collection.Limit(opts.Limit)
    
    // Filter anwenden
    for _, filter := range opts.Filters {
        // Automatische Firestore-Operator-Konvertierung
        q = q.Where(filter.Field, filter.Operator.ToFireStoreOperator(), filter.Value)
    }
    
    // Sortierung
    if opts.OrderBy != "" {
        direction := firestore.Asc
        if opts.OrderByDirection == query.Desc {
            direction = firestore.Desc
        }
        q = q.OrderBy(opts.OrderBy, direction)
    }
    
    // Paginierung
    if opts.Next != "" {
        doc, err := r.collection.Doc(opts.Next).Get(ctx)
        if err != nil {
            return nil, err
        }
        q = q.StartAfter(doc)
    }
    
    // Query ausführen
    docs, err := q.Documents(ctx).GetAll()
    if err != nil {
        return nil, err
    }
    
    // Ergebnisse verarbeiten...
    return &PaginationResult{
        Items: users,
        Limit: opts.Limit,
        Next:  nextPageToken,
        Prev:  prevPageToken,
    }, nil
}
```

### SQL Repository

```go
func (r *UserRepository) Get(ctx context.Context, opts *query.QueryOptions) (*PaginationResult, error) {
    var conditions []string
    var args []interface{}
    
    // Filter zu SQL-Bedingungen konvertieren
    for _, filter := range opts.Filters {
        switch filter.Operator {
        case query.Eq:
            conditions = append(conditions, fmt.Sprintf("%s = ?", filter.Field))
            args = append(args, filter.Value)
        case query.Gt:
            conditions = append(conditions, fmt.Sprintf("%s > ?", filter.Field))
            args = append(args, filter.Value)
        case query.Lt:
            conditions = append(conditions, fmt.Sprintf("%s < ?", filter.Field))
            args = append(args, filter.Value)
        // ... weitere Operatoren
        }
    }
    
    // SQL-Query aufbauen
    sql := "SELECT * FROM users"
    if len(conditions) > 0 {
        sql += " WHERE " + strings.Join(conditions, " AND ")
    }
    
    // Sortierung hinzufügen
    if opts.OrderBy != "" {
        direction := "ASC"
        if opts.OrderByDirection == query.Desc {
            direction = "DESC"
        }
        sql += fmt.Sprintf(" ORDER BY %s %s", opts.OrderBy, direction)
    }
    
    // Limit hinzufügen
    sql += fmt.Sprintf(" LIMIT %d", opts.Limit)
    
    rows, err := r.db.QueryContext(ctx, sql, args...)
    // ... query execution
}
```

## Erweiterte Features

### Pagination

```go
type PaginationResult[T any] struct {
    Items   []T             `json:"items"`
    Limit   int             `json:"limit"`
    Next    string          `json:"next"`
    Prev    string          `json:"prev"`
    Filters *[]query.Filter `json:"filters,omitempty"`
}

// Usage
func GetUsersWithPagination(c *gin.Context) {
    queryOpts, _ := query.NewQueryOptionsFromUrl(c.Request.URL)
    
    result, err := userRepo.Get(c.Request.Context(), &queryOpts)
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }
    
    // Client bekommt Next/Prev-Tokens für Navigation
    c.JSON(200, result)
}
```

### Custom Sorting

```go
// URL: /api/users?sort=-created_at,name
// Sortiert nach created_at descending, dann name ascending

func parseMultiSort(sortParam string) []query.SortOption {
    var sorts []query.SortOption
    
    for _, sort := range strings.Split(sortParam, ",") {
        direction := query.Asc
        field := sort
        
        if strings.HasPrefix(sort, "-") {
            direction = query.Desc
            field = sort[1:]
        }
        
        sorts = append(sorts, query.SortOption{
            Field:     field,
            Direction: direction,
        })
    }
    
    return sorts
}
```

### Validation und Sanitization

```go
func ValidateQueryOptions(opts *query.QueryOptions) error {
    // Limit validieren
    if opts.Limit < 1 || opts.Limit > 1000 {
        return errors.New("limit must be between 1 and 1000")
    }
    
    // Erlaubte Felder definieren
    allowedFields := map[string]bool{
        "name":       true,
        "email":      true,
        "created_at": true,
        "status":     true,
    }
    
    // Filter validieren 
    for _, filter := range opts.Filters {
        if !allowedFields[filter.Field] {
            return fmt.Errorf("field '%s' is not allowed for filtering", filter.Field)
        }
    }
    
    // Sortierung validieren
    if opts.OrderBy != "" && !allowedFields[opts.OrderBy] {
        return fmt.Errorf("field '%s' is not allowed for sorting", opts.OrderBy)
    }
    
    return nil
}
```

## API Referenz

### Filter

```go
type Filter struct {
    Field    string      // Feldname
    Operator Operator    // Vergleichsoperator
    Value    interface{} // Vergleichswert
}
```

### QueryOptions

```go
type QueryOptions struct {
    Limit            int         // Anzahl Ergebnisse (Standard: 30, Max: 100)
    Next             string      // Pagination-Token für Vorwärts-Navigation
    Previous         string      // Pagination-Token für Rückwärts-Navigation
    OrderBy          string      // Sortierfeld (Standard: "id")
    OrderByDirection Direction   // Sortierrichtung (asc/desc, Standard: desc)
    Filters          []Filter    // Angewandte Filter
}
```

### Funktionen

```go
// URL zu Filtern
func NewFiltersFromUrl(value *url.URL) ([]Filter, error)
func NewFilterFromUrlString(value string) ([]Filter, error)

// URL zu QueryOptions
func NewQueryOptionsFromUrl(value *url.URL) (QueryOptions, error)
func NewQueryOptionsFromUrlString(value string) (QueryOptions, error)

// Form zu QueryOptions
func NewQueryOptionsFromForm(values url.Values, fieldMappings ...map[string]string) (QueryOptions, error)
```

## URL-Format Beispiele

### Einfache Filter
```
/api/users?name=John&active=true
```

### Erweiterte Operatoren
```
/api/users?age__gt=18&created_at__gte=2023-01-01&tags__contains=admin
```

### Sortierung
```
/api/users?sort=-created_at        # Nach created_at absteigend
/api/users?sort=name               # Nach name aufsteigend
```

### Paginierung
```
/api/users?limit=50&next=abc123    # Nächste 50 Ergebnisse
/api/users?limit=25&prev=xyz789    # Vorherige 25 Ergebnisse
```

### Kombiniert
```
/api/users?name__like=john&age__gte=18&status__in=active,pending&sort=-created_at&limit=20
```

## Best Practices

### 1. Standardwerte definieren

```go
func GetDefaultQueryOptions() query.QueryOptions {
    return query.QueryOptions{
        Limit:            30,
        OrderBy:          "id",
        OrderByDirection: query.Desc,
        Filters:          []query.Filter{},
    }
}
```

### 2. Field-Validation

```go
func ValidateFields(allowedFields map[string]bool, filters []query.Filter) error {
    for _, filter := range filters {
        if !allowedFields[filter.Field] {
            return fmt.Errorf("field '%s' not allowed", filter.Field)
        }
    }
    return nil
}
```

### 3. Type-Safe Values

```go
func ConvertFilterValue(filter query.Filter, expectedType string) (interface{}, error) {
    switch expectedType {
    case "int":
        return strconv.Atoi(filter.Value.(string))
    case "bool":
        return strconv.ParseBool(filter.Value.(string))
    case "time":
        return time.Parse(time.RFC3339, filter.Value.(string))
    default:
        return filter.Value, nil
    }
}
```

## Testing

```go
func TestQueryParsing(t *testing.T) {
    // Test einfache Filter
    filters, err := query.NewFilterFromUrlString("?name=John&age=25")
    assert.NoError(t, err)
    assert.Len(t, filters, 2)
    assert.Equal(t, "name", filters[0].Field)
    assert.Equal(t, query.Eq, filters[0].Operator)
    assert.Equal(t, "John", filters[0].Value)
    
    // Test erweiterte Operatoren
    filters, err = query.NewFilterFromUrlString("?age__gt=18&status__contains=active")
    assert.NoError(t, err)
    assert.Equal(t, query.Gt, filters[0].Operator)
    assert.Equal(t, query.Contains, filters[1].Operator)
    
    // Test QueryOptions
    opts, err := query.NewQueryOptionsFromUrlString("?limit=50&sort=-created_at")
    assert.NoError(t, err)
    assert.Equal(t, 50, opts.Limit)
    assert.Equal(t, "created_at", opts.OrderBy)
    assert.Equal(t, query.Desc, opts.OrderByDirection)
}
```