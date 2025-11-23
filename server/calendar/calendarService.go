package calendar

import (
	"bytes"
	"database/sql"
	"encoding/hex"
	"crypto/rand"
	"time"
	"strings"
	"strconv"
)

// Event mirrors a calendar_index row.
type Event struct {
	UID              string
	UserID           string
	Kind             string
	SourceID         int64
	Summary          string
	DeadlineEpoch    sql.NullInt64
	Completed        bool
	LastModified     int64
	Seq              int
	CancelledAt      sql.NullInt64
}

// GetUserEvents loads all events (even completed) for a user.
func GetUserEvents(db *sql.DB, userID string) ([]Event, error) {
	rows, err := db.Query(`
		SELECT uid, user_id, kind, source_id, summary, deadline_epoch,
					 completed, last_modified_epoch, seq, cancelled_at
		FROM calendar_index
		WHERE user_id = ?
		ORDER BY
			-- show dated items first by date, then undated, then cancelled
			(deadline_epoch IS NULL) ASC,
			deadline_epoch ASC,
			kind ASC,
			source_id ASC;
	`, userID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []Event
	for rows.Next() {
		var e Event
		if err := rows.Scan(&e.UID, &e.UserID, &e.Kind, &e.SourceID, &e.Summary,
			&e.DeadlineEpoch, &e.Completed, &e.LastModified, &e.Seq, &e.CancelledAt); err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, nil
}


func BuildICS(events []Event) string {
	var buf bytes.Buffer
	w := func(s string) { buf.WriteString(s + "\r\n") }

	loc, _ := time.LoadLocation("Europe/London") // choose a canonical tz

	w("BEGIN:VCALENDAR")
	w("PRODID:-//YourApp//Calendar 1.0//EN")
	w("VERSION:2.0")
	w("CALSCALE:GREGORIAN")
	w("METHOD:PUBLISH")

	escape := func(s string) string {
		// iCalendar text escaping: backslash, semicolon, comma, newline
		s = strings.ReplaceAll(s, `\`, `\\`)
		s = strings.ReplaceAll(s, `;`, `\;`)
		s = strings.ReplaceAll(s, `,`, `\,`)
		s = strings.ReplaceAll(s, "\n", `\n`)
		return s
	}

	for _, e := range events {
		// skip items without a deadline
		if !e.DeadlineEpoch.Valid {
			continue
		}

		// Treat DeadlineEpoch as the END of a 1 hour event.
		endLocal := time.Unix(e.DeadlineEpoch.Int64, 0).In(loc)
		startLocal := endLocal.Add(-1 * time.Hour)

		startUTC := startLocal.UTC().Format("20060102T150405Z")
		endUTC := endLocal.UTC().Format("20060102T150405Z")
		dtstamp := time.Unix(e.LastModified, 0).UTC().Format("20060102T150405Z")

		status := "CONFIRMED"
		if e.CancelledAt.Valid {
			status = "CANCELLED"
		}

		w("BEGIN:VEVENT")
		w("UID:" + e.UID)
		w("SEQUENCE:" + strconv.Itoa(e.Seq))
		w("DTSTAMP:" + dtstamp)
		w("STATUS:" + status)
		w("SUMMARY:" + escape(e.Summary))
		w("DTSTART:" + startUTC)
		w("DTEND:" + endUTC)
		w("END:VEVENT")
	}

	w("END:VCALENDAR")
	return buf.String()
}


func GetOrCreateCalendarToken(db *sql.DB, userID string) (string, error) {
	var tok string

	// Fast path: already exists
	if err := db.QueryRow(`SELECT token FROM calendar_tokens WHERE user_id = ?`, userID).Scan(&tok); err == nil {
		return tok, nil
	} else if err != sql.ErrNoRows {
		return "", err
	}

	// Mint a new 256-bit token
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	tok = hex.EncodeToString(buf)

	// Try insert; if a concurrent insert won, read it back
	if _, err := db.Exec(`INSERT INTO calendar_tokens(token, user_id) VALUES(?, ?)`, tok, userID); err != nil {
		// SQLite unique constraint text is portable enough to check
		if strings.Contains(strings.ToLower(err.Error()), "unique") {
			if err2 := db.QueryRow(`SELECT token FROM calendar_tokens WHERE user_id = ?`, userID).Scan(&tok); err2 == nil {
				return tok, nil
			}
		}
		return "", err
	}
	return tok, nil
}


// RotateCalendarToken replaces the user's token and returns the new one.
func RotateCalendarToken(db *sql.DB, userID string) (string, error) {
    // mint new 256-bit token
    buf := make([]byte, 32)
    if _, err := rand.Read(buf); err != nil { return "", err }
    tok := hex.EncodeToString(buf)

    tx, err := db.Begin()
    if err != nil { return "", err }
    defer func() { _ = tx.Rollback() }()

    // try update first
    res, err := tx.Exec(`
        UPDATE calendar_tokens
        SET token = ?, created_at = strftime('%s','now'), last_used_at = NULL
        WHERE user_id = ?`, tok, userID)
    if err != nil { return "", err }

    n, _ := res.RowsAffected()
    if n == 0 {
        if _, err := tx.Exec(`INSERT INTO calendar_tokens(token, user_id) VALUES(?, ?)`, tok, userID); err != nil {
            return "", err
        }
    }
    if err := tx.Commit(); err != nil { return "", err }
    return tok, nil
}
