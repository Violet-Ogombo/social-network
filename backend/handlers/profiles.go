package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"social-network/backend/db"
	"social-network/backend/utils"
	"strconv"
)

// GET /api/profile?id=<id>  - if id omitted, returns current user's profile (requires auth cookie)
func GetProfileHandler(w http.ResponseWriter, r *http.Request) {
	idParam := r.URL.Query().Get("id")
	var targetID int64
	var err error
	if idParam == "" {
		// require auth
		uid := utils.GetUserIDFromContext(r)
		if uid == "" {
			utils.Error(w, http.StatusUnauthorized, "Unauthorized")
			return
		}
		targetID, err = strconv.ParseInt(uid, 10, 64)
		if err != nil {
			utils.Error(w, http.StatusUnauthorized, "Unauthorized")
			return
		}
	} else {
		targetID, err = strconv.ParseInt(idParam, 10, 64)
		if err != nil {
			utils.Error(w, http.StatusBadRequest, "Invalid id")
			return
		}
	}

	var user struct {
		ID          int64          `json:"id"`
		Email       sql.NullString `json:"email,omitempty"`
		FirstName   sql.NullString `json:"first_name,omitempty"`
		LastName    sql.NullString `json:"last_name,omitempty"`
		DateOfBirth sql.NullString `json:"date_of_birth,omitempty"`
		Avatar      sql.NullString `json:"avatar,omitempty"`
		Nickname    sql.NullString `json:"nickname,omitempty"`
		About       sql.NullString `json:"about,omitempty"`
		ProfileType sql.NullString `json:"profile_type,omitempty"`
		CreatedAt   sql.NullString `json:"created_at,omitempty"`
	}
	err = db.DB.QueryRow(`SELECT id, email, first_name, last_name, date_of_birth, avatar, nickname, about_me, profile_type, created_at FROM users WHERE id = ?`, targetID).
		Scan(&user.ID, &user.Email, &user.FirstName, &user.LastName, &user.DateOfBirth, &user.Avatar, &user.Nickname, &user.About, &user.ProfileType, &user.CreatedAt)
	if err != nil {
		utils.Error(w, http.StatusNotFound, "User not found")
		return
	}

	// build response with only non-sensitive fields
	resp := map[string]interface{}{
		"id":            user.ID,
		"first_name":    user.FirstName.String,
		"last_name":     user.LastName.String,
		"date_of_birth": user.DateOfBirth.String,
		"avatar":        user.Avatar.String,
		"nickname":      user.Nickname.String,
		"about":         user.About.String,
		"profile_type":  user.ProfileType.String,
		"created_at":    user.CreatedAt.String,
	}
	// only include email when requester is the same user
	if uid := utils.GetUserIDFromContext(r); uid != "" {
		if parsed, _ := strconv.ParseInt(uid, 10, 64); parsed == user.ID {
			resp["email"] = user.Email.String
		}
	}

	utils.JSON(w, http.StatusOK, resp)
}

// POST /api/profile/update - update current user's profile (protected)
func UpdateProfileHandler(w http.ResponseWriter, r *http.Request) {
	uid := utils.GetUserIDFromContext(r)
	if uid == "" {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	userID, err := strconv.ParseInt(uid, 10, 64)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	var payload struct {
		Nickname    string `json:"nickname"`
		FirstName   string `json:"first_name"`
		LastName    string `json:"last_name"`
		Avatar      string `json:"avatar"`
		About       string `json:"about"`
		ProfileType string `json:"profile_type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid input")
		return
	}
	// validate profile type
	pt := "public"
	if payload.ProfileType != "" {
		if payload.ProfileType == "private" {
			pt = "private"
		} else {
			pt = "public"
		}
	}
	_, err = db.DB.Exec(`UPDATE users SET nickname=?, first_name=?, last_name=?, avatar=?, about_me=?, profile_type=? WHERE id=?`, payload.Nickname, payload.FirstName, payload.LastName, payload.Avatar, payload.About, pt, userID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to update profile")
		return
	}
	utils.JSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

// GET /api/profile/followers?id=<id> - list followers for user (protected to check access if profile private)
func GetFollowersHandler(w http.ResponseWriter, r *http.Request) {
	idParam := r.URL.Query().Get("id")
	var targetID int64
	var err error
	if idParam == "" {
		uid := utils.GetUserIDFromContext(r)
		if uid == "" {
			utils.Error(w, http.StatusUnauthorized, "Unauthorized")
			return
		}
		targetID, _ = strconv.ParseInt(uid, 10, 64)
	} else {
		targetID, err = strconv.ParseInt(idParam, 10, 64)
		if err != nil {
			utils.Error(w, http.StatusBadRequest, "Invalid id")
			return
		}
	}

	// If target profile is private and requester is not the same user and not a follower, deny
	var profileType string
	_ = db.DB.QueryRow("SELECT profile_type FROM users WHERE id = ?", targetID).Scan(&profileType)
	requester := utils.GetUserIDFromContext(r)
	if profileType == "private" {
		// unauthenticated viewers are forbidden
		if requester == "" {
			utils.Error(w, http.StatusForbidden, "Profile is private")
			return
		}
		// if requester is not the owner, check if requester is follower
		reqID, _ := strconv.ParseInt(requester, 10, 64)
		if reqID != targetID {
			var cnt int
			_ = db.DB.QueryRow("SELECT COUNT(1) FROM followers WHERE follower_id=? AND followed_id=?", reqID, targetID).Scan(&cnt)
			if cnt == 0 {
				utils.Error(w, http.StatusForbidden, "Profile is private")
				return
			}
		}
	}

	rows, err := db.DB.Query("SELECT u.id, u.nickname, u.avatar FROM followers f JOIN users u ON f.follower_id = u.id WHERE f.followed_id = ? ORDER BY f.created_at DESC", targetID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to query followers")
		return
	}
	defer rows.Close()
	type uitem struct {
		ID       int64  `json:"id"`
		Nickname string `json:"nickname"`
		Avatar   string `json:"avatar"`
	}
	var out []uitem
	for rows.Next() {
		var it uitem
		var avatar sql.NullString
		if err := rows.Scan(&it.ID, &it.Nickname, &avatar); err != nil {
			utils.Error(w, http.StatusInternalServerError, "Failed to read follower")
			return
		}
		it.Avatar = avatar.String
		out = append(out, it)
	}
	utils.JSON(w, http.StatusOK, out)
}

// GET /api/profile/following?id=<id> - list users the target is following
func GetFollowingHandler(w http.ResponseWriter, r *http.Request) {
	idParam := r.URL.Query().Get("id")
	var targetID int64
	var err error
	if idParam == "" {
		uid := utils.GetUserIDFromContext(r)
		if uid == "" {
			utils.Error(w, http.StatusUnauthorized, "Unauthorized")
			return
		}
		targetID, _ = strconv.ParseInt(uid, 10, 64)
	} else {
		targetID, err = strconv.ParseInt(idParam, 10, 64)
		if err != nil {
			utils.Error(w, http.StatusBadRequest, "Invalid id")
			return
		}
	}
	rows, err := db.DB.Query("SELECT u.id, u.nickname, u.avatar FROM followers f JOIN users u ON f.followed_id = u.id WHERE f.follower_id = ? ORDER BY f.created_at DESC", targetID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to query following")
		return
	}
	defer rows.Close()
	type uitem struct {
		ID       int64  `json:"id"`
		Nickname string `json:"nickname"`
		Avatar   string `json:"avatar"`
	}
	var out []uitem
	for rows.Next() {
		var it uitem
		var avatar sql.NullString
		if err := rows.Scan(&it.ID, &it.Nickname, &avatar); err != nil {
			utils.Error(w, http.StatusInternalServerError, "Failed to read following")
			return
		}
		it.Avatar = avatar.String
		out = append(out, it)
	}
	utils.JSON(w, http.StatusOK, out)
}

// POST /api/profile/privacy - set profile privacy (protected)
func TogglePrivacyHandler(w http.ResponseWriter, r *http.Request) {
	uid := utils.GetUserIDFromContext(r)
	if uid == "" {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	userID, _ := strconv.ParseInt(uid, 10, 64)
	var payload struct {
		ProfileType string `json:"profile_type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid input")
		return
	}
	pt := "public"
	if payload.ProfileType == "private" {
		pt = "private"
	}
	_, err := db.DB.Exec("UPDATE users SET profile_type=? WHERE id=?", pt, userID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to update privacy")
		return
	}
	utils.JSON(w, http.StatusOK, map[string]string{"status": "updated", "profile_type": pt})
}
