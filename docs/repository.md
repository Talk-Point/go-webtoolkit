# Generic Repository Pattern

Das Repository-Modul bietet ein type-safe Repository-Pattern für Firestore mit Go Generics, das CRUD-Operationen, Paginierung und erweiterte Query-Funktionen vereinfacht.

## Features

- ✅ Type-safe Repository mit Go Generics  
- ✅ CRUD-Operationen mit automatischem ID-Management
- ✅ Paginierte Abfragen mit Next/Previous-Unterstützung
- ✅ Transaktionale Erstellung mit Duplikatsprüfung
- ✅ Integrierte Query-System-Unterstützung
- ✅ Automatisches Timestamp-Management
- ✅ Firestore-optimierte Implementierung

## Installation

```go
import "github.com/Talk-Point/go-webtoolkit/pkg/v2/repository"
```

## Entity Interface

Alle Entities müssen das Entity Interface implementieren:

```go
type Entity interface {
    DocId() string                          // Dokument-ID zurückgeben
    SetDocId(id string)                     // Dokument-ID setzen
    UniqFields() map[string]interface{}     // Eindeutige Felder für Duplikatsprüfung
}
```

### Entity Implementierung

```go
type User struct {
    ID        string    `json:"id" firestore:"-"`
    Email     string    `json:"email" firestore:"email"`
    Name      string    `json:"name" firestore:"name"`
    Active    bool      `json:"active" firestore:"active"`
    CreatedAt time.Time `json:"created_at" firestore:"created_at"`
    UpdatedAt time.Time `json:"updated_at" firestore:"updated_at"`
}

// Entity Interface implementieren
func (u *User) DocId() string {
    return u.ID
}

func (u *User) SetDocId(id string) {
    u.ID = id
}

func (u *User) UniqFields() map[string]interface{} {
    return map[string]interface{}{
        "email": u.Email,
    }
}
```

## Repository Erstellung

### Firestore Repository

```go
// Repository initialisieren
firestoreClient, err := firestore.NewClient(ctx, "project-id")
if err != nil {
    log.Fatal(err)
}

userRepo := repository.NewFirebaseRepository[User, User](firestoreClient, "User")
```

### Repository Interface

```go
type Repository[T Entity, TT any] interface {
    GetClient() *firestore.Client
    Get(ctx context.Context, opts *query.QueryOptions) (*PaginationResult[T], error)
    GetByID(ctx context.Context, id string) (*T, error)
    Create(ctx context.Context, obj T) (*string, error)
    CreateEasy(ctx context.Context, obj T) (*string, error)
    CreateQueryNotExists(ctx context.Context, obj T, funcQuery func(firestore.Query) firestore.Query) (*string, error)
    Update(ctx context.Context, id string, data map[string]interface{}) error
    Delete(ctx context.Context, id string) error
}
```

## CRUD Operationen

### Create (Erstellen)

```go
// Mit Duplikatsprüfung basierend auf UniqFields()
func CreateUser(ctx context.Context, userRepo repository.Repository[User, User]) {
    user := &User{
        Email:  "user@example.com",
        Name:   "John Doe",
        Active: true,
    }
    
    // Automatische Duplikatsprüfung und Timestamp-Setzung
    docID, err := userRepo.Create(ctx, *user)
    if err != nil {
        // Könnte ErrorAlreadyExists sein wenn Email bereits existiert
        log.Printf("Create failed: %v", err)
        return
    }
    
    log.Printf("User created with ID: %s", *docID)
}
```

### CreateEasy (Ohne Duplikatsprüfung)

```go
// Schnelle Erstellung ohne Duplikatsprüfung
func CreateUserEasy(ctx context.Context, userRepo repository.Repository[User, User]) {
    user := &User{
        Email:    "user@example.com",
        Name:     "John Doe",
        Active:   true,
    }
    
    docID, err := userRepo.CreateEasy(ctx, *user)
    if err != nil {
        log.Printf("CreateEasy failed: %v", err)
        return
    }
    
    log.Printf("User created with ID: %s", *docID)
}
```

### CreateQueryNotExists (Custom Query)

```go
// Erstellung mit benutzerdefinierter Duplikatsprüfung
func CreateUserWithCustomCheck(ctx context.Context, userRepo repository.Repository[User, User]) {
    user := &User{
        Email:    "user@example.com",
        Name:     "John Doe",
        Active:   true,
    }
    
    // Custom Query für Duplikatsprüfung
    checkQuery := func(q firestore.Query) firestore.Query {
        return q.Where("email", "==", user.Email).Where("active", "==", true)
    }
    
    docID, err := userRepo.CreateQueryNotExists(ctx, *user, checkQuery)
    if err != nil {
        log.Printf("CreateQueryNotExists failed: %v", err)
        return
    }
    
    log.Printf("User created with ID: %s", *docID)
}
```

### Read (Lesen)

```go
// Einzelnes Dokument laden
func GetUser(ctx context.Context, userRepo repository.Repository[User, User], userID string) {
    user, err := userRepo.GetByID(ctx, userID)
    if err != nil {
        log.Printf("GetByID failed: %v", err)
        return
    }
    
    log.Printf("User: %+v", user)
}

// Paginierte Liste laden
func GetUsers(ctx context.Context, userRepo repository.Repository[User, User]) {
    queryOpts := &query.QueryOptions{
        Limit:            20,
        OrderBy:          "created_at",
        OrderByDirection: query.Desc,
        Filters: []query.Filter{
            {Field: "active", Operator: query.Eq, Value: true},
        },
    }
    
    result, err := userRepo.Get(ctx, queryOpts)
    if err != nil {
        log.Printf("Get failed: %v", err)
        return
    }
    
    log.Printf("Found %d users", len(result.Items))
    log.Printf("Next page token: %s", result.Next)
}
```

### Update (Aktualisieren)

```go
// Felder aktualisieren
func UpdateUser(ctx context.Context, userRepo repository.Repository[User, User], userID string) {
    updates := map[string]interface{}{
        "name":       "Jane Doe",
        "updated_at": time.Now(),
        "active":     false,
    }
    
    err := userRepo.Update(ctx, userID, updates)
    if err != nil {
        log.Printf("Update failed: %v", err)
        return
    }
    
    log.Printf("User updated successfully")
}
```

### Delete (Löschen)

```go
// Dokument löschen
func DeleteUser(ctx context.Context, userRepo repository.Repository[User, User], userID string) {
    err := userRepo.Delete(ctx, userID)
    if err != nil {
        log.Printf("Delete failed: %v", err)
        return
    }
    
    log.Printf("User deleted successfully")
}
```

## Paginierung

### PaginationResult Struktur

```go
type PaginationResult[T any] struct {
    Items   []T             `json:"items"`    // Gefundene Elemente
    Limit   int             `json:"limit"`    // Anzahl pro Seite
    Next    string          `json:"next"`     // Token für nächste Seite
    Prev    string          `json:"prev"`     // Token für vorherige Seite
    Filters *[]query.Filter `json:"filters,omitempty"` // Angewandte Filter
}
```

### Paginierung verwenden

```go
// Erste Seite laden
func GetFirstPage(ctx context.Context, userRepo repository.Repository[User, User]) *repository.PaginationResult[User] {
    queryOpts := &query.QueryOptions{
        Limit:   20,
        OrderBy: "created_at",
        OrderByDirection: query.Desc,
    }
    
    result, err := userRepo.Get(ctx, queryOpts)
    if err != nil {
        log.Printf("Error: %v", err)
        return nil
    }
    
    return result
}

// Nächste Seite laden
func GetNextPage(ctx context.Context, userRepo repository.Repository[User, User], nextToken string) *repository.PaginationResult[User] {
    queryOpts := &query.QueryOptions{
        Limit:   20,
        Next:    nextToken,  // Token von vorheriger Seite
        OrderBy: "created_at",
        OrderByDirection: query.Desc,
    }
    
    result, err := userRepo.Get(ctx, queryOpts)
    if err != nil {
        log.Printf("Error: %v", err)
        return nil
    }
    
    return result
}
```

## Service Layer Integration

### User Service Beispiel

```go
type UserService struct {
    repo repository.Repository[User, User]
}

func NewUserService(firestoreClient *firestore.Client) *UserService {
    return &UserService{
        repo: repository.NewFirebaseRepository[User, User](firestoreClient, "User"),
    }
}

func (s *UserService) CreateUser(ctx context.Context, email, name string) (*User, error) {
    user := &User{
        Email:     email,
        Name:      name,
        Active:    true,
    }
    
    docID, err := s.repo.Create(ctx, *user)
    if err != nil {
        return nil, fmt.Errorf("failed to create user: %w", err)
    }
    
    // User mit ID zurückgeben
    user.SetDocId(*docID)
    return user, nil
}

func (s *UserService) GetUserByID(ctx context.Context, userID string) (*User, error) {
    user, err := s.repo.GetByID(ctx, userID)
    if err != nil {
        return nil, fmt.Errorf("failed to get user: %w", err)
    }
    
    return user, nil
}

func (s *UserService) SearchUsers(ctx context.Context, filters []query.Filter, limit int) (*repository.PaginationResult[User], error) {
    queryOpts := &query.QueryOptions{
        Filters: filters,
        Limit:   limit,
        OrderBy: "created_at",
        OrderByDirection: query.Desc,
    }
    
    result, err := s.repo.Get(ctx, queryOpts)
    if err != nil {
        return nil, fmt.Errorf("failed to search users: %w", err)
    }
    
    return result, nil
}

func (s *UserService) UpdateUserStatus(ctx context.Context, userID string, active bool) error {
    updates := map[string]interface{}{
        "active":     active,
        "updated_at": time.Now(),
    }
    
    err := s.repo.Update(ctx, userID, updates)
    if err != nil {
        return fmt.Errorf("failed to update user status: %w", err)
    }
    
    return nil
}
```

## HTTP Handler Integration

### REST API mit Repository

```go
type UserHandler struct {
    service *UserService
}

func (h *UserHandler) CreateUser(c *gin.Context) {
    var req struct {
        Email string `json:"email" binding:"required,email"`
        Name  string `json:"name" binding:"required"`
    }
    
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": "Invalid request"})
        return
    }
    
    user, err := h.service.CreateUser(c.Request.Context(), req.Email, req.Name)
    if err != nil {
        // Error-System Integration
        errorResponse, statusCode := errors.NewErrorResponse(err)
        c.JSON(statusCode, errorResponse)
        return
    }
    
    c.JSON(201, user)
}

func (h *UserHandler) GetUsers(c *gin.Context) {
    // Query-System Integration
    queryOpts, err := query.NewQueryOptionsFromUrl(c.Request.URL)
    if err != nil {
        c.JSON(400, gin.H{"error": "Invalid query parameters"})
        return
    }
    
    result, err := h.service.repo.Get(c.Request.Context(), &queryOpts)
    if err != nil {
        c.JSON(500, gin.H{"error": "Failed to fetch users"})
        return
    }
    
    c.JSON(200, result)
}

func (h *UserHandler) GetUser(c *gin.Context) {
    userID := c.Param("id")
    
    user, err := h.service.GetUserByID(c.Request.Context(), userID)
    if err != nil {
        errorResponse, statusCode := errors.NewErrorResponse(err)
        c.JSON(statusCode, errorResponse)
        return
    }
    
    c.JSON(200, user)
}

func (h *UserHandler) UpdateUser(c *gin.Context) {
    userID := c.Param("id")
    
    var updates map[string]interface{}
    if err := c.ShouldBindJSON(&updates); err != nil {
        c.JSON(400, gin.H{"error": "Invalid request"})
        return
    }
    
    // Timestamp automatisch hinzufügen
    updates["updated_at"] = time.Now()
    
    err := h.service.repo.Update(c.Request.Context(), userID, updates)
    if err != nil {
        errorResponse, statusCode := errors.NewErrorResponse(err)
        c.JSON(statusCode, errorResponse)
        return
    }
    
    c.JSON(200, gin.H{"message": "User updated successfully"})
}

func (h *UserHandler) DeleteUser(c *gin.Context) {
    userID := c.Param("id")
    
    err := h.service.repo.Delete(c.Request.Context(), userID)
    if err != nil {
        errorResponse, statusCode := errors.NewErrorResponse(err)
        c.JSON(statusCode, errorResponse)
        return
    }
    
    c.JSON(200, gin.H{"message": "User deleted successfully"})
}
```

## Erweiterte Features

### Automatisches Timestamp-Management

Das Repository setzt automatisch `CreatedAt` und `UpdatedAt` Felder:

```go
type User struct {
    ID        string    `firestore:"-"`
    Email     string    `firestore:"email"`
    Name      string    `firestore:"name"`
    CreatedAt time.Time `firestore:"created_at"`  // Automatisch gesetzt bei Create
    UpdatedAt time.Time `firestore:"updated_at"`  // Automatisch gesetzt bei Create
}
```

### Complex Queries

```go
// Erweiterte Query mit mehreren Filtern
func FindActiveUsersCreatedAfter(ctx context.Context, userRepo repository.Repository[User, User], after time.Time) (*repository.PaginationResult[User], error) {
    queryOpts := &query.QueryOptions{
        Filters: []query.Filter{
            {Field: "active", Operator: query.Eq, Value: true},
            {Field: "created_at", Operator: query.Gte, Value: after},
        },
        OrderBy:          "created_at",
        OrderByDirection: query.Desc,
        Limit:           50,
    }
    
    return userRepo.Get(ctx, queryOpts)
}
```

### Transaktionale Operationen

```go
// Firestore-Client für manuelle Transaktionen
func TransferUserData(ctx context.Context, userRepo repository.Repository[User, User], fromID, toID string) error {
    client := userRepo.GetClient()
    
    return client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
        // User A laden
        fromDoc, err := tx.Get(client.Collection("users").Doc(fromID))
        if err != nil {
            return err
        }
        
        // User B laden
        toDoc, err := tx.Get(client.Collection("users").Doc(toID))
        if err != nil {
            return err
        }
        
        // Beide User aktualisieren
        tx.Update(fromDoc.Ref, []firestore.Update{
            {Path: "status", Value: "transferred"},
            {Path: "updated_at", Value: time.Now()},
        })
        
        tx.Update(toDoc.Ref, []firestore.Update{
            {Path: "data_received", Value: true},
            {Path: "updated_at", Value: time.Now()},
        })
        
        return nil
    })
}
```

## Best Practices

### 1. Entity Design

```go
// Gut - Klare Trennung von ID und Daten
type User struct {
    ID        string    `json:"id" firestore:"-"`           // Firestore ID
    Email     string    `json:"email" firestore:"email"`    // Eindeutig
    Name      string    `json:"name" firestore:"name"`
    Profile   Profile   `json:"profile" firestore:"profile"` // Nested struct
    CreatedAt time.Time `json:"created_at" firestore:"created_at"`
    UpdatedAt time.Time `json:"updated_at" firestore:"updated_at"`
}

// Schlecht - ID im Daten-Struct
type BadUser struct {
    UserID string `firestore:"user_id"` // Redundant zu Firestore Doc ID
    // ...
}
```

### 2. Error Handling

```go
func (s *UserService) GetUser(ctx context.Context, userID string) (*User, error) {
    if userID == "" {
        return nil, &errors.ErrorBadRequest{
            ErrorDetail: errors.ErrorDetail{
                Resource: "User",
                Field:    "id", 
                Value:    "",
                Message:  "User ID is required",
            },
        }
    }
    
    user, err := s.repo.GetByID(ctx, userID)
    if err != nil {
        // Firestore errors zu custom errors konvertieren
        if status.Code(err) == codes.NotFound {
            return nil, &errors.ErrorNotFound{
                ErrorDetail: errors.ErrorDetail{
                    Resource: "User",
                    Field:    "id",
                    Value:    userID,
                    Message:  fmt.Sprintf("User with ID %s not found", userID),
                },
            }
        }
        return nil, err
    }
    
    return user, nil
}
```

### 3. Validation

```go
func (u *User) Validate() error {
    if u.Email == "" {
        return &errors.ErrorBadRequest{
            ErrorDetail: errors.ErrorDetail{
                Resource: "User",
                Field:    "email",
                Value:    "",
                Message:  "Email is required",
            },
        }
    }
    
    if !isValidEmail(u.Email) {
        return &errors.ErrorBadRequest{
            ErrorDetail: errors.ErrorDetail{
                Resource: "User",
                Field:    "email",
                Value:    u.Email,
                Message:  "Invalid email format",
            },
        }
    }
    
    return nil
}
```

## Testing

```go
func TestUserRepository(t *testing.T) {
    // Firestore Emulator für Tests
    ctx := context.Background()
    client, err := firestore.NewClient(ctx, "test-project")
    assert.NoError(t, err)
    defer client.Close()
    
    repo := repository.NewFirebaseRepository[User, User](client, "TestUser")
    
    // Test Create
    user := &User{
        Email: "test@example.com",
        Name:  "Test User",
        Active: true,
    }
    
    docID, err := repo.Create(ctx, *user)
    assert.NoError(t, err)
    assert.NotEmpty(t, *docID)
    
    // Test GetByID
    fetchedUser, err := repo.GetByID(ctx, *docID)
    assert.NoError(t, err)
    assert.Equal(t, user.Email, fetchedUser.Email)
    assert.Equal(t, *docID, fetchedUser.DocId())
    
    // Test Update
    updates := map[string]interface{}{
        "name": "Updated Name",
    }
    err = repo.Update(ctx, *docID, updates)
    assert.NoError(t, err)
    
    // Test Delete
    err = repo.Delete(ctx, *docID)
    assert.NoError(t, err)
    
    // Verify deletion
    _, err = repo.GetByID(ctx, *docID)
    assert.Error(t, err)
}
```