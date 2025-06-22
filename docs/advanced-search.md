# Advanced Search Engine

Das Advanced Search-Modul bietet eine leistungsstarke Such- und Filter-Engine für komplexe Abfragen mit textbasierter Query-Syntax und automatischer SQL-Generierung.

## Features

- ✅ Textbasierte Such-Queries mit SQL-Operatoren
- ✅ Unterstützung für verschiedene Datentypen (String, Int, Float, Bool, Date)
- ✅ Spalten-Aliasing und Field-Mapping
- ✅ Automatische SQL-WHERE- und ORDER BY-Generierung
- ✅ Type-Safe Value-Handling
- ✅ Flexible Sortierung mit mehreren Feldern

## Installation

```go
import "github.com/Talk-Point/go-webtoolkit/pkg/v2/advanced_search"
```

## Query-Syntax

### Basis-Syntax

```
field:operator:value
```

### Verfügbare Operatoren

```go
const (
    Equal              = "="        // Gleich
    NotEqual           = "!="       // Ungleich  
    GreaterThan        = ">"        // Größer als
    GreaterThanOrEqual = ">="       // Größer oder gleich
    LessThan           = "<"        // Kleiner als
    LessThanOrEqual    = "<="       // Kleiner oder gleich
    Null               = "null"     // IS NULL
    NotNull            = "notnull"  // IS NOT NULL
    Like               = "like"     // LIKE (mit automatischen Wildcards)
)
```

### Sortierungs-Syntax

```
sort:field1__asc,field2__desc
```

## Grundlegende Verwendung

### Einfache Queries

```go
// Einfache Suche
query := "name:=:John age:>:25"
search := advanced_search.NewAdvancedSearch(query)

clauses, sortClauses, err := search.Parse()
if err != nil {
    log.Fatal(err)
}

// Resultat:
// clauses[0] = Clause{Field: "name", Operator: Equal, Value: "John"}
// clauses[1] = Clause{Field: "age", Operator: GreaterThan, Value: "25"}
```

### Mit Sortierung

```go
// Query mit Sortierung
query := "status:=:active sort:created_at__desc,name__asc"
search := advanced_search.NewAdvancedSearch(query)

clauses, sortClauses, err := search.Parse()
if err != nil {
    log.Fatal(err)
}

// sortClauses[0] = SortClause{Field: "created_at", Direction: Desc}
// sortClauses[1] = SortClause{Field: "name", Direction: Asc}
```

### NULL-Checks

```go
// NULL und NOT NULL Checks
query := "deleted_at:null email:notnull"
search := advanced_search.NewAdvancedSearch(query)

clauses, _, err := search.Parse()
// clauses[0] = Clause{Field: "deleted_at", Operator: Null}
// clauses[1] = Clause{Field: "email", Operator: NotNull}
```

## SQL-Generierung

### Erweiterte SQL-Search

```go
// Spalten-Definitionen
columns := []advanced_search.Column{
    {Name: "user_name", Type: advanced_search.StringType, Aliases: []string{"name", "username"}},
    {Name: "user_email", Type: advanced_search.StringType, Aliases: []string{"email"}},
    {Name: "age", Type: advanced_search.IntType},
    {Name: "created_at", Type: advanced_search.DateTimeType, Aliases: []string{"created"}},
    {Name: "active", Type: advanced_search.BoolType},
}

// Query mit Spalten-Mapping
query := "name:=:John email:like:@gmail.com age:>:25 active:=:true sort:created__desc"
sqlSearch, err := advanced_search.NewAdvancedSqlSearch(query, columns...)
if err != nil {
    log.Fatal(err)
}

// SQL WHERE-Klausel generieren
whereClause, err := sqlSearch.WhereStatement()
if err != nil {
    log.Fatal(err)
}
// Resultat: "user_name = 'John' AND user_email LIKE '%@gmail.com%' AND age > 25 AND active = true"

// SQL ORDER BY-Klausel generieren
orderByClause, err := sqlSearch.SortStatement()
if err != nil {
    log.Fatal(err)
}
// Resultat: "created_at DESC"
```

### Vollständige SQL-Query

```go
// Komplette SQL-Query generieren
whereStmt, orderByStmt, err := sqlSearch.Sql()
if err != nil {
    log.Fatal(err)
}

// Finale SQL zusammenbauen
sql := "SELECT * FROM users"
if whereStmt != "" {
    sql += " WHERE " + whereStmt
}
if orderByStmt != "" {
    sql += " ORDER BY " + orderByStmt
}

// Resultat:
// "SELECT * FROM users WHERE user_name = 'John' AND user_email LIKE '%@gmail.com%' AND age > 25 AND active = true ORDER BY created_at DESC"
```

## Repository Integration

### SQL Repository mit Advanced Search

```go
type UserRepository struct {
    db      *sql.DB
    columns []advanced_search.Column
}

func NewUserRepository(db *sql.DB) *UserRepository {
    columns := []advanced_search.Column{
        {Name: "id", Type: advanced_search.IntType},
        {Name: "email", Type: advanced_search.StringType},
        {Name: "full_name", Type: advanced_search.StringType, Aliases: []string{"name", "username"}},
        {Name: "age", Type: advanced_search.IntType},
        {Name: "active", Type: advanced_search.BoolType},
        {Name: "created_at", Type: advanced_search.DateTimeType, Aliases: []string{"created", "date"}},
        {Name: "updated_at", Type: advanced_search.DateTimeType, Aliases: []string{"updated"}},
    }
    
    return &UserRepository{
        db:      db,
        columns: columns,
    }
}

func (r *UserRepository) Search(ctx context.Context, query string) ([]User, error) {
    // Query parsen und SQL generieren
    sqlSearch, err := advanced_search.NewAdvancedSqlSearch(query, r.columns...)
    if err != nil {
        return nil, fmt.Errorf("invalid search query: %w", err)
    }
    
    whereStmt, orderByStmt, err := sqlSearch.Sql()
    if err != nil {
        return nil, fmt.Errorf("failed to generate SQL: %w", err)
    }
    
    // SQL-Query aufbauen
    sql := "SELECT id, email, full_name, age, active, created_at, updated_at FROM users"
    if whereStmt != "" {
        sql += " WHERE " + whereStmt
    }
    if orderByStmt != "" {
        sql += " ORDER BY " + orderByStmt
    }
    
    // Query ausführen
    rows, err := r.db.QueryContext(ctx, sql)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    var users []User
    for rows.Next() {
        var user User
        err := rows.Scan(&user.ID, &user.Email, &user.Name, &user.Age, &user.Active, &user.CreatedAt, &user.UpdatedAt)
        if err != nil {
            return nil, err
        }
        users = append(users, user)
    }
    
    return users, nil
}
```

## HTTP Handler Integration

### REST API mit Advanced Search

```go
type SearchHandler struct {
    userRepo *UserRepository
}

func (h *SearchHandler) SearchUsers(c *gin.Context) {
    // Query-Parameter lesen
    query := c.Query("q")
    if query == "" {
        c.JSON(400, gin.H{"error": "Query parameter 'q' is required"})
        return
    }
    
    // Search ausführen
    users, err := h.userRepo.Search(c.Request.Context(), query)
    if err != nil {
        // Error-System Integration
        if searchErr, ok := err.(*advanced_search.ParseError); ok {
            c.JSON(400, gin.H{
                "error": "Invalid search query",
                "details": searchErr.Error(),
            })
            return
        }
        
        c.JSON(500, gin.H{"error": "Search failed"})
        return
    }
    
    c.JSON(200, gin.H{
        "users":  users,
        "count":  len(users),
        "query":  query,
    })
}

// URL-Beispiele:
// GET /api/users/search?q=name:=:John age:>:25
// GET /api/users/search?q=email:like:@gmail.com active:=:true sort:created__desc
// GET /api/users/search?q=created:>=:2023-01-01 sort:name__asc
```

### Form-basierte Suche

```go
func (h *SearchHandler) SearchForm(c *gin.Context) {
    // Form-Daten zu Query konvertieren
    var queryParts []string
    
    if name := c.PostForm("name"); name != "" {
        queryParts = append(queryParts, fmt.Sprintf("name:=:%s", name))
    }
    
    if email := c.PostForm("email"); email != "" {
        queryParts = append(queryParts, fmt.Sprintf("email:like:%s", email))
    }
    
    if ageMin := c.PostForm("age_min"); ageMin != "" {
        queryParts = append(queryParts, fmt.Sprintf("age:>=:%s", ageMin))
    }
    
    if ageMax := c.PostForm("age_max"); ageMax != "" {
        queryParts = append(queryParts, fmt.Sprintf("age:<=:%s", ageMax))
    }
    
    if active := c.PostForm("active"); active != "" {
        queryParts = append(queryParts, fmt.Sprintf("active:=:%s", active))
    }
    
    // Sortierung hinzufügen
    if sort := c.PostForm("sort"); sort != "" {
        queryParts = append(queryParts, fmt.Sprintf("sort:%s", sort))
    }
    
    query := strings.Join(queryParts, " ")
    
    // Search ausführen
    users, err := h.userRepo.Search(c.Request.Context(), query)
    if err != nil {
        c.JSON(500, gin.H{"error": "Search failed"})
        return
    }
    
    c.JSON(200, users)
}
```

## Erweiterte Features

### Custom Column Types

```go
// Eigene Datentypen definieren
const (
    EmailType   advanced_search.ColumnType = 100
    PhoneType   advanced_search.ColumnType = 101
    URLType     advanced_search.ColumnType = 102
)

// Custom SQL-Generierung für spezielle Typen
func CustomClauseSQL(clause advanced_search.Clause, column advanced_search.Column) string {
    switch column.Type {
    case EmailType:
        // Spezielle Email-Behandlung
        if clause.Operator == advanced_search.Like {
            return fmt.Sprintf("LOWER(%s) LIKE LOWER('%%%s%%')", column.Name, clause.Value)
        }
    case PhoneType:
        // Telefonnummer-Normalisierung
        normalizedValue := normalizePhoneNumber(clause.Value)
        return fmt.Sprintf("%s = '%s'", column.Name, normalizedValue)
    }
    
    // Fallback zu Standard-Implementation
    return clause.Sql(column.Name)
}
```

### Query Validation

```go
func ValidateSearchQuery(query string, allowedFields []string) error {
    search := advanced_search.NewAdvancedSearch(query)
    clauses, sortClauses, err := search.Parse()
    if err != nil {
        return err
    }
    
    allowedFieldsMap := make(map[string]bool)
    for _, field := range allowedFields {
        allowedFieldsMap[field] = true
    }
    
    // Validate filter fields
    for _, clause := range clauses {
        if !allowedFieldsMap[clause.Field] {
            return fmt.Errorf("field '%s' is not allowed for searching", clause.Field)
        }
    }
    
    // Validate sort fields
    for _, sortClause := range sortClauses {
        if !allowedFieldsMap[sortClause.Field] {
            return fmt.Errorf("field '%s' is not allowed for sorting", sortClause.Field)
        }
    }
    
    return nil
}
```

### Performance-Optimierung

```go
type SearchCache struct {
    cache map[string][]User
    mutex sync.RWMutex
    ttl   time.Duration
}

func (s *SearchCache) Get(query string) ([]User, bool) {
    s.mutex.RLock()
    defer s.mutex.RUnlock()
    
    if users, exists := s.cache[query]; exists {
        return users, true
    }
    return nil, false
}

func (s *SearchCache) Set(query string, users []User) {
    s.mutex.Lock()
    defer s.mutex.Unlock()
    
    s.cache[query] = users
    
    // TTL-basierte Bereinigung nach Delay
    time.AfterFunc(s.ttl, func() {
        s.mutex.Lock()
        delete(s.cache, query)
        s.mutex.Unlock()
    })
}

// Usage im Repository
func (r *UserRepository) SearchWithCache(ctx context.Context, query string) ([]User, error) {
    // Cache prüfen
    if users, found := r.cache.Get(query); found {
        return users, nil
    }
    
    // Search ausführen
    users, err := r.Search(ctx, query)
    if err != nil {
        return nil, err
    }
    
    // Cache setzen
    r.cache.Set(query, users)
    return users, nil
}
```

## API Referenz

### Column Definition

```go
type Column struct {
    Name    string        // Spaltenname in der Datenbank
    Type    ColumnType    // Datentyp der Spalte
    Aliases []string      // Alternative Namen für die Spalte
}

type ColumnType int
const (
    StringType   ColumnType = iota  // String/Text
    IntType                         // Integer
    FloatType                       // Float/Decimal
    BoolType                        // Boolean
    DateType                        // Date (YYYY-MM-DD)
    DateTimeType                    // DateTime (RFC3339)
)
```

### Clause Structure

```go
type Clause struct {
    Field    string    // Feldname
    Operator Operator  // SQL-Operator
    Value    string    // Vergleichswert
}

func (c Clause) Sql(name string) string // Generiert SQL für die Clause
```

### Sort Clause

```go
type SortClause struct {
    Field     string     // Feldname
    Direction Direction  // Sortierrichtung
}

type Direction int
const (
    Asc  Direction = iota  // Aufsteigend
    Desc                   // Absteigend
)
```

## Query-Beispiele

### Benutzer-Suche

```go
// Aktive Benutzer namens John
"name:=:John active:=:true"

// Benutzer mit Gmail-Adressen über 25 Jahre
"email:like:@gmail.com age:>:25"

// Kürzlich erstellte Benutzer, sortiert nach Name
"created:>=:2023-01-01 sort:name__asc"

// Benutzer ohne E-Mail oder gelöschte Accounts
"email:null deleted_at:notnull"
```

### E-Commerce-Suche

```go
// Produkte in Preisbereich
"price:>=:10 price:<=:100 category:=:electronics"

// Verfügbare Produkte, sortiert nach Beliebtheit
"stock:>:0 available:=:true sort:popularity__desc"

// Reduzierte Artikel
"discount:>:0 sort:discount__desc,price__asc"
```

### Content-Management

```go
// Veröffentlichte Artikel von Autor
"status:=:published author:=:john sort:published_at__desc"

// Artikel mit Tags, die Kommentare haben
"tags:like:javascript comments_count:>:0"

// Entwürfe der letzten Woche
"status:=:draft created:>=:2023-06-01 created:<=:2023-06-07"
```

## Best Practices

### 1. Column-Definitionen strukturieren

```go
// Gut - Klare Spalten-Definitionen mit Aliasing
columns := []advanced_search.Column{
    {
        Name:    "user_email",
        Type:    advanced_search.StringType,
        Aliases: []string{"email", "mail", "e-mail"},
    },
    {
        Name:    "full_name",
        Type:    advanced_search.StringType,
        Aliases: []string{"name", "username", "display_name"},
    },
}

// Schlecht - Keine Aliases, unklare Namen
columns := []advanced_search.Column{
    {Name: "col1", Type: advanced_search.StringType},
    {Name: "col2", Type: advanced_search.StringType},
}
```

### 2. Query-Validation

```go
func (h *SearchHandler) validateQuery(query string) error {
    // Maximale Query-Länge
    if len(query) > 1000 {
        return errors.New("query too long")
    }
    
    // Gefährliche SQL-Patterns prüfen
    dangerousPatterns := []string{"DROP", "DELETE", "UPDATE", "INSERT", "--", "/*"}
    queryUpper := strings.ToUpper(query)
    for _, pattern := range dangerousPatterns {
        if strings.Contains(queryUpper, pattern) {
            return fmt.Errorf("dangerous pattern detected: %s", pattern)
        }
    }
    
    return nil
}
```

### 3. Type-Safe Value-Handling

```go
func convertValue(value string, columnType advanced_search.ColumnType) (interface{}, error) {
    switch columnType {
    case advanced_search.IntType:
        return strconv.Atoi(value)
    case advanced_search.FloatType:
        return strconv.ParseFloat(value, 64)
    case advanced_search.BoolType:
        return strconv.ParseBool(value)
    case advanced_search.DateTimeType:
        return time.Parse(time.RFC3339, value)
    default:
        return value, nil
    }
}
```

## Testing

```go
func TestAdvancedSearch(t *testing.T) {
    // Test Query-Parsing
    query := "name:=:John age:>:25 sort:created__desc"
    search := advanced_search.NewAdvancedSearch(query)
    
    clauses, sortClauses, err := search.Parse()
    assert.NoError(t, err)
    assert.Len(t, clauses, 2)
    assert.Len(t, sortClauses, 1)
    
    // Test erste Clause
    assert.Equal(t, "name", clauses[0].Field)
    assert.Equal(t, advanced_search.Equal, clauses[0].Operator)
    assert.Equal(t, "John", clauses[0].Value)
    
    // Test Sortierung
    assert.Equal(t, "created", sortClauses[0].Field)
    assert.Equal(t, advanced_search.Desc, sortClauses[0].Direction)
}

func TestSQLGeneration(t *testing.T) {
    columns := []advanced_search.Column{
        {Name: "user_name", Type: advanced_search.StringType, Aliases: []string{"name"}},
        {Name: "age", Type: advanced_search.IntType},
    }
    
    query := "name:=:John age:>:25"
    sqlSearch, err := advanced_search.NewAdvancedSqlSearch(query, columns...)
    assert.NoError(t, err)
    
    whereStmt, err := sqlSearch.WhereStatement()
    assert.NoError(t, err)
    expected := "user_name = 'John' AND age > 25"
    assert.Equal(t, expected, whereStmt)
}
```