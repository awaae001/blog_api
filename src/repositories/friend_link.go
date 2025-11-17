package repositories

import (
	"blog_api/src/model"
	"database/sql"
	"fmt"
	"log"
	"strings"
)

// InsertFriendLinks inserts friend links from the configuration if they don't already exist.
func InsertFriendLinks(db *sql.DB, friendLinks []model.FriendWebsite) error {
	if len(friendLinks) == 0 {
		log.Println("No friend links to insert.")
		return nil
	}

	log.Println("[db][friend][init]Start inserting friend links...")

	stmt, err := db.Prepare("INSERT INTO friend_link (website_name, website_url, website_icon_url, description) VALUES (?, ?, ?, ?)")
	if err != nil {
		return fmt.Errorf("[db][friend][ERR]could not prepare insert statement: %w", err)
	}
	defer stmt.Close()

	for _, link := range friendLinks {
		var exists bool
		// Check if the link already exists
		err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM friend_link WHERE website_url = ?)", link.Link).Scan(&exists)
		if err != nil {
			log.Printf("[db][friend][ERR]Could not check for existing link %s: %v", link.Link, err)
			continue // Or return error, depending on desired strictness
		}

		if !exists {
			if _, err := stmt.Exec(link.Name, link.Link, link.Avatar, link.Info); err != nil {
				log.Printf("[db][friend][ERR]Could not insert friend link %s: %v", link.Name, err)
				// Decide if one failure should stop the whole process
			} else {
				log.Printf("[db][friend][init]Inserted friend link: %s", link.Name)
			}
		} else {
			log.Printf("[db][friend][init]Friend link %s already exists, skipping.", link.Name)
		}
	}

	log.Println("[db][friend][init]Friend links insertion process completed.")
	return nil
}

// GetAllFriendLinks retrieves all friend links from the database, excluding 'died' ones.
func GetAllFriendLinks(db *sql.DB) ([]model.FriendWebsite, error) {
	rows, err := db.Query("SELECT id, website_name, website_url, website_icon_url, description, times, status FROM friend_link WHERE status != 'died'")
	if err != nil {
		return nil, fmt.Errorf("could not query friend links: %w", err)
	}
	defer rows.Close()

	var links []model.FriendWebsite
	for rows.Next() {
		var link model.FriendWebsite
		if err := rows.Scan(&link.ID, &link.Name, &link.Link, &link.Avatar, &link.Info, &link.Times, &link.Status); err != nil {
			log.Printf("Could not scan friend link: %v", err)
			continue
		}
		links = append(links, link)
	}

	return links, nil
}

// GetFriendLinksWithFilter retrieves friend links with filtering and pagination support.
func GetFriendLinksWithFilter(db *sql.DB, status string, offset int, limit int) ([]model.FriendWebsite, error) {
	query := "SELECT id, website_name, website_url, website_icon_url, description, times, status FROM friend_link WHERE 1=1"
	args := []interface{}{}

	// Add status filter if provided
	if status != "" {
		query += " AND status = ?"
		args = append(args, status)
	}

	// Add pagination
	query += " ORDER BY updated_at DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("could not query friend links: %w", err)
	}
	defer rows.Close()

	var links []model.FriendWebsite
	for rows.Next() {
		var link model.FriendWebsite
		if err := rows.Scan(&link.ID, &link.Name, &link.Link, &link.Avatar, &link.Info, &link.Times, &link.Status); err != nil {
			log.Printf("Could not scan friend link: %v", err)
			continue
		}
		links = append(links, link)
	}

	return links, nil
}

// CountFriendLinks counts the total number of friend links matching the filter.
func CountFriendLinks(db *sql.DB, status string) (int, error) {
	query := "SELECT COUNT(*) FROM friend_link WHERE 1=1"
	args := []interface{}{}

	// Add status filter if provided
	if status != "" {
		query += " AND status = ?"
		args = append(args, status)
	}

	var count int
	err := db.QueryRow(query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("could not count friend links: %w", err)
	}

	return count, nil
}

// UpdateFriendLink updates the details of a friend link after crawling.
func UpdateFriendLink(db *sql.DB, link model.FriendWebsite, result model.CrawlResult) error {
	if result.Status == "survival" {
		link.Times = 0 // Reset times on success
	} else {
		link.Times++
	}

	if link.Times >= 4 {
		link.Status = "died"
	} else {
		link.Status = result.Status
	}

	// If there was a redirect, update the URL
	if result.RedirectURL != "" {
		link.Link = result.RedirectURL
	}

	stmt, err := db.Prepare(`
		UPDATE friend_link
		SET
			website_url = ?,
			description = CASE WHEN description = '' THEN ? ELSE description END,
			website_icon_url = CASE WHEN website_icon_url = '' THEN ? ELSE website_icon_url END,
			status = ?,
			times = ?,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`)
	if err != nil {
		return fmt.Errorf("could not prepare update statement: %w", err)
	}
	defer stmt.Close()

	if _, err := stmt.Exec(link.Link, result.Description, result.IconURL, link.Status, link.Times, link.ID); err != nil {
		return fmt.Errorf("could not update friend link with id %d: %w", link.ID, err)
	}

	log.Printf("Updated friend link with id %d. Status: %s, Times: %d", link.ID, link.Status, link.Times)
	return nil
}

// CreateFriendLink inserts a single new friend link into the database.
func CreateFriendLink(db *sql.DB, link model.FriendWebsite) (int64, error) {
	stmt, err := db.Prepare("INSERT INTO friend_link (website_name, website_url, website_icon_url, description, email, status) VALUES (?, ?, ?, ?, ?, 'pending')")
	if err != nil {
		return 0, fmt.Errorf("could not prepare insert statement for friend link: %w", err)
	}
	defer stmt.Close()

	result, err := stmt.Exec(link.Name, link.Link, link.Avatar, link.Info, link.Email)
	if err != nil {
		return 0, fmt.Errorf("could not execute insert statement for friend link: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("could not retrieve last insert ID for friend link: %w", err)
	}

	log.Printf("[db][friend] Inserted new friend link: %s with ID: %d", link.Name, id)
	return id, nil
}

// DeleteFriendLinksByID deletes friend links by their IDs and returns the deleted links.
func DeleteFriendLinksByID(db *sql.DB, ids []int) ([]model.FriendWebsite, error) {
	if len(ids) == 0 {
		return []model.FriendWebsite{}, nil
	}

	tx, err := db.Begin()
	if err != nil {
		return nil, fmt.Errorf("could not begin transaction: %w", err)
	}
	defer tx.Rollback() // Rollback on error

	// First, retrieve the records to be deleted
	query, args, err := buildInClause("SELECT id, website_name, website_url, website_icon_url, description, email, times, status FROM friend_link WHERE id IN (?)", ids)
	if err != nil {
		return nil, fmt.Errorf("could not build query: %w", err)
	}

	rows, err := tx.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("could not query friend links for deletion: %w", err)
	}
	defer rows.Close()

	var deletedLinks []model.FriendWebsite
	for rows.Next() {
		var link model.FriendWebsite
		if err := rows.Scan(&link.ID, &link.Name, &link.Link, &link.Avatar, &link.Info, &link.Email, &link.Times, &link.Status); err != nil {
			log.Printf("Could not scan friend link for deletion: %v", err)
			continue
		}
		deletedLinks = append(deletedLinks, link)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error during rows iteration: %w", err)
	}

	// Now, delete the records
	query, args, err = buildInClause("DELETE FROM friend_link WHERE id IN (?)", ids)
	if err != nil {
		return nil, fmt.Errorf("could not build delete query: %w", err)
	}

	result, err := tx.Exec(query, args...)
	if err != nil {
		return nil, fmt.Errorf("could not delete friend links: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("could not get rows affected: %w", err)
	}

	log.Printf("[db][friend] Deleted %d friend links.", rowsAffected)

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("could not commit transaction: %w", err)
	}

	return deletedLinks, nil
}

// UpdateFriendLinkByID updates a friend link by its ID based on the provided data.
func UpdateFriendLinkByID(db *sql.DB, req model.EditFriendLinkReq) (int64, error) {
	if len(req.Data) == 0 {
		return 0, fmt.Errorf("no data provided for update")
	}

	// Whitelist of updatable columns
	updatableColumns := map[string]bool{
		"website_name":     true,
		"website_url":      true,
		"website_icon_url": true,
		"description":      true,
		"email":            true,
		"status":           true,
	}

	query := "UPDATE friend_link SET "
	args := []interface{}{}
	var setClauses []string

	for col, val := range req.Data {
		if !updatableColumns[col] {
			log.Printf("[db][friend][WARN] Attempted to update non-updatable column: %s", col)
			continue
		}

		if !req.Opt.OverwriteIfBlank {
			if s, ok := val.(string); ok && s == "" {
				continue
			}
		}

		setClauses = append(setClauses, fmt.Sprintf("%s = ?", col))
		args = append(args, val)
	}

	if len(setClauses) == 0 {
		log.Println("[db][friend] No valid fields to update after filtering.")
		return 0, nil // Nothing to update
	}

	query += fmt.Sprintf("%s, updated_at = CURRENT_TIMESTAMP WHERE id = ?", strings.Join(setClauses, ", "))
	args = append(args, req.ID)

	result, err := db.Exec(query, args...)
	if err != nil {
		return 0, fmt.Errorf("could not execute update statement for friend link with id %d: %w", req.ID, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("could not get rows affected: %w", err)
	}

	log.Printf("[db][friend] Updated friend link with ID: %d. Rows affected: %d", req.ID, rowsAffected)
	return rowsAffected, nil
}

// buildInClause is a helper function to build SQL IN clauses.
func buildInClause(query string, params []int) (string, []interface{}, error) {
	if len(params) == 0 {
		return "", nil, fmt.Errorf("no parameters provided")
	}

	placeholders := ""
	args := make([]interface{}, len(params))
	for i, id := range params {
		if i > 0 {
			placeholders += ","
		}
		placeholders += "?"
		args[i] = id
	}

	// Correctly replace the single '?' with the generated placeholders.
	// This is a simplified approach; for multiple '?' in the query, a more robust method is needed.
	finalQuery := ""
	inClauseStarted := false
	for _, r := range query {
		if r == '?' && !inClauseStarted {
			finalQuery += placeholders
			inClauseStarted = true
		} else {
			finalQuery += string(r)
		}
	}

	return finalQuery, args, nil
}
