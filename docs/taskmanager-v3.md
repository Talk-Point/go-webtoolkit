# TaskManager v3 (Empfohlen)

TaskManager v3 ist die neueste und empfohlene Version für Google Cloud Tasks Integration. Sie bietet eine verbesserte API mit funktionalen Optionen, erweiterte Queue-Konfiguration und bessere Error-Behandlung.

## Features

- ✅ Verbesserte API mit funktionalen Optionen
- ✅ Erweiterte Queue-Konfiguration (Rate Limits, Concurrency, Retries)
- ✅ Delay-Unterstützung für zeitgesteuerte Tasks
- ✅ Automatische Queue-Erstellung bei Bedarf
- ✅ Bessere Error-Behandlung und Logging
- ✅ Vollständige Queue-Management-Operationen
- ✅ Test-Mode für Entwicklung

## Installation

```go
import "github.com/Talk-Point/go-webtoolkit/pkg/taskmanager/v3/taskmanager"
```

## Grundlegende Verwendung

### TaskManager initialisieren

```go
// Standard-Konfiguration
tm, err := taskmanager.NewTaskManager(&taskmanager.TaskManagerOptions{
    Project:  "my-gcp-project",
    Location: "europe-west1",
    AuthKey:  "your-api-key",
    BaseUrl:  "https://api.example.com",
})
if err != nil {
    log.Fatal(err)
}
defer tm.Close()

// Mit erweiterten Queue-Optionen
tm, err := taskmanager.NewTaskManager(&taskmanager.TaskManagerOptions{
    Project:  "my-gcp-project",
    Location: "europe-west1",
    AuthKey:  "your-api-key",
    BaseUrl:  "https://api.example.com",
    QueueOpts: &taskmanager.QueueOptions{
        MaxDispatchesPerSecond:  10.0,  // 10 Tasks pro Sekunde
        MaxConcurrentDispatches: 5,     // 5 parallele Ausführungen
        MaxAttempts:            3,      // 3 Wiederholungsversuche
    },
})
```

### Task ausführen

```go
// Einfache Task-Ausführung
err := tm.Run("user-notifications", "/api/webhooks/send-welcome-email")
if err != nil {
    log.Printf("Failed to run task: %v", err)
}

// Mit Payload
payload := []byte(`{"user_id": "123", "action": "welcome_email"}`)
err := tm.Run("user-notifications", "/api/webhooks/process-user",
    taskmanager.WithPayload(payload),
)

// Mit allen Optionen
err := tm.Run("user-notifications", "/api/webhooks/delayed-task",
    taskmanager.WithPayload(payload),
    taskmanager.WithMethod("POST"),
    taskmanager.WithDelay(5*time.Minute),
)
```

## Task-Optionen

### Verfügbare Optionen

```go
// Payload setzen
taskmanager.WithPayload([]byte(`{"key": "value"}`))

// HTTP-Method setzen (Standard: POST)
taskmanager.WithMethod("PUT")

// Delay für zeitgesteuerte Ausführung
taskmanager.WithDelay(30*time.Minute)
```

### Funktionale Optionen verwenden

```go
func ScheduleUserWelcome(tm *taskmanager.TaskManager, userID, email string) error {
    payload := map[string]interface{}{
        "user_id":   userID,
        "email":     email,
        "timestamp": time.Now().Format(time.RFC3339),
    }
    
    payloadBytes, err := json.Marshal(payload)
    if err != nil {
        return err
    }
    
    return tm.Run("user-notifications", "/api/webhooks/welcome-email",
        taskmanager.WithPayload(payloadBytes),
        taskmanager.WithMethod("POST"),
    )
}

func ScheduleDelayedBilling(tm *taskmanager.TaskManager, userID string, delay time.Duration) error {
    payload := map[string]interface{}{
        "user_id":    userID,
        "action":     "billing_reminder",
        "scheduled":  time.Now().Add(delay).Format(time.RFC3339),
    }
    
    payloadBytes, _ := json.Marshal(payload)
    
    return tm.Run("billing-tasks", "/api/webhooks/billing-reminder",
        taskmanager.WithPayload(payloadBytes),
        taskmanager.WithDelay(delay),
    )
}
```

## Queue-Management

### Queue erstellen

```go
// Queue mit Standard-Konfiguration
err := tm.CreateQueue("notifications")
if err != nil {
    log.Printf("Failed to create queue: %v", err)
}

// Queue wird automatisch bei Run() erstellt wenn nicht vorhanden
err := tm.Run("auto-created-queue", "/webhook")
// Queue wird automatisch erstellt falls sie nicht existiert
```

### Queue-Operationen

```go
// Queue pausieren
err := tm.PauseQueue("notifications")
if err != nil {
    log.Printf("Failed to pause queue: %v", err)
}

// Queue fortsetzen
err := tm.ResumeQueue("notifications")
if err != nil {
    log.Printf("Failed to resume queue: %v", err)
}

// Queue löschen
err := tm.DeleteQueue("notifications")
if err != nil {
    log.Printf("Failed to delete queue: %v", err)
}

// Queue-Details abrufen
queue, err := tm.GetQueue("notifications")
if err != nil {
    log.Printf("Failed to get queue: %v", err)
} else {
    log.Printf("Queue: %+v", queue)
}

// Alle Queues auflisten
queues, err := tm.ListQueues()
if err != nil {
    log.Printf("Failed to list queues: %v", err)
} else {
    for _, queue := range queues {
        log.Printf("Queue: %s", queue.Name)
    }
}

// Queue leeren
err := tm.PurgeQueue("notifications")
if err != nil {
    log.Printf("Failed to purge queue: %v", err)
}
```

### Queue-Konfiguration aktualisieren

```go
// Queue-Einstellungen ändern
newOpts := &taskmanager.QueueOptions{
    MaxDispatchesPerSecond:  20.0,  // Erhöhe auf 20 Tasks/s
    MaxConcurrentDispatches: 10,    // Erhöhe Parallelität auf 10
    MaxAttempts:            5,      // Mehr Wiederholungsversuche
}

err := tm.UpdateQueue("notifications", newOpts)
if err != nil {
    log.Printf("Failed to update queue: %v", err)
}
```

## Service Integration

### User Service mit v3

```go
type UserService struct {
    taskManager *taskmanager.TaskManager
    userRepo    UserRepository
}

func NewUserService(tm *taskmanager.TaskManager, repo UserRepository) *UserService {
    return &UserService{
        taskManager: tm,
        userRepo:    repo,
    }
}

func (s *UserService) CreateUser(ctx context.Context, user *User) error {
    // User in Datenbank erstellen
    err := s.userRepo.Create(ctx, user)
    if err != nil {
        return err
    }
    
    // Welcome-Email sofort senden
    err = s.scheduleWelcomeEmail(user.ID, user.Email)
    if err != nil {
        log.Printf("Failed to schedule welcome email: %v", err)
        // Nicht kritisch - User wurde erfolgreich erstellt
    }
    
    // Follow-up Email nach 3 Tagen
    err = s.scheduleFollowupEmail(user.ID, user.Email, 3*24*time.Hour)
    if err != nil {
        log.Printf("Failed to schedule followup email: %v", err)
    }
    
    return nil
}

func (s *UserService) scheduleWelcomeEmail(userID, email string) error {
    payload := map[string]interface{}{
        "user_id": userID,
        "email":   email,
        "type":    "welcome",
    }
    
    payloadBytes, err := json.Marshal(payload)
    if err != nil {
        return err
    }
    
    return s.taskManager.Run("email-notifications", "/api/webhooks/send-email",
        taskmanager.WithPayload(payloadBytes),
    )
}

func (s *UserService) scheduleFollowupEmail(userID, email string, delay time.Duration) error {
    payload := map[string]interface{}{
        "user_id": userID,
        "email":   email,
        "type":    "followup",
    }
    
    payloadBytes, err := json.Marshal(payload)
    if err != nil {
        return err
    }
    
    return s.taskManager.Run("email-notifications", "/api/webhooks/send-email",
        taskmanager.WithPayload(payloadBytes),
        taskmanager.WithDelay(delay),
    )
}

func (s *UserService) ScheduleBillingReminder(userID string, reminderDate time.Time) error {
    delay := time.Until(reminderDate)
    if delay < 0 {
        return errors.New("reminder date is in the past")
    }
    
    payload := map[string]interface{}{
        "user_id":      userID,
        "type":         "billing_reminder",
        "scheduled_at": reminderDate.Format(time.RFC3339),
    }
    
    payloadBytes, _ := json.Marshal(payload)
    
    return s.taskManager.Run("billing-tasks", "/api/webhooks/billing-reminder",
        taskmanager.WithPayload(payloadBytes),
        taskmanager.WithDelay(delay),
    )
}
```

## HTTP Handler Integration

### Webhook Handler für v3

```go
type WebhookHandler struct {
    userService  *UserService
    emailService *EmailService
}

func (h *WebhookHandler) SendEmail(c *gin.Context) {
    // API-Key validieren
    apiKey := c.GetHeader("X-API-Key")
    if apiKey != os.Getenv("WEBHOOK_API_KEY") {
        c.JSON(401, gin.H{"error": "Unauthorized"})
        return
    }
    
    var payload struct {
        UserID string `json:"user_id"`
        Email  string `json:"email"`
        Type   string `json:"type"`
    }
    
    if err := c.ShouldBindJSON(&payload); err != nil {
        c.JSON(400, gin.H{"error": "Invalid payload"})
        return
    }
    
    // Email basierend auf Typ senden
    var err error
    switch payload.Type {
    case "welcome":
        err = h.emailService.SendWelcomeEmail(payload.UserID, payload.Email)
    case "followup":
        err = h.emailService.SendFollowupEmail(payload.UserID, payload.Email)
    default:
        c.JSON(400, gin.H{"error": "Unknown email type"})
        return
    }
    
    if err != nil {
        log.Printf("Failed to send %s email to %s: %v", payload.Type, payload.Email, err)
        c.JSON(500, gin.H{"error": "Failed to send email"})
        return
    }
    
    c.JSON(200, gin.H{
        "status": "success",
        "type":   payload.Type,
        "email":  payload.Email,
    })
}

func (h *WebhookHandler) BillingReminder(c *gin.Context) {
    var payload struct {
        UserID      string `json:"user_id"`
        Type        string `json:"type"`
        ScheduledAt string `json:"scheduled_at"`
    }
    
    if err := c.ShouldBindJSON(&payload); err != nil {
        c.JSON(400, gin.H{"error": "Invalid payload"})
        return
    }
    
    // Billing-Erinnerung verarbeiten
    err := h.userService.ProcessBillingReminder(payload.UserID)
    if err != nil {
        log.Printf("Failed to process billing reminder for user %s: %v", payload.UserID, err)
        c.JSON(500, gin.H{"error": "Failed to process billing reminder"})
        return
    }
    
    c.JSON(200, gin.H{
        "status":     "processed",
        "user_id":    payload.UserID,
        "processed_at": time.Now().Format(time.RFC3339),
    })
}
```

## Batch Operations

### Bulk Task Creation

```go
func (s *UserService) SendBulkNotifications(userIDs []string, message string) error {
    for _, userID := range userIDs {
        payload := map[string]interface{}{
            "user_id": userID,
            "message": message,
            "type":    "notification",
        }
        
        payloadBytes, err := json.Marshal(payload)
        if err != nil {
            log.Printf("Failed to marshal payload for user %s: %v", userID, err)
            continue
        }
        
        // Task für jeden User erstellen
        err = s.taskManager.Run("notifications", "/api/webhooks/send-notification",
            taskmanager.WithPayload(payloadBytes),
        )
        if err != nil {
            log.Printf("Failed to schedule notification for user %s: %v", userID, err)
            continue
        }
    }
    
    return nil
}
```

### Scheduled Batch Processing

```go
func (s *UserService) ScheduleDailyReports() error {
    // Täglich um 9:00 Uhr ausführen
    now := time.Now()
    nextRun := time.Date(now.Year(), now.Month(), now.Day()+1, 9, 0, 0, 0, now.Location())
    delay := time.Until(nextRun)
    
    payload := map[string]interface{}{
        "type":        "daily_report",
        "scheduled_at": nextRun.Format(time.RFC3339),
    }
    
    payloadBytes, _ := json.Marshal(payload)
    
    return s.taskManager.Run("reports", "/api/webhooks/generate-daily-report",
        taskmanager.WithPayload(payloadBytes),
        taskmanager.WithDelay(delay),
    )
}
```

## Test Mode

### Development Setup

```go
// Test Mode aktivieren
os.Setenv("TEST_MODE", "true")

tm, err := taskmanager.NewTaskManager(&taskmanager.TaskManagerOptions{
    Project:  "test-project",
    Location: "test-location",
    AuthKey:  "test-key",
    BaseUrl:  "http://localhost:8080",
})

// Tasks werden nur geloggt, nicht an Cloud Tasks gesendet
err = tm.Run("test-queue", "/test-webhook",
    taskmanager.WithPayload([]byte(`{"test": true}`)),
)
// Output: Log-Message mit Task-Details
```

### Testing Utilities

```go
func TestUserServiceTasks(t *testing.T) {
    // Test Environment setup
    os.Setenv("TEST_MODE", "true")
    
    tm, err := taskmanager.NewTaskManager(&taskmanager.TaskManagerOptions{
        Project:  "test",
        Location: "test",
        AuthKey:  "test",
        BaseUrl:  "http://localhost",
    })
    assert.NoError(t, err)
    defer tm.Close()
    
    userService := NewUserService(tm, mockUserRepo)
    
    user := &User{
        ID:    "123",
        Email: "test@example.com",
        Name:  "Test User",
    }
    
    // User erstellen - Tasks werden nur geloggt
    err = userService.CreateUser(context.Background(), user)
    assert.NoError(t, err)
    
    // Verify dass Tasks geplant wurden (via Logs)
    // In Test Mode werden keine echten Cloud Tasks erstellt
}
```

## Erweiterte Konfiguration

### Environment-basierte Konfiguration

```go
type Config struct {
    TaskManager TaskManagerConfig `mapstructure:"task_manager"`
}

type TaskManagerConfig struct {
    Project   string       `mapstructure:"project"`
    Location  string       `mapstructure:"location"`
    BaseUrl   string       `mapstructure:"base_url"`
    AuthKey   string       `mapstructure:"auth_key"`
    QueueOpts QueueOptions `mapstructure:"queue_options"`
}

type QueueOptions struct {
    MaxDispatchesPerSecond  float64 `mapstructure:"max_dispatches_per_second"`
    MaxConcurrentDispatches int32   `mapstructure:"max_concurrent_dispatches"`
    MaxAttempts            int32   `mapstructure:"max_attempts"`
}

func NewTaskManagerFromConfig(config TaskManagerConfig) (*taskmanager.TaskManager, error) {
    return taskmanager.NewTaskManager(&taskmanager.TaskManagerOptions{
        Project:   config.Project,
        Location:  config.Location,
        BaseUrl:   config.BaseUrl,
        AuthKey:   config.AuthKey,
        QueueOpts: &taskmanager.QueueOptions{
            MaxDispatchesPerSecond:  config.QueueOpts.MaxDispatchesPerSecond,
            MaxConcurrentDispatches: config.QueueOpts.MaxConcurrentDispatches,
            MaxAttempts:            config.QueueOpts.MaxAttempts,
        },
    })
}
```

### YAML-Konfiguration

```yaml
task_manager:
  project: "my-gcp-project"
  location: "europe-west1"
  base_url: "https://api.example.com"
  auth_key: "${WEBHOOK_API_KEY}"
  queue_options:
    max_dispatches_per_second: 10.0
    max_concurrent_dispatches: 5
    max_attempts: 3
```

## Monitoring und Observability

### Custom Logging

```go
type TaskManagerWrapper struct {
    tm     *taskmanager.TaskManager
    logger *log.Logger
}

func (w *TaskManagerWrapper) Run(queue, path string, options ...taskmanager.TaskOption) error {
    start := time.Now()
    
    err := w.tm.Run(queue, path, options...)
    
    duration := time.Since(start)
    
    if err != nil {
        w.logger.WithFields(log.Fields{
            "queue":    queue,
            "path":     path,
            "duration": duration,
            "error":    err,
        }).Error("Task scheduling failed")
    } else {
        w.logger.WithFields(log.Fields{
            "queue":    queue,
            "path":     path,
            "duration": duration,
        }).Info("Task scheduled successfully")
    }
    
    return err
}
```

### Metrics Collection

```go
type TaskMetrics struct {
    TasksScheduled prometheus.Counter
    TasksDuration  prometheus.Histogram
    TasksErrors    prometheus.Counter
}

func (m *TaskMetrics) RecordTask(queue, path string, duration time.Duration, err error) {
    m.TasksScheduled.WithLabelValues(queue, path).Inc()
    m.TasksDuration.WithLabelValues(queue, path).Observe(duration.Seconds())
    
    if err != nil {
        m.TasksErrors.WithLabelValues(queue, path).Inc()
    }
}
```

## Best Practices

### 1. Queue-Organisation

```go
// Gut - Logische Queue-Gruppierung
const (
    QueueEmail        = "email-notifications"
    QueueBilling      = "billing-tasks"
    QueueDataExport   = "data-export"
    QueueImageProcess = "image-processing"
)

// Verschiedene Konfigurationen pro Queue-Typ
func ConfigureQueues(tm *taskmanager.TaskManager) error {
    // Schnelle Email-Queue
    emailOpts := &taskmanager.QueueOptions{
        MaxDispatchesPerSecond:  50.0,
        MaxConcurrentDispatches: 20,
        MaxAttempts:            3,
    }
    
    // Langsamere Billing-Queue (wegen externer APIs)
    billingOpts := &taskmanager.QueueOptions{
        MaxDispatchesPerSecond:  5.0,
        MaxConcurrentDispatches: 2,
        MaxAttempts:            5,
    }
    
    // Setup queues...
    return nil
}
```

### 2. Error Handling und Retries

```go
func (s *UserService) scheduleTaskWithRetry(queue, path string, payload []byte, maxRetries int) error {
    var lastErr error
    
    for i := 0; i < maxRetries; i++ {
        err := s.taskManager.Run(queue, path,
            taskmanager.WithPayload(payload),
        )
        if err == nil {
            return nil
        }
        
        lastErr = err
        
        // Exponential backoff
        waitTime := time.Duration(math.Pow(2, float64(i))) * time.Second
        time.Sleep(waitTime)
        
        log.Printf("Task scheduling attempt %d failed: %v", i+1, err)
    }
    
    return fmt.Errorf("failed to schedule task after %d attempts: %w", maxRetries, lastErr)
}
```

### 3. Payload Design

```go
// Standard Task Payload Structure
type TaskPayload struct {
    Type      string                 `json:"type"`
    UserID    string                 `json:"user_id,omitempty"`
    Data      map[string]interface{} `json:"data"`
    Metadata  TaskMetadata          `json:"metadata"`
}

type TaskMetadata struct {
    ScheduledAt string `json:"scheduled_at"`
    Source      string `json:"source"`
    TraceID     string `json:"trace_id,omitempty"`
}

func NewTaskPayload(taskType, userID string, data map[string]interface{}) []byte {
    payload := TaskPayload{
        Type:   taskType,
        UserID: userID,
        Data:   data,
        Metadata: TaskMetadata{
            ScheduledAt: time.Now().Format(time.RFC3339),
            Source:      "user-service",
            TraceID:     generateTraceID(),
        },
    }
    
    bytes, _ := json.Marshal(payload)
    return bytes
}
```

## Performance-Optimierung

### Connection Pooling

```go
// TaskManager wiederverwendbar machen
type TaskManagerPool struct {
    managers map[string]*taskmanager.TaskManager
    mutex    sync.RWMutex
}

func (p *TaskManagerPool) GetManager(project string) (*taskmanager.TaskManager, error) {
    p.mutex.RLock()
    tm, exists := p.managers[project]
    p.mutex.RUnlock()
    
    if exists {
        return tm, nil
    }
    
    // Create new manager
    p.mutex.Lock()
    defer p.mutex.Unlock()
    
    // Double-check
    if tm, exists := p.managers[project]; exists {
        return tm, nil
    }
    
    tm, err := taskmanager.NewTaskManager(&taskmanager.TaskManagerOptions{
        Project:  project,
        Location: "europe-west1",
        // ... other options
    })
    if err != nil {
        return nil, err
    }
    
    p.managers[project] = tm
    return tm, nil
}
```

## Migration von v2

### API-Unterschiede

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

### Migration Utility

```go
func MigrateToV3(oldTm *taskmanagerv2.TaskManager, newTm *taskmanager.TaskManager, task *taskmanagerv2.Task) error {
    var options []taskmanager.TaskOption
    
    if task.Payload != nil {
        options = append(options, taskmanager.WithPayload(task.Payload))
    }
    
    if task.Method != "" {
        options = append(options, taskmanager.WithMethod(task.Method))
    }
    
    return newTm.Run(task.Queue, task.Path, options...)
}
```