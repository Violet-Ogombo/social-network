package main

import (
	"net/http"
	"os"
	"social-network/backend/handlers"
)

func registerRoutes(mux *http.ServeMux) {
	// Serve production build if present, otherwise the dev public folder
	if _, err := os.Stat("./frontend/dist"); err == nil {
		mux.Handle("/", http.FileServer(http.Dir("./frontend/dist")))
	} else {
		mux.Handle("/", http.FileServer(http.Dir("./frontend/public")))
	}

	// Websocket endpoint (protected by auth middleware so context contains user ID)
	mux.Handle("/ws", AuthMiddleware(http.HandlerFunc(HandleWebSocket)))

	// API endpoints
	mux.HandleFunc("/register", handlers.RegisterHandler)
	mux.HandleFunc("/login", handlers.LoginHandler)
	mux.HandleFunc("/logout", handlers.LogoutHandler)
	mux.HandleFunc("/api/check-session", handlers.CheckSessionHandler)
	// follower endpoints
	mux.Handle("/api/follow", AuthMiddleware(http.HandlerFunc(handlers.FollowHandler)))     // POST to follow
	mux.Handle("/api/unfollow", AuthMiddleware(http.HandlerFunc(handlers.UnfollowHandler))) // POST to unfollow
	mux.Handle("/api/follow/accept", AuthMiddleware(http.HandlerFunc(handlers.AcceptFollowHandler)))
	mux.Handle("/api/follow/decline", AuthMiddleware(http.HandlerFunc(handlers.DeclineFollowHandler)))
	mux.Handle("/api/follow/requests", AuthMiddleware(http.HandlerFunc(handlers.ListRequests))) // GET list pending requests

	// profile endpoints
	mux.HandleFunc("/api/profile", handlers.GetProfileHandler) // GET public profile (id optional) or current if authenticated
	mux.Handle("/api/profile/update", AuthMiddleware(http.HandlerFunc(handlers.UpdateProfileHandler)))
	mux.Handle("/api/profile/followers", AuthMiddleware(http.HandlerFunc(handlers.GetFollowersHandler)))
	mux.Handle("/api/profile/following", AuthMiddleware(http.HandlerFunc(handlers.GetFollowingHandler)))
	mux.Handle("/api/profile/privacy", AuthMiddleware(http.HandlerFunc(handlers.TogglePrivacyHandler)))

	// posts
	mux.Handle("/api/posts/create", AuthMiddleware(http.HandlerFunc(handlers.CreatePostHandler)))
	mux.HandleFunc("/api/posts", handlers.ListFeedHandler)

	// notifications
	mux.Handle("/api/notifications", AuthMiddleware(http.HandlerFunc(handlers.ListNotificationsHandler)))
	mux.Handle("/api/notifications/mark-read", AuthMiddleware(http.HandlerFunc(handlers.MarkNotificationsReadHandler)))
	mux.Handle("/api/group/create", AuthMiddleware(http.HandlerFunc(handlers.CreateGroupHandler)))
	mux.HandleFunc("/api/groups", handlers.ListGroupsHandler)
	mux.HandleFunc("/api/group", handlers.GetGroupHandler)
	mux.Handle("/api/group/invite", AuthMiddleware(http.HandlerFunc(handlers.InviteHandler)))
	mux.Handle("/api/group/invite/respond", AuthMiddleware(http.HandlerFunc(handlers.RespondInviteHandler)))
	mux.Handle("/api/group/post/create", AuthMiddleware(http.HandlerFunc(handlers.CreateGroupPostHandler)))
	mux.HandleFunc("/api/group/posts", handlers.ListGroupPostsHandler)
	mux.Handle("/api/group/comment", AuthMiddleware(http.HandlerFunc(handlers.AddGroupCommentHandler)))
	mux.Handle("/api/group/event/create", AuthMiddleware(http.HandlerFunc(handlers.CreateEventHandler)))
	mux.Handle("/api/group/event/vote", AuthMiddleware(http.HandlerFunc(handlers.VoteEventHandler)))
	mux.Handle("/api/posts/comment", AuthMiddleware(http.HandlerFunc(handlers.AddCommentHandler)))

	// serve uploaded images
	mux.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir("uploads"))))

}
