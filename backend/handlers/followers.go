package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"social-network/backend/db"
	"social-network/backend/utils"
	"strconv"
	"strings"
)

// POST /api/follow - send follow request (handles public/private profile logic)
func FollowHandler(w http.ResponseWriter, r *http.Request) {
	userIDStr := utils.GetUserIDFromContext(r)
	if userIDStr == "" {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	var payload struct {
		TargetID int64 `json:"target_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid input")
		return
	}
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Check target profile type
	var profileType string
	err = db.DB.QueryRow("SELECT profile_type FROM users WHERE id = ?", payload.TargetID).Scan(&profileType)
	if err != nil {
		utils.Error(w, http.StatusNotFound, "User not found")
		return
	}

	profileType = strings.ToLower(profileType)
	if profileType == "public" {
		// Auto-follow
		_, err := db.DB.Exec("INSERT OR IGNORE INTO followers (follower_id, followed_id, created_at) VALUES (?, ?, datetime('now'))", userID, payload.TargetID)
		if err != nil {
			utils.Error(w, http.StatusInternalServerError, "Failed to follow")
			return
		}
		// create notification for the target user about the new follower
		_ = CreateNotification(payload.TargetID, userID, "new_follower", "")
		utils.JSON(w, http.StatusOK, map[string]string{"status": "followed"})
		return
	}
	// Private: create follow request
	_, err = db.DB.Exec("INSERT OR IGNORE INTO follow_requests (sender_id, receiver_id, status, created_at) VALUES (?, ?, 'pending', datetime('now'))", userID, payload.TargetID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to send request")
		return
	}
	// notify target about follow request
	_ = CreateNotification(payload.TargetID, userID, "follow_request", "")
	utils.JSON(w, http.StatusOK, map[string]string{"status": "requested"})
}

// POST /api/follow/accept - accept request
func AcceptFollowHandler(w http.ResponseWriter, r *http.Request) {
	userIDStr := utils.GetUserIDFromContext(r)
	if userIDStr == "" {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	var payload struct {
		SenderID int64 `json:"sender_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid input")
		return
	}
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	// Accept request
	res, err := db.DB.Exec("UPDATE follow_requests SET status='accepted' WHERE sender_id=? AND receiver_id=? AND status='pending'", payload.SenderID, userID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed")
		return
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		utils.Error(w, http.StatusBadRequest, "No pending request")
		return
	}
	// Add to followers table
	_, err = db.DB.Exec("INSERT OR IGNORE INTO followers (follower_id, followed_id, created_at) VALUES (?, ?, datetime('now'))", payload.SenderID, userID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed")
		return
	}
	// notify sender their request was accepted
	_ = CreateNotification(payload.SenderID, userID, "follow_request_accepted", "")
	utils.JSON(w, http.StatusOK, map[string]string{"status": "accepted"})
}

// POST /api/follow/decline - decline request
func DeclineFollowHandler(w http.ResponseWriter, r *http.Request) {
	userIDStr := utils.GetUserIDFromContext(r)
	if userIDStr == "" {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	var payload struct {
		SenderID int64 `json:"sender_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid input")
		return
	}
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	res, err := db.DB.Exec("UPDATE follow_requests SET status='declined' WHERE sender_id=? AND receiver_id=? AND status='pending'", payload.SenderID, userID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed")
		return
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		utils.Error(w, http.StatusBadRequest, "No pending request")
		return
	}
	// notify sender their request was declined
	_ = CreateNotification(payload.SenderID, userID, "follow_request_declined", "")
	utils.JSON(w, http.StatusOK, map[string]string{"status": "declined"})
}

// POST /api/unfollow - unfollow a user
func UnfollowHandler(w http.ResponseWriter, r *http.Request) {
	userIDStr := utils.GetUserIDFromContext(r)
	if userIDStr == "" {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	var payload struct {
		TargetID int64 `json:"target_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid input")
		return
	}
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	_, err = db.DB.Exec("DELETE FROM followers WHERE follower_id=? AND followed_id=?", userID, payload.TargetID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed")
		return
	}
	utils.JSON(w, http.StatusOK, map[string]string{"status": "unfollowed"})
}

// GET /api/follow/requests - list pending follow requests for current user
func ListRequests(w http.ResponseWriter, r *http.Request) {
	userIDStr := utils.GetUserIDFromContext(r)
	if userIDStr == "" {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	rows, err := db.DB.Query("SELECT id, sender_id, created_at FROM follow_requests WHERE receiver_id=? AND status='pending' ORDER BY created_at DESC", userID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to query requests")
		return
	}
	defer rows.Close()

	type req struct {
		ID       int64  `json:"id"`
		SenderID int64  `json:"sender_id"`
		Created  string `json:"created_at"`
	}
	var out []req
	for rows.Next() {
		var ritem req
		var created sql.NullString
		if err := rows.Scan(&ritem.ID, &ritem.SenderID, &created); err != nil {
			utils.Error(w, http.StatusInternalServerError, "Failed to read request")
			return
		}
		ritem.Created = created.String
		out = append(out, ritem)
	}
	utils.JSON(w, http.StatusOK, out)
}
