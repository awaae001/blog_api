package repositories

import (
	"blog_api/src/model"
	"database/sql"
	"fmt"
	"log"
)

// InsertFriendLinks inserts friend links from the configuration if they don't already exist.
func InsertFriendLinks(db *sql.DB, friendLinks []model.FriendWebsite) error {
	if len(friendLinks) == 0 {
		log.Println("No friend links to insert.")
		return nil
	}

	log.Println("Start inserting friend links...")

	stmt, err := db.Prepare("INSERT INTO friend_link (website_name, website_url, website_icon_url, description) VALUES (?, ?, ?, ?)")
	if err != nil {
		return fmt.Errorf("could not prepare insert statement: %w", err)
	}
	defer stmt.Close()

	for _, link := range friendLinks {
		var exists bool
		// Check if the link already exists
		err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM friend_link WHERE website_url = ?)", link.Link).Scan(&exists)
		if err != nil {
			log.Printf("Could not check for existing link %s: %v", link.Link, err)
			continue // Or return error, depending on desired strictness
		}

		if !exists {
			if _, err := stmt.Exec(link.Name, link.Link, link.Avatar, link.Info); err != nil {
				log.Printf("Could not insert friend link %s: %v", link.Name, err)
				// Decide if one failure should stop the whole process
			} else {
				log.Printf("Inserted friend link: %s", link.Name)
			}
		} else {
			log.Printf("Friend link %s already exists, skipping.", link.Name)
		}
	}

	log.Println("Friend links insertion process completed.")
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

	stmt, err := db.Prepare("UPDATE friend_link SET website_url = ?, description = ?, website_icon_url = ?, status = ?, times = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?")
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
