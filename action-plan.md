## Action Plan

- Services setup
    - [X] Setup logger, configure routing, make env variables accessable
    - [X] Figure out how to connect to jellyfin server
    - [X] Configure nats messaging system (removed, don't need the complexity for a small single server)
    - [X] Room service to handle all of the apps rooms

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
    - [ ] Clean up rooms after 30s of being empty?
    - [ ] Delete room if host leaves or pass host?

- Game flow
    - [X] move from lobby state to movies state
    - [ ] People vote (don't need SSE, everyone can have own instance)
    - [ ] Submit button to submit choices
    - [ ] View results!

- Deploy
    - [ ] Deploy this as a docker container so people can download and use with their Jellyfin servers
    - [ ] Figure out how to set env variables and upload stuff to dockerhub for unRAID

- Side quests
    - [ ] Figure out if there's a way to remove javascript for setting theme on page load in Layout
    if there is, whole page can be replaced with SSE, and I can remove header and footer when game
        starts
    - [ ] make skeleton loader for movies?? 

- Stretch goals
    - [ ] create DB to store results of finished games
    - [ ] let individual users log in and generate JWT token
    - [ ] save users selections over time in the DB
    - [ ] cache movies somehow

- Stretchier goals
    - [ ] Add an input field for the host of each room to choose an actor
    - [ ] Generate LLM prompt or use a preconfigured one, that makes the LLM a host for the game
    - [ ] Have the LLM say something funny or witty in-betwixt rounds, deliver with SSE???
    - [ ] Generate a tournament bracket
    - [ ] People vote on the faceoffs 1 at a time
