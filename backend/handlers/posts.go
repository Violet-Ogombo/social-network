package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"social-network/backend/db"
	"social-network/backend/utils"
	"strconv"
	"strings"
)

// CreatePostHandler handles text + optional image posts
func CreatePostHandler(w http.ResponseWriter, r *http.Request) {
	uid := utils.GetUserIDFromContext(r)
	if uid == "" {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	userID, _ := strconv.ParseInt(uid, 10, 64)

	// parse multipart form for optional file
	r.ParseMultipartForm(10 << 20) // 10MB
	content := r.FormValue("content")
	privacy := r.FormValue("privacy")
	allowed := r.FormValue("allowed") // comma-separated ids for private

	imageURL := ""
	file, fh, err := r.FormFile("image")
	if err == nil && file != nil {
		defer file.Close()
		os.MkdirAll("uploads", 0755)
		fname := fmt.Sprintf("%d_%s", userID, filepath.Base(fh.Filename))
		dst, _ := os.Create(filepath.Join("uploads", fname))
		defer dst.Close()
		io.Copy(dst, file)
		imageURL = "/uploads/" + fname
	}

	if privacy == "" {
		privacy = "public"
	}

	_, err = db.DB.Exec("INSERT INTO posts (author_id, content, image_url, privacy, allowed_user_ids) VALUES (?, ?, ?, ?, ?)", userID, content, imageURL, privacy, allowed)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to create post")
		return
	}
	utils.JSON(w, http.StatusOK, map[string]string{"status": "created"})
}

// ListFeedHandler returns posts visible to the requester
func ListFeedHandler(w http.ResponseWriter, r *http.Request) {
	// optional ?user_id to list a user's posts
	viewer := utils.GetUserIDFromContext(r)
	var viewerID int64
	if viewer != "" {
		viewerID, _ = strconv.ParseInt(viewer, 10, 64)
	}

	qUser := r.URL.Query().Get("user_id")
	var rows *sql.Rows
	var err error
	if qUser != "" {
		// list posts by a specific user, but apply privacy
		tid, _ := strconv.ParseInt(qUser, 10, 64)
		rows, err = db.DB.Query("SELECT id, author_id, content, image_url, privacy, allowed_user_ids, created_at FROM posts WHERE author_id = ? ORDER BY created_at DESC", tid)
	} else {
		// feed: show public posts + posts from followed users + own private posts where allowed
		rows, err = db.DB.Query("SELECT id, author_id, content, image_url, privacy, allowed_user_ids, created_at FROM posts ORDER BY created_at DESC")
	}
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to load posts")
		return
	}
	defer rows.Close()

	type P struct {
		ID       int64  `json:"id"`
		AuthorID int64  `json:"author_id"`
		Content  string `json:"content"`
		ImageURL string `json:"image_url"`
		Privacy  string `json:"privacy"`
		Allowed  string `json:"allowed_user_ids"`
		Created  string `json:"created_at"`
	}
	var out []P
	for rows.Next() {
		var p P
		var allowed sql.NullString
		if err := rows.Scan(&p.ID, &p.AuthorID, &p.Content, &p.ImageURL, &p.Privacy, &allowed, &p.Created); err != nil {
			continue
		}
		p.Allowed = allowed.String
		// privacy enforcement: minimalistic
		visible := false
		if p.Privacy == "public" {
			visible = true
		} else if p.Privacy == "followers" {
			if viewerID > 0 {
				var cnt int
				db.DB.QueryRow("SELECT COUNT(1) FROM followers WHERE follower_id=? AND followed_id=?", viewerID, p.AuthorID).Scan(&cnt)
				if cnt > 0 || viewerID == p.AuthorID {
					visible = true
				}
			}
		} else if p.Privacy == "private" {
			if viewerID == p.AuthorID {
				visible = true
			} else if p.Allowed != "" {
				parts := strings.Split(p.Allowed, ",")
				for _, s := range parts {
					if s == "" {
						continue
					}
					if vid, _ := strconv.ParseInt(strings.TrimSpace(s), 10, 64); vid == viewerID {
						visible = true
						break
					}
				}
			}
		}
		if visible {
			out = append(out, p)
		}
	}
	utils.JSON(w, http.StatusOK, out)
}

// AddCommentHandler adds a comment to a post (respecting post visibility implicitly by assuming front-end only shows allowed posts)
func AddCommentHandler(w http.ResponseWriter, r *http.Request) {
	uid := utils.GetUserIDFromContext(r)
	if uid == "" {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	userID, _ := strconv.ParseInt(uid, 10, 64)
	var payload struct {
		PostID  int64  `json:"post_id"`
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid input")
		return
	}
	_, err := db.DB.Exec("INSERT INTO comments (post_id, user_id, content) VALUES (?, ?, ?)", payload.PostID, userID, payload.Content)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to add comment")
		return
	}
	utils.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
