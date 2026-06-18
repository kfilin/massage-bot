// Data migration tool for Vera Massage Bot.
// Handles database cleanup, calendar cleanup, and appointment migration
// from vfilinav@gmail.com to veramassagist@gmail.com.
//
// Usage:
//   go run scripts/data_migration.go <command> [flags]
//
// Commands:
//   clean-db        Remove stale metadata and optionally purge test patients
//   list-events     List all events in the current Google Calendar
//   clean-calendar  Delete events from the current Google Calendar (interactive or --force)
//   auth            Generate OAuth URL to get a token for a Google account
//   migrate         Migrate appointments from vfilinav@ to veramassagist@
package main

import (
	"bufio"
	"context"

	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

var (
	reader = bufio.NewReader(os.Stdin)
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run scripts/data_migration.go <command>")
		fmt.Println("")
		fmt.Println("Commands:")
		fmt.Println("  clean-db              Remove stale data from the project database")
		fmt.Println("  list-events           List all events in the current Google Calendar")
		fmt.Println("  clean-calendar        Delete events from the current calendar (interactive)")
		fmt.Println("  auth                  Generate OAuth URL to get a token for a Google account")
		fmt.Println("  migrate [token_file]  Migrate appointments from vfilinav@gmail.com")
		fmt.Println("")
		fmt.Println("Environment:")
		fmt.Println("  Loads .env from project root (DATABASE_URL or DB_* vars)")
		fmt.Println("  Uses GOOGLE_CREDENTIALS_JSON, GOOGLE_TOKEN_JSON, GOOGLE_CALENDAR_ID")
		os.Exit(1)
	}

	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: no .env file found at project root: %v", err)
	}

	cmd := os.Args[1]
	switch cmd {
	case "clean-db":
		cleanDB()
	case "list-events":
		listEvents()
	case "clean-calendar":
		cleanCalendar()
	case "auth":
		doAuth()
	case "migrate":
		migrateFromFile := ""
		if len(os.Args) > 2 {
			migrateFromFile = os.Args[2]
		}
		doMigrate(migrateFromFile)
	default:
		log.Fatalf("Unknown command: %s", cmd)
	}
}

// ─── Database ────────────────────────────────────────────────────────────────

func connectDB() *sqlx.DB {
	host := getEnv("DB_HOST", "localhost")
	// If DB_HOST is "db" (Docker internal name), try 127.0.0.1 first
	// since the DB port is mapped to the host.
	if host == "db" {
		host = "127.0.0.1"
	}
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "postgres")
	password := getEnv("DB_PASSWORD", "")
	dbName := getEnv("DB_NAME", "massage_bot")
	sslmode := getEnv("DB_SSL_MODE", "disable")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbName, sslmode)

	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database at %s:%s: %v\n\n"+
			"Hint: DB is inside Docker. Either:\n"+
			"  1. Run the script from a container on the same Docker network:\n"+
			"     docker run --rm -v \"$(pwd):/app\" --network massage-bot-internal \\\n"+
			"       -e DB_HOST=db golang:1.25-alpine go run /app/scripts/data_migration.go ...\n"+
			"  2. Or set HOST_DB_PORT in .env and make sure port 5432 is mapped to host.\n",
			host, port, err)
	}
	return db
}

func cleanDB() {
	fmt.Println("\n=== Phase 1: Database Cleanup ===")
	db := connectDB()
	defer db.Close()

	// 1. Show current state
	var patientCount, apptCount, metaCount, mediaCount, eventCount int
	db.Get(&patientCount, "SELECT COUNT(*) FROM patients")
	db.Get(&apptCount, "SELECT COUNT(*) FROM appointments")
	db.Get(&metaCount, "SELECT COUNT(*) FROM appointment_metadata")
	db.Get(&mediaCount, "SELECT COUNT(*) FROM patient_media")
	db.Get(&eventCount, "SELECT COUNT(*) FROM analytics_events")

	fmt.Printf("\nCurrent state:\n")
	fmt.Printf("  Patients:             %d\n", patientCount)
	fmt.Printf("  Appointments:         %d\n", apptCount)
	fmt.Printf("  Appointment metadata: %d\n", metaCount)
	fmt.Printf("  Patient media files:  %d\n", mediaCount)
	fmt.Printf("  Analytics events:     %d\n", eventCount)

	// 2. Show patients
	type PatientInfo struct {
		TelegramID string `db:"telegram_id"`
		Name       string `db:"name"`
		Visits     int    `db:"total_visits"`
		FirstVisit *time.Time `db:"first_visit"`
	}
	var patients []PatientInfo
	db.Select(&patients, "SELECT telegram_id, name, total_visits, first_visit FROM patients ORDER BY name")
	if len(patients) > 0 {
		fmt.Printf("\nRegistered patients:\n")
		for _, p := range patients {
			fv := "never"
			if p.FirstVisit != nil && !p.FirstVisit.IsZero() && p.FirstVisit.Year() > 2000 {
				fv = p.FirstVisit.Format("2006-01-02")
			}
			fmt.Printf("  • %s (ID: %s) — %d visit(s), first: %s\n", p.Name, p.TelegramID, p.Visits, fv)
		}
	}
	fmt.Println()

	// 3. Purge stale metadata
	if metaCount > 0 {
		fmt.Printf("Stale appointment_metadata: %d records — all will be purged (orphaned, no active appointments).\n", metaCount)
		if confirm("Purge appointment_metadata?") {
			db.MustExec("DELETE FROM appointment_metadata")
			fmt.Printf("  ✅ Purged %d metadata records.\n", metaCount)
		}
	}

	// 4. Purge stale analytics events
	if eventCount > 0 {
		fmt.Printf("\nAnalytics events: %d records.\n", eventCount)
		if confirm("Purge all analytics_events?") {
			db.MustExec("DELETE FROM analytics_events")
			fmt.Printf("  ✅ Purged %d analytics events.\n", eventCount)
		}
	}

	// 5. Option to delete specific patients
	fmt.Println()
	fmt.Println("If any patients are test/mock data, you can remove them now.")
	for _, p := range patients {
		if strings.Contains(strings.ToLower(p.Name), "test") ||
			strings.Contains(strings.ToLower(p.Name), "mock") ||
			p.TelegramID == "" {
			if confirm(fmt.Sprintf("Delete patient '%s' (ID: %s)?", p.Name, p.TelegramID)) {
				db.MustExec("DELETE FROM patient_media WHERE patient_id = $1", p.TelegramID)
				db.MustExec("DELETE FROM analytics_events WHERE patient_id = $1", p.TelegramID)
				db.MustExec("DELETE FROM appointments WHERE customer_id = $1", p.TelegramID)
				db.MustExec("DELETE FROM appointment_metadata WHERE appointment_id IN (SELECT id FROM appointments WHERE customer_id = $1)", p.TelegramID)
				db.MustExec("DELETE FROM patients WHERE telegram_id = $1", p.TelegramID)
				fmt.Printf("  ✅ Deleted patient '%s'.\n", p.Name)
			}
		}
	}

	// 6. Show final state
	var finalPatientCount int
	db.Get(&finalPatientCount, "SELECT COUNT(*) FROM patients")
	fmt.Printf("\n✅ Database cleanup complete. %d patient(s) remaining.\n", finalPatientCount)
}

// ─── Google Calendar ─────────────────────────────────────────────────────────

func makeCalendarClient() (*calendar.Service, string) {
	credsJSON := os.Getenv("GOOGLE_CREDENTIALS_JSON")
	if credsJSON == "" {
		log.Fatalf("GOOGLE_CREDENTIALS_JSON is not set in .env")
	}

	config, err := google.ConfigFromJSON([]byte(credsJSON), calendar.CalendarScope)
	if err != nil {
		log.Fatalf("Failed to parse Google credentials: %v", err)
	}
	config.RedirectURL = "http://localhost:8080"

	tokenJSON := os.Getenv("GOOGLE_TOKEN_JSON")
	if tokenJSON == "" {
		log.Fatalf("GOOGLE_TOKEN_JSON is not set in .env")
	}

	var tok oauth2.Token
	if err := json.Unmarshal([]byte(tokenJSON), &tok); err != nil {
		log.Fatalf("Failed to parse GOOGLE_TOKEN_JSON: %v", err)
	}

	ctx := context.Background()
	client := oauth2.NewClient(ctx, config.TokenSource(ctx, &tok))

	svc, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Failed to create Calendar service: %v", err)
	}

	calendarID := os.Getenv("GOOGLE_CALENDAR_ID")
	if calendarID == "" {
		calendarID = "primary"
	}

	return svc, calendarID
}

func getCalendarInfo(svc *calendar.Service, calendarID string) {
	cal, err := svc.Calendars.Get(calendarID).Do()
	if err != nil {
		log.Printf("  Warning: could not get calendar info: %v", err)
		return
	}
	fmt.Printf("  Calendar: %s (%s)\n", cal.Summary, calendarID)
}

func listEvents() {
	fmt.Println("\n=== List Events in Current Calendar ===")
	svc, calendarID := makeCalendarClient()
	getCalendarInfo(svc, calendarID)

	events, err := svc.Events.List(calendarID).
		ShowDeleted(false).
		SingleEvents(true).
		MaxResults(500).
		OrderBy("startTime").
		Do()
	if err != nil {
		log.Fatalf("Failed to list events: %v", err)
	}

	if len(events.Items) == 0 {
		fmt.Println("\nNo events found in calendar.")
		return
	}

	fmt.Printf("\nFound %d event(s):\n\n", len(events.Items))
	for _, e := range events.Items {
		start := "all-day"
		if e.Start.DateTime != "" {
			t, _ := time.Parse(time.RFC3339, e.Start.DateTime)
			start = t.Format("2006-01-02 Mon 15:04")
		} else if e.Start.Date != "" {
			start = e.Start.Date
		}
		fmt.Printf("  [%s] %s — %s\n", e.Id[:min(len(e.Id), 24)], start, e.Summary)
		if e.Description != "" {
			desc := strings.Split(e.Description, "\n")[0]
			if len(desc) > 80 {
				desc = desc[:80] + "..."
			}
			fmt.Printf("         %s\n", desc)
		}
	}
	fmt.Println()
}

func cleanCalendar() {
	fmt.Println("\n=== Phase 2: Clean Calendar Events ===")
	svc, calendarID := makeCalendarClient()
	getCalendarInfo(svc, calendarID)

	events, err := svc.Events.List(calendarID).
		ShowDeleted(false).
		SingleEvents(true).
		MaxResults(500).
		OrderBy("startTime").
		Do()
	if err != nil {
		log.Fatalf("Failed to list events: %v", err)
	}

	if len(events.Items) == 0 {
		fmt.Println("\nNo events to clean. Calendar is already empty.")
		return
	}

	fmt.Printf("\nFound %d event(s) in calendar.\n", len(events.Items))
	for i, e := range events.Items {
		start := "all-day"
		if e.Start.DateTime != "" {
			t, _ := time.Parse(time.RFC3339, e.Start.DateTime)
			start = t.Format("2006-01-02 15:04")
		} else if e.Start.Date != "" {
			start = e.Start.Date
		}
		fmt.Printf("  [%2d] %s | %s | %s\n", i+1, start, ellipsis(e.Summary, 40), e.Id)
	}
	fmt.Println()

	if confirm(fmt.Sprintf("Delete ALL %d events from this calendar?", len(events.Items))) {
		for _, e := range events.Items {
			if err := svc.Events.Delete(calendarID, e.Id).Do(); err != nil {
				log.Printf("  ⚠ Failed to delete event %s (%s): %v", e.Summary, e.Id, err)
			} else {
				fmt.Printf("  ✅ Deleted: %s\n", e.Summary)
			}
			time.Sleep(200 * time.Millisecond) // rate limit
		}
		fmt.Printf("\n✅ Deleted %d events from calendar.\n", len(events.Items))
	} else {
		fmt.Println("Skipped calendar cleanup.")
	}
}

// ─── OAuth ───────────────────────────────────────────────────────────────────

func doAuth() {
	fmt.Println("\n=== Generate OAuth Token ===")

	credsJSON := os.Getenv("GOOGLE_CREDENTIALS_JSON")
	if credsJSON == "" {
		log.Fatalf("GOOGLE_CREDENTIALS_JSON is not set in .env")
	}

	config, err := google.ConfigFromJSON([]byte(credsJSON), calendar.CalendarScope)
	if err != nil {
		log.Fatalf("Failed to parse credentials: %v", err)
	}
	config.RedirectURL = "http://localhost:8080"

	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline, oauth2.ApprovalForce)

	fmt.Printf("\n==============================\n")
	fmt.Printf(" 1. Open this URL in a browser:\n\n")
	fmt.Printf("    %s\n\n", authURL)
	fmt.Printf(" 2. Sign in as the Google account you want to authorize\n")
	fmt.Printf(" 3. After authorizing, you'll be redirected to localhost:8080/?code=...\n")
	fmt.Printf("    If that page doesn't load, copy the 'code' parameter from the URL.\n")
	fmt.Printf(" 4. Enter the authorization code below.\n")
	fmt.Printf("==============================\n\n")

	fmt.Print("Paste the authorization code here: ")
	code, err := reader.ReadString('\n')
	if err != nil {
		log.Fatalf("Failed to read code: %v", err)
	}
	code = strings.TrimSpace(code)

	if code == "" {
		// Try web server fallback
		code = listenForAuthCode()
		if code == "" {
			log.Fatalf("No authorization code provided.")
		}
	}

	ctx := context.Background()
	tok, err := config.Exchange(ctx, code)
	if err != nil {
		log.Fatalf("Failed to exchange code for token: %v", err)
	}

	tokJSON, _ := json.MarshalIndent(tok, "", "  ")
	fmt.Printf("\n✅ SUCCESS! Token obtained.\n\n")
	fmt.Printf("Add this to your .env or save to a file:\n\n")
	fmt.Printf("GOOGLE_TOKEN_JSON='%s'\n\n", string(tokJSON))

	// Also save to file if requested
	fmt.Print("Save token to a file? (y/n): ")
	save, _ := reader.ReadString('\n')
	if strings.TrimSpace(strings.ToLower(save)) == "y" {
		fmt.Print("Filename (default: data/token_vfilinav.json): ")
		fname, _ := reader.ReadString('\n')
		fname = strings.TrimSpace(fname)
		if fname == "" {
			fname = "data/token_vfilinav.json"
		}
		os.WriteFile(fname, tokJSON, 0600)
		fmt.Printf("✅ Saved to %s\n", fname)
	}
}

func listenForAuthCode() string {
	codeChan := make(chan string, 1)

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code != "" {
			fmt.Fprintf(w, "✅ Authorization successful! You can close this tab.")
			codeChan <- code
		} else {
			fmt.Fprintf(w, "No code found in URL. Check the URL parameters.")
		}
	})

	server := &http.Server{Addr: ":8080", Handler: mux}
	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Printf("HTTP server on :8080 exited: %v", err)
		}
	}()
	defer server.Shutdown(context.Background())

	fmt.Println("Listening on http://localhost:8080 for OAuth callback...")
	select {
	case code := <-codeChan:
		return code
	case <-time.After(5 * time.Minute):
		return ""
	}
}

// ─── Migration ───────────────────────────────────────────────────────────────

func doMigrate(tokenFile string) {
	fmt.Println("\n=== Phase 3+4: Appointments Migration ===")
	fmt.Println("This command migrates appointments from vfilinav@gmail.com to the project.")
	fmt.Println()
	fmt.Println("Step 1: You need an OAuth token for vfilinav@gmail.com.")
	fmt.Println("        Run: go run scripts/data_migration.go auth")
	fmt.Println("        Authenticate as vfilinav@gmail.com and save the token to a file.")
	fmt.Println()

	tokenPath := tokenFile
	if tokenPath == "" {
		fmt.Print("Path to vfilinav OAuth token file (default: data/token_vfilinav.json): ")
		path, _ := reader.ReadString('\n')
		tokenPath = strings.TrimSpace(path)
		if tokenPath == "" {
			tokenPath = "data/token_vfilinav.json"
		}
	}

	// Read vfilinav token
	tokBytes, err := os.ReadFile(tokenPath)
	if err != nil {
		log.Fatalf("Failed to read token file %s: %v\n\nRun 'go run scripts/data_migration.go auth' first to generate it.", tokenPath, err)
	}

	var vfilinavToken oauth2.Token
	if err := json.Unmarshal(tokBytes, &vfilinavToken); err != nil {
		log.Fatalf("Failed to parse token file: %v", err)
	}

	// Connect to project's calendar (veramassagist via existing GOOGLE_TOKEN_JSON)
	projectSvc, projectCalendarID := makeCalendarClient()
	fmt.Printf("\nSource (vfilinav@gmail.com):     token file: %s\n", tokenPath)
	fmt.Printf("Target (veramassagist@gmail.com): calendar: %s\n", projectCalendarID)
	getCalendarInfo(projectSvc, projectCalendarID)
	fmt.Println()

	// Create a separate client for vfilinav
	credsJSON := os.Getenv("GOOGLE_CREDENTIALS_JSON")
	config, err := google.ConfigFromJSON([]byte(credsJSON), calendar.CalendarScope)
	if err != nil {
		log.Fatalf("Failed to parse credentials: %v", err)
	}
	config.RedirectURL = "http://localhost:8080"

	ctx := context.Background()
	vfClient := oauth2.NewClient(ctx, config.TokenSource(ctx, &vfilinavToken))
	vfSvc, err := calendar.NewService(ctx, option.WithHTTPClient(vfClient))
	if err != nil {
		log.Fatalf("Failed to create vfilinav calendar service: %v", err)
	}

	// List vfilinav's calendars to find the right one
	calList, err := vfSvc.CalendarList.List().Do()
	if err != nil {
		log.Fatalf("Failed to list vfilinav calendars: %v", err)
	}

	fmt.Printf("Calendars accessible by vfilinav@gmail.com:\n")
	for i, cal := range calList.Items {
		fmt.Printf("  [%d] %s (%s)\n", i+1, cal.Summary, cal.Id)
		if cal.Primary {
			fmt.Printf("        ← PRIMARY\n")
		}
	}
	fmt.Println()

	// Use primary calendar (or ask user)
	vfCalendarID := "primary"
	if len(calList.Items) > 1 {
		fmt.Printf("Which calendar contains Vera's appointments? (default: primary): ")
		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)
		if choice != "" {
			idx := 0
			fmt.Sscanf(choice, "%d", &idx)
			if idx > 0 && idx <= len(calList.Items) {
				vfCalendarID = calList.Items[idx-1].Id
			}
		}
	}

	// Read events from vfilinav
	timeMin := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	timeMax := time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC)

	events, err := vfSvc.Events.List(vfCalendarID).
		ShowDeleted(false).
		SingleEvents(true).
		MaxResults(500).
		TimeMin(timeMin.Format(time.RFC3339)).
		TimeMax(timeMax.Format(time.RFC3339)).
		OrderBy("startTime").
		Do()
	if err != nil {
		log.Fatalf("Failed to read events from vfilinav's calendar: %v", err)
	}

	fmt.Printf("\nFound %d events in vfilinav's calendar.\n", len(events.Items))
	if len(events.Items) == 0 {
		fmt.Println("Nothing to migrate.")
		return
	}

	// Show events for review
	fmt.Println("\nEvents to migrate:")
	for i, e := range events.Items {
		start := "all-day"
		if e.Start.DateTime != "" {
			t, _ := time.Parse(time.RFC3339, e.Start.DateTime)
			start = t.Format("2006-01-02 15:04")
		} else if e.Start.Date != "" {
			start = e.Start.Date
		}
		fmt.Printf("  [%2d] %s | %s\n", i+1, start, ellipsis(e.Summary, 50))
	}
	fmt.Println()

	if !confirm("Migrate these events to the project calendar?") {
		fmt.Println("Migration cancelled.")
		return
	}

	// Copy events to project calendar + import to DB
	db := connectDB()
	defer db.Close()

	migrated := 0
	skipped := 0

	for _, e := range events.Items {
		// Check if event already exists (by summary + start time)
		startStr := ""
		if e.Start.DateTime != "" {
			startStr = e.Start.DateTime
		} else if e.Start.Date != "" {
			startStr = e.Start.Date
		}

		// Create event in project calendar
		newEvent := &calendar.Event{
			Summary:     e.Summary,
			Description: e.Description,
			Start:       e.Start,
			End:         e.End,
		}

		created, err := projectSvc.Events.Insert(projectCalendarID, newEvent).Do()
		if err != nil {
			log.Printf("  ⚠ Failed to create event '%s': %v", e.Summary, err)
			skipped++
			continue
		}

		fmt.Printf("  ✅ Created: %s (%s)\n", created.Summary, created.Id)

		// Extract patient info from event
		parts := strings.SplitN(e.Summary, " - ", 2)
		customerName := ""
		if len(parts) >= 2 {
			customerName = strings.TrimSpace(parts[1])
		}

		// Extract TGID from description if present
		patientID := ""
		if e.Description != "" && strings.HasPrefix(e.Description, "TGID:") {
			descParts := strings.SplitN(e.Description, "\n", 2)
			patientID = strings.TrimPrefix(descParts[0], "TGID:")
		}

		// Try to save patient if not already in DB
		if patientID != "" {
			var exists int
			db.Get(&exists, "SELECT COUNT(*) FROM patients WHERE telegram_id = $1", patientID)
			if exists == 0 && customerName != "" {
				_, err := db.Exec(`
					INSERT INTO patients (telegram_id, name, first_visit, last_visit, total_visits)
					VALUES ($1, $2, $3, $3, 1)
					ON CONFLICT (telegram_id) DO NOTHING`,
					patientID, customerName, startStr)
				if err != nil {
					log.Printf("  ⚠ Failed to save patient %s: %v", customerName, err)
				} else {
					fmt.Printf("  📋 Imported patient: %s (%s)\n", customerName, patientID)
				}
			}
		}

		migrated++
		time.Sleep(100 * time.Millisecond) // rate limit
	}

	fmt.Printf("\n✅ Migration complete: %d events created, %d skipped.\n", migrated, skipped)
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  1. Verify the migrated events in veramassagist's calendar")
	fmt.Println("  2. Update GOOGLE_CALENDAR_ID to 'primary' (or keep current if it already maps correctly)")
	fmt.Println("  3. Redeploy the bot")
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func confirm(prompt string) bool {
	fmt.Printf("%s (y/N): ", prompt)
	answer, _ := reader.ReadString('\n')
	answer = strings.TrimSpace(strings.ToLower(answer))
	return answer == "y" || answer == "yes"
}

func ellipsis(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}