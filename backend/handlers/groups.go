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
)

// CreateGroupHandler - POST { name, description }
func CreateGroupHandler(w http.ResponseWriter, r *http.Request) {
	uid := utils.GetUserIDFromContext(r)
	if uid == "" {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	userID, _ := strconv.ParseInt(uid, 10, 64)
	var payload struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid input")
		return
	}
	res, err := db.DB.Exec("INSERT INTO groups (owner_id, name, description) VALUES (?, ?, ?)", userID, payload.Name, payload.Description)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to create group")
		return
	}
	id, _ := res.LastInsertId()
	// add owner as member
	db.DB.Exec("INSERT OR IGNORE INTO group_members (group_id, user_id, role) VALUES (?, ?, 'owner')", id, userID)
	utils.JSON(w, http.StatusOK, map[string]interface{}{"status": "created", "group_id": id})
}

// ListGroupsHandler - GET /api/groups
func ListGroupsHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.DB.Query("SELECT id, owner_id, name, description, created_at FROM groups ORDER BY created_at DESC")
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to list groups")
		return
	}
	defer rows.Close()
	type G struct {
		ID          int64  `json:"id"`
		OwnerID     int64  `json:"owner_id"`
		Name        string `json:"name"`
		Description string `json:"description"`
		Created     string `json:"created_at"`
	}
	var out []G
	for rows.Next() {
		var g G
		rows.Scan(&g.ID, &g.OwnerID, &g.Name, &g.Description, &g.Created)
		out = append(out, g)
	}
	utils.JSON(w, http.StatusOK, out)
}

// GetGroupHandler - GET /api/group?id=<id>
func GetGroupHandler(w http.ResponseWriter, r *http.Request) {
	idParam := r.URL.Query().Get("id")
	if idParam == "" {
		utils.Error(w, http.StatusBadRequest, "Missing id")
		return
	}
	gid, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid id")
		return
	}
	var g struct {
		ID          int64          `json:"id"`
		OwnerID     int64          `json:"owner_id"`
		Name        string         `json:"name"`
		Description sql.NullString `json:"description"`
		Created     string         `json:"created_at"`
	}
	err = db.DB.QueryRow("SELECT id, owner_id, name, description, created_at FROM groups WHERE id = ?", gid).Scan(&g.ID, &g.OwnerID, &g.Name, &g.Description, &g.Created)
	if err != nil {
		utils.Error(w, http.StatusNotFound, "Group not found")
		return
	}
	// get members count
	var memberCount int
	db.DB.QueryRow("SELECT COUNT(1) FROM group_members WHERE group_id = ?", gid).Scan(&memberCount)
	resp := map[string]interface{}{
		"group":   g,
		"members": memberCount,
	}
	utils.JSON(w, http.StatusOK, resp)
}

// InviteHandler - POST { group_id, invitee_id }
func InviteHandler(w http.ResponseWriter, r *http.Request) {
	uid := utils.GetUserIDFromContext(r)
	if uid == "" {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	inviter, _ := strconv.ParseInt(uid, 10, 64)
	var payload struct {
		GroupID   int64 `json:"group_id"`
		InviteeID int64 `json:"invitee_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid input")
		return
	}
	_, err := db.DB.Exec("INSERT INTO group_invites (group_id, inviter_id, invitee_id) VALUES (?, ?, ?)", payload.GroupID, inviter, payload.InviteeID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to invite")
		return
	}
	utils.JSON(w, http.StatusOK, map[string]string{"status": "invited"})
}

// RespondInviteHandler - POST { invite_id, action: accept|decline }
func RespondInviteHandler(w http.ResponseWriter, r *http.Request) {
	uid := utils.GetUserIDFromContext(r)
	if uid == "" {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	userID, _ := strconv.ParseInt(uid, 10, 64)
	var payload struct {
		InviteID int64  `json:"invite_id"`
		Action   string `json:"action"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid input")
		return
	}
	var invite struct {
		GroupID   int64
		InviteeID int64
	}
	err := db.DB.QueryRow("SELECT group_id, invitee_id FROM group_invites WHERE id = ? AND status = 'pending'", payload.InviteID).Scan(&invite.GroupID, &invite.InviteeID)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid invite")
		return
	}
	if invite.InviteeID != userID {
		utils.Error(w, http.StatusForbidden, "Not allowed")
		return
	}
	if payload.Action == "accept" {
		db.DB.Exec("UPDATE group_invites SET status='accepted' WHERE id=?", payload.InviteID)
		db.DB.Exec("INSERT OR IGNORE INTO group_members (group_id, user_id) VALUES (?,?)", invite.GroupID, userID)
		utils.JSON(w, http.StatusOK, map[string]string{"status": "accepted"})
		return
	}
	db.DB.Exec("UPDATE group_invites SET status='declined' WHERE id=?", payload.InviteID)
	utils.JSON(w, http.StatusOK, map[string]string{"status": "declined"})
}

// CreateGroupPostHandler - POST multipart/form with content & optional image
func CreateGroupPostHandler(w http.ResponseWriter, r *http.Request) {
	uid := utils.GetUserIDFromContext(r)
	if uid == "" {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	userID, _ := strconv.ParseInt(uid, 10, 64)
	r.ParseMultipartForm(10 << 20)
	gidStr := r.FormValue("group_id")
	gid, _ := strconv.ParseInt(gidStr, 10, 64)
	content := r.FormValue("content")
	imageURL := ""
	file, fh, err := r.FormFile("image")
	if err == nil && file != nil {
		defer file.Close()
		// save to uploads
		os.MkdirAll("uploads", 0755)
		fname := fmt.Sprintf("group_%d_%s", gid, filepath.Base(fh.Filename))
		dst, _ := os.Create(filepath.Join("uploads", fname))
		defer dst.Close()
		io.Copy(dst, file)
		imageURL = "/uploads/" + fname
	}
	// ensure user is a member
	var cnt int
	db.DB.QueryRow("SELECT COUNT(1) FROM group_members WHERE group_id=? AND user_id=?", gid, userID).Scan(&cnt)
	if cnt == 0 {
		utils.Error(w, http.StatusForbidden, "Not a member")
		return
	}
	_, err = db.DB.Exec("INSERT INTO group_posts (group_id, author_id, content, image_url) VALUES (?, ?, ?, ?)", gid, userID, content, imageURL)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to create post")
		return
	}
	utils.JSON(w, http.StatusOK, map[string]string{"status": "created"})
}

// ListGroupPostsHandler - GET /api/group/posts?group_id=<id>
func ListGroupPostsHandler(w http.ResponseWriter, r *http.Request) {
	gidStr := r.URL.Query().Get("group_id")
	if gidStr == "" {
		utils.Error(w, http.StatusBadRequest, "Missing group_id")
		return
	}
	gid, _ := strconv.ParseInt(gidStr, 10, 64)
	rows, err := db.DB.Query("SELECT id, group_id, author_id, content, image_url, created_at FROM group_posts WHERE group_id = ? ORDER BY created_at DESC", gid)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed")
		return
	}
	defer rows.Close()
	type P struct {
		ID       int64  `json:"id"`
		GroupID  int64  `json:"group_id"`
		AuthorID int64  `json:"author_id"`
		Content  string `json:"content"`
		Image    string `json:"image_url"`
		Created  string `json:"created_at"`
	}
	var out []P
	for rows.Next() {
		var p P
		rows.Scan(&p.ID, &p.GroupID, &p.AuthorID, &p.Content, &p.Image, &p.Created)
		out = append(out, p)
	}
	utils.JSON(w, http.StatusOK, out)
}

// AddGroupCommentHandler - POST { post_id, content }
func AddGroupCommentHandler(w http.ResponseWriter, r *http.Request) {
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
	// check membership by looking up post's group
	var gid int64
	err := db.DB.QueryRow("SELECT group_id FROM group_posts WHERE id = ?", payload.PostID).Scan(&gid)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid post")
		return
	}
	var cnt int
	db.DB.QueryRow("SELECT COUNT(1) FROM group_members WHERE group_id=? AND user_id=?", gid, userID).Scan(&cnt)
	if cnt == 0 {
		utils.Error(w, http.StatusForbidden, "Not a member")
		return
	}
	_, err = db.DB.Exec("INSERT INTO group_comments (post_id, user_id, content) VALUES (?, ?, ?)", payload.PostID, userID, payload.Content)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed")
		return
	}
	utils.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// CreateEventHandler - POST { group_id, title, description, event_time }
func CreateEventHandler(w http.ResponseWriter, r *http.Request) {
	uid := utils.GetUserIDFromContext(r)
	if uid == "" {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	userID, _ := strconv.ParseInt(uid, 10, 64)
	var payload struct {
		GroupID     int64  `json:"group_id"`
		Title       string `json:"title"`
		Description string `json:"description"`
		EventTime   string `json:"event_time"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid input")
		return
	}
	// ensure creator is member
	var cnt int
	db.DB.QueryRow("SELECT COUNT(1) FROM group_members WHERE group_id=? AND user_id=?", payload.GroupID, userID).Scan(&cnt)
	if cnt == 0 {
		utils.Error(w, http.StatusForbidden, "Not a member")
		return
	}
	_, err := db.DB.Exec("INSERT INTO events (group_id, creator_id, title, description, event_time) VALUES (?, ?, ?, ?, ?)", payload.GroupID, userID, payload.Title, payload.Description, payload.EventTime)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to create event")
		return
	}
	utils.JSON(w, http.StatusOK, map[string]string{"status": "created"})
}

// VoteEventHandler - POST { event_id, vote }
func VoteEventHandler(w http.ResponseWriter, r *http.Request) {
	uid := utils.GetUserIDFromContext(r)
	if uid == "" {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	userID, _ := strconv.ParseInt(uid, 10, 64)
	var payload struct {
		EventID int64  `json:"event_id"`
		Vote    string `json:"vote"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid input")
		return
	}
	// upsert vote
	db.DB.Exec("INSERT OR REPLACE INTO event_votes (id, event_id, user_id, vote) VALUES ((SELECT id FROM event_votes WHERE event_id=? AND user_id=?), ?, ?, ?)", payload.EventID, userID, payload.EventID, userID, payload.Vote)
	utils.JSON(w, http.StatusOK, map[string]string{"status": "voted"})
}
