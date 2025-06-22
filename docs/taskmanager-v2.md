# TaskManager v2 (Legacy)

TaskManager v2 bietet eine einfache Integration mit Google Cloud Tasks für die Verarbeitung von Hintergrund-Jobs. Diese Version wird als Legacy betrachtet - für neue Projekte wird TaskManager v3 empfohlen.

## Features

- ✅ Einfache Task-Erstellung und -Verwaltung
- ✅ Queue-Management (Erstellen, Pausieren, Fortsetzen, Löschen)
- ✅ Test-Mode für Entwicklung und Testing
- ✅ Automatische Queue-Erstellung bei Bedarf
- ✅ HTTP-basierte Task-Ausführung
- ✅ Integriertes Logging

## Installation

```go
import "github.com/Talk-Point/go-webtoolkit/pkg/taskmanager/v2/taskmanager"
```

## Grundlegende Verwendung

### TaskManager initialisieren

```go
// Für Produktion mit Google Cloud Tasks
options := &taskmanager.Options{
    Project:  "my-gcp-project",
    Location: "europe-west1",
    BaseURL:  "https://api.example.com",
    AuthKey:  "your-api-key",
}

tm := taskmanager.NewCloudTaskManager(options)

// Für Tests
tm := taskmanager.NewTestTaskManager()
```

### Task erstellen und ausführen

```go
// Task-Struktur
task := &taskmanager.Task{
    Method:  "POST",
    Path:    "/api/webhooks/process-user",
    Payload: []byte(`{"user_id": "123", "action": "welcome_email"}`),
    Queue:   "user-notifications",
}

// Task zur Ausführung hinzufügen
err := tm.AddTask(task)
if err != nil {
    log.Printf("Failed to add task: %v", err)
}
```

### Task mit automatischer Queue-Erstellung

```go
// Queue wird automatisch erstellt wenn sie nicht existiert
err := tm.AddTaskAndCreateQueueWhenNotExists(task)
if err != nil {
    log.Printf("Failed to add task: %v", err)
}
```

## Queue-Management

### Queue manuell erstellen

```go
// Queue mit Standard-Konfiguration erstellen
err := tm.QueueCreate("user-notifications")
if err != nil {
    log.Printf("Failed to create queue: %v", err)
}
```

### Queue pausieren/fortsetzen

```go
// Queue pausieren
err := tm.PauseQueue("user-notifications")
if err != nil {
    log.Printf("Failed to pause queue: %v", err)
}

// Queue fortsetzen
err := tm.ResumeQueue("user-notifications")
if err != nil {
    log.Printf("Failed to resume queue: %v", err)
}
```

### Queue löschen

```go
err := tm.QueueRemove("user-notifications")
if err != nil {
    log.Printf("Failed to remove queue: %v", err)
}
```

## Task Handler

### Cloud Task Handler (Standard)

```go
// Automatisch verwendet bei NewCloudTaskManager
// Sendet HTTP-Requests an die konfigurierte BaseURL
func CloudTaskHandler(manager *TaskManager, task Task) error {
    // Erstellt HTTP POST Request an BaseURL + task.Path
    // Mit Headers: Content-Type: application/json, X-API-Key: AuthKey
    // Und Body: task.Payload
}
```

### Test Task Handler

```go
// Verwendet bei NewTestTaskManager
// Loggt nur die Task-Details ohne tatsächliche Ausführung
func TestTaskHandler(manager *TaskManager, task Task) error {
    // Loggt Task-Informationen für Debugging
    return nil
}
```

### Custom Task Handler

```go
// Eigenen Task Handler implementieren
func CustomTaskHandler(manager *taskmanager.TaskManager, task taskmanager.Task) error {
    log.Printf("Processing task: %s %s", task.Method, task.Path)
    
    // Custom Logic hier
    switch task.Path {
    case "/api/webhooks/send-email":
        return sendEmail(task.Payload)
    case "/api/webhooks/process-payment":
        return processPayment(task.Payload)
    default:
        return fmt.Errorf("unknown task path: %s", task.Path)
    }
}

// TaskManager mit Custom Handler
tm := &taskmanager.TaskManager{
    Project:     "my-project",
    Location:    "europe-west1", 
    BaseURL:     "https://api.example.com",
    AuthKey:     "api-key",
    TaskHandler: CustomTaskHandler,
}
```

## Service Integration

### User Service mit Tasks

```go
type UserService struct {
    taskManager *taskmanager.TaskManager
    userRepo    UserRepository
}

func (s *UserService) CreateUser(ctx context.Context, user *User) error {
    // User in Datenbank erstellen
    err := s.userRepo.Create(ctx, user)
    if err != nil {
        return err
    }
    
    // Welcome-Email Task erstellen
    welcomeTask := &taskmanager.Task{
        Method:  "POST",
        Path:    "/api/webhooks/send-welcome-email",
        Payload: []byte(fmt.Sprintf(`{"user_id": "%s", "email": "%s"}`, user.ID, user.Email)),
        Queue:   "email-notifications",
    }
    
    // Task asynchron ausführen
    err = s.taskManager.AddTaskAndCreateQueueWhenNotExists(welcomeTask)
    if err != nil {
        log.Printf("Failed to schedule welcome email: %v", err)
        // Fehler nicht weiterleiten - User wurde erfolgreich erstellt
    }
    
    return nil
}

func (s *UserService) ProcessSubscription(ctx context.Context, userID string, planID string) error {
    // Subscription verarbeiten...
    
    // Billing-Task für später planen
    billingTask := &taskmanager.Task{
        Method:  "POST",
        Path:    "/api/webhooks/process-billing",
        Payload: []byte(fmt.Sprintf(`{"user_id": "%s", "plan_id": "%s", "timestamp": "%s"}`, userID, planID, time.Now().Format(time.RFC3339))),
        Queue:   "billing-tasks",
    }
    
    return s.taskManager.AddTaskAndCreateQueueWhenNotExists(billingTask)
}
```

## HTTP Handler Integration

### Webhook Handler

```go
type WebhookHandler struct {
    userService  *UserService
    emailService *EmailService
}

// Handler für Task-Webhooks
func (h *WebhookHandler) ProcessWelcomeEmail(c *gin.Context) {
    // API-Key validieren
    apiKey := c.GetHeader("X-API-Key")
    if apiKey != os.Getenv("WEBHOOK_API_KEY") {
        c.JSON(401, gin.H{"error": "Unauthorized"})
        return
    }
    
    var payload struct {
        UserID string `json:"user_id"`
        Email  string `json:"email"`
    }
    
    if err := c.ShouldBindJSON(&payload); err != nil {
        c.JSON(400, gin.H{"error": "Invalid payload"})
        return
    }
    
    // Welcome-Email senden
    err := h.emailService.SendWelcomeEmail(payload.UserID, payload.Email)
    if err != nil {
        log.Printf("Failed to send welcome email: %v", err)
        c.JSON(500, gin.H{"error": "Failed to send email"})
        return
    }
    
    c.JSON(200, gin.H{"status": "success"})
}

// Handler für Billing-Tasks
func (h *WebhookHandler) ProcessBilling(c *gin.Context) {
    var payload struct {
        UserID    string `json:"user_id"`
        PlanID    string `json:"plan_id"`
        Timestamp string `json:"timestamp"`
    }
    
    if err := c.ShouldBindJSON(&payload); err != nil {
        c.JSON(400, gin.H{"error": "Invalid payload"})
        return
    }
    
    // Billing verarbeiten
    err := h.userService.ProcessBillingForUser(payload.UserID, payload.PlanID)
    if err != nil {
        log.Printf("Billing failed for user %s: %v", payload.UserID, err)
        c.JSON(500, gin.H{"error": "Billing failed"})
        return
    }
    
    c.JSON(200, gin.H{"status": "billing processed"})
}
```

### Gin Router Setup

```go
func SetupTaskRoutes(router *gin.Engine, webhookHandler *WebhookHandler) {
    webhooks := router.Group("/api/webhooks")
    {
        webhooks.POST("/send-welcome-email", webhookHandler.ProcessWelcomeEmail)
        webhooks.POST("/process-billing", webhookHandler.ProcessBilling)
        webhooks.POST("/process-payment", webhookHandler.ProcessPayment)
        webhooks.POST("/send-notification", webhookHandler.SendNotification)
    }
}
```

## Konfiguration

### Environment Variables

```bash
# Google Cloud Configuration
GOOGLE_CLOUD_PROJECT=my-gcp-project
GOOGLE_APPLICATION_CREDENTIALS=/path/to/service-account.json

# TaskManager Configuration
TASK_MANAGER_LOCATION=europe-west1
TASK_MANAGER_BASE_URL=https://api.example.com
TASK_MANAGER_AUTH_KEY=your-webhook-api-key

# Test Mode
TEST_MODE=true  # Für Entwicklung und Tests
```

### Konfiguration im Code

```go
type TaskManagerConfig struct {
    Project  string
    Location string
    BaseURL  string
    AuthKey  string
    TestMode bool
}

func NewTaskManagerFromConfig(config TaskManagerConfig) *taskmanager.TaskManager {
    if config.TestMode {
        return taskmanager.NewTestTaskManager()
    }
    
    options := &taskmanager.Options{
        Project:  config.Project,
        Location: config.Location,
        BaseURL:  config.BaseURL,
        AuthKey:  config.AuthKey,
    }
    
    return taskmanager.NewCloudTaskManager(options)
}
```

## Test Mode

### Development Setup

```go
// In Tests und Entwicklung
os.Setenv("TEST_MODE", "true")

tm := taskmanager.NewTestTaskManager()

// Tasks werden nur geloggt, nicht ausgeführt
task := &taskmanager.Task{
    Method:  "POST",
    Path:    "/api/webhooks/test",
    Payload: []byte(`{"test": true}`),
    Queue:   "test-queue",
}

err := tm.AddTask(task)
// Loggt: "Task fired" mit Task-Details
```

### Testing Utilities

```go
func TestUserCreation(t *testing.T) {
    // Test TaskManager verwenden
    tm := taskmanager.NewTestTaskManager()
    
    userService := &UserService{
        taskManager: tm,
        userRepo:    mockUserRepo,
    }
    
    user := &User{
        Email: "test@example.com",
        Name:  "Test User",
    }
    
    err := userService.CreateUser(context.Background(), user)
    assert.NoError(t, err)
    
    // In Test Mode werden Tasks nur geloggt
    // Keine echten HTTP-Requests oder Cloud Tasks
}
```

## Queue-Konfiguration

### Standard Queue-Settings

```go
// Standard-Konfiguration für Queues
queue := &taskspb.Queue{
    Name: queuePath,
    RateLimits: &taskspb.RateLimits{
        MaxDispatchesPerSecond:  1,    // 1 Task pro Sekunde
        MaxConcurrentDispatches: 1,    // 1 parallele Ausführung
    },
    RetryConfig: &taskspb.RetryConfig{
        MaxAttempts: 5,                // 5 Wiederholungsversuche
    },
}
```

### Queue-Monitoring

```go
// Queue-Status überwachen (erfordert manuelle Implementation)
func MonitorQueues(tm *taskmanager.TaskManager, queues []string) {
    for _, queueName := range queues {
        // Queue-Status abrufen
        // (Requires additional Google Cloud Tasks API calls)
        
        log.Printf("Monitoring queue: %s", queueName)
        // Custom monitoring logic...
    }
}
```

## Best Practices

### 1. Queue-Naming

```go
// Gut - Beschreibende Queue-Namen
"user-notifications"
"billing-tasks"
"email-delivery"
"image-processing"

// Schlecht - Unklare Namen
"queue1"
"tasks"
"background"
```

### 2. Error Handling

```go
func (s *UserService) scheduleTask(task *taskmanager.Task) {
    err := s.taskManager.AddTaskAndCreateQueueWhenNotExists(task)
    if err != nil {
        // Task-Fehler loggen aber nicht Business-Logic blockieren
        log.WithFields(log.Fields{
            "queue": task.Queue,
            "path":  task.Path,
            "error": err,
        }).Error("Failed to schedule task")
        
        // Optional: Fallback-Mechanismus
        s.scheduleTaskForRetry(task)
    }
}
```

### 3. Payload Design

```go
// Gut - Strukturierte Payloads
type WelcomeEmailPayload struct {
    UserID    string `json:"user_id"`
    Email     string `json:"email"`
    Language  string `json:"language"`
    Timestamp string `json:"timestamp"`
}

payload, _ := json.Marshal(WelcomeEmailPayload{
    UserID:    user.ID,
    Email:     user.Email,
    Language:  user.Language,
    Timestamp: time.Now().Format(time.RFC3339),
})

// Schlecht - Unstrukturierte Strings
payload := fmt.Sprintf("user:%s,email:%s", user.ID, user.Email)
```

## Migration zu v3

TaskManager v3 bietet verbesserte Features. Migration:

```go
// v2 (Alt)
tm := taskmanager.NewCloudTaskManager(&taskmanager.Options{
    Project:  "project",
    Location: "location", 
    BaseURL:  "url",
    AuthKey:  "key",
})

task := &taskmanager.Task{
    Method:  "POST",
    Path:    "/webhook",
    Payload: payload,
    Queue:   "queue-name",
}

err := tm.AddTask(task)

// v3 (Neu)
tm, err := taskmanager.NewTaskManager(&taskmanager.TaskManagerOptions{
    Project:  "project",
    Location: "location",
    BaseUrl:  "url", 
    AuthKey:  "key",
})

err = tm.Run("queue-name", "/webhook", 
    taskmanager.WithPayload(payload),
    taskmanager.WithMethod("POST"),
)
```

## Fehlerbehebung

### Häufige Probleme

1. **"Queue does not exist" Error**
   ```go
   // Lösung: AddTaskAndCreateQueueWhenNotExists verwenden
   err := tm.AddTaskAndCreateQueueWhenNotExists(task)
   ```

2. **Authentication Errors**
   ```bash
   # Service Account Key setzen
   export GOOGLE_APPLICATION_CREDENTIALS=/path/to/key.json
   ```

3. **Webhook 401 Errors**
   ```go
   // API-Key in Headers prüfen
   apiKey := c.GetHeader("X-API-Key")
   if apiKey != expectedKey {
       c.JSON(401, gin.H{"error": "Unauthorized"})
       return
   }
   ```

## Logging

TaskManager v2 verwendet sirupsen/logrus für Logging:

```go
// Task-Ausführung wird automatisch geloggt
log.WithFields(log.Fields{
    "task":      endpoint,
    "test-mode": "false", 
}).Debug("Task fired")

// Queue-Operationen werden geloggt
log.WithFields(log.Fields{
    "queue": queue,
}).Info("TaskManager Queue created")
```