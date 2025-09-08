package api

import ()

// APIHandler handles Api interactions concerning non-HTML related responses
// type APIHandler struct {
// 	messageService *services.MessageService
// }
//
// func NewAPIHandler(messageService *services.MessageService) *APIHandler {
// 	return &APIHandler{
// 		messageService: messageService,
// 	}
// }

// Sets up all API Routes through Chi Router.
// API Routes should return non-web elements, or perform server actions (I.E. JSON, send chat messages)
// func (h *APIHandler) SetupRoutes(r chi.Router) {
//
// 	// Protected API routes
// 	r.Group(func(r chi.Router) {
// 		r.Use(RequireUsername)
//
// 		// r.Post("/message", h.PublishChatMessage)
// 		// r.Post("/rooms/{roomName}/join", h.JoinRoom)
// 		// r.Post("/rooms/{roomName}/leave", h.LeaveRoom)
// 	})
// }

// func (h *APIHandler) PublishChatMessage(w http.ResponseWriter, r *http.Request) {
// 	var req types.Message
//
// 	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
// 		utils.WriteJSONError(w, http.StatusBadRequest, "Invalid Request Body")
// 		return
// 	}
//
// 	username := utils.GetUsernameFromCookie(r)
// 	if username == "" {
// 		utils.WriteJSONError(w, http.StatusBadRequest, "Username not found")
// 		return
// 	}
//
// 	// Extract room from subject (chat.roomName -> roomName)
// 	roomName := strings.TrimPrefix(req.Subject, "chat.")
//
// 	// Call service - it handles everything
// 	if err := h.messageService.SendChatMessage(req.Room, username, req.Message); err != nil {
// 		utils.WriteJSONError(w, http.StatusInternalServerError, "Failed to send message")
// 		return
// 	}
//
// 	utils.WriteJSONResponse(w, http.StatusOK, map[string]any{
// 		"ok":   true,
// 		"room": roomName,
// 	})
// }

// func (h *APIHandler) JoinRoom(w http.ResponseWriter, r *http.Request) {
// 	var req types.Message
//
// 	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
// 		utils.WriteJSONError(w, http.StatusBadRequest, "Invalid Request Body")
// 		return
// 	}
//
// 	username := utils.GetUsernameFromCookie(r)
// 	if username == "" {
// 		utils.WriteJSONError(w, http.StatusBadRequest, "Username not found")
// 		return
// 	}
//
// 	fmt.Printf("req:%+v", req)
//
// 	// Extract room from subject (chat.roomName -> roomName)
//
// 	// Call service - it handles everything
// 	if err := h.SendChatMessage(req.Room, username, req.Message); err != nil {
// 		utils.WriteJSONError(w, http.StatusInternalServerError, "Failed to send message")
// 		return
// 	}
//
// 	utils.WriteJSONResponse(w, http.StatusOK, map[string]any{
// 		"ok":   true,
// 		"room": req.Room,
// 	})
// }

// func (h *APIHandler) SendChatMessage(roomName, username, message string) error {
// 	chatMsg := types.Message{
// 		Username: username,
// 		Message:  message,
// 	}
//
// 	msgBytes, err := json.Marshal(chatMsg)
// 	if err != nil {
// 		return fmt.Errorf("failed to marshal message: %w", err)
// 	}
//
// 	msgString := string(msgBytes)
//
// 	// Store message
// 	h.mu.Lock()
// 	h.roomMessages[roomName] = append(h.roomMessages[roomName], msgString)
// 	h.mu.Unlock()
//
// 	// Broadcast to SSE clients
// 	h.sseBroadcaster.BroadcastToRoom(roomName, msgString)
//
// 	h.logger.Info("Chat message sent", "room", roomName, "username", username)
// 	return nil
// }
