package web

import (
	"fmt"
	"net/http"
	"strings"
	"time"
	"watchma/view/steps"

	"github.com/go-chi/chi/v5"
	"github.com/starfederation/datastar-go/datastar"
)

func (h *WebHandler) Announce(w http.ResponseWriter, r *http.Request) {
	roomName := chi.URLParam(r, "roomName")
	room, _ := h.services.RoomService.GetRoom(roomName)

	sorted := SortMoviesByVotes(room.Game.Votes)
	winners := GetWinnerMovies(sorted, room)

	finalMessage := "And the winner is...\n"
	winnerNum := len(winners)
	if winnerNum > 1 {
		finalMessage = "And the winners are...\n"
	}

	buildGptMessage := fmt.Sprintf(`You are writing a reveal scene for: %s

  Write a dialogue-only scene using this EXACT format:

  **[Character Name]:** *"their dialogue"*
  **[Different Character]:** *"their response"*

  RULES:
  1. Use 3-5 characters from the movie
  2. Each character speaks 2-3 times
  3. Include character catchphrases naturally
  4. Build suspense without saying the movie title
  5. NO actor names, NO spoilers, NO narration
  6. Match the movie's genre/tone

  EXAMPLE (for a different movie):
  **Morpheus:** *"What if I told you... we're the ones they chose?"*
  **Trinity:** *"The question isn't how, it's why."*

  NOW write the scene for:`, winners[0].Movie.Name)

	if h.services.OpenAiProvider != nil {
		var err error
		finalMessage, err = h.services.OpenAiProvider.FetchAiResponse(buildGptMessage)
		if err != nil {
			h.logger.Error("AI request failed", "error", err)
		}
	}

	lines := strings.SplitSeq(finalMessage, "\n")

	var strStream []string
	for line := range lines {
		strStream = append(strStream, line)
		announce := steps.AiAnnounce(room, strStream)
		if err := datastar.NewSSE(w, r).PatchElementTempl(announce); err != nil {
			h.logger.Error("Error Rendering Results Page Page", "Error", err)
		}
		time.Sleep(2000 * time.Millisecond)
	}

	strStream = append(strStream, finalMessage)
	announce := steps.AiAnnounce(room, strStream)
	if err := datastar.NewSSE(w, r).PatchElementTempl(announce); err != nil {
		h.logger.Error("Error Rendering Results Page Page", "Error", err)
	}
	time.Sleep(5000 * time.Millisecond)

	h.services.RoomService.FinishGame(roomName)
}
