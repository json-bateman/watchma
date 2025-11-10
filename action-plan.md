## Action Plan

- Services setup
    - [X] Setup logger, configure routing, make env variables accessable
    - [X] Figure out how to connect to jellyfin server
    - [X] Configure nats messaging system (removed, don't need the complexity for a small single server)
    - [X] Room service to handle all of the apps rooms

- Nats again
    - [X] Reintroduce NATS, so I have a simple library and one location for all events

- Room flow
    - [X] host page to host a room (done with form)
    - [X] table for all active rooms
    - [X] others can join room, assign users to room
    - [X] When host makes room, after leaving, doesn't remove him from room? Something is wrong with the
    join / leave logic (fixed)
    - [X] Make Ready Up! button work, user struct - ready: true? maybe not, but sure seems like a flag
    - [X] Start Game! Button, should push users into a new endpoint maybe? /{room}/vote? Not sure
    how that would work, might have to just replace the entire screen via SSE? (I solved this by
    just overwriting the whole page with an SSE event)
    - [X] Make Join table update with kept alive SSE connection (see side quest)
    - [X] Clean up rooms after all players leave.
    - [X] Delete room if host leaves or pass host? (Pass host to random person)

- Game flow
    - [X] move from lobby state to movies state
    - [X] People vote (don't need SSE, everyone can have own instance)
    - [X] Submit button to submit choices
    - [X] View results!
    - [X] Fix map sorting, when there's ties, users can get different results on final screen
    (sort of fixed, now just displays ties)
    - [X] Change game flow, start game --> players add movies, host chooses the max
    number of movies people can add, maybe set timer to 1 minute
    - [X] cache movies somehow - (Cached GET req for 1 minute)
    - [X] players add movies --> Aggregate all choices to vote on.
    - [X] Right now everyone gets their own unique streamed announcement response which I don't want // I want to stream a single message, need to save the res in room.Game.Announce
    - [ ] Veto round
    - [ ] Somehow lock users to current game, if they accidentally navigate away they can rejoin. 
    - [ ] Ending alternative, Host presses play, SSE event pushes everyone to a playing session of the movie in a 
    `<video></video>` player. Now that would be sweet. `https://api.jellyfin.org/` might have endponts to do this, I think 
    `GET/POST ---  http://localhost/Items/{itemId}/PlaybackInfo` this might be the api call

- Deploy
    - [ ] Deploy this as a docker container so people can download and use with their Jellyfin servers
    - [ ] Figure out how to set env variables and upload stuff to dockerhub for unRAID

- Side quests
    - [X] Figure out if there's a way to remove javascript for setting theme on page load in Layou
    if there is, whole page can be replaced with SSE, and I can remove header and footer when game
        starts

- Stretch goals
    - [X] let individual users log in
    - [ ] Multiple Rounds of voting to narrow down to a single winner
    - [ ] create DB to store results of finished games (Have DB need to make table and save it)
    - [ ] save users selections over time in the DB

- Stretchier goals
    - [ ] Generate LLM prompt or use a preconfigured one, that makes the LLM a host for the game
    - [X] Have the LLM say something funny or witty in-betwixt rounds, deliver with SSE???
    - [ ] Generate a tournament bracket?
    - [ ] People vote on the faceoffs 1 at a time
